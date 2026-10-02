CREATE TABLE IF NOT EXISTS payment_releases (
  payment_release_id TEXT PRIMARY KEY,
  order_id TEXT NOT NULL UNIQUE,
  payment_intent_id TEXT NOT NULL,
  seller_user_id TEXT NOT NULL,
  amount_cents BIGINT NOT NULL,
  currency TEXT NOT NULL,
  idempotency_key TEXT NOT NULL,
  stripe_transfer_id TEXT,
  status TEXT NOT NULL,
  failure_reason TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_releases_payment_intent_id ON payment_releases(payment_intent_id);
CREATE INDEX IF NOT EXISTS idx_payment_releases_seller_user_id ON payment_releases(seller_user_id);
