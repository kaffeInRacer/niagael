-- Performance indexes for product_service
-- Addresses: Product List JOIN+sort, Category ListWithProductCount, Variant triple-JOIN, ILIKE search

-- 1. Enable pg_trgm for GIN trigram indexes (ILIKE '%...%' support)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 2. GIN trigram untuk product name ILIKE search
CREATE INDEX idx_product_name_trgm ON product USING gin (name gin_trgm_ops) WHERE deleted_at IS NULL;

-- 3. Covering index untuk Product List query (reduces Seq Scan + improves sort)
CREATE INDEX idx_product_list_cover ON product (deleted_at, is_active, category_id, created_at DESC)
  INCLUDE (name, slug, description, price, stock, stock_reserved, is_promo_excluded, updated_at);

-- 4. Composite index untuk Product ListCount (same WHERE without SELECT columns)
CREATE INDEX idx_product_category_active ON product (category_id, is_active) WHERE deleted_at IS NULL;

-- 5. Composite index untuk Variant ListActiveByProductId (triple JOIN: variant→product→category)
CREATE INDEX idx_variant_product_active ON product_variant (product_id, is_active, deleted_at)
  INCLUDE (name, price, stock, stock_reserved, attributes, created_at, updated_at);

-- 6. GIN trigram untuk category name ILIKE search
CREATE INDEX idx_category_name_trgm ON category USING gin (name gin_trgm_ops) WHERE deleted_at IS NULL;

-- 7. Composite index untuk Product image batch lookup
CREATE INDEX idx_product_image_product ON product_image (product_id, deleted_at)
  INCLUDE (file_name, sort_order, created_at);
