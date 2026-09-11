CREATE TABLE IF NOT EXISTS product (
    id              uuid PRIMARY KEY,
    category_id     uuid NOT NULL REFERENCES category(id),
    name            varchar(255) NOT NULL,
    slug            varchar(255) NOT NULL,
    description     text NOT NULL DEFAULT '',
    price           bigint NOT NULL DEFAULT 0,
    stock           bigint NOT NULL DEFAULT 0,
    stock_reserved  bigint NOT NULL DEFAULT 0,
    is_active       boolean NOT NULL DEFAULT true,
    is_promo_excluded boolean NOT NULL DEFAULT false,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz,
    deleted_at      timestamptz,
    CONSTRAINT ux_product_slug UNIQUE (slug)
);

CREATE INDEX ix_product_deleted_at ON product(deleted_at);
CREATE INDEX idx_product_category_id ON product(category_id);
CREATE INDEX idx_product_slug ON product(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_product_is_active ON product(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_product_price ON product(price) WHERE deleted_at IS NULL;
CREATE INDEX idx_product_name_trgm ON product USING gin (name gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_product_list_cover ON product (deleted_at, is_active, category_id, created_at DESC)
    INCLUDE (name, slug, description, price, stock, stock_reserved, is_promo_excluded, updated_at);
CREATE INDEX idx_product_category_active ON product (category_id, is_active) WHERE deleted_at IS NULL;
