package repository

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"kaffein/dynamic-pricing-service/internal/domain"
	"kaffein/dynamic-pricing-service/internal/interfaces/IRepository"
	"kaffein/dynamic-pricing-service/pkg/postgresql"
	"kaffein/dynamic-pricing-service/utils"
	"kaffein/dynamic-pricing-service/utils/constants"
)

type (
	PricingAllocationItem = IRepository.PricingAllocationItem
	PricingAllocation     = IRepository.PricingAllocation
)

type pricingAllocationRepository struct {
	store *postgresql.Store
}

func NewPricingAllocationRepository(store *postgresql.Store) IRepository.PricingAllocationRepository {
	return &pricingAllocationRepository{store: store}
}

func (r *pricingAllocationRepository) AllocateWithTx(ctx context.Context, allocation PricingAllocation) (*PricingAllocation, error) {
	allocation.RequestHash = allocationFingerprint(allocation)
	var result *PricingAllocation
	err := r.store.ExecTx(ctx, func(tx postgresql.DBTX) error {
		existing, err := readAllocation(ctx, tx, allocation.OrderID)
		if err != nil {
			return err
		}
		if existing != nil {
			if existing.UserID != allocation.UserID || existing.RequestHash != allocation.RequestHash {
				return errors.New(constants.ErrAllocationConflict)
			}
			result = existing
			return nil
		}

		result = &allocation
		sort.Slice(result.Items, func(i, j int) bool {
			if result.Items[i].FlashSaleID == result.Items[j].FlashSaleID {
				return result.Items[i].ItemID < result.Items[j].ItemID
			}
			return result.Items[i].FlashSaleID < result.Items[j].FlashSaleID
		})

		hasFlashSale := false
		for i := range result.Items {
			item := &result.Items[i]
			if item.Quantity <= 0 || item.FlashSaleID == "" {
				continue
			}
			var stock int64
			var maxPerUser int
			var used int
			err := tx.QueryRow(ctx, `
				SELECT fs.name, fs.discount_percent, fs.stock, fs.max_per_user
				FROM flash_sale fs
				WHERE fs.id = $1 AND fs.is_active = true AND fs.deleted_at IS NULL
				  AND NOW() BETWEEN fs.start_time AND fs.end_time
				FOR UPDATE`, item.FlashSaleID).Scan(
				&item.Name, &item.DiscountPercent, &stock, &maxPerUser,
			)
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			if err != nil {
				return err
			}
			err = tx.QueryRow(ctx, `
				SELECT COALESCE((SELECT quantity FROM flash_sale_usage WHERE flash_sale_id = $1 AND user_id = $2), 0)`,
				item.FlashSaleID, allocation.UserID).Scan(&used)
			if err != nil {
				return err
			}

			quantity := domain.AllocatableQuantity(item.Quantity, stock, maxPerUser, used)
			if quantity <= 0 {
				continue
			}

			command, err := tx.Exec(ctx, `UPDATE flash_sale SET stock = stock - $2, updated_at = NOW() WHERE id = $1 AND stock >= $2`, item.FlashSaleID, quantity)
			if err != nil {
				return err
			}
			if command.RowsAffected() != 1 {
				return errors.New(constants.ErrFlashSaleStockInsufficient)
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO flash_sale_usage (id, flash_sale_id, user_id, quantity, created_at)
				VALUES (gen_random_uuid(), $1, $2, $3, NOW())
				ON CONFLICT (flash_sale_id, user_id) DO UPDATE
				SET quantity = flash_sale_usage.quantity + EXCLUDED.quantity, updated_at = NOW()`, item.FlashSaleID, allocation.UserID, quantity); err != nil {
				return err
			}
			item.AllocatedQuantity = quantity
			stock -= int64(quantity)
			used += quantity
			hasFlashSale = true
		}

		adjustedSubtotal := allocation.Subtotal
		for _, item := range result.Items {
			discountedUnitPrice := utils.DiscountedUnitPrice(item.UnitPrice, item.DiscountPercent)
			adjustedSubtotal -= (item.UnitPrice - discountedUnitPrice) * int64(item.AllocatedQuantity)
		}

		if allocation.PromoCode != "" {
			var quantity, usedCount, maxPerUser, userUsed int
			var discountValue, minPurchase, maxDiscount int64
			var active, canCombine bool
			var startDate, endDate time.Time
			err := tx.QueryRow(ctx, `
				SELECT p.id, p.name, p.discount_type, p.discount_value, p.min_purchase,
				       p.max_discount, p.quantity, p.used_count, p.max_usage_per_user,
				       p.can_combine_flash_sale, p.start_date, p.end_date, p.is_active
				FROM promo p
				WHERE p.code = $1 AND p.deleted_at IS NULL
				FOR UPDATE`, allocation.PromoCode).Scan(
				&result.PromoID, &result.PromoName, &result.PromoDiscountType, &discountValue,
				&minPurchase, &maxDiscount, &quantity, &usedCount, &maxPerUser, &canCombine,
				&startDate, &endDate, &active,
			)
			if errors.Is(err, pgx.ErrNoRows) {
				return errors.New(constants.ErrPromoNotFound)
			}
			if err != nil {
				return err
			}
			err = tx.QueryRow(ctx, `
				SELECT COALESCE((SELECT quantity FROM promo_usage WHERE promo_id = $1 AND user_id = $2), 0)`,
				result.PromoID, allocation.UserID).Scan(&userUsed)
			if err != nil {
				return err
			}
			result.PromoDiscountAmount, err = domain.ValidatePromoAllocation(domain.PromoAllocationInput{
				Active:              active,
				StartDate:           startDate,
				EndDate:             endDate,
				Quantity:            quantity,
				UsedCount:           usedCount,
				MaxPerUser:          maxPerUser,
				UserUsed:            userUsed,
				CanCombineFlashSale: canCombine,
				MinPurchase:         minPurchase,
				HasFlashSale:        hasFlashSale,
				Subtotal:            adjustedSubtotal,
				DiscountType:        result.PromoDiscountType,
				DiscountValue:       discountValue,
				MaxDiscount:         maxDiscount,
			}, time.Now())
			if err != nil {
				return err
			}

			command, err := tx.Exec(ctx, `
				UPDATE promo SET used_count = used_count + 1, updated_at = NOW()
				WHERE id = $1 AND (quantity = 0 OR used_count < quantity)`, result.PromoID)
			if err != nil {
				return err
			}
			if command.RowsAffected() != 1 {
				return errors.New(constants.ErrPromoLimitReached)
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO promo_usage (id, promo_id, user_id, quantity, created_at)
				VALUES (gen_random_uuid(), $1, $2, 1, NOW())
				ON CONFLICT (promo_id, user_id) DO UPDATE
				SET quantity = promo_usage.quantity + 1, updated_at = NOW()`, result.PromoID, allocation.UserID); err != nil {
				return err
			}
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO pricing_allocation
			(order_id, user_id, promo_id, promo_code, promo_name, promo_discount_type, promo_discount_amount, request_hash, status)
			VALUES ($1, $2, NULLIF($3, '')::uuid, $4, $5, $6, $7, $8, 'allocated')`,
			result.OrderID, result.UserID, result.PromoID, result.PromoCode, result.PromoName,
			result.PromoDiscountType, result.PromoDiscountAmount, result.RequestHash)
		if err != nil {
			return err
		}
		for _, item := range result.Items {
			if item.AllocatedQuantity == 0 {
				continue
			}
			_, err := tx.Exec(ctx, `
				INSERT INTO pricing_allocation_flash_sale
				(order_id, item_id, flash_sale_id, quantity, name, discount_percent)
				VALUES ($1, $2, $3, $4, $5, $6)`, result.OrderID, item.ItemID,
				item.FlashSaleID, item.AllocatedQuantity, item.Name, item.DiscountPercent)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return result, err
}

func (r *pricingAllocationRepository) ReleaseWithTx(ctx context.Context, orderID string, userID string) error {
	return r.store.ExecTx(ctx, func(tx postgresql.DBTX) error {
		command, err := tx.Exec(ctx, `
			INSERT INTO pricing_allocation (order_id, user_id, status, released_at)
			VALUES ($1, $2, 'released', NOW())
			ON CONFLICT (order_id) DO NOTHING`, orderID, userID)
		if err != nil {
			return err
		}
		if command.RowsAffected() == 1 {
			return nil
		}

		var allocatedUserID, promoID, allocationStatus string
		err = tx.QueryRow(ctx, `
			SELECT user_id, COALESCE(promo_id::text, ''), status
			FROM pricing_allocation WHERE order_id = $1 FOR UPDATE`, orderID).Scan(&allocatedUserID, &promoID, &allocationStatus)
		if err != nil {
			return err
		}
		if allocatedUserID != userID {
			return errors.New(constants.ErrAllocationOwnerMismatch)
		}
		if allocationStatus == "released" {
			return nil
		}

		rows, err := tx.Query(ctx, `
			SELECT flash_sale_id, quantity
			FROM pricing_allocation_flash_sale
			WHERE order_id = $1
			ORDER BY flash_sale_id, item_id`, orderID)
		if err != nil {
			return err
		}
		type releasedItem struct {
			id       string
			quantity int
		}
		var items []releasedItem
		for rows.Next() {
			var item releasedItem
			if err := rows.Scan(&item.id, &item.quantity); err != nil {
				rows.Close()
				return err
			}
			items = append(items, item)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		for _, item := range items {
			command, err := tx.Exec(ctx, `UPDATE flash_sale SET stock = stock + $2, updated_at = NOW() WHERE id = $1`, item.id, item.quantity)
			if err != nil {
				return err
			}
			if command.RowsAffected() != 1 {
				return errors.New(constants.ErrFlashSaleReleaseTargetMissing)
			}
			command, err = tx.Exec(ctx, `UPDATE flash_sale_usage SET quantity = quantity - $3, updated_at = NOW() WHERE flash_sale_id = $1 AND user_id = $2 AND quantity >= $3`, item.id, allocatedUserID, item.quantity)
			if err != nil {
				return err
			}
			if command.RowsAffected() != 1 {
				return errors.New(constants.ErrFlashSaleUsageTargetMissing)
			}
		}
		if promoID != "" {
			command, err := tx.Exec(ctx, `UPDATE promo SET used_count = used_count - 1, updated_at = NOW() WHERE id = $1 AND used_count > 0`, promoID)
			if err != nil {
				return err
			}
			if command.RowsAffected() != 1 {
				return errors.New(constants.ErrPromoReleaseTargetMissing)
			}
			command, err = tx.Exec(ctx, `UPDATE promo_usage SET quantity = quantity - 1, updated_at = NOW() WHERE promo_id = $1 AND user_id = $2 AND quantity > 0`, promoID, allocatedUserID)
			if err != nil {
				return err
			}
			if command.RowsAffected() != 1 {
				return errors.New(constants.ErrPromoUsageTargetMissing)
			}
		}
		command, err = tx.Exec(ctx, `UPDATE pricing_allocation SET status = 'released', released_at = NOW() WHERE order_id = $1 AND status = 'allocated'`, orderID)
		if err != nil {
			return err
		}
		if command.RowsAffected() != 1 {
			return errors.New(constants.ErrAllocationReleaseTargetMissing)
		}
		return nil
	})
}

func readAllocation(ctx context.Context, db postgresql.DBTX, orderID string) (*PricingAllocation, error) {
	var result PricingAllocation
	var status string
	err := db.QueryRow(ctx, `
		SELECT order_id, user_id, promo_code, COALESCE(promo_id::text, ''), promo_name, promo_discount_type,
		       promo_discount_amount, request_hash, status
		FROM pricing_allocation WHERE order_id = $1 FOR UPDATE`, orderID).Scan(
		&result.OrderID, &result.UserID, &result.PromoCode, &result.PromoID, &result.PromoName,
		&result.PromoDiscountType, &result.PromoDiscountAmount, &result.RequestHash, &status,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if status != "allocated" {
		return nil, errors.New(constants.ErrAllocationAlreadyReleased)
	}
	rows, err := db.Query(ctx, `
		SELECT item_id, flash_sale_id, quantity, name, discount_percent
		FROM pricing_allocation_flash_sale WHERE order_id = $1 ORDER BY item_id`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item PricingAllocationItem
		if err := rows.Scan(&item.ItemID, &item.FlashSaleID, &item.AllocatedQuantity, &item.Name, &item.DiscountPercent); err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	return &result, rows.Err()
}

func allocationFingerprint(allocation PricingAllocation) string {
	items := make([]domain.AllocationFingerprintItem, 0, len(allocation.Items))
	for _, item := range allocation.Items {
		items = append(items, domain.AllocationFingerprintItem{
			ItemID:      item.ItemID,
			FlashSaleID: item.FlashSaleID,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		})
	}
	return domain.AllocationFingerprint(allocation.UserID, allocation.PromoCode, allocation.Subtotal, items)
}
