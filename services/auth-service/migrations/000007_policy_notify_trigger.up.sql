CREATE OR REPLACE FUNCTION notify_policy_change() RETURNS trigger AS $$
DECLARE
    changed_service VARCHAR(64);
BEGIN
    changed_service := CASE WHEN TG_OP = 'DELETE' THEN OLD.service ELSE NEW.service END;
    PERFORM pg_notify('casbin_policy_changed', changed_service);
    IF TG_OP = 'UPDATE' AND OLD.service <> NEW.service THEN
        PERFORM pg_notify('casbin_policy_changed', OLD.service);
    END IF;
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS casbin_rule_notify ON casbin_rule;
CREATE TRIGGER casbin_rule_notify
AFTER INSERT OR UPDATE OR DELETE ON casbin_rule
FOR EACH ROW EXECUTE FUNCTION notify_policy_change();
