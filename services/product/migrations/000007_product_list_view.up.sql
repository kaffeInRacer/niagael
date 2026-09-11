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

CREATE UNIQUE INDEX idx_plv_id ON product_list_view (id);
CREATE INDEX idx_plv_slug ON product_list_view (slug);
CREATE INDEX idx_plv_category_id ON product_list_view (category_id);
CREATE INDEX idx_plv_is_active ON product_list_view (is_active);
CREATE INDEX idx_plv_price ON product_list_view (price);
CREATE INDEX idx_plv_created_at ON product_list_view (created_at DESC);
CREATE INDEX idx_plv_name ON product_list_view (name);
CREATE INDEX idx_plv_updated_at ON product_list_view (updated_at DESC);
CREATE INDEX idx_plv_category_active ON product_list_view (category_id, is_active);
CREATE INDEX idx_plv_name_trgm ON product_list_view USING gin (name gin_trgm_ops);

GRANT SELECT ON product_list_view TO postgres;
