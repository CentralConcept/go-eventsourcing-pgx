package sql

import (
	"context"
	"github.com/CentralConcept/go-eventsourcing-pgx/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	migrations = []migrate.Migration{
		migrate.NewMigration("create events table",
			func(ctx context.Context, cmd migrate.Commands) error {
				if _, err := cmd.Exec(ctx, `create table if not exists snapshots (id VARCHAR NOT NULL, type VARCHAR, version INTEGER, global_version INTEGER, state bytea);`); err != nil {
					return err
				}
				return nil
			}),
	}
)

// Migrate the database
func (p *PostgresPgx) Migrate(s *pgxpool.Pool) error {
	migrator, err := migrate.NewMigrationManager(
		migrate.Migrations(migrations...),
	)
	if err != nil {
		return err
	}
	return migrator.Migrate(context.Background(), s)
}
