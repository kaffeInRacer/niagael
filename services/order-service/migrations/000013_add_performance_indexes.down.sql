-- Revert performance indexes for order_service
DROP INDEX IF EXISTS idx_cart_item_lookup;
DROP INDEX IF EXISTS idx_order_item_order;
DROP INDEX IF EXISTS idx_payment_order_status;
DROP INDEX IF EXISTS idx_orders_user_created;
DROP INDEX IF EXISTS idx_orders_id_trgm;
DROP INDEX IF EXISTS idx_orders_user_id_trgm;
DROP EXTENSION IF EXISTS pg_trgm;
