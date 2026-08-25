package ui

import (
	"context"
	"database/sql"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/dhth/punchout/internal/issuecache"
	svc "github.com/dhth/punchout/internal/service"
	"github.com/dhth/punchout/internal/ui/theme"
)

func RenderUI(
	ctx context.Context,
	db *sql.DB,
	worklogStore WorklogStore,
	jiraSvc svc.Jira,
	issueStore issuecache.Store,
	opts Options,
	thm theme.Theme,
) error {
	tuiCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	debug := os.Getenv("DEBUG") == "1"
	if debug {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			return err
		}
		defer f.Close()
	}

	p := tea.NewProgram(
		InitialModel(tuiCtx, db, worklogStore, jiraSvc, issueStore, opts, thm, debug),
		tea.WithContext(tuiCtx),
	)
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
