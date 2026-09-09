CREATE TABLE pricing_allocation (
    order_id                 uuid PRIMARY KEY,
    user_id                  uuid NOT NULL,
    promo_id                 uuid REFERENCES promo(id),
    promo_code               varchar(50) NOT NULL DEFAULT '',
    promo_name               varchar(255) NOT NULL DEFAULT '',
    promo_discount_type      varchar(20) NOT NULL DEFAULT '',
    promo_discount_amount    bigint NOT NULL DEFAULT 0,
    request_hash             varchar(64) NOT NULL DEFAULT '',
    status                   varchar(20) NOT NULL CHECK (status IN ('allocated', 'released')),
    created_at               timestamptz NOT NULL DEFAULT now(),
    released_at              timestamptz
);

CREATE TABLE pricing_allocation_flash_sale (
    order_id          uuid NOT NULL REFERENCES pricing_allocation(order_id),
    item_id           uuid NOT NULL,
    flash_sale_id     uuid NOT NULL REFERENCES flash_sale(id),
    quantity          int NOT NULL CHECK (quantity > 0),
    name              varchar(255) NOT NULL,
    discount_percent  int NOT NULL,
    PRIMARY KEY (order_id, item_id)
);

CREATE INDEX ix_pricing_allocation_flash_sale_id
    ON pricing_allocation_flash_sale(flash_sale_id);
