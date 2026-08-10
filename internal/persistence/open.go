package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	errCouldntOpenDB       = errors.New("couldn't open punchout database")
	errCouldntInitializeDB = errors.New("couldn't initialize database")
)

func GetDB(dbpath string) (*sql.DB, error) {
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

	err = InitDB(db)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntInitializeDB, err.Error())
	}

	return db, nil
}
