package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dhth/punchout/internal/domain"
)

var ErrIssueHasNoActiveWorklog = errors.New("issue doesn't have an active worklog")

func (s *SQLiteStore) ActiveWorklog(ctx context.Context) (*domain.InProgressWorklog, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT
    issue_key,
    begin_ts,
    comment
FROM
    issue_log
WHERE
    active = true
ORDER BY
    begin_ts DESC
LIMIT
    1;
`)

	var worklog domain.InProgressWorklog
	var comment sql.NullString
	if err := row.Scan(&worklog.IssueKey, &worklog.BeginTS, &comment); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	worklog.BeginTS = worklog.BeginTS.Local()
	worklog.Comment = comment.String

	return &worklog, nil
}

func (s *SQLiteStore) StartWorklog(ctx context.Context, issueKey string, beginTS time.Time) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO
    issue_log (issue_key, begin_ts, active, synced)
VALUES
    (?, ?, true, false);
`, issueKey, beginTS.UTC())

	return err
}

func (s *SQLiteStore) FinishWorklog(ctx context.Context, worklog domain.Worklog) error {
	result, err := s.db.ExecContext(ctx, `
UPDATE
    issue_log
SET
    active = false,
    begin_ts = ?,
    end_ts = ?,
    comment = ?
WHERE
    issue_key = ?
    AND active = true;
`, worklog.BeginTS.UTC(), worklog.EndTS.UTC(), worklog.Comment, worklog.IssueKey)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: %s", ErrIssueHasNoActiveWorklog, worklog.IssueKey)
	}

	return nil
}

func (s *SQLiteStore) AddWorklog(ctx context.Context, worklog domain.Worklog) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO
    issue_log (
        issue_key,
        begin_ts,
        end_ts,
        comment,
        active,
        synced
    )
VALUES
    (?, ?, ?, ?, false, false);
`,
		worklog.IssueKey,
		worklog.BeginTS.UTC(),
		worklog.EndTS.UTC(),
		worklog.Comment,
	)

	return err
}

func (s *SQLiteStore) UnsyncedWorklogs(ctx context.Context) ([]domain.StoredWorklog, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
    ID,
    issue_key,
    begin_ts,
    end_ts,
    comment
FROM
    issue_log
WHERE
    active = false
    AND synced = false
ORDER BY
    end_ts DESC;
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	worklogs := make([]domain.StoredWorklog, 0)
	for rows.Next() {
		var worklog domain.StoredWorklog
		var comment sql.NullString
		if err := rows.Scan(
			&worklog.ID,
			&worklog.IssueKey,
			&worklog.BeginTS,
			&worklog.EndTS,
			&comment,
		); err != nil {
			return nil, err
		}

		worklog.BeginTS = worklog.BeginTS.Local()
		worklog.EndTS = worklog.EndTS.Local()
		worklog.Comment = comment.String
		worklogs = append(worklogs, worklog)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return worklogs, nil
}

func (s *SQLiteStore) MarkWorklogSynced(ctx context.Context, id int, comment *string) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE
    issue_log
SET
    synced = true,
    comment = COALESCE(?, comment)
WHERE
    ID = ?;
`, comment, id)

	return err
}
