package repository

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

// MigrationsDir is the default relative path to migration scripts.
const MigrationsDir = "migrations"

// RunMigrations applies all up migrations in dir.
func RunMigrations(db *sql.DB, dir string) error {
	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}
	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

// RollbackMigrations rolls back the most recent migration.
func RollbackMigrations(db *sql.DB, dir string) error {
	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}
	if err := goose.Down(db, dir); err != nil {
		return fmt.Errorf("goose down: %w", err)
	}
	return nil
}

// ResetMigrations rolls back all migrations.
func ResetMigrations(db *sql.DB, dir string) error {
	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}
	if err := goose.Reset(db, dir); err != nil {
		return fmt.Errorf("goose reset: %w", err)
	}
	return nil
}
