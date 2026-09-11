CREATE TABLE stock_reservation (
    order_id TEXT PRIMARY KEY,
    request_hash TEXT NOT NULL,
    items JSONB NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('reserved', 'confirmed', 'released')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
