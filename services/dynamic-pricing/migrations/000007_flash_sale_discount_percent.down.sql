ALTER TABLE flash_sale
DROP COLUMN IF EXISTS discount_percent;

ALTER TABLE flash_sale
ADD COLUMN IF NOT EXISTS discount_price bigint NOT NULL DEFAULT 0;
