-- Performance indexes for order_service
-- Addresses: Order List JOIN+ILIKE, ProcessCallback FOR UPDATE, Cart lookups

-- 1. Enable pg_trgm for GIN trigram indexes (ILIKE '%...%' support)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 2. GIN trigram untuk orders user_id ILIKE search (Order List)
CREATE INDEX idx_orders_user_id_trgm ON orders USING gin (user_id gin_trgm_ops) WHERE deleted_at IS NULL;

-- 3. GIN trigram untuk orders id text ILIKE search (Order List search by order ID)
CREATE INDEX idx_orders_id_trgm ON orders USING gin ((id::text) gin_trgm_ops) WHERE deleted_at IS NULL;

-- 4. Composite index untuk Order List pagination (cursor-based)
CREATE INDEX idx_orders_user_created ON orders (user_id, created_at DESC) WHERE deleted_at IS NULL;

-- 5. Composite index untuk ProcessCallback (payment + order lock)
CREATE INDEX idx_payment_order_status ON payment (order_id, status);

-- 6. Composite index untuk OrderItem by order_id (sudah ada partial, tambah coverage)
CREATE INDEX idx_order_item_order ON order_item (order_id, created_at ASC) WHERE deleted_at IS NULL;

-- 7. Composite index untuk Cart item lookup by cart_id
CREATE INDEX idx_cart_item_lookup ON cart_item (cart_id, created_at ASC);
