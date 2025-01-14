package cmd

import (
	"github.com/BurntSushi/toml"
)

const (
	jiraInstallationTypeOnPremise = "onpremise"
	jiraInstallationTypeCloud     = "cloud"
)

type JiraConfig struct {
	InstallationType  string  `toml:"installation_type"`
	JiraURL           *string `toml:"jira_url"`
	JQL               *string
	JiraTimeDeltaMins int     `toml:"jira_time_delta_mins"`
	JiraToken         *string `toml:"jira_token"`
	JiraUsername      *string `toml:"jira_username"`
	FallbackComment   *string `toml:"fallback_comment"`
}

type POConfig struct {
	Jira JiraConfig
}

func getConfig(filePath string) (POConfig, error) {
	var config POConfig
	_, err := toml.DecodeFile(filePath, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
