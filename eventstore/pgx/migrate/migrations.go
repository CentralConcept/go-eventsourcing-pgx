package migrate

import (
	"context"
	"fmt"
)

// MigrationCommands is a wrapper around a function so that it implements the Migration interface.
type MigrationCommands func(context.Context, Commands) error

// Migration is the migration interface that all migrations must implement.
type Migration interface {
	fmt.Stringer
	Run(context.Context, Commands) error
}

// migrationFunctionAction is a struct that holds a migration function and its name.
type migrationFunctionAction struct {
	Fn   MigrationCommands // The migration function to be executed.
	Name string            // The name of the migration.
}

// Run executes the migration function.
func (m *migrationFunctionAction) Run(ctx context.Context, tx Commands) error {
	return m.Fn(ctx, tx)
}

// String returns the name of the migration.
func (m *migrationFunctionAction) String() string {
	return m.Name
}

// NewMigration creates a migration from a function.
// name: The name of the migration.
// fn: The migration function to be executed.
func NewMigration(name string, fn MigrationCommands) Migration {
	return &migrationFunctionAction{
		Name: name,
		Fn:   fn,
	}
}

// NewRawMigration creates a migration from a raw SQL string.
// name: The name of the migration.
// sql: The raw SQL string to be executed as the migration.
func NewRawMigration(name, sql string) Migration {
	return &migrationFunctionAction{
		Name: name,
		Fn:   func(ctx context.Context, tx Commands) error { _, err := tx.Exec(ctx, sql); return err },
	}
}
