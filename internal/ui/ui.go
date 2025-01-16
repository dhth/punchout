package ui

import (
	"database/sql"
	"os"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	tea "github.com/charmbracelet/bubbletea"
)

func RenderUI(db *sql.DB, jiraClient *jira.Client, installationType JiraInstallationType, jql string, jiraTimeDeltaMins int, fallbackComment *string) error {
	debug := os.Getenv("DEBUG") == "1"
	if debug {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			return err
		}
		defer f.Close()
	}

	p := tea.NewProgram(InitialModel(db, jiraClient, installationType, jql, jiraTimeDeltaMins, fallbackComment, debug), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
