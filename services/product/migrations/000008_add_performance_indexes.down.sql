-- Revert performance indexes for product_service
DROP INDEX IF EXISTS idx_product_image_product;
DROP INDEX IF EXISTS idx_category_name_trgm;
DROP INDEX IF EXISTS idx_variant_product_active;
DROP INDEX IF EXISTS idx_product_category_active;
DROP INDEX IF EXISTS idx_product_list_cover;
DROP INDEX IF EXISTS idx_product_name_trgm;
DROP EXTENSION IF EXISTS pg_trgm;
