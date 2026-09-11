CREATE TABLE IF NOT EXISTS cart (
    id          uuid PRIMARY KEY,
    user_id     varchar(36) NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz,
    deleted_at  timestamptz
);

CREATE TABLE IF NOT EXISTS cart_item (
    id          uuid PRIMARY KEY,
    cart_id     uuid NOT NULL REFERENCES cart(id),
    product_id  uuid NOT NULL,
    variant_id  uuid,
    quantity    int NOT NULL DEFAULT 1,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz,
    deleted_at  timestamptz
);

CREATE INDEX ix_cart_user_id ON cart(user_id) WHERE deleted_at IS NULL;
CREATE INDEX ix_cart_item_cart_id ON cart_item(cart_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_cart_item_lookup ON cart_item (cart_id, created_at ASC);
