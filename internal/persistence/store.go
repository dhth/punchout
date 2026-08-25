package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // sqlite driver
)

var (
	errCouldntOpenDB       = errors.New("couldn't open punchout database")
	errCouldntInitializeDB = errors.New("couldn't initialize database")
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbpath string) (*SQLiteStore, error) {
	dbDir := filepath.Dir(dbpath)
	if err := os.MkdirAll(dbDir, 0o700); err != nil {
		return nil, fmt.Errorf("couldn't create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbpath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntOpenDB, err.Error())
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := initDB(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("%w: %s", errCouldntInitializeDB, err.Error())
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// DB temporarily exposes the underlying database during the TUI store migration.
func (s *SQLiteStore) DB() *sql.DB {
	return s.db
}
