CREATE EXTENSION IF NOT EXISTS citext;
CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY,
    email citext NOT NULL UNIQUE,
    password_hash text NOT NULL,
    role text NOT NULL CHECK (role IN ('tenant', 'staff', 'admin')),
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

INSERT INTO users (id, email, password_hash, role)
VALUES
--   user1234
    ('10000000-0000-0000-0000-000000000001', 'tenant@demo.com', '$2a$12$J1ysIESBsfj.yptSVJ0PYeepTrHLhBUCLvPXlLBmSVzJ84xd/BoOC', 'tenant'),
--     staff1234
    ('10000000-0000-0000-0000-000000000002', 'staff@demo.com', '$2a$12$J1ysIESBsfj.yptSVJ0PYeepTrHLhBUCLvPXlLBmSVzJ84xd/BoOC', 'staff'),
--     admin1234
    ('10000000-0000-0000-0000-000000000003', 'admin@demo.com', '$2a$12$J1ysIESBsfj.yptSVJ0PYeepTrHLhBUCLvPXlLBmSVzJ84xd/BoOC', 'admin')
ON CONFLICT (email) DO NOTHING;
