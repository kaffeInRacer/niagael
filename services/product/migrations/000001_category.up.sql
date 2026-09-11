CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS category(
    id          uuid PRIMARY KEY,
    name        varchar(255) NOT NULL,
    slug        varchar(255) NOT NULL,
    description text NOT NULL DEFAULT '',
    is_active   boolean NOT NULL DEFAULT true,
    deleted_at  timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz,
    CONSTRAINT ux_category_slug UNIQUE (slug)
);

CREATE INDEX ix_category_deleted_at ON category (deleted_at);
CREATE INDEX idx_category_slug ON category(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_category_name_trgm ON category USING gin (name gin_trgm_ops) WHERE deleted_at IS NULL;
