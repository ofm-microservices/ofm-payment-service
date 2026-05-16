CREATE TABLE IF NOT EXISTS payment_intents (
    payment_intent_id UUID PRIMARY KEY,
    order_id UUID NOT NULL,
    provider TEXT NOT NULL,
    provider_intent_id TEXT,
    checkout_url TEXT,
    amount_cents BIGINT NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_intents_order_id ON payment_intents(order_id);
CREATE INDEX IF NOT EXISTS idx_payment_intents_status ON payment_intents(status);
