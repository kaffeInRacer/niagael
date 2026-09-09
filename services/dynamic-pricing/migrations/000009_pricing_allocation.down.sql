DO $$
BEGIN
    IF to_regclass('pricing_allocation') IS NULL THEN
        RETURN;
    END IF;

    IF to_regclass('pricing_allocation_flash_sale') IS NOT NULL AND EXISTS (
        SELECT 1
        FROM pricing_allocation pa
        JOIN pricing_allocation_flash_sale pafs ON pafs.order_id = pa.order_id
        LEFT JOIN flash_sale_usage fsu
            ON fsu.flash_sale_id = pafs.flash_sale_id
           AND fsu.user_id = pa.user_id
        WHERE pa.status = 'allocated'
        GROUP BY pafs.flash_sale_id, pa.user_id, fsu.quantity
        HAVING fsu.quantity IS NULL OR fsu.quantity < SUM(pafs.quantity)
    ) THEN
        RAISE EXCEPTION USING
            MESSAGE = 'cannot roll back pricing allocations: flash-sale usage counters are inconsistent',
            HINT = 'Repair missing or insufficient flash_sale_usage rows before retrying rollback.';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pricing_allocation pa
        JOIN promo p ON p.id = pa.promo_id
        WHERE pa.status = 'allocated' AND pa.promo_id IS NOT NULL
        GROUP BY pa.promo_id, p.used_count
        HAVING p.used_count < COUNT(*)
    ) THEN
        RAISE EXCEPTION USING
            MESSAGE = 'cannot roll back pricing allocations: promo counters are inconsistent',
            HINT = 'Repair insufficient promo.used_count values before retrying rollback.';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pricing_allocation pa
        LEFT JOIN promo_usage pu
            ON pu.promo_id = pa.promo_id
           AND pu.user_id = pa.user_id
        WHERE pa.status = 'allocated' AND pa.promo_id IS NOT NULL
        GROUP BY pa.promo_id, pa.user_id, pu.quantity
        HAVING pu.quantity IS NULL OR pu.quantity < COUNT(*)
    ) THEN
        RAISE EXCEPTION USING
            MESSAGE = 'cannot roll back pricing allocations: promo usage counters are inconsistent',
            HINT = 'Repair missing or insufficient promo_usage rows before retrying rollback.';
    END IF;

    IF to_regclass('pricing_allocation_flash_sale') IS NOT NULL THEN
        UPDATE flash_sale fs
        SET stock = fs.stock + restored.quantity,
            updated_at = now()
        FROM (
            SELECT pafs.flash_sale_id, SUM(pafs.quantity)::bigint AS quantity
            FROM pricing_allocation pa
            JOIN pricing_allocation_flash_sale pafs ON pafs.order_id = pa.order_id
            WHERE pa.status = 'allocated'
            GROUP BY pafs.flash_sale_id
        ) restored
        WHERE fs.id = restored.flash_sale_id;

        UPDATE flash_sale_usage fsu
        SET quantity = fsu.quantity - restored.quantity,
            updated_at = now()
        FROM (
            SELECT pafs.flash_sale_id, pa.user_id, SUM(pafs.quantity)::int AS quantity
            FROM pricing_allocation pa
            JOIN pricing_allocation_flash_sale pafs ON pafs.order_id = pa.order_id
            WHERE pa.status = 'allocated'
            GROUP BY pafs.flash_sale_id, pa.user_id
        ) restored
        WHERE fsu.flash_sale_id = restored.flash_sale_id
          AND fsu.user_id = restored.user_id;
    END IF;

    UPDATE promo p
    SET used_count = p.used_count - restored.quantity,
        updated_at = now()
    FROM (
        SELECT pa.promo_id, COUNT(*)::int AS quantity
        FROM pricing_allocation pa
        WHERE pa.status = 'allocated' AND pa.promo_id IS NOT NULL
        GROUP BY pa.promo_id
    ) restored
    WHERE p.id = restored.promo_id;

    UPDATE promo_usage pu
    SET quantity = pu.quantity - restored.quantity,
        updated_at = now()
    FROM (
        SELECT pa.promo_id, pa.user_id, COUNT(*)::int AS quantity
        FROM pricing_allocation pa
        WHERE pa.status = 'allocated' AND pa.promo_id IS NOT NULL
        GROUP BY pa.promo_id, pa.user_id
    ) restored
    WHERE pu.promo_id = restored.promo_id
      AND pu.user_id = restored.user_id;
END;
$$;

DROP TABLE IF EXISTS pricing_allocation_flash_sale;
DROP TABLE IF EXISTS pricing_allocation;
