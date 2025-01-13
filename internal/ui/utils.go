package ui

import (
	"database/sql"
	"fmt"
	"math"
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

func insertNewEntry(db *sql.DB, issueKey string, beginTs time.Time) error {
	stmt, err := db.Prepare(`
    INSERT INTO issue_log (issue_key, begin_ts, active, synced)
    VALUES (?, ?, ?, ?);
`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(issueKey, beginTs, true, 0)
	if err != nil {
		return err
	}

	return nil
}

func updateLastEntry(db *sql.DB, issueKey, comment string, beginTs, endTs time.Time) error {
	stmt, err := db.Prepare(`
UPDATE issue_log
SET active = 0,
    begin_ts = ?,
    end_ts = ?,
    comment = ?
WHERE issue_key = ?
AND active = 1;
`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(beginTs, endTs, comment, issueKey)
	if err != nil {
		return err
	}

	return nil
}

func stopCurrentlyActiveEntry(db *sql.DB, issueKey string, endTs time.Time) error {
	stmt, err := db.Prepare(`
UPDATE issue_log
SET active = 0,
    end_ts = ?,
    comment = ''
WHERE issue_key = ?
AND active = 1;
`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(endTs, issueKey)
	if err != nil {
		return err
	}

	return nil
}

func fetchEntries(db *sql.DB) ([]worklogEntry, error) {
	var logEntries []worklogEntry

	rows, err := db.Query(`
SELECT ID, issue_key, begin_ts, end_ts, comment, active, synced
FROM issue_log
WHERE active=false AND synced=false
ORDER by end_ts DESC;
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry worklogEntry
		err = rows.Scan(&entry.ID,
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

func fetchSyncedEntries(db *sql.DB) ([]syncedWorklogEntry, error) {
	var logEntries []syncedWorklogEntry

	rows, err := db.Query(`
SELECT ID, issue_key, begin_ts, end_ts, comment
FROM issue_log
WHERE active=false AND synced=true
ORDER by end_ts DESC LIMIT 30;
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry syncedWorklogEntry
		err = rows.Scan(&entry.ID,
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

func humanizeDuration(durationInSecs int) string {
	duration := time.Duration(durationInSecs) * time.Second

	if duration.Seconds() < 60 {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}

	if duration.Minutes() < 60 {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}

	modMins := int(math.Mod(duration.Minutes(), 60))

	if modMins == 0 {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	}

	return fmt.Sprintf("%dh %dm", int(duration.Hours()), modMins)
}
