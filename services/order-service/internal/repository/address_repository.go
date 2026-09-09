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

type addressRepository struct {
	db postgresql.DBTX
}

func NewAddressRepository(store *postgresql.Store) IRepository.AddressRepository {
	return &addressRepository{
		db: store,
	}
}

func (r *addressRepository) List(ctx context.Context, params dto.ListAddressParams) ([]domain.Address, error) {
	const query = `
		SELECT id, user_id, street, city, province, postal_code, country, is_default, created_at, updated_at
		FROM address
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY is_default DESC, created_at DESC
	`
	rows, err := r.db.Query(ctx, query, params.UserId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []domain.Address
	for rows.Next() {
		var a domain.Address
		if err := rows.Scan(
			&a.Id,
			&a.UserId,
			&a.Street,
			&a.City,
			&a.Province,
			&a.PostalCode,
			&a.Country,
			&a.IsDefault,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		addresses = append(addresses, a)
	}

	return addresses, rows.Err()
}

func (r *addressRepository) Create(ctx context.Context, args dto.CreateAddressDto) error {
	const query = `
		INSERT INTO address (id, user_id, street, city, province, postal_code, country, is_default, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`
	_, err := r.db.Exec(ctx, query, args.Id, args.UserId, args.Street, args.City, args.Province, args.PostalCode, args.Country, args.IsDefault)
	if err != nil {
		return err
	}

	return nil
}

func (r *addressRepository) Update(ctx context.Context, id string, args dto.UpdateAddressDto) error {
	const query = `
		UPDATE address
		SET street = $3, city = $4, province = $5, postal_code = $6, country = $7, is_default = $8, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	cmdTag, err := r.db.Exec(ctx, query, id, args.UserId, args.Street, args.City, args.Province, args.PostalCode, args.Country, args.IsDefault)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New(constants.ErrAddressNotFound)
	}

	return nil
}

func (r *addressRepository) Delete(ctx context.Context, id string) error {
	const query = `
		UPDATE address
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New(constants.ErrAddressNotFound)
	}

	return nil
}

func (r *addressRepository) ReadById(ctx context.Context, id string) (*domain.Address, error) {
	const query = `
		SELECT id, user_id, street, city, province, postal_code, country, is_default, created_at, updated_at
		FROM address
		WHERE id = $1 AND deleted_at IS NULL
	`

	var a domain.Address
	err := r.db.QueryRow(ctx, query, id).Scan(
		&a.Id,
		&a.UserId,
		&a.Street,
		&a.City,
		&a.Province,
		&a.PostalCode,
		&a.Country,
		&a.IsDefault,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &a, nil
}
