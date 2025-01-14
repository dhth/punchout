package persistence

import (
	"database/sql"
	"errors"
	"time"

	c "github.com/dhth/punchout/internal/common"
)

var (
	ErrNoTaskIsActive           = errors.New("no task is active")
	ErrCouldntStopActiveTask    = errors.New("couldn't stop active task")
	ErrCouldntStartTrackingTask = errors.New("couldn't start tracking task")
)

func getNumActiveIssues(db *sql.DB) (int, error) {
	row := db.QueryRow(`
SELECT COUNT(*)
from issue_log
WHERE active=1
`)
	var numActiveIssues int
	err := row.Scan(&numActiveIssues)
	return numActiveIssues, err
}

func getWorklogEntriesForIssue(db *sql.DB, issueKey string) ([]c.WorklogEntry, error) {
	var logEntries []c.WorklogEntry

	rows, err := db.Query(`
SELECT ID, issue_key, begin_ts, end_ts, comment, active, synced
FROM issue_log
WHERE issue_key=?
ORDER by end_ts DESC;
`, issueKey)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var entry c.WorklogEntry
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

	if iterErr := rows.Err(); iterErr != nil {
		return nil, iterErr
	}

	return logEntries, nil
}

func InsertNewEntry(db *sql.DB, issueKey string, beginTs time.Time) error {
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

func UpdateLastEntry(db *sql.DB, issueKey, comment string, beginTs, endTs time.Time) error {
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

func StopCurrentlyActiveEntry(db *sql.DB, issueKey string, endTs time.Time) error {
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

func FetchEntries(db *sql.DB) ([]c.WorklogEntry, error) {
	var logEntries []c.WorklogEntry

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
		var entry c.WorklogEntry
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

	if iterErr := rows.Err(); iterErr != nil {
		return nil, iterErr
	}

	return logEntries, nil
}

func FetchSyncedEntries(db *sql.DB) ([]c.SyncedWorklogEntry, error) {
	var logEntries []c.SyncedWorklogEntry

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
		var entry c.SyncedWorklogEntry
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

	if iterErr := rows.Err(); iterErr != nil {
		return nil, iterErr
	}

	return logEntries, nil
}

func DeleteEntry(db *sql.DB, id int) error {
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

func UpdateSyncStatus(db *sql.DB, id int) error {
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

func DeleteActiveLogInDB(db *sql.DB) error {
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

func GetActiveIssue(db *sql.DB) (string, error) {
	row := db.QueryRow(`
SELECT issue_key
from issue_log
WHERE active=1
ORDER BY begin_ts DESC
LIMIT 1
`)
	var activeIssue string
	err := row.Scan(&activeIssue)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNoTaskIsActive
	} else if err != nil {
		return "", err
	}
	return activeIssue, nil
}

func QuickSwitchActiveIssue(db *sql.DB, currentIssue, selectedIssue string, currentTime time.Time) error {
	err := StopCurrentlyActiveEntry(db, currentIssue, currentTime)
	if err != nil {
		return ErrCouldntStopActiveTask
	}

	return InsertNewEntry(db, selectedIssue, currentTime)
}
