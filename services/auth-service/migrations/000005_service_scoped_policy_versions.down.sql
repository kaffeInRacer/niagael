DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM casbin_rule
        GROUP BY ptype, v0, v1, v2, v3, v4, v5
        HAVING COUNT(DISTINCT service) > 1
    ) THEN
        RAISE EXCEPTION USING
            MESSAGE = 'cannot roll back service-scoped policies: policy keys exist in multiple services',
            HINT = 'Remove or merge cross-service policy duplicates explicitly before retrying rollback.';
    END IF;
END;
$$;

DROP TRIGGER IF EXISTS casbin_rule_policy_version_truncate ON casbin_rule;
DROP TRIGGER IF EXISTS casbin_rule_policy_version ON casbin_rule;
DROP FUNCTION IF EXISTS bump_policy_versions_after_truncate();
DROP FUNCTION IF EXISTS bump_policy_version();
DROP TABLE IF EXISTS policy_versions;

ALTER TABLE casbin_rule DROP CONSTRAINT IF EXISTS casbin_rule_unique;
ALTER TABLE casbin_rule
    ADD CONSTRAINT casbin_rule_unique UNIQUE (ptype, v0, v1, v2, v3, v4, v5);
