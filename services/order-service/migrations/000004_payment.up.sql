CREATE TABLE IF NOT EXISTS payment (
    id         uuid PRIMARY KEY,
    order_id   uuid NOT NULL REFERENCES orders(id),
    amount     bigint NOT NULL,
    method     varchar(50) NOT NULL,
    status     varchar(20) NOT NULL DEFAULT 'pending',
    snap_token text,
    va_number  varchar(50),
    paid_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz
);

CREATE INDEX ix_payment_order_id ON payment(order_id);
CREATE INDEX ix_payment_status ON payment(status);
