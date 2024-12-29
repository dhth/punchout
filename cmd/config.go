package cmd

import (
	"github.com/BurntSushi/toml"
)

type JiraConfig struct {
	JiraURL           *string `toml:"jira_url"`
	Jql               *string
	JiraTimeDeltaMins int     `toml:"jira_time_delta_mins"`
	JiraToken         *string `toml:"jira_token"`
	JiraCloudToken    *string `toml:"jira_cloud_token"`
	JiraCloudUsername *string `toml:"jira_cloud_username"`
}

type POConfig struct {
	Jira JiraConfig
}

func readConfig(filePath string) (POConfig, error) {

	var config POConfig
	_, err := toml.DecodeFile(expandTilde(filePath), &config)
	if err != nil {
		return config, err
	}

	return config, nil

}
