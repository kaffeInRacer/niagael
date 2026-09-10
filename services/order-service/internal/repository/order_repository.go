package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"kaffein/order-service/internal/domain"
	"kaffein/order-service/internal/dto"
	"kaffein/order-service/internal/interfaces/IRepository"
	"kaffein/order-service/pkg/postgresql"
	"kaffein/order-service/utils/constants"
)

type orderRepository struct {
	db    postgresql.DBTX
	store *postgresql.Store
}

func NewOrderRepository(store *postgresql.Store) IRepository.OrderRepository {
	return &orderRepository{
		db:    store,
		store: store,
	}
}

func (r *orderRepository) CreateWithItemsWithTx(ctx context.Context, order domain.Order, items []domain.OrderItem) error {
	return r.store.ExecTx(ctx, func(tx postgresql.DBTX) error {
		return insertOrderWithItems(ctx, tx, order, items)
	})
}

func insertOrderWithItems(ctx context.Context, tx postgresql.DBTX, order domain.Order, items []domain.OrderItem) error {
	const orderQuery = `
		INSERT INTO orders (id, order_ref, user_id, address_id, total_amount, status, snap_token, expire_time, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`
	if _, err := tx.Exec(ctx, orderQuery,
		order.Id,
		order.OrderRef,
		order.UserId,
		order.AddressId,
		order.TotalAmount,
		order.Status,
		order.SnapToken,
		order.ExpireTime,
	); err != nil {
		return err
	}

	const itemQuery = `
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
	for _, item := range items {
		if _, err := tx.Exec(ctx, itemQuery,
			item.Id, item.OrderId, item.ProductId, item.VariantId,
			item.ProductName, item.ProductPrice, item.Quantity,
			item.FlashSaleId, item.FlashSaleName, item.FlashSaleDiscountPercent, item.FlashSaleDiscountPrice, item.FlashSaleOriginalPrice, item.FlashSaleQuantity,
			item.PromoId, item.PromoCode, item.PromoName, item.PromoDiscountType, item.PromoDiscountAmount,
			item.FinalPrice,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *orderRepository) Create(ctx context.Context, args domain.Order) error {
	const query = `
		INSERT INTO orders (id, order_ref, user_id, address_id, total_amount, status, snap_token, expire_time, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`
	_, err := r.db.Exec(ctx, query,
		args.Id,
		args.OrderRef,
		args.UserId,
		args.AddressId,
		args.TotalAmount,
		args.Status,
		args.SnapToken,
		args.ExpireTime,
	)

	return err
}

func (r *orderRepository) Update(ctx context.Context, id string, args domain.Order) error {
	const query = `
		UPDATE orders
		SET total_amount = $2, status = $3, snap_token = $4, expire_time = $5, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, id, args.TotalAmount, args.Status, args.SnapToken, args.ExpireTime)
	return err
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id string, fromStatuses []string, status string) error {
	const query = `
		UPDATE orders
		SET status = $3, updated_at = NOW()
		WHERE id = $1
		  AND status = ANY($2::text[])
		  AND deleted_at IS NULL
	`
	cmdTag, err := r.db.Exec(ctx, query, id, fromStatuses, status)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New(constants.ErrInvalidOrderStatusTransition)
	}

	return nil
}

func (r *orderRepository) ReadById(ctx context.Context, id string) (*domain.Order, error) {
	const query = `
		SELECT id, order_ref, user_id, address_id, total_amount, status, snap_token, expire_time, created_at, updated_at
		FROM orders
		WHERE id = $1 AND deleted_at IS NULL
	`

	var o domain.Order
	err := r.db.QueryRow(ctx, query, id).Scan(
		&o.Id,
		&o.OrderRef,
		&o.UserId,
		&o.AddressId,
		&o.TotalAmount,
		&o.Status,
		&o.SnapToken,
		&o.ExpireTime,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (r *orderRepository) ReadByUserId(ctx context.Context, userId string, before string, pageSize int32) ([]domain.Order, error) {
	const query = `
		SELECT id, order_ref, user_id, address_id, total_amount, status, snap_token, expire_time, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		AND (NULLIF($2::text, '') IS NULL OR created_at < NULLIF($2::text, '')::timestamptz)
		AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $3::int
	`
	rows, err := r.db.Query(ctx, query, userId, before, pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(
			&o.Id,
			&o.OrderRef,
			&o.UserId,
			&o.AddressId,
			&o.TotalAmount,
			&o.Status,
			&o.SnapToken,
			&o.ExpireTime,
			&o.CreatedAt,
			&o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, rows.Err()
}

func (r *orderRepository) ReadItemsByOrderId(ctx context.Context, orderId string) ([]domain.OrderItem, error) {
	const query = `
		SELECT id, order_id, product_id, variant_id,
		       product_name, product_price, quantity,
		       flash_sale_id, flash_sale_name, flash_sale_discount_percent, flash_sale_discount_price, flash_sale_original_price, flash_sale_quantity,
		       promo_id, promo_code, promo_name, promo_discount_type, promo_discount_amount,
		       final_price, created_at, updated_at
		FROM order_item
		WHERE order_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
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
			&item.FlashSaleId, &item.FlashSaleName, &item.FlashSaleDiscountPercent, &item.FlashSaleDiscountPrice, &item.FlashSaleOriginalPrice, &item.FlashSaleQuantity,
			&item.PromoId, &item.PromoCode, &item.PromoName, &item.PromoDiscountType, &item.PromoDiscountAmount,
			&item.FinalPrice, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *orderRepository) List(ctx context.Context, params dto.ListOrderParams) ([]domain.Order, error) {
	const query = `
		SELECT o.id, o.order_ref, o.user_id, o.address_id, o.total_amount, o.status, o.snap_token, o.expire_time, o.created_at, o.updated_at, COALESCE(b.email, '')
		FROM orders o
		LEFT JOIN buyers b ON b.id::text = o.user_id
		WHERE (
			NULLIF($1::text, '') IS NULL
			OR o.order_ref ILIKE '%' || $1::text || '%'
			OR b.email::text ILIKE '%' || $1::text || '%'
		)
		AND (NULLIF($2::text, '') IS NULL OR o.user_id = $2::text)
		AND (NULLIF($3::text, '') IS NULL OR o.status = $3::text)
		AND o.deleted_at IS NULL
		ORDER BY
		  CASE WHEN $4::text = 'total_amount' AND $5::text = 'asc'  THEN total_amount END ASC,
		  CASE WHEN $4::text = 'created_at'   AND $5::text = 'asc'  THEN created_at   END ASC,
		  CASE WHEN $4::text = 'total_amount' AND $5::text = 'desc' THEN total_amount END DESC,
		  CASE WHEN $4::text = 'created_at'   AND $5::text = 'desc' THEN created_at   END DESC,
		created_at DESC
		LIMIT $6::int
		OFFSET $7::int
	`
	rows, err := r.db.Query(ctx, query,
		params.Search,
		params.UserId,
		params.Status,
		params.OrderBy,
		params.OrderDir,
		params.PageSize,
		params.PageOffset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(
			&o.Id,
			&o.OrderRef,
			&o.UserId,
			&o.AddressId,
			&o.TotalAmount,
			&o.Status,
			&o.SnapToken,
			&o.ExpireTime,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.BuyerEmail,
		); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, rows.Err()
}

func (r *orderRepository) ListCount(ctx context.Context, params dto.ListOrderParams) (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM orders o
		LEFT JOIN buyers b ON b.id::text = o.user_id
		WHERE (
			NULLIF($1::text, '') IS NULL
			OR o.order_ref ILIKE '%' || $1::text || '%'
			OR b.email::text ILIKE '%' || $1::text || '%'
		)
		AND (NULLIF($2::text, '') IS NULL OR o.user_id = $2::text)
		AND (NULLIF($3::text, '') IS NULL OR o.status = $3::text)
		AND o.deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query, params.Search, params.UserId, params.Status).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
