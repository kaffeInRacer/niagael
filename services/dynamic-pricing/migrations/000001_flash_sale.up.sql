CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS flash_sale (
    id              uuid PRIMARY KEY,
    name            varchar(255) NOT NULL,
    product_id      uuid NOT NULL,
    variant_id      uuid,
    discount_percent int NOT NULL DEFAULT 0,
    stock           bigint NOT NULL DEFAULT 0,
    max_per_user    int NOT NULL DEFAULT 0,
    start_time      timestamptz NOT NULL,
    end_time        timestamptz NOT NULL,
    is_active       boolean NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz,
    deleted_at      timestamptz,
    CONSTRAINT flash_sale_discount_percent_range CHECK (discount_percent BETWEEN 0 AND 100)
);

CREATE INDEX ix_flash_sale_deleted_at ON flash_sale(deleted_at);
CREATE INDEX idx_flash_sale_product_id ON flash_sale(product_id);
CREATE INDEX idx_flash_sale_start_end ON flash_sale(start_time, end_time);
CREATE INDEX idx_flash_sale_is_active ON flash_sale(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_flash_sale_name ON flash_sale(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_flash_sale_variant_id ON flash_sale(variant_id);
CREATE INDEX idx_flash_sale_name_trgm ON flash_sale USING gin (name gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_flash_sale_active_time ON flash_sale (is_active, start_time, end_time, stock) WHERE deleted_at IS NULL;
CREATE INDEX idx_flash_sale_product_active ON flash_sale (product_id, is_active, start_time, end_time, stock) WHERE deleted_at IS NULL;
CREATE INDEX idx_flash_sale_product_variant ON flash_sale (product_id, variant_id, is_active, start_time, end_time, stock) WHERE deleted_at IS NULL;
