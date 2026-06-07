ALTER TABLE payment_webhook_events
    DROP CONSTRAINT IF EXISTS fk_payment_webhook_events_payment_intent_id;

ALTER TABLE payment_releases
    DROP CONSTRAINT IF EXISTS fk_payment_releases_payment_intent_id;

ALTER TABLE payment_releases
    ALTER COLUMN seller_user_id TYPE TEXT USING seller_user_id::text,
    ALTER COLUMN payment_intent_id TYPE TEXT USING payment_intent_id::text,
    ALTER COLUMN order_id TYPE TEXT USING order_id::text,
    ALTER COLUMN payment_release_id TYPE TEXT USING payment_release_id::text;
