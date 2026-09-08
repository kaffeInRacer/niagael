CREATE TABLE IF NOT EXISTS product_variant (
    id              uuid PRIMARY KEY,
    product_id      uuid NOT NULL REFERENCES product(id),
    name            varchar(255) NOT NULL,
    price           bigint NOT NULL DEFAULT 0,
    stock           bigint NOT NULL DEFAULT 0,
    stock_reserved  bigint NOT NULL DEFAULT 0,
    attributes      jsonb NOT NULL DEFAULT '{}',
    is_active       boolean NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz,
    deleted_at      timestamptz
);

CREATE INDEX ix_product_variant_deleted_at ON product_variant(deleted_at);
CREATE INDEX idx_product_variant_product_id ON product_variant(product_id);
