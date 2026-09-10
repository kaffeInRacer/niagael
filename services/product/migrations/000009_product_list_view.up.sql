-- =============================================
-- Materialized View: product_list_view
-- Pre-joins product + category for fast listing
-- =============================================

-- 1. Create materialized view
CREATE MATERIALIZED VIEW product_list_view AS
SELECT
    p.id,
    p.category_id,
    c.name AS category_name,
    c.is_active AS category_active,
    p.name,
    p.slug,
    p.description,
    p.price,
    p.stock,
    p.stock_reserved,
    p.is_active,
    p.is_promo_excluded,
    p.created_at,
    p.updated_at
FROM product p
LEFT JOIN category c ON p.category_id = c.id
WHERE p.deleted_at IS NULL;

-- 2. Create indexes for common query patterns
-- Primary key for single row lookups
CREATE UNIQUE INDEX idx_plv_id ON product_list_view (id);

-- For slug-based lookups
CREATE INDEX idx_plv_slug ON product_list_view (slug);

-- For category filtering
CREATE INDEX idx_plv_category_id ON product_list_view (category_id);

-- For is_active filtering
CREATE INDEX idx_plv_is_active ON product_list_view (is_active);

-- For price range queries
CREATE INDEX idx_plv_price ON product_list_view (price);

-- For created_at sorting (most common sort)
CREATE INDEX idx_plv_created_at ON product_list_view (created_at DESC);

-- For name sorting
CREATE INDEX idx_plv_name ON product_list_view (name);

-- For updated_at sorting
CREATE INDEX idx_plv_updated_at ON product_list_view (updated_at DESC);

-- Composite index for category + active filter (ListCount)
CREATE INDEX idx_plv_category_active ON product_list_view (category_id, is_active);

-- GIN trigram for name search (ILIKE)
CREATE INDEX idx_plv_name_trgm ON product_list_view USING gin (name gin_trgm_ops);

-- 3. Grant access
GRANT SELECT ON product_list_view TO postgres;
