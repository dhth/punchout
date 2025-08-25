package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	jiraCloud "github.com/andygrunwald/go-jira/v2/cloud"
	jiraOnPremise "github.com/andygrunwald/go-jira/v2/onpremise"
	pers "github.com/dhth/punchout/internal/persistence"
	"github.com/dhth/punchout/internal/ui"
	"github.com/spf13/cobra"
)

const (
	configFileName = "punchout/punchout.toml"
)

var (
	errCouldntGetHomeDir       = errors.New("couldn't get your home directory")
	errCouldntGetConfigDir     = errors.New("couldn't get your config directory")
	errConfigFilePathEmpty     = errors.New("config file path cannot be empty")
	errDBPathEmpty             = errors.New("db file path cannot be empty")
	errCouldntInitializeDB     = errors.New("couldn't initialize database")
	errTimeDeltaIncorrect      = errors.New("couldn't convert time delta to a number")
	errCouldntParseConfigFile  = errors.New("couldn't parse config file")
	errInvalidInstallationType = fmt.Errorf("invalid value for jira installation type (allowed values: [%s, %s])", jiraInstallationTypeOnPremise, jiraInstallationTypeCloud)
	errCouldntCreateDB         = errors.New("couldn't create punchout database")
	errCouldntCreateJiraClient = errors.New("couldn't create JIRA client")
)

func Execute() error {
	rootCmd, err := NewRootCommand()
	if err != nil {
		return err
	}

	return rootCmd.Execute()
}

func NewRootCommand() (*cobra.Command, error) {
	var (
		configFilePath       string
		dbPath               string
		jiraInstallationType string
		jiraURL              string
		jiraToken            string
		jiraUsername         string
		jql                  string
		fallbackComment      string
		jiraTimeDeltaMinsStr string
		listConfig           bool
	)

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntGetHomeDir, err.Error())
	}

	defaultConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntGetConfigDir, err.Error())
	}

	rootCmd := &cobra.Command{
		Use:           "punchout",
		Short:         "punchout takes the suck out of logging time on JIRA.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			if configFilePath == "" {
				return errConfigFilePathEmpty
			}

			if dbPath == "" {
				return errDBPathEmpty
			}

			dbPathFull := expandTilde(dbPath, userHomeDir)

			var jiraTimeDeltaMins int
			if jiraTimeDeltaMinsStr != "" {
				jiraTimeDeltaMins, err = strconv.Atoi(jiraTimeDeltaMinsStr)
				if err != nil {
					return fmt.Errorf("%w: %s", errTimeDeltaIncorrect, err.Error())
				}
			}

			configPathFull := expandTilde(configFilePath, userHomeDir)

			cfg, err := getConfig(configPathFull)
			if err != nil {
				return fmt.Errorf("%w: %s", errCouldntParseConfigFile, err.Error())
			}

			if jiraInstallationType != "" {
				cfg.Jira.InstallationType = jiraInstallationType
			}

			if jiraURL != "" {
				cfg.Jira.JiraURL = &jiraURL
			}

			if jiraToken != "" {
				cfg.Jira.JiraToken = &jiraToken
			}

			if jiraUsername != "" {
				cfg.Jira.JiraUsername = &jiraUsername
			}

			if jql != "" {
				cfg.Jira.JQL = &jql
			}

			if jiraTimeDeltaMinsStr != "" {
				cfg.Jira.JiraTimeDeltaMins = jiraTimeDeltaMins
			}

			if fallbackComment != "" {
				cfg.Jira.FallbackComment = &fallbackComment
			}

			// validations
			var installationType ui.JiraInstallationType
			switch cfg.Jira.InstallationType {
			case "", jiraInstallationTypeOnPremise: // "" to maintain backwards compatibility
				installationType = ui.OnPremiseInstallation
				cfg.Jira.InstallationType = jiraInstallationTypeOnPremise
			case jiraInstallationTypeCloud:
				installationType = ui.CloudInstallation
			default:
				return errInvalidInstallationType
			}

			if cfg.Jira.JiraURL == nil || *cfg.Jira.JiraURL == "" {
				return fmt.Errorf("jira-url cannot be empty")
			}

			if cfg.Jira.JQL == nil || *cfg.Jira.JQL == "" {
				return fmt.Errorf("jql cannot be empty")
			}

			if cfg.Jira.JiraToken == nil || *cfg.Jira.JiraToken == "" {
				return fmt.Errorf("jira-token cannot be empty")
			}

			if installationType == ui.CloudInstallation && (cfg.Jira.JiraUsername == nil || *cfg.Jira.JiraUsername == "") {
				return fmt.Errorf("jira-username cannot be empty for cloud installation")
			}

			if cfg.Jira.FallbackComment != nil && strings.TrimSpace(*cfg.Jira.FallbackComment) == "" {
				return fmt.Errorf("fallback-comment cannot be empty")
			}

			if listConfig {
				fmt.Fprintf(os.Stdout, `Config:

Config File Path                        %s
DB File Path                            %s
JIRA Installation Type                  %s
JIRA URL                                %s
JIRA Token                              %s
JQL                                     %s
JIRA Time Delta Mins                    %d
`,
					configPathFull,
					dbPathFull,
					cfg.Jira.InstallationType,
					*cfg.Jira.JiraURL,
					*cfg.Jira.JiraToken,
					*cfg.Jira.JQL,
					cfg.Jira.JiraTimeDeltaMins)

				if installationType == ui.CloudInstallation {
					fmt.Fprintf(os.Stdout, "JIRA Username                           %s\n", *cfg.Jira.JiraUsername)
				}

				if cfg.Jira.FallbackComment != nil {
					fmt.Fprintf(os.Stdout, "Fallback Comment                        %s\n", *cfg.Jira.FallbackComment)
				}
				return nil
			}

			db, err := pers.GetDB(dbPathFull)
			if err != nil {
				return fmt.Errorf("%w: %s", errCouldntCreateDB, err.Error())
			}

			err = pers.InitDB(db)
			if err != nil {
				return fmt.Errorf("%w: %s", errCouldntInitializeDB, err.Error())
			}

			var httpClient *http.Client
			switch installationType {
			case ui.OnPremiseInstallation:
				tp := jiraOnPremise.BearerAuthTransport{
					Token: *cfg.Jira.JiraToken,
				}
				httpClient = tp.Client()
			case ui.CloudInstallation:
				tp := jiraCloud.BasicAuthTransport{
					Username: *cfg.Jira.JiraUsername,
					APIToken: *cfg.Jira.JiraToken,
				}
				httpClient = tp.Client()
			}

			// Using the on-premise client regardless of the user's installation type
			// The APIs between the two installation types seem to differ, but this
			// seems to be alright for punchout's use case. If this situation changes,
			// this will need to be refactored.
			// https://github.com/andygrunwald/go-jira/issues/473
			cl, err := jiraOnPremise.NewClient(*cfg.Jira.JiraURL, httpClient)
			if err != nil {
				return fmt.Errorf("%w: %s", errCouldntCreateJiraClient, err.Error())
			}

			return ui.RenderUI(db, cl, installationType, *cfg.Jira.JQL, cfg.Jira.JiraTimeDeltaMins, cfg.Jira.FallbackComment)
		},
	}

	ros := runtime.GOOS
	var defaultConfigFilePath string

	switch ros {
	case "darwin":
		// This is to maintain backwards compatibility with a decision made in the first release of punchout
		defaultConfigFilePath = filepath.Join(userHomeDir, ".config", configFileName)
	default:
		defaultConfigFilePath = filepath.Join(defaultConfigDir, configFileName)
	}

	dbFileName := fmt.Sprintf("punchout.v%s.db", pers.DBVersion)
	defaultDBPath := filepath.Join(userHomeDir, dbFileName)

	rootCmd.Flags().StringVarP(&configFilePath, "config-file-path", "", defaultConfigFilePath, "location of punchout's config file")
	rootCmd.Flags().StringVarP(&dbPath, "db-path", "", defaultDBPath, "location of punchout's local database")
	rootCmd.Flags().StringVarP(&jiraInstallationType, "jira-installation-type", "", "", "JIRA installation type; allowed values: [cloud, onpremise]")
	rootCmd.Flags().StringVarP(&jiraURL, "jira-url", "", "", "URL of the JIRA server")
	rootCmd.Flags().StringVarP(&jiraToken, "jira-token", "", "", "jira token (PAT for on-premise installation, API token for cloud installation)")
	rootCmd.Flags().StringVarP(&jiraUsername, "jira-username", "", "", "username for authentication (for cloud installation)")
	rootCmd.Flags().StringVarP(&jql, "jql", "", "", "JQL to use to query issues")
	rootCmd.Flags().StringVarP(&fallbackComment, "fallback-comment", "", "", "fallback comment to use for worklog entries")
	rootCmd.Flags().StringVarP(&jiraTimeDeltaMinsStr, "jira-time-delta-mins", "", "", "time delta (in minutes) between your timezone and the timezone of the JIRA server; can be +/-")
	rootCmd.Flags().BoolVarP(&listConfig, "list-config", "", false, "print the config that punchout will use")

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	return rootCmd, nil
}
