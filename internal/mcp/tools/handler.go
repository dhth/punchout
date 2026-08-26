package tools

import (
	"github.com/dhth/punchout/internal/config"
	svc "github.com/dhth/punchout/internal/service"
)

type Handler struct {
	store    WorklogStore
	jiraSvc  svc.Jira
	jiraOpts config.JiraOptions
}

func NewHandler(store WorklogStore, jiraSvc svc.Jira, jiraOpts config.JiraOptions) *Handler {
	return &Handler{
		store:    store,
		jiraSvc:  jiraSvc,
		jiraOpts: jiraOpts,
	}
}
