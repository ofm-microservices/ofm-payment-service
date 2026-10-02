package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/ofm-microservices/ofm-common/pkg/migration/events"
	"github.com/ofm-microservices/ofm-common/pkg/resilience"
	"payment-service/config"
	app "payment-service/internal/application"
	eventbroker "payment-service/internal/presentation/event_broker"
)

// RecoverySubscriber applies payment fallback commands through payment application use-cases.
type RecoverySubscriber interface{ Subscribe(context.Context) error }
type recoverySubscriber struct {
	broker eventbroker.EventBroker
	svc    app.Service
	cfg    config.KafkaConfig
	log    logging.Logger
}

// NewRecoverySubscriber constructs the payment recovery Kafka adapter.
func NewRecoverySubscriber(b eventbroker.EventBroker, svc app.Service, cfg config.KafkaConfig, log logging.Logger) (RecoverySubscriber, error) {
	if b == nil || svc == nil || log == nil {
		return nil, errors.New("invalid payment recovery subscriber dependency")
	}
	return &recoverySubscriber{broker: b, svc: svc, cfg: cfg, log: log.With(logging.String("module", "kafka-payment-recovery-subscriber"))}, nil
}
func (s *recoverySubscriber) Subscribe(ctx context.Context) error {
	return s.broker.Subscribe(ctx, s.cfg.RecoveryTopic, s.handle)
}
func (s *recoverySubscriber) handle(ctx context.Context, _ string, raw []byte) error {
	var cmd events.Envelope
	if err := json.Unmarshal(raw, &cmd); err != nil {
		return fmt.Errorf("decode payment recovery command: %w", err)
	}
	if !strings.EqualFold(cmd.AggregateType, "payment") {
		return fmt.Errorf("unsupported payment recovery aggregate_type=%q", cmd.AggregateType)
	}
	var result any
	var aggregateID string
	var err error
	if strings.HasSuffix(strings.ToLower(cmd.CommandPath), "/freelancer/onboarding/start") {
		var c app.StartFreelancerOnboardingCommand
		if err = json.Unmarshal(cmd.Payload, &c); err != nil {
			return fmt.Errorf("decode payment onboarding recovery payload: %w", err)
		}
		if c.UserID == "" {
			c.UserID = cmd.RecoveryPrincipalID
		}
		if strings.TrimSpace(c.UserID) == "" {
			return resilience.Permanent(fmt.Errorf("payment recovery command %s is missing user_id", cmd.CommandID))
		}
		if c.IdempotencyKey == "" {
			c.IdempotencyKey = cmd.IdempotencyKey
		}
		onboarding, callErr := s.svc.StartFreelancerOnboarding(ctx, c)
		if callErr != nil {
			s.log.Error("payment recovery onboarding failed",
				logging.Operation("migration.recovery.payment.failed"),
				logging.String("command_id", cmd.CommandID),
				logging.String("test_run_id", cmd.TestRunID),
				logging.String("user_id", c.UserID),
				logging.Err(callErr),
			)
			return callErr
		}
		result = onboarding
		aggregateID = onboarding.UserID
	} else {
		var c app.CreateIntentCommand
		if err = json.Unmarshal(cmd.Payload, &c); err != nil {
			return fmt.Errorf("decode payment recovery payload: %w", err)
		}
		for field, value := range map[string]string{"payment_intent_id": c.IntentID, "order_id": c.OrderID} {
			if strings.TrimSpace(value) == "" {
				return resilience.Permanent(fmt.Errorf("payment recovery command %s is missing %s", cmd.CommandID, field))
			}
		}
		intent, callErr := s.svc.CreateIntent(ctx, c)
		if callErr != nil {
			s.log.Error("payment recovery intent failed",
				logging.Operation("migration.recovery.payment.failed"),
				logging.String("command_id", cmd.CommandID),
				logging.String("test_run_id", cmd.TestRunID),
				logging.Err(callErr),
			)
			return callErr
		}
		result = intent
		aggregateID = cmd.AggregateID
	}
	completed := events.Envelope{EventID: cmd.EventID + ".completed", CommandID: cmd.CommandID, CorrelationID: cmd.CorrelationID, CausationID: cmd.EventID, IdempotencyKey: cmd.IdempotencyKey, TestRunID: cmd.TestRunID, EventType: "migration.recovery.completed", Operation: cmd.Operation, SchemaVersion: 1, AggregateType: "payment", AggregateID: cmd.AggregateID, SourceService: "payment-service-recovery", OccurredAt: time.Now().UTC(), Payload: marshal(result)}
	completed.AggregateID = aggregateID
	body, e := json.Marshal(completed)
	if e != nil {
		return e
	}
	return s.broker.Publish(context.WithoutCancel(ctx), s.cfg.RecoveryCompletedTopic, body)
}
func marshal(v any) []byte { b, _ := json.Marshal(v); return b }
