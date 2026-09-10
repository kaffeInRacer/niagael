-- Revert performance indexes for dynamic_pricing
DROP INDEX IF EXISTS idx_promo_usage_lookup;
DROP INDEX IF EXISTS idx_flash_sale_usage_lookup;
DROP INDEX IF EXISTS idx_promo_active_time;
DROP INDEX IF EXISTS idx_promo_search_trgm;
DROP INDEX IF EXISTS idx_flash_sale_product_variant;
DROP INDEX IF EXISTS idx_flash_sale_product_active;
DROP INDEX IF EXISTS idx_flash_sale_active_time;
DROP INDEX IF EXISTS idx_flash_sale_name_trgm;
DROP EXTENSION IF EXISTS pg_trgm;
