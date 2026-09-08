CREATE TABLE IF NOT EXISTS flash_sale (
    id              uuid PRIMARY KEY,
    name            varchar(255) NOT NULL,
    product_id      uuid NOT NULL,
    discount_price  bigint NOT NULL,
    stock           bigint NOT NULL DEFAULT 0,
    max_per_user    int NOT NULL DEFAULT 0,
    start_time      timestamptz NOT NULL,
    end_time        timestamptz NOT NULL,
    is_active       boolean NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz,
    deleted_at      timestamptz
);

CREATE INDEX ix_flash_sale_deleted_at ON flash_sale(deleted_at);
CREATE INDEX idx_flash_sale_product_id ON flash_sale(product_id);
CREATE INDEX idx_flash_sale_start_end ON flash_sale(start_time, end_time);
CREATE INDEX idx_flash_sale_is_active ON flash_sale(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_flash_sale_name ON flash_sale(name) WHERE deleted_at IS NULL;
