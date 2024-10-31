package migrate

import (
	"context"
	"fmt"
	"github.com/CentralConcept/go-eventsourcing-pgx/eventstore/pgx/logging"
	"github.com/jackc/pgx/v5"
	"time"
)

type MigrationManager struct {
	tableName  string
	migrations []Migration
	logger     logging.Logger
}

type Option func(*MigrationManager)

func TableName(tableName string) Option {
	return func(m *MigrationManager) {
		m.tableName = tableName
	}
}

func Log(logger logging.Logger) Option {
	return func(m *MigrationManager) {
		m.logger = logger
	}
}

func Migrations(migrations ...Migration) Option {
	return func(m *MigrationManager) {
		m.migrations = migrations
	}
}

func NewMigrationManager(opts ...Option) (*MigrationManager, error) {
	m := &MigrationManager{
		logger:    logging.LoggerFunc(logging.DefaultLogger),
		tableName: defaultTableName,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m, nil
}
func (m *MigrationManager) Migrate(ctx context.Context, db Conn) error {
	appliedCount, err := m.getAppliedCount(ctx, db)
	if err != nil {
		return err
	}

	if err = m.checkAppliedCount(appliedCount); err != nil {
		return err
	}

	missingMigrations := m.calculateMissingMigrations(appliedCount)
	m.logMissingMigrations(missingMigrations)

	return m.applyMissingMigrations(ctx, db, appliedCount)
}

func (m *MigrationManager) getAppliedCount(ctx context.Context, db Conn) (int, error) {
	return m.countApplied(ctx, db)
}

func (m *MigrationManager) checkAppliedCount(appliedCount int) error {
	if appliedCount > len(m.migrations) {
		return fmt.Errorf(ErrTooManyAppliedMigrations)
	}
	return nil
}

func (m *MigrationManager) calculateMissingMigrations(appliedCount int) int {
	return len(m.migrations) - appliedCount
}

func (m *MigrationManager) logMissingMigrations(missingMigrations int) {
	m.logger.Log("Initiating the execution of missing database migrations...", map[string]any{"missing": missingMigrations})
}

func (m *MigrationManager) applyMissingMigrations(ctx context.Context, db Conn, appliedCount int) error {
	for i, migration := range m.migrations[appliedCount:] {
		if err := m.applyMigration(ctx, db, migration, i+appliedCount); err != nil {
			return err
		}
	}
	return nil
}

func (m *MigrationManager) applyMigration(ctx context.Context, db Conn, migration Migration, version int) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf(logStartingTransactionTemplate, err)
	}
	defer tx.Rollback(ctx)

	if err = m.runMigration(ctx, tx, migration, version); err != nil {
		return fmt.Errorf("migrationManager: error while running migrations: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("migrationManager: transaction commit failed due to an error: %w", err)
	}
	return nil
}

func (m *MigrationManager) Pending(ctx context.Context, db Conn) ([]Migration, error) {
	appliedCount, err := m.countApplied(ctx, db)
	if err != nil {
		return nil, err
	}
	return m.migrations[appliedCount:], nil
}

func (m *MigrationManager) countApplied(ctx context.Context, db Conn) (int, error) {
	if _, err := db.Exec(ctx, fmt.Sprintf(createMigrationsTableQuery, m.tableName)); err != nil {
		return 0, err
	}
	var count int
	row := db.QueryRow(ctx, fmt.Sprintf(countAppliedMigrationsQuery, m.tableName))
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (m *MigrationManager) runMigration(ctx context.Context, tx pgx.Tx, migration Migration, id int) error {
	m.logger.Log(logRunningMigration, map[string]any{
		"id":   id,
		"name": migration.String(),
	})

	start := time.Now()
	if err := migration.Run(ctx, tx); err != nil {
		return fmt.Errorf(logErrorMigrationTemplate, migration.String(), err)
	}

	if _, err := tx.Exec(ctx, fmt.Sprintf(insertVersionQuery, m.tableName), id, migration.String()); err != nil {
		return fmt.Errorf(logUpdateVersionTemplate, err)
	}

	m.logger.Log(logAppliedMigration, map[string]any{
		"id":   id,
		"name": migration.String(),
		"took": time.Since(start),
	})
	return nil
}
