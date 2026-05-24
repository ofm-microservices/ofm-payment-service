package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"payment-service/internal/domain"

	"github.com/google/uuid"
	"github.com/ofm-microservices/ofm-common/pkg/logging"
	paymentflowv1 "github.com/ofm-microservices/ofm-common/proto/paymentflow/v1"
	"github.com/stripe/stripe-go/v85"
	checkoutsession "github.com/stripe/stripe-go/v85/checkout/session"
	"google.golang.org/protobuf/encoding/protojson"
)

type service struct {
	intents  IntentRepository
	webhooks WebhookRepository
	read     PaymentIntentReadRepository
	accounts ConnectAccountRepository
	releases PaymentReleaseRepository
	broker   EventBroker
	stripe   StripeConnectGateway
	log      Logger
}

// New constructs the payment application service.
func New(intents IntentRepository, webhooks WebhookRepository, read PaymentIntentReadRepository, accounts ConnectAccountRepository, releases PaymentReleaseRepository, broker EventBroker, stripe StripeConnectGateway, log Logger) (Service, error) {
	if intents == nil {
		return nil, ErrNilIntentRepository
	}
	if webhooks == nil {
		return nil, ErrNilWebhookRepository
	}
	if read == nil {
		return nil, ErrNilReadRepository
	}
	if accounts == nil {
		return nil, ErrNilConnectAccountRepository
	}
	if releases == nil {
		return nil, ErrNilPaymentReleaseRepository
	}
	if broker == nil {
		return nil, ErrNilEventBroker
	}
	if stripe == nil {
		return nil, ErrNilStripeConnectGateway
	}
	if log == nil {
		return nil, ErrNilLogger
	}
	return &service{intents: intents, webhooks: webhooks, read: read, accounts: accounts, releases: releases, broker: broker, stripe: stripe, log: log.With(logging.String("module", "application"))}, nil
}

func (s *service) CreateIntent(ctx context.Context, cmd CreateIntentCommand) (*CreateIntentResult, error) {
	now := time.Now().UTC()
	intent := domain.PaymentIntent{
		IntentID:    cmd.IntentID,
		OrderID:     cmd.OrderID,
		Provider:    cmd.Provider,
		AmountCents: cmd.AmountCents,
		Currency:    cmd.Currency,
		Status:      domain.PaymentStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := s.intents.Create(ctx, intent); err != nil {
		return nil, err
	}
	session, err := checkoutsession.New(&stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:        stripe.String("http://localhost:3000/orders/checkout/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:         stripe.String("http://localhost:3000/orders/checkout/cancel?session_id={CHECKOUT_SESSION_ID}"),
		ClientReferenceID: stripe.String(cmd.OrderID),
		CustomerCreation:  stripe.String(string(stripe.CheckoutSessionCustomerCreationAlways)),
		PaymentMethodTypes: []*string{
			stripe.String("card"),
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(1),
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:   stripe.String(strings.ToLower(strings.TrimSpace(cmd.Currency))),
					UnitAmount: stripe.Int64(cmd.AmountCents),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(cmd.OrderID),
						Description: stripe.String("OFM order checkout"),
						Metadata: map[string]string{
							"order_id": cmd.OrderID,
							"saga_id":  cmd.SagaID,
						},
					},
				},
			},
		},
		Metadata: map[string]string{
			"order_id": cmd.OrderID,
			"saga_id":  cmd.SagaID,
		},
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{
				"order_id": cmd.OrderID,
				"saga_id":  cmd.SagaID,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("stripe checkout session not created")
	}
	intent.CheckoutURL = session.URL
	intent.ProviderIntentID = session.ID
	_ = s.intents.UpdateCheckoutURL(ctx, intent.IntentID, intent.CheckoutURL)
	_ = s.intents.UpdateStatus(ctx, intent.IntentID, domain.PaymentStatusIntentCreated)
	if created, err := s.intents.GetByID(ctx, intent.IntentID); err == nil {
		created.CheckoutURL = intent.CheckoutURL
		created.ProviderIntentID = intent.ProviderIntentID
		created.Status = domain.PaymentStatusIntentCreated
		_ = s.read.Upsert(ctx, created)
	}
	res := &CreateIntentResult{
		IntentID:         cmd.IntentID,
		SagaID:           cmd.SagaID,
		OrderID:          cmd.OrderID,
		ProviderIntentID: session.ID,
		CheckoutURL:      session.URL,
		Status:           domain.PaymentStatusIntentCreated,
		OccurredAt:       now.Format(time.RFC3339Nano),
	}
	payload, err := protojson.Marshal(&paymentflowv1.PaymentIntentResult{
		SagaId:           cmd.SagaID,
		OrderId:          cmd.OrderID,
		PaymentIntentId:  cmd.IntentID,
		CheckoutUrl:      session.URL,
		ProviderIntentId: session.ID,
		Status:           domain.PaymentStatusIntentCreated,
		OccurredAt:       now.Format(time.RFC3339Nano),
	})
	if err == nil {
		_ = s.broker.Publish(ctx, "payment.intent.result", payload)
	}
	return res, nil
}

func (s *service) HandleWebhook(ctx context.Context, evt WebhookCommand) error {
	if exists, _ := s.webhooks.Exists(ctx, evt.EventID); exists {
		return nil
	}
	if evt.OrderID == "" {
		return errors.New("missing order id in payment webhook")
	}
	intent, err := s.intents.UpdateWebhookPaymentIntent(ctx, evt.OrderID, evt.ProviderIntentID, evt.Status)
	if err != nil {
		return err
	}
	evt.IntentID = intent.IntentID
	evt.OrderID = intent.OrderID
	if _, err := s.webhooks.Create(ctx, domain.WebhookEvent{
		EventID:   evt.EventID,
		Provider:  evt.Provider,
		OrderID:   evt.OrderID,
		IntentID:  evt.IntentID,
		Status:    evt.Status,
		Payload:   evt.PayloadJSON,
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		return err
	}
	intent.ProviderIntentID = evt.ProviderIntentID
	intent.Status = evt.Status
	intent.UpdatedAt = time.Now().UTC()
	_ = s.read.Upsert(ctx, intent)
	switch evt.Status {
	case domain.PaymentStatusCaptured:
		completed, _ := protojson.Marshal(&paymentflowv1.PaymentStatusEvent{
			SagaId:          evt.SagaID,
			OrderId:         evt.OrderID,
			PaymentIntentId: evt.IntentID,
			Status:          "payment.order_payment_succeeded",
			OccurredAt:      time.Now().UTC().Format(time.RFC3339Nano),
		})
		return s.broker.Publish(ctx, "payment.order_payment_succeeded", completed)
	case domain.PaymentStatusFailed:
		failed, _ := protojson.Marshal(&paymentflowv1.PaymentStatusEvent{
			SagaId:          evt.SagaID,
			OrderId:         evt.OrderID,
			PaymentIntentId: evt.IntentID,
			Status:          "payment.order_payment_failed",
			OccurredAt:      time.Now().UTC().Format(time.RFC3339Nano),
		})
		return s.broker.Publish(ctx, "payment.order_payment_failed", failed)
	default:
		return nil
	}
}

func (s *service) Subscribe(ctx context.Context) error {
	return s.broker.Subscribe(ctx, "payment.intent", func(ctx context.Context, subject string, payload []byte) error {
		var cmd paymentflowv1.PaymentIntentCommand
		if err := protojson.Unmarshal(payload, &cmd); err != nil {
			return err
		}
		_, err := s.CreateIntent(ctx, CreateIntentCommand{
			IntentID:       cmd.GetPaymentIntentId(),
			SagaID:         cmd.GetSagaId(),
			OrderID:        cmd.GetOrderId(),
			AmountCents:    cmd.GetAmountCents(),
			Currency:       cmd.GetCurrency(),
			Provider:       cmd.GetProvider(),
			IdempotencyKey: cmd.GetIdempotencyKey(),
			RequestedAt:    cmd.GetRequestedAt(),
		})
		return err
	})
}

func (s *service) StartFreelancerOnboarding(ctx context.Context, cmd StartFreelancerOnboardingCommand) (*StartFreelancerOnboardingResult, error) {
	log := logging.WithContext(ctx, s.log)
	userID := cmd.UserID
	if userID == "" {
		return nil, domain.ErrInvalidID
	}
	log.Info("connect onboarding started",
		logging.Operation("application.connect_onboarding.start"),
		logging.String("user_id", userID),
	)

	now := time.Now().UTC()
	prevStatus := ""
	account, err := s.accounts.GetByUserID(ctx, userID)
	if err != nil && !errors.Is(err, domain.ErrConnectAccountNotFound) {
		return nil, err
	}
	if account != nil {
		prevStatus = account.Status
	}
	if account == nil {
		account, err = s.stripe.CreateAccount(ctx, cmd)
		if err != nil {
			return nil, err
		}
		log.Info("connect account created",
			logging.Operation("application.connect_onboarding.start"),
			logging.String("user_id", userID),
			logging.String("stripe_account_id", account.StripeAccountID),
		)
	}

	if account.Status == "" {
		account.Status = domain.ConnectStatusPending
	}

	if account.Status != domain.ConnectStatusCompleted {
		url, err := s.stripe.CreateOnboardingLink(ctx, account.StripeAccountID, cmd.ReturnURL, cmd.RefreshURL)
		if err != nil {
			return nil, err
		}
		account.OnboardingURL = url
		log.Info("onboarding link created",
			logging.Operation("application.connect_onboarding.start"),
			logging.String("user_id", userID),
			logging.String("stripe_account_id", account.StripeAccountID),
		)
	}

	account.UpdatedAt = now
	persisted, err := s.accounts.Upsert(ctx, *account)
	if err != nil {
		return nil, err
	}
	if prevStatus != "" && persisted.Status != prevStatus {
		log.Info("connect account status changed",
			logging.Operation("application.connect_onboarding.start"),
			logging.String("user_id", persisted.UserID),
			logging.String("stripe_account_id", persisted.StripeAccountID),
			logging.String("from", prevStatus),
			logging.String("to", persisted.Status),
		)
	}
	if prevStatus != domain.ConnectStatusCompleted && persisted.Status == domain.ConnectStatusCompleted {
		log.Info("connect onboarding completed",
			logging.Operation("application.connect_onboarding.start"),
			logging.String("user_id", persisted.UserID),
			logging.String("stripe_account_id", persisted.StripeAccountID),
		)
	}

	return &StartFreelancerOnboardingResult{
		UserID:           persisted.UserID,
		StripeAccountID:  persisted.StripeAccountID,
		OnboardingURL:    persisted.OnboardingURL,
		Status:           persisted.Status,
		DetailsSubmitted: persisted.DetailsSubmitted,
		ChargesEnabled:   persisted.ChargesEnabled,
		PayoutsEnabled:   persisted.PayoutsEnabled,
		DisabledReason:   persisted.DisabledReason,
		OccurredAt:       now.Format(time.RFC3339Nano),
	}, nil
}

func (s *service) GetConnectStatus(ctx context.Context, userID string) (*GetConnectStatusResult, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, domain.ErrInvalidID
	}
	account, err := s.accounts.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &GetConnectStatusResult{
		UserID:          account.UserID,
		StripeAccountID: account.StripeAccountID,
		Status:          account.Status,
		DisabledReason:  account.DisabledReason,
		OccurredAt:      time.Now().UTC().Format(time.RFC3339Nano),
	}, nil
}

func (s *service) ReleaseFunds(ctx context.Context, cmd ReleaseFundsCommand) (*ReleaseFundsResult, error) {
	account, err := s.accounts.GetByUserID(ctx, cmd.SellerUserID)
	if err != nil {
		return nil, err
	}
	if account.Status != domain.ConnectStatusCompleted || !account.PayoutsEnabled {
		return nil, errors.New("seller connect account is not ready for payouts")
	}
	if existing, err := s.releases.GetByOrderID(ctx, cmd.OrderID); err == nil && existing != nil {
		switch existing.Status {
		case domain.PaymentReleaseStatusReleased:
			return &ReleaseFundsResult{
				OrderID:          existing.OrderID,
				PaymentReleaseID: existing.ReleaseID,
				StripeTransferID: existing.StripeTransferID,
				Status:           existing.Status,
				OccurredAt:       existing.UpdatedAt.UTC().Format(time.RFC3339Nano),
			}, nil
		case domain.PaymentReleaseStatusPending:
			if existing.StripeTransferID != "" {
				return &ReleaseFundsResult{
					OrderID:          existing.OrderID,
					PaymentReleaseID: existing.ReleaseID,
					StripeTransferID: existing.StripeTransferID,
					Status:           domain.PaymentReleaseStatusReleased,
					OccurredAt:       existing.UpdatedAt.UTC().Format(time.RFC3339Nano),
				}, nil
			}
		}
	}
	now := time.Now().UTC()
	release := domain.PaymentRelease{
		ReleaseID:      uuid.NewString(),
		OrderID:        cmd.OrderID,
		PaymentID:      cmd.PaymentID,
		SellerUserID:   cmd.SellerUserID,
		AmountCents:    cmd.AmountCents,
		Currency:       strings.ToUpper(strings.TrimSpace(cmd.Currency)),
		IdempotencyKey: strings.TrimSpace(cmd.IdempotencyKey),
		Status:         domain.PaymentReleaseStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	persisted, err := s.releases.Create(ctx, release)
	if err != nil {
		return nil, err
	}
	transferID, err := s.stripe.CreateTransfer(ctx, cmd, account.StripeAccountID)
	if err != nil {
		_ = s.releases.UpdateStatus(ctx, persisted.ReleaseID, domain.PaymentReleaseStatusFailed, "", err.Error())
		return nil, err
	}
	if err := s.releases.UpdateStatus(ctx, persisted.ReleaseID, domain.PaymentReleaseStatusReleased, transferID, ""); err != nil {
		return nil, err
	}
	return &ReleaseFundsResult{
		OrderID:          persisted.OrderID,
		PaymentReleaseID: persisted.ReleaseID,
		StripeTransferID: transferID,
		Status:           domain.PaymentReleaseStatusReleased,
		OccurredAt:       now.Format(time.RFC3339Nano),
	}, nil
}

func (s *service) GetReleaseByOrderID(ctx context.Context, orderID string) (*GetReleaseByOrderResult, error) {
	release, err := s.releases.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return &GetReleaseByOrderResult{
		OrderID:          release.OrderID,
		PaymentReleaseID: release.ReleaseID,
		PaymentID:        release.PaymentID,
		SellerUserID:     release.SellerUserID,
		AmountCents:      release.AmountCents,
		Currency:         release.Currency,
		IdempotencyKey:   release.IdempotencyKey,
		StripeTransferID: release.StripeTransferID,
		Status:           release.Status,
		FailureReason:    release.FailureReason,
		OccurredAt:       release.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}, nil
}

func (s *service) HandleConnectWebhook(ctx context.Context, evt ConnectWebhookCommand) error {
	log := logging.WithContext(ctx, s.log)
	if exists, _ := s.webhooks.Exists(ctx, evt.EventID); exists {
		log.Warn("webhook duplicate ignored",
			logging.Operation("application.connect_webhook.handle"),
			logging.String("event_id", evt.EventID),
			logging.String("stripe_account_id", evt.StripeAccountID),
		)
		return nil
	}
	stored, err := s.accounts.GetByStripeAccountID(ctx, evt.StripeAccountID)
	if err != nil && !errors.Is(err, domain.ErrConnectAccountNotFound) {
		return err
	}
	if stored == nil {
		stored = &domain.ConnectAccount{StripeAccountID: evt.StripeAccountID, UserID: evt.UserID}
	}
	incomingStatus := deriveConnectStatus(evt.DetailsSubmitted, evt.ChargesEnabled, evt.PayoutsEnabled, evt.DisabledReason)
	if stored.Status == domain.ConnectStatusCompleted && incomingStatus != domain.ConnectStatusCompleted {
		log.Warn("webhook event out of order",
			logging.Operation("application.connect_webhook.handle"),
			logging.String("event_id", evt.EventID),
			logging.String("stripe_account_id", evt.StripeAccountID),
			logging.String("current_status", stored.Status),
			logging.String("incoming_status", incomingStatus),
		)
	}
	if strings.TrimSpace(evt.UserID) == "" {
		log.Warn("stripe account missing user_id metadata",
			logging.Operation("http.connect_webhook.extract"),
			logging.String("event_id", evt.EventID),
			logging.String("stripe_account_id", evt.StripeAccountID),
		)
	}
	if strings.TrimSpace(evt.DisabledReason) != "" && strings.TrimSpace(stored.DisabledReason) != strings.TrimSpace(evt.DisabledReason) {
		log.Warn("connect account restricted",
			logging.Operation("application.connect_webhook.handle"),
			logging.String("event_id", evt.EventID),
			logging.String("stripe_account_id", evt.StripeAccountID),
			logging.String("disabled_reason", evt.DisabledReason),
		)
	}
	stored.DetailsSubmitted = evt.DetailsSubmitted || stored.DetailsSubmitted
	stored.ChargesEnabled = evt.ChargesEnabled || stored.ChargesEnabled
	stored.PayoutsEnabled = evt.PayoutsEnabled || stored.PayoutsEnabled
	stored.DisabledReason = evt.DisabledReason
	stored.Status = incomingStatus
	stored.UpdatedAt = time.Now().UTC()
	if _, err := s.accounts.Upsert(ctx, *stored); err != nil {
		s.log.Error("connect webhook upsert failed",
			logging.Operation("application.connect_webhook.handle"),
			logging.String("event_id", evt.EventID),
			logging.String("stripe_account_id", stored.StripeAccountID),
			logging.Err(err),
		)
		return err
	}
	return nil
}

func connectDisabledReason(account *stripe.Account) string {
	if account == nil || account.Requirements == nil {
		return ""
	}
	return string(account.Requirements.DisabledReason)
}

func deriveConnectStatus(detailsSubmitted, chargesEnabled, payoutsEnabled bool, disabledReason string) string {
	if detailsSubmitted && chargesEnabled && payoutsEnabled {
		return domain.ConnectStatusCompleted
	}
	if strings.TrimSpace(disabledReason) != "" {
		return domain.ConnectStatusDisabled
	}
	return domain.ConnectStatusPending
}
