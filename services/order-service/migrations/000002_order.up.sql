CREATE TABLE IF NOT EXISTS orders (
    id           uuid PRIMARY KEY,
    user_id      varchar(36) NOT NULL,
    address_id   uuid NOT NULL REFERENCES address(id),
    total_amount bigint NOT NULL,
    status       varchar(20) NOT NULL DEFAULT 'pending',
    snap_token   text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz,
    deleted_at   timestamptz
);

CREATE INDEX ix_orders_deleted_at ON orders(deleted_at);
CREATE INDEX ix_orders_user_id ON orders(user_id) WHERE deleted_at IS NULL;
CREATE INDEX ix_orders_status ON orders(status) WHERE deleted_at IS NULL;
