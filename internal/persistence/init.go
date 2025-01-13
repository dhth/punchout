package persistence

import "database/sql"

const (
	DBVersion = "1"
)

func InitDB(db *sql.DB) error {
	_, err := db.Exec(`
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
`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
DELETE from issue_log 
WHERE end_ts < DATE('now', '-60 days');
`)
	return err
}
