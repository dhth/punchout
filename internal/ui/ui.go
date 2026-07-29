package ui

import (
	"database/sql"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/dhth/punchout/internal/config"
	svc "github.com/dhth/punchout/internal/service"
)

func RenderUI(db *sql.DB, jiraSvc svc.Jira, jiraOpts config.JiraOptions) error {
	debug := os.Getenv("DEBUG") == "1"
	if debug {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			return err
		}
		defer f.Close()
	}

	p := tea.NewProgram(InitialModel(db, jiraSvc, jiraOpts, debug))
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
