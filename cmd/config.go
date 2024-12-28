package cmd

import (
	"github.com/BurntSushi/toml"
)

type JiraConfig struct {
	JiraURL           *string `toml:"jira_url"`
	JiraToken         *string `toml:"jira_token"`
	Jql               *string
	JiraTimeDeltaMins int `toml:"jira_time_delta_mins"`
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
