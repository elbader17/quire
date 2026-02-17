// Package quire provides a database-like interface for Google Sheets.
// It allows CRUD operations, querying, and type-safe data mapping.
package quire

import (
	"context"
	"fmt"
)

// DB represents a database connection to a Google Sheet.
type DB struct {
	spreadsheetID string
	client        SheetsClient
}

// SheetsClient defines the interface for Google Sheets operations.
type SheetsClient interface {
	Read(ctx context.Context, range_ string) ([][]interface{}, error)
	Write(ctx context.Context, range_ string, values [][]interface{}) error
	Append(ctx context.Context, range_ string, values [][]interface{}) error
	Clear(ctx context.Context, range_ string) error
}

// Config holds database configuration.
type Config struct {
	SpreadsheetID string
	Credentials   []byte // Service account JSON
}

// New creates a new DB instance with the provided configuration.
func New(cfg Config) (*DB, error) {
	if cfg.SpreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet ID is required")
	}

	if len(cfg.Credentials) == 0 {
		return nil, fmt.Errorf("credentials are required")
	}

	client, err := newSheetsClient(cfg.Credentials, cfg.SpreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets client: %w", err)
	}

	return &DB{
		spreadsheetID: cfg.SpreadsheetID,
		client:        client,
	}, nil
}

// Table returns a Table handle for the specified sheet name.
func (db *DB) Table(name string) *Table {
	return &Table{
		db:   db,
		name: name,
	}
}

// Close releases any resources held by the database.
func (db *DB) Close() error {
	return nil
}
