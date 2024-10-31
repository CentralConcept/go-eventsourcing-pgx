package pgx

import (
	"context"
	"errors"
	"fmt"
	"github.com/CentralConcept/eventsourcing/pgx/logging"
	"github.com/hallgren/eventsourcing/core"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"sync"
	"time"
)

// PostgresPgx is a struct that holds a connection pool to a PostgresSQL database and a mutex for locking.
type PostgresPgx struct {
	db     *pgxpool.Pool
	lock   *sync.Mutex
	logger logging.Logger
}

type PgxOption func(*PostgresPgx)

func WithLogger(logger logging.Logger) PgxOption {
	return func(p *PostgresPgx) {
		p.logger = logger
	}
}

func WithLock(lock *sync.Mutex) PgxOption {
	return func(p *PostgresPgx) {
		p.lock = lock
	}
}

func NewPostgresPgx(db *pgxpool.Pool, opts ...PgxOption) *PostgresPgx {
	p := &PostgresPgx{
		db:     db,
		logger: logging.LoggerFunc(logging.DefaultLogger),
	}
	for _, o := range opts {
		o(&PostgresPgx{})
	}
	return p
}

func (p *PostgresPgx) Save(events []core.Event) error {
	if len(events) == 0 {
		return nil
	}

	ctx := context.Background()

	if p.lock != nil {
		p.lock.Lock()
		defer p.lock.Unlock()
	}

	aggregateID := events[0].AggregateID
	aggregateType := events[0].AggregateType

	tx, txErr := p.db.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return fmt.Errorf("could not start a write transaction: %v", txErr)
	}
	defer func(tx pgx.Tx, ctx context.Context) {

		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			p.logger.Log("could not rollback transaction", map[string]interface{}{"error": err})
		}
	}(tx, ctx)

	var version int
	rowErr := tx.QueryRow(
		ctx,
		`SELECT version FROM events WHERE id = $1 AND "type" = $2 ORDER BY version DESC LIMIT 1`,
		aggregateID,
		aggregateType,
	).Scan(&version)
	if rowErr != nil && rowErr != pgx.ErrNoRows {
		return rowErr
	}
	currentVersion := core.Version(version)

	if core.Version(currentVersion)+1 != events[0].Version {
		return core.ErrConcurrency
	}

	insert := `INSERT INTO events (id, version, reason, "type", timestamp, data, metadata) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING seq`
	for i, event := range events {
		err := tx.QueryRow(
			ctx,
			insert,
			event.AggregateID,
			event.Version,
			event.Reason,
			event.AggregateType,
			event.Timestamp.Format(time.RFC3339),
			event.Data,
			event.Metadata,
		).Scan(&events[i].GlobalVersion)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// Get retrieves events from the PostgreSQL database based on the given aggregate ID, type, and version.
// It returns an iterator over the events and an error if any occurs.
func (p *PostgresPgx) Get(ctx context.Context, id string, aggregateType string, afterVersion core.Version) (core.Iterator, error) {
	selectStm := `SELECT seq, id, version, reason, type, timestamp, data, metadata FROM events WHERE id = $1 AND "type" = $2 AND version > $3 ORDER BY version`
	rows, err := p.db.Query(ctx, selectStm, id, aggregateType, afterVersion)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return &iterator{rows: rows}, nil
}

// All returns a function that iterates over all events in the GlobalEvents order starting from a given version.
// The function returns an iterator over the events and an error if any occurs.
func (p *PostgresPgx) All(ctx context.Context, start core.Version, count uint64) func() (core.Iterator, error) {
	return func() (core.Iterator, error) {
		rows, err := p.db.Query(
			ctx,
			`SELECT seq, id, version, reason, type, timestamp, data, metadata FROM events WHERE seq >= $1 ORDER BY seq LIMIT $2`,
			start,
			count,
		)
		if err != nil {
			return nil, err
		}
		return &iterator{rows: rows}, nil
	}
}
