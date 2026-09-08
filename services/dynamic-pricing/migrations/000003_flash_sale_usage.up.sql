CREATE TABLE IF NOT EXISTS flash_sale_usage (
    id              uuid PRIMARY KEY,
    flash_sale_id   uuid NOT NULL REFERENCES flash_sale(id),
    user_id         uuid NOT NULL,
    quantity        int NOT NULL DEFAULT 0,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz,
    CONSTRAINT ux_flash_sale_usage UNIQUE (flash_sale_id, user_id)
);

CREATE INDEX ix_flash_sale_usage_flash_sale_id ON flash_sale_usage(flash_sale_id);
CREATE INDEX ix_flash_sale_usage_user_id ON flash_sale_usage(user_id);
