CREATE TABLE IF NOT EXISTS connect_accounts (
    user_id UUID PRIMARY KEY,
    stripe_account_id TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL,
    details_submitted BOOLEAN NOT NULL DEFAULT FALSE,
    charges_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    payouts_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    disabled_reason TEXT,
    onboarding_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_connect_accounts_stripe_account_id ON connect_accounts(stripe_account_id);
