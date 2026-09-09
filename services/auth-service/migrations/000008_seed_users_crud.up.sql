-- Seed users CRUD permissions for admin role
DELETE FROM casbin_rule WHERE ptype = 'p' AND v1 = 'users';

INSERT INTO casbin_rule (ptype, v0, v1, v2, service) VALUES
    ('p', 'admin', 'users', 'read', 'auth'),
    ('p', 'admin', 'users', 'create', 'auth'),
    ('p', 'admin', 'users', 'update', 'auth'),
    ('p', 'admin', 'users', 'delete', 'auth')
ON CONFLICT (ptype, v0, v1, v2, v3, v4, v5, service) DO NOTHING;
