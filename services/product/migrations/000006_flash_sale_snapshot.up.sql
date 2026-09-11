CREATE TABLE IF NOT EXISTS flash_sale_snapshot (
    product_id       UUID NOT NULL,
    variant_id       TEXT NOT NULL DEFAULT '',
    name             VARCHAR(255) NOT NULL DEFAULT '',
    discount_percent INT NOT NULL DEFAULT 0,
    start_time       TIMESTAMPTZ NOT NULL,
    end_time         TIMESTAMPTZ NOT NULL,
    is_active        BOOLEAN NOT NULL DEFAULT true,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (product_id, variant_id)
);
