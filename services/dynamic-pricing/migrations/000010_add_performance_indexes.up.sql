-- Performance indexes for dynamic_pricing
-- Addresses: Flash Sale List ILIKE+sort, Promo List ILIKE+sort, active time-range lookups

-- 1. Enable pg_trgm for GIN trigram indexes (ILIKE '%...%' support)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 2. GIN trigram untuk flash_sale name ILIKE search
CREATE INDEX idx_flash_sale_name_trgm ON flash_sale USING gin (name gin_trgm_ops) WHERE deleted_at IS NULL;

-- 3. Composite index untuk flash_sale active + time range lookup (ReadByProductId, List current-only)
CREATE INDEX idx_flash_sale_active_time ON flash_sale (is_active, start_time, end_time, stock)
  WHERE deleted_at IS NULL;

-- 4. Composite index untuk flash_sale product + active + time (ReadByProductIds batch)
CREATE INDEX idx_flash_sale_product_active ON flash_sale (product_id, is_active, start_time, end_time, stock)
  WHERE deleted_at IS NULL;

-- 5. Composite index untuk flash_sale variant pair lookup (ReadByVariantIds)
CREATE INDEX idx_flash_sale_product_variant ON flash_sale (product_id, variant_id, is_active, start_time, end_time, stock)
  WHERE deleted_at IS NULL;

-- 6. GIN trigram untuk promo name+code ILIKE search
CREATE INDEX idx_promo_search_trgm ON promo USING gin (name gin_trgm_ops, code gin_trgm_ops) WHERE deleted_at IS NULL;

-- 7. Composite index untuk promo active + time range (List current-only)
CREATE INDEX idx_promo_active_time ON promo (is_active, start_date, end_date, quantity, used_count)
  WHERE deleted_at IS NULL;

-- 8. Composite index untuk flash_sale_usage lookup (subquery in allocation)
CREATE INDEX idx_flash_sale_usage_lookup ON flash_sale_usage (flash_sale_id, user_id, quantity);

-- 9. Composite index untuk promo_usage lookup (subquery in allocation)
CREATE INDEX idx_promo_usage_lookup ON promo_usage (promo_id, user_id, quantity);
