-- Revert performance indexes for auth_service
DROP INDEX IF EXISTS idx_users_email_trgm;
DROP INDEX IF EXISTS idx_casbin_rule_lookup;
DROP INDEX IF EXISTS idx_users_created_at;
DROP EXTENSION IF EXISTS pg_trgm;
