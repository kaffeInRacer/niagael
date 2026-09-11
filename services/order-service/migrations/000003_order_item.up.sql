CREATE TABLE IF NOT EXISTS order_item (
    id                        uuid PRIMARY KEY,
    order_id                  uuid NOT NULL REFERENCES orders(id),
    product_id                varchar(36) NOT NULL,
    variant_id                varchar(36),
    product_name              varchar(255) NOT NULL,
    product_price             bigint NOT NULL,
    quantity                  int NOT NULL,
    flash_sale_id             varchar(36),
    flash_sale_name           varchar(255),
    flash_sale_discount_price bigint,
    flash_sale_original_price bigint,
    flash_sale_quantity       int DEFAULT 0,
    flash_sale_discount_percent int,
    promo_id                  varchar(36),
    promo_code                varchar(50),
    promo_name                varchar(255),
    promo_discount_type       varchar(20),
    promo_discount_amount     bigint DEFAULT 0,
    final_price               bigint NOT NULL,
    created_at                timestamptz NOT NULL DEFAULT now(),
    updated_at                timestamptz,
    deleted_at                timestamptz,
    CONSTRAINT order_item_flash_sale_discount_percent_range CHECK (flash_sale_discount_percent BETWEEN 0 AND 100)
);

CREATE INDEX ix_order_item_deleted_at ON order_item(deleted_at);
CREATE INDEX ix_order_item_order_id ON order_item(order_id) WHERE deleted_at IS NULL;
CREATE INDEX ix_order_item_product_id ON order_item(product_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_order_item_order ON order_item (order_id, created_at ASC) WHERE deleted_at IS NULL;
