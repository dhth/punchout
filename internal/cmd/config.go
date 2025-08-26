package cmd

import (
	"github.com/BurntSushi/toml"
)

const (
	jiraInstallationTypeOnPremise = "onpremise"
	jiraInstallationTypeCloud     = "cloud"
)

type jiraConfig struct {
	InstallationType  string  `toml:"installation_type"`
	JiraURL           *string `toml:"jira_url"`
	JQL               *string
	JiraTimeDeltaMins int     `toml:"jira_time_delta_mins"`
	JiraToken         *string `toml:"jira_token"`
	JiraUsername      *string `toml:"jira_username"`
	FallbackComment   *string `toml:"fallback_comment"`
}

type config struct {
	Jira jiraConfig `toml:"jira"`
}

func getConfig(filePath string) (config, error) {
	var config config
	_, err := toml.DecodeFile(filePath, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
