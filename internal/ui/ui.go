package ui

import (
	"database/sql"
	"os"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	tea "github.com/charmbracelet/bubbletea"
)

func RenderUI(db *sql.DB, jiraClient *jira.Client, installationType JiraInstallationType, jql string, jiraTimeDeltaMins int) error {
	if len(os.Getenv("DEBUG_LOG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			return err
		}
		defer f.Close()
	}

	debug := os.Getenv("DEBUG") == "true"
	p := tea.NewProgram(InitialModel(db, jiraClient, installationType, jql, jiraTimeDeltaMins, debug), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
