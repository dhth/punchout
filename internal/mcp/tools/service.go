package tools

import (
	"database/sql"

	"github.com/dhth/punchout/internal/config"
	svc "github.com/dhth/punchout/internal/service"
)

type Handler struct {
	DB       *sql.DB
	JiraSvc  svc.Jira
	JiraOpts config.JiraOptions
}
