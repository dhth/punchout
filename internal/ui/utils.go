package ui

import (
	"database/sql"
	"strings"
	"time"
)

func RightPadTrim(s string, length int) string {
	if len(s) >= length {
		if length > 3 {
			return s[:length-3] + "..."
		}
		return s[:length]
	}
	return s + strings.Repeat(" ", length-len(s))
}

func Trim(s string, length int) string {
	if len(s) >= length {
		if length > 3 {
			return s[:length-3] + "..."
		}
		return s[:length]
	}
	return s
}

func insertNewEntry(db *sql.DB, issueKey string) error {

	stmt, err := db.Prepare(`
    INSERT INTO issue_log (issue_key, begin_ts, active, synced)
    VALUES (?, ?, ?, ?);
    `)

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(issueKey, time.Now(), true, 0)
	if err != nil {
		return err
	}

	return nil
}

func updateLastEntry(db *sql.DB, issueKey, comment string) error {
	stmt, err := db.Prepare(`
UPDATE issue_log
SET active = 0,
    end_ts = ?,
    comment = ?
WHERE issue_key = ?
AND active = 1;
`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), comment, issueKey)
	if err != nil {
		return err
	}

	return nil

}

func fetchEntries(db *sql.DB) ([]WorklogEntry, error) {

	var logEntries []WorklogEntry

	rows, err := db.Query(`
SELECT ID, issue_key, begin_ts, end_ts, comment, active, synced
FROM issue_log
WHERE active=false AND synced=false
ORDER by begin_ts DESC;
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry WorklogEntry
		err = rows.Scan(&entry.Id,
			&entry.IssueKey,
			&entry.BeginTS,
			&entry.EndTS,
			&entry.Comment,
			&entry.Active,
			&entry.Synced,
		)
		if err != nil {
			return nil, err
		}
		logEntries = append(logEntries, entry)

	}
	return logEntries, nil
}

func fetchSyncedEntries(db *sql.DB) ([]SyncedWorklogEntry, error) {

	var logEntries []SyncedWorklogEntry

	rows, err := db.Query(`
SELECT ID, issue_key, begin_ts, end_ts, comment
FROM issue_log
WHERE active=false AND synced=true
ORDER by begin_ts DESC LIMIT 30;
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry SyncedWorklogEntry
		err = rows.Scan(&entry.Id,
			&entry.IssueKey,
			&entry.BeginTS,
			&entry.EndTS,
			&entry.Comment,
		)
		if err != nil {
			return nil, err
		}
		logEntries = append(logEntries, entry)

	}
	return logEntries, nil
}

func deleteEntry(db *sql.DB, id int) error {

	stmt, err := db.Prepare(`
DELETE from issue_log
WHERE ID=?;
`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	return nil
}

func updateSyncStatus(db *sql.DB, id int) error {
	stmt, err := db.Prepare(`
UPDATE issue_log
SET synced = 1
WHERE id = ?;
`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	return nil

}
