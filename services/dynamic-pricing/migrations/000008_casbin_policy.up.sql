CREATE TABLE IF NOT EXISTS casbin_rule (
    id BIGSERIAL PRIMARY KEY,
    ptype VARCHAR(16) NOT NULL,
    v0 VARCHAR(255) NOT NULL DEFAULT '',
    v1 VARCHAR(255) NOT NULL DEFAULT '',
    v2 VARCHAR(255) NOT NULL DEFAULT '',
    v3 VARCHAR(255) NOT NULL DEFAULT '',
    v4 VARCHAR(255) NOT NULL DEFAULT '',
    v5 VARCHAR(255) NOT NULL DEFAULT '',
    CONSTRAINT casbin_rule_unique UNIQUE (ptype, v0, v1, v2, v3, v4, v5)
);

INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'tenant', 'promo', 'apply'),
    ('p', 'staff', 'flash_sale', 'read'),
    ('p', 'staff', 'flash_sale', 'create'),
    ('p', 'staff', 'flash_sale', 'update'),
    ('p', 'staff', 'promo', 'read'),
    ('p', 'staff', 'promo', 'create'),
    ('p', 'staff', 'promo', 'update'),
    ('p', 'admin', 'flash_sale', 'read'),
    ('p', 'admin', 'flash_sale', 'create'),
    ('p', 'admin', 'flash_sale', 'update'),
    ('p', 'admin', 'flash_sale', 'delete'),
    ('p', 'admin', 'promo', 'read'),
    ('p', 'admin', 'promo', 'create'),
    ('p', 'admin', 'promo', 'update'),
    ('p', 'admin', 'promo', 'delete'),
    ('p', 'admin', 'promo', 'apply')
ON CONFLICT (ptype, v0, v1, v2, v3, v4, v5) DO NOTHING;
