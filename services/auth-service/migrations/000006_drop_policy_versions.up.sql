DROP TRIGGER IF EXISTS casbin_rule_policy_version ON casbin_rule;
DROP FUNCTION IF EXISTS bump_policy_version();
DROP TABLE IF EXISTS policy_versions;
