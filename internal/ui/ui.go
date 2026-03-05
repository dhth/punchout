package ui

import (
	"database/sql"
	"os"

	tea "charm.land/bubbletea/v2"
	d "github.com/dhth/punchout/internal/domain"
	svc "github.com/dhth/punchout/internal/service"
)

func RenderUI(db *sql.DB, jiraSvc svc.Jira, jiraCfg d.JiraConfig) error {
	debug := os.Getenv("DEBUG") == "1"
	if debug {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			return err
		}
		defer f.Close()
	}

	p := tea.NewProgram(InitialModel(db, jiraSvc, jiraCfg, debug))
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
