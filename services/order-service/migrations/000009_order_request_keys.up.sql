CREATE TABLE IF NOT EXISTS order_request_keys (
    key        VARCHAR(64) PRIMARY KEY,
    order_id   UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
