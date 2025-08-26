package persistence

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	errCouldntOpenDB       = errors.New("couldn't open punchout database")
	errCouldntInitializeDB = errors.New("couldn't initialize database")
)

func GetDB(dbpath string) (*sql.DB, error) {
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
