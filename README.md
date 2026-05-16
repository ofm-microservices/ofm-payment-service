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
