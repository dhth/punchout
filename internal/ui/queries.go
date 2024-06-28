package ui

import (
	"database/sql"
)

func deleteActiveLogInDB(db *sql.DB) error {

	stmt, err := db.Prepare(`
DELETE FROM issue_log
WHERE active=true;
`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()

	return err
}
