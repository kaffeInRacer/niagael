DROP INDEX IF EXISTS ux_orders_order_ref;
ALTER TABLE orders
DROP COLUMN IF EXISTS order_ref;
