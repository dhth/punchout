package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

type IssueLogEntry struct {
	Id          int
	IssueKey    string
	BeginTS     time.Time
	EndTS       time.Time
	Active      bool
	Synced      bool
	LastUpdated time.Time
}

func main() {
	db, err := sql.Open("sqlite", "punchoutdb")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if _, err = db.Exec(`
CREATE TABLE IF NOT EXISTS issue_log (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    issue_key TEXT NOT NULL,
    begin_ts TIMESTAMP NOT NULL,
    end_ts TIMESTAMP,
    active BOOLEAN NOT NULL,
    synced BOOLEAN NOT NULL,
    last_updated TIMESTAMP NOT NULL
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
		fmt.Println(err.Error())
		os.Exit(1)
	}

	entry := IssueLogEntry{
		IssueKey:    "WEBENG-1099",
		BeginTS:     time.Now().Add(time.Hour * -3),
		Active:      true,
		Synced:      false,
		LastUpdated: time.Now(),
	}

	insertEntry(db, entry)
	time.Sleep(time.Second * 10)
	entry.EndTS = time.Now()
	entry.Active = false
	entry.LastUpdated = time.Now()
	updateEntry(db, entry)

}

func insertEntry(db *sql.DB, entry IssueLogEntry) {
	stmt, err := db.Prepare(`
    INSERT INTO issue_log (issue_key, begin_ts, active, synced, last_updated)
    VALUES (?, ?, ?, ?, ?);
    `)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer stmt.Close()

	_, err = stmt.Exec(entry.IssueKey, entry.BeginTS, entry.Active, entry.Synced, entry.LastUpdated)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func updateEntry(db *sql.DB, entry IssueLogEntry) {
	stmt, err := db.Prepare(`
UPDATE issue_log
SET active = 0,
    end_ts = ?,
    last_updated = ?
WHERE issue_key = ?
AND active = 1;
    `)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer stmt.Close()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer stmt.Close()

	_, err = stmt.Exec(entry.EndTS, entry.LastUpdated, entry.IssueKey)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
