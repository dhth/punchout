package persistence

import (
	"database/sql"

	d "github.com/dhth/punchout/internal/domain"
)

func getNumActiveIssuesFromDB(db *sql.DB) (int, error) {
	row := db.QueryRow(`
SELECT
    COUNT(*)
FROM
    issue_log
WHERE
    active = 1
`)
	var numActiveIssues int
	err := row.Scan(&numActiveIssues)
	return numActiveIssues, err
}

func getWorkLogsForIssueFromDB(db *sql.DB, issueKey string) ([]d.StoredWorklog, error) {
	logEntries := make([]d.StoredWorklog, 0)

	rows, err := db.Query(`
SELECT
    ID,
    issue_key,
    begin_ts,
    end_ts,
    COMMENT,
    synced
FROM
    issue_log
WHERE
    issue_key =?
    AND active = false
ORDER BY
    end_ts DESC;
`, issueKey)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var entry d.StoredWorklog
		var comment sql.NullString
		err = rows.Scan(
			&entry.ID,
			&entry.IssueKey,
			&entry.BeginTS,
			&entry.EndTS,
			&comment,
			&entry.Synced,
		)
		if err != nil {
			return nil, err
		}
		entry.BeginTS = entry.BeginTS.Local()
		entry.EndTS = entry.EndTS.Local()
		entry.Comment = comment.String
		logEntries = append(logEntries, entry)
	}

	if iterErr := rows.Err(); iterErr != nil {
		return nil, iterErr
	}

	return logEntries, nil
}
