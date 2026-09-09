ALTER TABLE payment ADD COLUMN IF NOT EXISTS redirect_url text;

CREATE TABLE IF NOT EXISTS payment_duplicate_archive (
    LIKE payment INCLUDING DEFAULTS INCLUDING CONSTRAINTS INCLUDING STORAGE INCLUDING COMMENTS
);
ALTER TABLE payment_duplicate_archive ADD COLUMN IF NOT EXISTS archived_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE payment_duplicate_archive ADD COLUMN IF NOT EXISTS archive_reason text NOT NULL DEFAULT 'duplicate_order_id';
ALTER TABLE payment_duplicate_archive ADD COLUMN IF NOT EXISTS kept_payment_id uuid NOT NULL;

DO $$
DECLARE
    duplicate_count bigint;
    archived_count bigint;
BEGIN
    CREATE TEMP TABLE payment_duplicate_preflight ON COMMIT DROP AS
    SELECT id, first_value(id) OVER (
        PARTITION BY order_id
        ORDER BY
            CASE
                WHEN status = 'paid' THEN 0
                WHEN status = 'pending' AND NULLIF(snap_token, '') IS NOT NULL THEN 1
                WHEN NULLIF(snap_token, '') IS NOT NULL THEN 2
                ELSE 3
            END,
            created_at DESC,
            id DESC
    ) AS kept_payment_id,
    row_number() OVER (
        PARTITION BY order_id
        ORDER BY
            CASE
                WHEN status = 'paid' THEN 0
                WHEN status = 'pending' AND NULLIF(snap_token, '') IS NOT NULL THEN 1
                WHEN NULLIF(snap_token, '') IS NOT NULL THEN 2
                ELSE 3
            END,
            created_at DESC,
            id DESC
    ) AS duplicate_rank
    FROM payment;

    SELECT count(*) INTO duplicate_count
    FROM payment_duplicate_preflight
    WHERE duplicate_rank > 1;

    INSERT INTO payment_duplicate_archive (
        id, order_id, amount, method, status, snap_token, va_number, paid_at,
        created_at, updated_at, redirect_url, kept_payment_id
    )
    SELECT p.id, p.order_id, p.amount, p.method, p.status, p.snap_token, p.va_number, p.paid_at,
           p.created_at, p.updated_at, p.redirect_url, d.kept_payment_id
    FROM payment p
    JOIN payment_duplicate_preflight d ON d.id = p.id
    WHERE d.duplicate_rank > 1;

    GET DIAGNOSTICS archived_count = ROW_COUNT;
    IF archived_count <> duplicate_count THEN
        RAISE EXCEPTION 'payment duplicate preflight failed: expected to archive %, archived %', duplicate_count, archived_count;
    END IF;

    DELETE FROM payment p
    USING payment_duplicate_preflight d
    WHERE d.id = p.id AND d.duplicate_rank > 1;
END $$;

DROP INDEX IF EXISTS ix_payment_order_id;
CREATE UNIQUE INDEX ux_payment_order_id ON payment(order_id);
