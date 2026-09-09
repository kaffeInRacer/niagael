package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"kaffein/order-service/pkg/postgresql"
)

type OutboxRepository struct {
	db    *pgxpool.Pool
	store *postgresql.Store
}

func NewOutboxRepository(pool *pgxpool.Pool, store *postgresql.Store) *OutboxRepository {
	return &OutboxRepository{db: pool, store: store}
}

type OutboxEvent struct {
	ID      int64
	Topic   string
	Key     string
	Payload []byte
}

const insertOutboxQuery = `
	INSERT INTO outbox_events (topic, key, payload)
	VALUES ($1::varchar, $2::varchar, $3::jsonb)
`

func (r *OutboxRepository) InsertWithTx(ctx context.Context, tx postgresql.DBTX, topic, key string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal outbox payload: %w", err)
	}
	_, err = tx.Exec(ctx, insertOutboxQuery, topic, key, data)
	return err
}

func (r *OutboxRepository) Insert(ctx context.Context, topic, key string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal outbox payload: %w", err)
	}
	_, err = r.db.Exec(ctx, insertOutboxQuery, topic, key, data)
	return err
}

func (r *OutboxRepository) ListUnprocessed(ctx context.Context, limit int) ([]OutboxEvent, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, topic, key, payload FROM outbox_events ORDER BY id LIMIT $1 FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []OutboxEvent
	for rows.Next() {
		var e OutboxEvent
		if err := rows.Scan(&e.ID, &e.Topic, &e.Key, &e.Payload); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *OutboxRepository) MarkProcessed(ctx context.Context, ids []int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM outbox_events WHERE id = ANY($1::bigint[])`, ids)
	return err
}
