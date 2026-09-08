ALTER TABLE order_item
ADD COLUMN IF NOT EXISTS flash_sale_discount_percent int;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'order_item_flash_sale_discount_percent_range'
    ) THEN
        ALTER TABLE order_item
        ADD CONSTRAINT order_item_flash_sale_discount_percent_range
        CHECK (flash_sale_discount_percent BETWEEN 0 AND 100);
    END IF;
END $$;
