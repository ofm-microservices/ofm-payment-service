ALTER TABLE payment_releases
  DROP COLUMN IF EXISTS stripe_refund_id,
  DROP COLUMN IF EXISTS customer_amount_cents,
  DROP COLUMN IF EXISTS freelancer_amount_cents,
  DROP COLUMN IF EXISTS customer_percentage,
  DROP COLUMN IF EXISTS freelancer_percentage;
