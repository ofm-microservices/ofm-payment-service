DROP TRIGGER IF EXISTS payment_intents_outbox ON payment_intents;
DROP TRIGGER IF EXISTS payment_webhook_events_outbox ON payment_webhook_events;
DROP TRIGGER IF EXISTS connect_accounts_outbox ON connect_accounts;
DROP TRIGGER IF EXISTS payment_releases_outbox ON payment_releases;
DROP FUNCTION IF EXISTS capture_payment_outbox_event();
DROP TABLE IF EXISTS outbox_events;
