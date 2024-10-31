package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/CentralConcept/go-eventsourcing-pgx/logging"
	"github.com/hallgren/eventsourcing/core"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"sync"
)

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

// Save persists the snapshot
func (s *PostgresPgx) Save(snapshot core.Snapshot) error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return errors.New(fmt.Sprintf("could not start a write transaction, %v", err))
	}
	defer tx.Rollback(ctx)

	statement := `SELECT id from snapshots where id=$1 AND "type"=$2 LIMIT 1`
	var id string
	err = tx.QueryRow(ctx, statement, snapshot.ID, snapshot.Type).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		// insert
		statement = `INSERT INTO snapshots (state, id, type, version, global_version) VALUES ($1, $2, $3, $4, $5)`
		_, err = tx.Exec(ctx, statement, string(snapshot.State), snapshot.ID, snapshot.Type, snapshot.Version, snapshot.GlobalVersion)
		if err != nil {
			return err
		}
	} else {
		// update
		statement = `UPDATE snapshots set state=$1, version=$2, global_version=$3 where id=$4 AND type=$5`
		_, err = tx.Exec(ctx, statement, string(snapshot.State), snapshot.Version, snapshot.GlobalVersion, snapshot.ID, snapshot.Type)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// Get return the snapshot data from the database
func (s *PostgresPgx) Get(ctx context.Context, aggregateID, aggregateType string) (core.Snapshot, error) {
	var globalVersion core.Version
	var version core.Version
	var state []byte

	selectStm := `Select version, global_version, state from snapshots where id=? and type=?`
	err := s.db.QueryRow(ctx, selectStm, aggregateID, aggregateType).
		Scan(&version, &globalVersion, &state)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return core.Snapshot{}, core.ErrSnapshotNotFound
	} else if err != nil {
		return core.Snapshot{}, err
	}

	return core.Snapshot{
		ID:            aggregateID,
		Type:          aggregateType,
		Version:       version,
		GlobalVersion: globalVersion,
		State:         state,
	}, nil
}
