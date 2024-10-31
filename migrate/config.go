package migrate

const (
	defaultTableName               = "evolution"
	ErrTooManyAppliedMigrations    = "too many applied migrations"
	createMigrationsTableQuery     = `CREATE TABLE IF NOT EXISTS %s (id INT8 NOT NULL, version VARCHAR(255) NOT NULL, PRIMARY KEY (id));`
	countAppliedMigrationsQuery    = "SELECT count(*) FROM %s"
	insertVersionQuery             = "INSERT INTO %s (id, version) VALUES ($1, $2)"
	logStartingTransactionTemplate = "migrator: error while starting transaction: %w"
	logRunningMigration            = "applying migration"
	logAppliedMigration            = "applied migration"
	logErrorMigrationTemplate      = "error executing golang migration %s: %w"
	logUpdateVersionTemplate       = "error updating migration versions: %w"
)
