# OFM Payment Service

## Purpose

The Payment Service owns payment intents, normalized payment webhook state, Connect accounts, releases, transfers, and refunds. It is the payment source of truth and does not own order workflow decisions. Status: active.

## Flow and fake Stripe mode

Order-saga commands create or release payment state. Real Stripe webhooks and the enabled fake-Stripe adapter enter the same application handlers, persistence, idempotency checks, Kafka publication, logs, metrics, and traces.

Set STRIPE_FAKE_ENABLED=true for automated local experiments. Fake mode provides deterministic webhook routes and does not call Stripe. It must not be confused with skipping settlement: payment records still move through the release/transfer/refund state machine.

## Configuration

.env.example groups are DB_*, migrations, Stripe credentials and webhook settings, fake-mode flags, NATS/Kafka, listener ports, and observability. Stripe values select real test-mode credentials; fake mode selects the local adapter; webhook settings control signature validation. Never commit secrets.

## Local development

    cp .env.example .env
    set -a && source .env && set +a
    go run ./cmd/payment-service
    go test ./...

## Build and operations

Dockerfile builds ofm/payment-service:<tag>. Helm deploys the service. Diagnose payment intents, webhook events, release/transfer/refund rows, saga acknowledgments, and Kafka projection state together.

