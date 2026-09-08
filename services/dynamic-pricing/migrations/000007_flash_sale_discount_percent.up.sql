ALTER TABLE flash_sale
ADD COLUMN IF NOT EXISTS discount_percent int;

UPDATE flash_sale
SET discount_percent = 0
WHERE discount_percent IS NULL;

ALTER TABLE flash_sale
ALTER COLUMN discount_percent SET DEFAULT 0,
ALTER COLUMN discount_percent SET NOT NULL;

ALTER TABLE flash_sale
DROP COLUMN IF EXISTS discount_price;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'flash_sale_discount_percent_range'
    ) THEN
        ALTER TABLE flash_sale
        ADD CONSTRAINT flash_sale_discount_percent_range
        CHECK (discount_percent BETWEEN 0 AND 100);
    END IF;
END $$;
