CREATE TABLE IF NOT EXISTS promo (
    id                      uuid PRIMARY KEY,
    name                    varchar(255) NOT NULL,
    code                    varchar(50) NOT NULL,
    description             text NOT NULL DEFAULT '',
    discount_type           varchar(20) NOT NULL,
    discount_value          bigint NOT NULL,
    min_purchase            bigint NOT NULL DEFAULT 0,
    max_discount            bigint NOT NULL DEFAULT 0,
    quantity                int NOT NULL DEFAULT 0,
    used_count              int NOT NULL DEFAULT 0,
    max_usage_per_user      int NOT NULL DEFAULT 0,
    can_combine_flash_sale  boolean NOT NULL DEFAULT false,
    start_date              timestamptz NOT NULL,
    end_date                timestamptz NOT NULL,
    is_active               boolean NOT NULL DEFAULT true,
    created_at              timestamptz NOT NULL DEFAULT now(),
    updated_at              timestamptz,
    deleted_at              timestamptz,
    CONSTRAINT ux_promo_code UNIQUE (code)
);

CREATE INDEX ix_promo_deleted_at ON promo(deleted_at);
CREATE INDEX idx_promo_start_end ON promo(start_date, end_date);
CREATE INDEX idx_promo_is_active ON promo(is_active) WHERE deleted_at IS NULL;
