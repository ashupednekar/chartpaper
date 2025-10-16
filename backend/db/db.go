package db

import (
	"context"
	"database/sql"
	"fmt"
)

// DBTX is an interface for database operations
// type DBTX interface {
// 	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
// 	PrepareContext(context.Context, string) (*sql.Stmt, error)
// 	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
// 	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
// }

// New creates a new Queries object and sets the search_path
func New(db *sql.DB) *Queries {
	// Set search_path for the connection
	_, err := db.Exec("SET search_path TO chartpaper")
	if err != nil {
		// This is a critical error, so we might want to panic
		panic(fmt.Sprintf("Failed to set search_path: %v", err))
	}
	return &Queries{db: db}
}

// Queries wraps the database connection
// type Queries struct {
// 	db DBTX
// }

// NewTx creates a new Queries object from a transaction
func NewTx(tx *sql.Tx) *Queries {
	return &Queries{db: tx}
}

// WithTx returns a new Queries object with the given transaction
// func (q *Queries) WithTx(tx *sql.Tx) *Queries {
// 	return &Queries{
// 		db: tx,
// 	}
// }
