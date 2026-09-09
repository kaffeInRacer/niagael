DROP INDEX IF EXISTS ux_payment_order_id;
CREATE INDEX IF NOT EXISTS ix_payment_order_id ON payment(order_id);

-- Archived duplicate payments and redirect_url remain intentionally: rollback must not discard financial records.
