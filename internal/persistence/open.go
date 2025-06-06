package persistence

import (
	"database/sql"
)

func GetDB(dbpath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbpath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	return db, nil
}
