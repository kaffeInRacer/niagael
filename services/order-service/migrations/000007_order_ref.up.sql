ALTER TABLE orders
ADD COLUMN IF NOT EXISTS order_ref varchar(32);

UPDATE orders
SET order_ref = 'ORD-'
    || upper(substring(md5(id::text) from 1 for 5))
    || '-'
    || lpad((('x' || substring(md5(id::text) from 6 for 3))::bit(24)::bigint % 100000)::text, 5, '0')
WHERE order_ref IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_orders_order_ref ON orders(order_ref);
