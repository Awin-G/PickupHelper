package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// DBTX is the common interface between *sqlx.DB and *sqlx.Tx. Repository
// methods accept DBTX so the service layer can pass either a plain
// connection or a transaction started via WithTx. Both *sqlx.DB and
// *sqlx.Tx satisfy this interface.
type DBTX interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

// BeginTx starts a new transaction with the default isolation level.
func BeginTx(ctx context.Context, db *sqlx.DB) (*sqlx.Tx, error) {
	return db.BeginTxx(ctx, nil)
}

// WithTx runs fn inside a transaction. If fn returns an error or panics,
// the transaction is rolled back; otherwise it is committed. The provided
// db is used only to begin the transaction — fn receives the *sqlx.Tx
// (which also implements DBTX) and should pass it to repository methods
// that participate in the tx.
func WithTx(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) (err error) {
	if db == nil {
		return fmt.Errorf("repository.WithTx: nil db")
	}
	tx, err := BeginTx(ctx, db)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
			return
		}
		if cerr := tx.Commit(); cerr != nil {
			err = fmt.Errorf("commit tx: %w", cerr)
		}
	}()
	err = fn(tx)
	return err
}
