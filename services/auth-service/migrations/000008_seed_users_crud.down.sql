DELETE FROM casbin_rule WHERE ptype = 'p' AND v1 = 'users' AND v2 IN ('create', 'delete');

INSERT INTO casbin_rule (ptype, v0, v1, v2, service) VALUES
    ('p', 'admin', 'users', 'read', 'auth'),
    ('p', 'admin', 'users', 'update', 'auth')
ON CONFLICT (ptype, v0, v1, v2, v3, v4, v5, service) DO NOTHING;
