# OFM Payment Service

Payment service owns payment intent orchestration, Stripe webhook normalization, and freelancer Stripe Connect onboarding.

## Local Run

```bash
cp .env.example .env
set -a && source .env && set +a && go run ./cmd/payment-service
```

## Notes

- `STRIPE_SECRET_KEY` must be a Stripe secret key.
- `STRIPE_FREELANCER_ONBOARDING_WEBHOOK_SECRET` comes from the Stripe Dashboard webhook endpoint.
- `STRIPE_CONNECT_RETURN_URL` and `STRIPE_CONNECT_REFRESH_URL` are your app URLs, not Stripe values.

For automated integration and load tests, set `STRIPE_FAKE_ENABLED=true`. The
payment service then uses a deterministic in-process adapter and does not call
Stripe or require browser onboarding. Leave it `false` (the default) for real
Stripe test-mode flows.

When fake mode is enabled, test drivers may submit normalized webhook commands
to `POST /v1/fake-stripe/webhooks/payment` and
`POST /v1/fake-stripe/webhooks/connect`. These routes are not registered when
fake mode is disabled; they use the same application handlers, persistence,
Kafka publication, idempotency, logging, metrics, and tracing as real Stripe
webhooks.
