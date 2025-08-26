package mcp

import (
	"database/sql"
	"fmt"

	d "github.com/dhth/punchout/internal/domain"
	svc "github.com/dhth/punchout/internal/service"
)

func Serve(_ *sql.DB, _ svc.Jira, _ d.JiraConfig) error {
	fmt.Println("serving...")
	return nil
}
