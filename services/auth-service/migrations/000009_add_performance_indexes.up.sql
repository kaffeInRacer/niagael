-- Performance indexes for auth_service
-- Addresses: UserList ORDER BY sort, GetAllPolicies scan, GetResources DISTINCT

-- 1. Index untuk UserList ORDER BY created_at DESC (menghindari Sort di memory)
CREATE INDEX idx_users_created_at ON users (created_at DESC) WHERE deleted_at IS NULL;

-- 2. Composite index untuk GetAllPolicies (ptype + service filter + sort columns)
CREATE INDEX idx_casbin_rule_lookup ON casbin_rule (ptype, service, v0, v1, v2);

-- 3. GIN trigram untuk email search (jika diperlukan di masa depan)
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_users_email_trgm ON users USING gin (email gin_trgm_ops);
