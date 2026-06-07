ALTER TABLE payment_releases
    ALTER COLUMN payment_release_id TYPE UUID USING payment_release_id::uuid,
    ALTER COLUMN order_id TYPE UUID USING order_id::uuid,
    ALTER COLUMN payment_intent_id TYPE UUID USING payment_intent_id::uuid,
    ALTER COLUMN seller_user_id TYPE UUID USING seller_user_id::uuid;

ALTER TABLE payment_releases
    ADD CONSTRAINT fk_payment_releases_payment_intent_id
        FOREIGN KEY (payment_intent_id)
        REFERENCES payment_intents(payment_intent_id)
        ON DELETE RESTRICT;

ALTER TABLE payment_webhook_events
    ADD CONSTRAINT fk_payment_webhook_events_payment_intent_id
        FOREIGN KEY (payment_intent_id)
        REFERENCES payment_intents(payment_intent_id)
        ON DELETE RESTRICT;
