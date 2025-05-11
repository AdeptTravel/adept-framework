package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Store struct{ DB *sql.DB }

func (s Store) Add(ctx context.Context, tx *sql.Tx, topic string, evt any) error {
	bytes, _ := json.Marshal(evt)
	_, err := tx.ExecContext(ctx,
		`INSERT INTO event_outbox(id,topic,payload) VALUES (?,?,?)`,
		uuid.New().String(), topic, bytes)
	return err
}

// Pump moves unpublished rows to the in-memory bus.
func (s Store) Pump(ctx context.Context, bus interface {
	Publish(context.Context, string, any)
}) {
	tick := time.NewTicker(200 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{})
			if err != nil {
				log.Error().Err(err).Msg("begin")
				continue
			}
			rows, _ := tx.QueryContext(ctx,
				`SELECT id,topic,payload FROM event_outbox
				  WHERE published_at IS NULL
				  FOR UPDATE SKIP LOCKED LIMIT 100`)
			var (
				id, topic string
				data      []byte
			)
			for rows.Next() {
				_ = rows.Scan(&id, &topic, &data)
				var anyJson map[string]any
				_ = json.Unmarshal(data, &anyJson)
				bus.Publish(ctx, topic, anyJson)
				_, _ = tx.ExecContext(ctx,
					`UPDATE event_outbox SET published_at=NOW() WHERE id=?`, id)
			}
			_ = tx.Commit()
		}
	}
}
