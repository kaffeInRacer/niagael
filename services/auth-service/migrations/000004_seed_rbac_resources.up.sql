-- Clean up old policies
DELETE FROM casbin_rule WHERE ptype = 'p';

-- Product Service: categories, products
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
    ('p', 'admin', 'products', 'delete', 'product')
ON CONFLICT (ptype, v0, v1, v2, v3, v4, v5) DO NOTHING;

-- Dynamic Pricing Service: flash-sales, promos
INSERT INTO casbin_rule (ptype, v0, v1, v2, service) VALUES
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
    ('p', 'admin', 'promos', 'delete', 'dynamic-pricing')
ON CONFLICT (ptype, v0, v1, v2, v3, v4, v5) DO NOTHING;

-- Order Service: orders
INSERT INTO casbin_rule (ptype, v0, v1, v2, service) VALUES
    ('p', 'staff', 'orders', 'read', 'order'),
    ('p', 'admin', 'orders', 'read', 'order'),
    ('p', 'admin', 'orders', 'update', 'order')
ON CONFLICT (ptype, v0, v1, v2, v3, v4, v5) DO NOTHING;

-- Auth Service: users, policies
INSERT INTO casbin_rule (ptype, v0, v1, v2, service) VALUES
    ('p', 'admin', 'users', 'read', 'auth'),
    ('p', 'admin', 'users', 'update', 'auth'),
    ('p', 'admin', 'policies', 'read', 'auth'),
    ('p', 'admin', 'policies', 'create', 'auth'),
    ('p', 'admin', 'policies', 'delete', 'auth')
ON CONFLICT (ptype, v0, v1, v2, v3, v4, v5) DO NOTHING;
