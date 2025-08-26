package cmd

import (
	"github.com/BurntSushi/toml"
)

const (
	jiraInstallationTypeOnPremise = "onpremise"
	jiraInstallationTypeCloud     = "cloud"
)

type userJiraConfig struct {
	InstallationType  string  `toml:"installation_type"`
	JiraURL           *string `toml:"jira_url"`
	JQL               *string
	JiraTimeDeltaMins int     `toml:"jira_time_delta_mins"`
	JiraToken         *string `toml:"jira_token"`
	JiraUsername      *string `toml:"jira_username"`
	FallbackComment   *string `toml:"fallback_comment"`
}

type userConfig struct {
	Jira userJiraConfig `toml:"jira"`
}

func getConfig(filePath string) (userConfig, error) {
	var config userConfig
	_, err := toml.DecodeFile(filePath, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
