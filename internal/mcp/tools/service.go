package tools

import (
	"database/sql"

	d "github.com/dhth/punchout/internal/domain"
	svc "github.com/dhth/punchout/internal/service"
)

type Handler struct {
	DB      *sql.DB
	JiraSvc svc.Jira
	JiraCfg d.JiraConfig
}
