ALTER TABLE casbin_rule ADD COLUMN IF NOT EXISTS service VARCHAR(64) NOT NULL DEFAULT '';

UPDATE casbin_rule SET service = 'product' WHERE v1 IN ('category', 'product', 'variant', 'product_image') AND service = '';
UPDATE casbin_rule SET service = 'dynamic-pricing' WHERE v1 = 'promo' AND service = '';
UPDATE casbin_rule SET service = 'order' WHERE v1 = 'order' AND service = '';
