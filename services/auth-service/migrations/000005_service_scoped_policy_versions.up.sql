ALTER TABLE casbin_rule DROP CONSTRAINT IF EXISTS casbin_rule_unique;
ALTER TABLE casbin_rule
    ADD CONSTRAINT casbin_rule_unique UNIQUE (ptype, v0, v1, v2, v3, v4, v5, service);

CREATE TABLE IF NOT EXISTS policy_versions (
    service VARCHAR(64) PRIMARY KEY,
    version BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO policy_versions (service)
VALUES ('auth'), ('product'), ('order'), ('dynamic-pricing')
ON CONFLICT (service) DO NOTHING;

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

CREATE OR REPLACE FUNCTION bump_policy_versions_after_truncate() RETURNS trigger AS $$
BEGIN
    UPDATE policy_versions
    SET version = version + 1, updated_at = now();
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS casbin_rule_policy_version ON casbin_rule;
CREATE TRIGGER casbin_rule_policy_version
AFTER INSERT OR UPDATE OR DELETE ON casbin_rule
FOR EACH ROW EXECUTE FUNCTION bump_policy_version();

DROP TRIGGER IF EXISTS casbin_rule_policy_version_truncate ON casbin_rule;
CREATE TRIGGER casbin_rule_policy_version_truncate
AFTER TRUNCATE ON casbin_rule
FOR EACH STATEMENT EXECUTE FUNCTION bump_policy_versions_after_truncate();

INSERT INTO casbin_rule (ptype, v0, v1, v2, service) VALUES
    ('p', 'admin', 'users', 'read', 'auth'),
    ('p', 'admin', 'users', 'update', 'auth'),
    ('p', 'admin', 'policies', 'read', 'auth'),
    ('p', 'admin', 'policies', 'create', 'auth'),
    ('p', 'admin', 'policies', 'delete', 'auth'),

    ('p', 'staff', 'categories', 'read', 'product'),
    ('p', 'staff', 'products', 'read', 'product'),
    ('p', 'staff', 'products', 'create', 'product'),
    ('p', 'staff', 'products', 'update', 'product'),
    ('p', 'staff', 'variants', 'read', 'product'),
    ('p', 'staff', 'variants', 'create', 'product'),
    ('p', 'staff', 'variants', 'update', 'product'),
    ('p', 'staff', 'product-images', 'read', 'product'),
    ('p', 'staff', 'product-images', 'create', 'product'),
    ('p', 'staff', 'product-images', 'delete', 'product'),
    ('p', 'admin', 'categories', 'read', 'product'),
    ('p', 'admin', 'categories', 'create', 'product'),
    ('p', 'admin', 'categories', 'update', 'product'),
    ('p', 'admin', 'categories', 'delete', 'product'),
    ('p', 'admin', 'products', 'read', 'product'),
    ('p', 'admin', 'products', 'create', 'product'),
    ('p', 'admin', 'products', 'update', 'product'),
    ('p', 'admin', 'products', 'delete', 'product'),
    ('p', 'admin', 'variants', 'read', 'product'),
    ('p', 'admin', 'variants', 'create', 'product'),
    ('p', 'admin', 'variants', 'update', 'product'),
    ('p', 'admin', 'variants', 'delete', 'product'),
    ('p', 'admin', 'product-images', 'read', 'product'),
    ('p', 'admin', 'product-images', 'create', 'product'),
    ('p', 'admin', 'product-images', 'delete', 'product'),

    ('p', 'tenant', 'carts', 'read', 'order'),
    ('p', 'tenant', 'carts', 'create', 'order'),
    ('p', 'tenant', 'carts', 'update', 'order'),
    ('p', 'tenant', 'carts', 'delete', 'order'),
    ('p', 'tenant', 'addresses', 'read', 'order'),
    ('p', 'tenant', 'addresses', 'create', 'order'),
    ('p', 'tenant', 'addresses', 'update', 'order'),
    ('p', 'tenant', 'addresses', 'delete', 'order'),
    ('p', 'tenant', 'orders', 'read', 'order'),
    ('p', 'tenant', 'orders', 'create', 'order'),
    ('p', 'tenant', 'payments', 'read', 'order'),
    ('p', 'tenant', 'payments', 'create', 'order'),
    ('p', 'staff', 'orders', 'read', 'order'),
    ('p', 'staff', 'orders', 'update', 'order'),
    ('p', 'admin', 'carts', 'read', 'order'),
    ('p', 'admin', 'carts', 'create', 'order'),
    ('p', 'admin', 'carts', 'update', 'order'),
    ('p', 'admin', 'carts', 'delete', 'order'),
    ('p', 'admin', 'addresses', 'read', 'order'),
    ('p', 'admin', 'addresses', 'create', 'order'),
    ('p', 'admin', 'addresses', 'update', 'order'),
    ('p', 'admin', 'addresses', 'delete', 'order'),
    ('p', 'admin', 'orders', 'read', 'order'),
    ('p', 'admin', 'orders', 'create', 'order'),
    ('p', 'admin', 'orders', 'update', 'order'),
    ('p', 'admin', 'payments', 'read', 'order'),
    ('p', 'admin', 'payments', 'create', 'order'),

    ('p', 'tenant', 'promos', 'apply', 'dynamic-pricing'),
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
    ('p', 'admin', 'promos', 'apply', 'dynamic-pricing')
ON CONFLICT (ptype, v0, v1, v2, v3, v4, v5, service) DO NOTHING;
