package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

type Store struct {
	DBTX
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		DBTX: pool,
		pool: pool,
	}
}

func (s *Store) ExecTx(ctx context.Context, fn func(DBTX) error) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		return fn(tx)
	})
}

func (s *Store) ExecTxWithOptions(ctx context.Context, options pgx.TxOptions, fn func(DBTX) error) error {
	return pgx.BeginTxFunc(ctx, s.pool, options, func(tx pgx.Tx) error {
		return fn(tx)
	})
}
