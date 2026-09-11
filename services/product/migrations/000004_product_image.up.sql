CREATE TABLE IF NOT EXISTS product_image (
    id          uuid PRIMARY KEY,
    product_id  uuid NOT NULL REFERENCES product(id),
    file_name   varchar(500) NOT NULL,
    sort_order  int NOT NULL DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

CREATE INDEX ix_product_image_deleted_at ON product_image(deleted_at);
CREATE INDEX idx_product_image_product_id ON product_image(product_id);
CREATE INDEX idx_product_image_product ON product_image (product_id, deleted_at)
    INCLUDE (file_name, sort_order, created_at);
