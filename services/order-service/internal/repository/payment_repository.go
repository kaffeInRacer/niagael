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

type paymentRepository struct {
	db    postgresql.DBTX
	store *postgresql.Store
}

func NewPaymentRepository(store *postgresql.Store) IRepository.PaymentRepository {
	return &paymentRepository{
		db:    store,
		store: store,
	}
}

func (r *paymentRepository) GetOrCreate(ctx context.Context, args domain.Payment) (*domain.Payment, error) {
	const query = `
		INSERT INTO payment (id, order_id, amount, method, status, snap_token, va_number, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (order_id) DO NOTHING
	`
	if _, err := r.db.Exec(ctx, query, args.Id, args.OrderId, args.Amount, args.Method, args.Status, args.SnapToken, args.VaNumber); err != nil {
		return nil, err
	}
	return r.ReadByOrderId(ctx, args.OrderId)
}

func (r *paymentRepository) InitiateWithTx(ctx context.Context, orderId string, createTransaction func(*domain.Payment) (string, string, error)) (string, error) {
	var redirectUrl string
	err := r.store.ExecTx(ctx, func(tx postgresql.DBTX) error {
		const lockQuery = `
			SELECT id, order_id, amount, method, status, COALESCE(snap_token, ''), COALESCE(redirect_url, ''), va_number, paid_at, created_at, updated_at
			FROM payment
			WHERE order_id = $1
			FOR UPDATE
		`
		var payment domain.Payment
		err := tx.QueryRow(ctx, lockQuery, orderId).Scan(
			&payment.Id, &payment.OrderId, &payment.Amount, &payment.Method, &payment.Status,
			&payment.SnapToken, &payment.RedirectUrl, &payment.VaNumber, &payment.PaidAt, &payment.CreatedAt, &payment.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New(constants.ErrPaymentNotFound)
		}

		if err != nil {
			return err
		}

		if payment.RedirectUrl != "" {
			redirectUrl = payment.RedirectUrl
			return nil
		}

		token, url, err := createTransaction(&payment)
		if err != nil {
			return err
		}
		const updateQuery = `
			UPDATE payment
			SET snap_token = $2, redirect_url = $3, updated_at = NOW()
			WHERE id = $1
		`
		if _, err := tx.Exec(ctx, updateQuery, payment.Id, token, url); err != nil {
			return err
		}
		const orderQuery = `
			UPDATE orders
			SET snap_token = $2, updated_at = NOW()
			WHERE id = $1 AND status = 'pending' AND deleted_at IS NULL
		`
		tag, err := tx.Exec(ctx, orderQuery, orderId, token)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return errors.New(constants.ErrInvalidOrderStatusTransition)
		}
		redirectUrl = url
		return nil
	})
	return redirectUrl, err
}

func (r *paymentRepository) ReadByOrderId(ctx context.Context, orderId string) (*domain.Payment, error) {
	const query = `
		SELECT id, order_id, amount, method, status, COALESCE(snap_token, ''), COALESCE(redirect_url, ''), va_number, paid_at, created_at, updated_at
		FROM payment
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var p domain.Payment
	err := r.db.QueryRow(ctx, query, orderId).Scan(
		&p.Id, &p.OrderId, &p.Amount, &p.Method, &p.Status,
		&p.SnapToken, &p.RedirectUrl, &p.VaNumber, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
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
		SELECT id, order_id, amount, method, status, COALESCE(snap_token, ''), COALESCE(redirect_url, ''), va_number, paid_at, created_at, updated_at
		FROM payment
		WHERE id = $1
	`

	var p domain.Payment
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.Id, &p.OrderId, &p.Amount, &p.Method, &p.Status,
		&p.SnapToken, &p.RedirectUrl, &p.VaNumber, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *paymentRepository) ProcessCallbackWithTx(ctx context.Context, orderId string, amount int64, paymentStatus string, stockAction func(tx postgresql.DBTX) error) (bool, error) {
	processed := false
	err := r.store.ExecTx(ctx, func(tx postgresql.DBTX) error {
		const lockQuery = `
			SELECT p.id, p.amount, p.status, o.total_amount, o.status
			FROM payment p
			JOIN orders o ON o.id = p.order_id
			WHERE p.order_id = $1 AND o.deleted_at IS NULL
			FOR UPDATE OF p, o
		`

		var paymentId, currentPaymentStatus, currentOrderStatus string
		var paymentAmount, orderAmount int64
		err := tx.QueryRow(ctx, lockQuery, orderId).Scan(
			&paymentId, &paymentAmount, &currentPaymentStatus, &orderAmount, &currentOrderStatus,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New(constants.ErrPaymentNotFound)
		}

		if err != nil {
			return err
		}
		if amount != paymentAmount || amount != orderAmount {
			return domain.ErrInvalidPaymentAmount
		}
		if !domain.PaymentTransitionAllowed(currentPaymentStatus, paymentStatus) || currentOrderStatus != "pending" {
			return nil
		}

		if err := stockAction(tx); err != nil {
			return err
		}

		orderStatus := "cancelled"
		if paymentStatus == domain.PaymentStatusPaid {
			orderStatus = "paid"
		}
		const paymentQuery = `
			UPDATE payment
			SET status = $2::varchar, paid_at = CASE WHEN $2::varchar = 'paid' THEN NOW() ELSE paid_at END, updated_at = NOW()
			WHERE id = $1 AND status = 'pending'
		`
		paymentTag, err := tx.Exec(ctx, paymentQuery, paymentId, paymentStatus)
		if err != nil {
			return err
		}

		if paymentTag.RowsAffected() != 1 {
			return errors.New(constants.ErrInvalidOrderStatusTransition)
		}
		const orderQuery = `
			UPDATE orders
			SET status = $2, updated_at = NOW()
			WHERE id = $1 AND status = 'pending' AND deleted_at IS NULL
		`
		orderTag, err := tx.Exec(ctx, orderQuery, orderId, orderStatus)
		if err != nil {
			return err
		}

		if orderTag.RowsAffected() != 1 {
			return errors.New(constants.ErrInvalidOrderStatusTransition)
		}
		processed = true
		return nil
	})
	return processed, err
}
