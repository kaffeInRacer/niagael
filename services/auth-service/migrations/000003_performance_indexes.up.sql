CREATE INDEX idx_users_created_at ON users (created_at DESC) WHERE deleted_at IS NULL;

CREATE INDEX idx_casbin_rule_lookup ON casbin_rule (ptype, service, v0, v1, v2);

CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_users_email_trgm ON users USING gin (email gin_trgm_ops);
