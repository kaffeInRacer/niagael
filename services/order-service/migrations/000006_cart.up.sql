CREATE TABLE IF NOT EXISTS cart (
    id         uuid PRIMARY KEY,
    user_id    varchar(36) NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz
);

CREATE TABLE IF NOT EXISTS cart_item (
    id         uuid PRIMARY KEY,
    cart_id    uuid NOT NULL REFERENCES cart(id) ON DELETE CASCADE,
    product_id varchar(36) NOT NULL,
    variant_id varchar(36),
    quantity   int NOT NULL CHECK (quantity > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz,
    UNIQUE NULLS NOT DISTINCT (cart_id, product_id, variant_id)
);

CREATE INDEX ix_cart_item_cart_id ON cart_item(cart_id);
