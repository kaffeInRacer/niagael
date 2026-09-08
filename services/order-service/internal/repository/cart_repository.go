package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"kaffein/order-service/internal/domain"
	"kaffein/order-service/internal/interfaces/IRepository"
	"kaffein/order-service/pkg/postgresql"
	"kaffein/order-service/utils/constants"
)

type cartRepository struct {
	db    postgresql.DBTX
	store *postgresql.Store
}

func NewCartRepository(store *postgresql.Store) IRepository.CartRepository {
	return &cartRepository{db: store, store: store}
}

func (r *cartRepository) ReadByUserId(ctx context.Context, userId string) (*domain.Cart, error) {
	const cartQuery = `
		SELECT id, user_id, created_at, updated_at
		FROM cart
		WHERE user_id = $1
	`

	var cart domain.Cart
	if err := r.db.QueryRow(ctx, cartQuery, userId).Scan(
		&cart.Id,
		&cart.UserId,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	const itemsQuery = `
		SELECT id, cart_id, product_id, variant_id, quantity, created_at, updated_at
		FROM cart_item
		WHERE cart_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, itemsQuery, cart.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.CartItem
		if err := rows.Scan(
			&item.Id,
			&item.CartId,
			&item.ProductId,
			&item.VariantId,
			&item.Quantity,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		cart.Items = append(cart.Items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) AddItem(ctx context.Context, userId string, cartId string, item domain.CartItem, maxQuantity int) error {
	return r.store.ExecTx(ctx, func(tx postgresql.DBTX) error {
		const cartQuery = `
			INSERT INTO cart (id, user_id, created_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (user_id) DO UPDATE SET updated_at = NOW()
			RETURNING id
		`
		if err := tx.QueryRow(ctx, cartQuery, cartId, userId).Scan(&item.CartId); err != nil {
			return err
		}

		const itemQuery = `
			INSERT INTO cart_item (id, cart_id, product_id, variant_id, quantity, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (cart_id, product_id, variant_id) DO UPDATE
			SET quantity = cart_item.quantity + EXCLUDED.quantity,
				updated_at = NOW()
			WHERE cart_item.quantity + EXCLUDED.quantity <= $6
		`
		cmdTag, err := tx.Exec(ctx, itemQuery, item.Id, item.CartId, item.ProductId, item.VariantId, item.Quantity, maxQuantity)
		if err != nil {
			return err
		}
		if cmdTag.RowsAffected() == 0 {
			return errors.New(constants.ErrCartQuantityLimit)
		}
		return nil
	})
}

func (r *cartRepository) UpdateItem(ctx context.Context, itemId string, userId string, quantity int, maxQuantity int) error {
	const query = `
		UPDATE cart_item AS ci
		SET quantity = $3::int, updated_at = NOW()
		FROM cart AS c
		WHERE ci.id = $1::uuid
		  AND ci.cart_id = c.id
		  AND c.user_id = $2::text
		  AND $3::int <= $4::int
	`
	cmdTag, err := r.db.Exec(ctx, query, itemId, userId, quantity, maxQuantity)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New(constants.ErrCartItemNotFound)
	}
	return nil
}

func (r *cartRepository) DeleteItem(ctx context.Context, itemId string, userId string) error {
	const query = `
		DELETE FROM cart_item AS ci
		USING cart AS c
		WHERE ci.id = $1
		  AND ci.cart_id = c.id
		  AND c.user_id = $2
	`
	cmdTag, err := r.db.Exec(ctx, query, itemId, userId)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New(constants.ErrCartItemNotFound)
	}
	return nil
}

func (r *cartRepository) Clear(ctx context.Context, userId string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM cart WHERE user_id = $1`, userId)
	return err
}
