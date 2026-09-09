CREATE TABLE IF NOT EXISTS policy_versions (
    service VARCHAR(64) PRIMARY KEY,
    version BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION bump_policy_version() RETURNS trigger AS $$
DECLARE
    changed_service VARCHAR(64);
BEGIN
    changed_service := CASE WHEN TG_OP = 'DELETE' THEN OLD.service ELSE NEW.service END;
    INSERT INTO policy_versions (service, version, updated_at)
    VALUES (changed_service, 1, now())
    ON CONFLICT (service) DO UPDATE
    SET version = policy_versions.version + 1, updated_at = now();

    IF TG_OP = 'UPDATE' AND OLD.service <> NEW.service THEN
        INSERT INTO policy_versions (service, version, updated_at)
        VALUES (OLD.service, 1, now())
        ON CONFLICT (service) DO UPDATE
        SET version = policy_versions.version + 1, updated_at = now();
    END IF;
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER casbin_rule_policy_version
AFTER INSERT OR UPDATE OR DELETE ON casbin_rule
FOR EACH ROW EXECUTE FUNCTION bump_policy_version();
