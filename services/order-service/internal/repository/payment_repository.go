package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"kaffein/order-service/internal/domain"
	"kaffein/order-service/internal/interfaces/IRepository"
	"kaffein/order-service/pkg/postgresql"
)

type paymentRepository struct {
	db postgresql.DBTX
}

func NewPaymentRepository(store *postgresql.Store) IRepository.PaymentRepository {
	return &paymentRepository{
		db: store,
	}
}

func (r *paymentRepository) Create(ctx context.Context, args domain.Payment) error {
	const query = `
		INSERT INTO payment (id, order_id, amount, method, status, snap_token, va_number, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	_, err := r.db.Exec(ctx, query, args.Id, args.OrderId, args.Amount, args.Method, args.Status, args.SnapToken, args.VaNumber)
	return err
}

func (r *paymentRepository) ReadByOrderId(ctx context.Context, orderId string) (*domain.Payment, error) {
	const query = `
		SELECT id, order_id, amount, method, status, snap_token, va_number, paid_at, created_at, updated_at
		FROM payment
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var p domain.Payment
	err := r.db.QueryRow(ctx, query, orderId).Scan(
		&p.Id, &p.OrderId, &p.Amount, &p.Method, &p.Status,
		&p.SnapToken, &p.VaNumber, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *paymentRepository) ReadById(ctx context.Context, id string) (*domain.Payment, error) {
	const query = `
		SELECT id, order_id, amount, method, status, snap_token, va_number, paid_at, created_at, updated_at
		FROM payment
		WHERE id = $1
	`

	var p domain.Payment
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.Id, &p.OrderId, &p.Amount, &p.Method, &p.Status,
		&p.SnapToken, &p.VaNumber, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	const query = `
		UPDATE payment
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id, status)
	return err
}
