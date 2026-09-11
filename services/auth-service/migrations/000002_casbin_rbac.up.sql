CREATE TABLE IF NOT EXISTS casbin_rule (
    id BIGSERIAL PRIMARY KEY,
    ptype VARCHAR(16) NOT NULL,
    v0 VARCHAR(255) NOT NULL DEFAULT '',
    v1 VARCHAR(255) NOT NULL DEFAULT '',
    v2 VARCHAR(255) NOT NULL DEFAULT '',
    v3 VARCHAR(255) NOT NULL DEFAULT '',
    v4 VARCHAR(255) NOT NULL DEFAULT '',
    v5 VARCHAR(255) NOT NULL DEFAULT '',
    service VARCHAR(64) NOT NULL DEFAULT '',
    CONSTRAINT casbin_rule_unique UNIQUE (ptype, v0, v1, v2, v3, v4, v5, service)
);

INSERT INTO casbin_rule (ptype, v0, v1, v2, service) VALUES
    ('p', 'staff', 'categories', 'read', 'product'),
    ('p', 'staff', 'products', 'read', 'product'),
    ('p', 'staff', 'products', 'create', 'product'),
    ('p', 'staff', 'products', 'update', 'product'),
    ('p', 'admin', 'categories', 'read', 'product'),
    ('p', 'admin', 'categories', 'create', 'product'),
    ('p', 'admin', 'categories', 'update', 'product'),
    ('p', 'admin', 'categories', 'delete', 'product'),
    ('p', 'admin', 'products', 'read', 'product'),
    ('p', 'admin', 'products', 'create', 'product'),
    ('p', 'admin', 'products', 'update', 'product'),
    ('p', 'admin', 'products', 'delete', 'product'),
    ('p', 'staff', 'flash-sales', 'read', 'dynamic-pricing'),
    ('p', 'staff', 'flash-sales', 'create', 'dynamic-pricing'),
    ('p', 'staff', 'promos', 'read', 'dynamic-pricing'),
    ('p', 'staff', 'promos', 'create', 'dynamic-pricing'),
    ('p', 'admin', 'flash-sales', 'read', 'dynamic-pricing'),
    ('p', 'admin', 'flash-sales', 'create', 'dynamic-pricing'),
    ('p', 'admin', 'flash-sales', 'update', 'dynamic-pricing'),
    ('p', 'admin', 'flash-sales', 'delete', 'dynamic-pricing'),
    ('p', 'admin', 'promos', 'read', 'dynamic-pricing'),
    ('p', 'admin', 'promos', 'create', 'dynamic-pricing'),
    ('p', 'admin', 'promos', 'update', 'dynamic-pricing'),
    ('p', 'admin', 'promos', 'delete', 'dynamic-pricing'),
    ('p', 'staff', 'orders', 'read', 'order'),
    ('p', 'admin', 'orders', 'read', 'order'),
    ('p', 'admin', 'orders', 'update', 'order'),
    ('p', 'admin', 'users', 'read', 'auth'),
    ('p', 'admin', 'users', 'create', 'auth'),
    ('p', 'admin', 'users', 'update', 'auth'),
    ('p', 'admin', 'users', 'delete', 'auth'),
    ('p', 'admin', 'policies', 'read', 'auth'),
    ('p', 'admin', 'policies', 'create', 'auth'),
    ('p', 'admin', 'policies', 'delete', 'auth')
ON CONFLICT (ptype, v0, v1, v2, v3, v4, v5, service) DO NOTHING;

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
