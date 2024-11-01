package pgx

import (
	"context"
	"github.com/CentralConcept/go-eventsourcing-pgx/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	migrations = []migrate.Migration{
		migrate.NewMigration("create events table",
			func(ctx context.Context, cmd migrate.Commands) error {
				if _, err := cmd.Exec(ctx, `create table events (seq SERIAL PRIMARY KEY, id VARCHAR NOT NULL, version INTEGER, reason VARCHAR, "type" VARCHAR, timestamp VARCHAR, data bytea, metadata bytea);`); err != nil {
					return err
				}
				return nil
			}),
		migrate.NewRawMigration(
			"create indices",
			`create unique index id_type_version on events (id, type, version);create index id_type on events (id, type);`,
		),
		migrate.NewMigration("Rename column type to aggregate_type",
			func(ctx context.Context, cmd migrate.Commands) error {
				if _, err := cmd.Exec(ctx, `alter table events rename column type to aggregate_type;`); err != nil {
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
