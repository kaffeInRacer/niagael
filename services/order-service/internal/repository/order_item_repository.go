package repository

import (
	"context"
	"kaffein/order-service/internal/domain"
	"kaffein/order-service/internal/interfaces/IRepository"
	"kaffein/order-service/pkg/postgresql"
)

type orderItemRepository struct {
	db postgresql.DBTX
}

func NewOrderItemRepository(store *postgresql.Store) IRepository.OrderItemRepository {
	return &orderItemRepository{
		db: store,
	}
}

func (r *orderItemRepository) Create(ctx context.Context, args domain.OrderItem) error {
	const query = `
		INSERT INTO order_item (
			id, order_id, product_id, variant_id,
			product_name, product_price, quantity,
			flash_sale_id, flash_sale_name, flash_sale_discount_percent, flash_sale_discount_price, flash_sale_original_price, flash_sale_quantity,
			promo_id, promo_code, promo_name, promo_discount_type, promo_discount_amount,
			final_price, created_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18,
			$19, NOW()
		)
	`
	_, err := r.db.Exec(ctx, query,
		args.Id, args.OrderId, args.ProductId, args.VariantId,
		args.ProductName, args.ProductPrice, args.Quantity,
		args.FlashSaleId, args.FlashSaleName, args.FlashSaleDiscountPercent, args.FlashSaleDiscountPrice, args.FlashSaleOriginalPrice, args.FlashSaleQuantity,
		args.PromoId, args.PromoCode, args.PromoName, args.PromoDiscountType, args.PromoDiscountAmount,
		args.FinalPrice,
	)
	return err
}

func (r *orderItemRepository) ReadByOrderId(ctx context.Context, orderId string) ([]domain.OrderItem, error) {
	const query = `
		SELECT id, order_id, product_id, variant_id,
		       product_name, product_price, quantity,
		       flash_sale_id, flash_sale_name, flash_sale_discount_percent, flash_sale_discount_price, flash_sale_original_price, flash_sale_quantity,
		       promo_id, promo_code, promo_name, promo_discount_type, promo_discount_amount,
		       final_price, created_at, updated_at
		FROM order_item
		WHERE order_id = $1 AND deleted_at IS NULL
	`
	rows, err := r.db.Query(ctx, query, orderId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem

		if err := rows.Scan(
			&item.Id, &item.OrderId, &item.ProductId, &item.VariantId,
			&item.ProductName, &item.ProductPrice, &item.Quantity,
			&item.FlashSaleId, &item.FlashSaleName, &item.FlashSaleDiscountPercent,
			&item.FlashSaleDiscountPrice, &item.FlashSaleOriginalPrice, &item.FlashSaleQuantity,
			&item.PromoId, &item.PromoCode, &item.PromoName, &item.PromoDiscountType, &item.PromoDiscountAmount,
			&item.FinalPrice, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
