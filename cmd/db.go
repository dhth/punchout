package cmd

import "database/sql"

const (
	PUNCHOUT_DB_VERSION = "1"
)

func setupDB(dbpath string) (*sql.DB, error) {

	db, err := sql.Open("sqlite", dbpath)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err != nil {
		return nil, err
	}

	if _, err = db.Exec(`
CREATE TABLE IF NOT EXISTS issue_log (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    issue_key TEXT NOT NULL,
    begin_ts TIMESTAMP NOT NULL,
    end_ts TIMESTAMP,
    comment VARCHAR(255),
    active BOOLEAN NOT NULL,
    synced BOOLEAN NOT NULL
);

CREATE TRIGGER IF NOT EXISTS prevent_duplicate_active_insert
BEFORE INSERT ON issue_log
BEGIN
    SELECT CASE
        WHEN EXISTS (SELECT 1 FROM issue_log WHERE active = 1)
        THEN RAISE(ABORT, 'Only one row with active=1 is allowed')
    END;
END;
`); err != nil {
		return nil, err
	}
	return db, nil
}
