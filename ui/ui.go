package ui

import (
	"database/sql"
	"fmt"
	"os"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	tea "github.com/charmbracelet/bubbletea"
)

func RenderUI(db *sql.DB, jiraClient *jira.Client, jql string, jiraTimeDeltaMins int) {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	p := tea.NewProgram(InitialModel(db, jiraClient, jql, jiraTimeDeltaMins), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there has been an error: %v", err)
		os.Exit(1)
	}
}
