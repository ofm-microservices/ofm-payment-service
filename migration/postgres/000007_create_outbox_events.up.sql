CREATE TABLE IF NOT EXISTS outbox_events (
    event_id UUID PRIMARY KEY,
    aggregate_type TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    operation TEXT NOT NULL,
    schema_version INT NOT NULL DEFAULT 1,
    payload JSONB NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_outbox_events_aggregate
    ON outbox_events (aggregate_type, aggregate_id, created_at);

CREATE OR REPLACE FUNCTION capture_payment_outbox_event() RETURNS trigger
LANGUAGE plpgsql AS $$
DECLARE row_data JSONB; operation_value TEXT := CASE TG_OP WHEN 'INSERT' THEN 'created' WHEN 'DELETE' THEN 'deactivated' ELSE 'updated' END; aggregate_type_value TEXT;
BEGIN
    row_data := CASE WHEN TG_OP = 'DELETE' THEN to_jsonb(OLD) ELSE to_jsonb(NEW) END;
    aggregate_type_value := CASE WHEN TG_TABLE_NAME = 'payment_webhook_events' THEN 'payment_webhook_events' WHEN TG_TABLE_NAME = 'payment_releases' THEN 'payment_releases' WHEN TG_TABLE_NAME = 'connect_accounts' THEN 'connect_accounts' ELSE 'payment_intents' END;
    INSERT INTO outbox_events(event_id, aggregate_type, aggregate_id, event_type, operation, payload)
    VALUES (md5(clock_timestamp()::TEXT || random()::TEXT)::UUID, aggregate_type_value, COALESCE(row_data->>'payment_intent_id', row_data->>'payment_release_id', row_data->>'event_id', row_data->>'user_id', row_data->>'id'), 'payment-service.' || aggregate_type_value || '.changed', operation_value, row_data);
    IF TG_OP = 'DELETE' THEN RETURN OLD; ELSE RETURN NEW; END IF;
END;
$$;

DROP TRIGGER IF EXISTS payment_intents_outbox ON payment_intents;
CREATE TRIGGER payment_intents_outbox AFTER INSERT OR UPDATE OR DELETE ON payment_intents FOR EACH ROW EXECUTE FUNCTION capture_payment_outbox_event();
DROP TRIGGER IF EXISTS payment_webhook_events_outbox ON payment_webhook_events;
CREATE TRIGGER payment_webhook_events_outbox AFTER INSERT OR UPDATE OR DELETE ON payment_webhook_events FOR EACH ROW EXECUTE FUNCTION capture_payment_outbox_event();
DROP TRIGGER IF EXISTS connect_accounts_outbox ON connect_accounts;
CREATE TRIGGER connect_accounts_outbox AFTER INSERT OR UPDATE OR DELETE ON connect_accounts FOR EACH ROW EXECUTE FUNCTION capture_payment_outbox_event();
DROP TRIGGER IF EXISTS payment_releases_outbox ON payment_releases;
CREATE TRIGGER payment_releases_outbox AFTER INSERT OR UPDATE OR DELETE ON payment_releases FOR EACH ROW EXECUTE FUNCTION capture_payment_outbox_event();
