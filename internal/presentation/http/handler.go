package http

import (
	"encoding/json"
	"fmt"
	"payment-service/internal/application"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/webhook"
)

func (s *server) handleFakePaymentWebhook(c *fiber.Ctx) error {
	var evt application.WebhookCommand
	if err := json.Unmarshal(c.Body(), &evt); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid fake webhook"})
	}
	if evt.EventID == "" {
		evt.EventID = "evt_fake_payment"
	}
	if evt.Provider == "" {
		evt.Provider = "fake-stripe"
	}
	if err := s.app.HandleWebhook(c.Context(), evt); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusAccepted)
}

func (s *server) handleFakePaymentLifecycle(c *fiber.Ctx) error {
	var cmd application.CreateIntentCommand
	if err := json.Unmarshal(c.Body(), &cmd); err != nil || cmd.OrderID == "" || cmd.IntentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "intent_id and order_id are required"})
	}
	if cmd.AmountCents <= 0 {
		cmd.AmountCents = 100
	}
	if cmd.Currency == "" {
		cmd.Currency = "usd"
	}
	if cmd.Provider == "" {
		cmd.Provider = "fake-stripe"
	}
	created, err := s.app.CreateIntent(c.Context(), cmd)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	status := c.Query("status", "captured")
	if err := s.app.HandleWebhook(c.Context(), application.WebhookCommand{
		EventID:  "evt_fake_" + cmd.IntentID,
		Provider: "fake-stripe", EventType: "payment_intent." + status,
		IntentID: created.ProviderIntentID, ProviderIntentID: created.ProviderIntentID,
		OrderID: cmd.OrderID, SagaID: cmd.SagaID, Status: status,
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"intent": created, "webhook_status": status})
}

func (s *server) handleFakeConnectWebhook(c *fiber.Ctx) error {
	var evt application.ConnectWebhookCommand
	if err := json.Unmarshal(c.Body(), &evt); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid fake webhook"})
	}
	if evt.EventID == "" {
		evt.EventID = "evt_fake_connect"
	}
	if evt.Provider == "" {
		evt.Provider = "fake-stripe"
	}
	if err := s.app.HandleConnectWebhook(c.Context(), evt); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusAccepted)
}

func (s *server) handleConnectReturn(c *fiber.Ctx) error {
	s.log.Info("stripe connect return callback received",
		logging.Operation("http.connect.return"),
		logging.String("status", "returned"),
	)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "returned",
		"message": "Stripe Connect onboarding returned. Account status is finalized by webhook.",
	})
}

func (s *server) handleConnectRefresh(c *fiber.Ctx) error {
	s.log.Info("stripe connect refresh callback received",
		logging.Operation("http.connect.refresh"),
		logging.String("status", "refresh_required"),
	)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "refresh_required",
		"message": "Create a new Stripe Connect onboarding link and retry.",
	})
}

func (s *server) handleWebhook(c *fiber.Ctx) error {
	event, err := webhook.ConstructEvent(c.Body(), c.Get("Stripe-Signature"), s.stripe.CheckoutWebhookSecret)
	if err != nil {
		s.log.Error("stripe webhook verification failed", logging.Err(err))
		return c.Status(fiber.StatusBadRequest).SendString("invalid webhook signature")
	}

	var evt application.WebhookCommand
	evt.EventID = event.ID
	evt.Provider = "stripe"
	evt.EventType = string(event.Type)
	evt.PayloadJSON = string(c.Body())

	if err := extractStripeWebhook(event, &evt); err != nil {
		s.log.Error("stripe webhook parse failed", logging.Err(err))
		return c.Status(fiber.StatusBadRequest).SendString("invalid webhook payload")
	}

	if err := s.app.HandleWebhook(c.Context(), evt); err != nil {
		s.log.Error("webhook handling failed", logging.Err(err))
		return c.Status(fiber.StatusInternalServerError).SendString("webhook failed")
	}
	return c.SendStatus(fiber.StatusAccepted)
}

func (s *server) handleConnectWebhook(c *fiber.Ctx) error {
	log := logging.WithContext(c.Context(), s.log)
	event, err := webhook.ConstructEvent(c.Body(), c.Get("Stripe-Signature"), s.stripe.FreelancerOnboardingWebhookSecret)
	if err != nil {
		log.Error("stripe connect webhook verification failed",
			logging.Operation("http.connect_webhook.verify_signature"),
			logging.String("route", "/v1/freelancer/onboarding/webhook"),
			logging.Err(err),
		)
		return c.Status(fiber.StatusBadRequest).SendString("invalid webhook signature")
	}
	var evt application.ConnectWebhookCommand
	evt.EventID = event.ID
	evt.Provider = "stripe"
	evt.EventType = string(event.Type)
	evt.PayloadJSON = string(c.Body())

	if err := extractStripeConnectWebhook(event, &evt, log); err != nil {
		log.Error("stripe connect webhook parse failed",
			logging.Operation("http.connect_webhook.extract"),
			logging.String("event_id", evt.EventID),
			logging.String("event_account", event.Account),
			logging.Err(err),
		)
		return c.Status(fiber.StatusBadRequest).SendString("invalid webhook payload")
	}
	if err := s.app.HandleConnectWebhook(c.Context(), evt); err != nil {
		log.Error("connect webhook handling failed",
			logging.Operation("application.connect_webhook.handle"),
			logging.String("event_id", evt.EventID),
			logging.String("stripe_account_id", evt.StripeAccountID),
			logging.Err(err),
		)
		return c.Status(fiber.StatusInternalServerError).SendString("webhook failed")
	}
	return c.SendStatus(fiber.StatusAccepted)
}

func extractStripeWebhook(event stripe.Event, evt *application.WebhookCommand) error {
	switch event.Type {
	case "payment_intent.succeeded", "payment_intent.payment_failed", "payment_intent.requires_action", "payment_intent.canceled":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return err
		}
		evt.IntentID = pi.ID
		evt.ProviderIntentID = pi.ID
		evt.OrderID = metadataValue(pi.Metadata, "order_id")
		evt.SagaID = metadataValue(pi.Metadata, "saga_id")
		evt.OccurredAt = time.Unix(event.Created, 0).UTC().Format(time.RFC3339Nano)
		evt.Status = mapStripeEventStatus(string(event.Type), string(pi.Status))
		return nil
	default:
		return fmt.Errorf("unsupported stripe event %s", event.Type)
	}
}

func mapStripeEventStatus(eventType, providerStatus string) string {
	switch {
	case strings.Contains(eventType, "succeeded"):
		return "captured"
	case strings.Contains(eventType, "failed"), strings.Contains(eventType, "canceled"):
		return "failed"
	case strings.Contains(eventType, "requires_action"):
		return "requires_action"
	default:
		if providerStatus != "" {
			return providerStatus
		}
		return "pending"
	}
}

func metadataValue(metadata map[string]string, key string) string {
	if metadata == nil {
		return ""
	}
	return metadata[key]
}

func extractStripeConnectWebhook(event stripe.Event, evt *application.ConnectWebhookCommand, log logging.Logger) error {
	switch event.Type {
	case "account.updated", "account.application.deauthorized":
		var acct stripe.Account
		if err := json.Unmarshal(event.Data.Raw, &acct); err != nil {
			return err
		}
		if strings.TrimSpace(event.Account) != "" {
			evt.StripeAccountID = event.Account
		} else {
			evt.StripeAccountID = acct.ID
		}
		evt.UserID = metadataValue(acct.Metadata, "user_id")
		evt.DetailsSubmitted = acct.DetailsSubmitted
		evt.ChargesEnabled = acct.ChargesEnabled
		evt.PayoutsEnabled = acct.PayoutsEnabled
		evt.DisabledReason = connectDisabledReason(&acct)
		return nil
	default:
		return fmt.Errorf("unsupported stripe connect event %s", event.Type)
	}
}

func connectDisabledReason(account *stripe.Account) string {
	if account == nil || account.Requirements == nil {
		return ""
	}
	return string(account.Requirements.DisabledReason)
}

func connectMetadataValue(metadata map[string]string, key string) string {
	if metadata == nil {
		return ""
	}
	return metadata[key]
}
