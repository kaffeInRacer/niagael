ALTER TABLE flash_sale ADD COLUMN variant_id uuid;

CREATE INDEX idx_flash_sale_variant_id ON flash_sale(variant_id);
