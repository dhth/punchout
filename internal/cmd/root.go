package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	d "github.com/dhth/punchout/internal/domain"
	"github.com/dhth/punchout/internal/mcp"
	pers "github.com/dhth/punchout/internal/persistence"
	svc "github.com/dhth/punchout/internal/service"
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
		flagConfigFilePath       string
		flagDBPath               string
		flagFallbackComment      string
		flagJiraInstallationType string
		flagJiraTimeDeltaMinsStr string
		flagJiraToken            string
		flagJiraURL              string
		flagJiraUsername         string
		flagJQL                  string
		flagListConfig           bool

		cfg              config
		configPathFull   string
		dbPathFull       string
		installationType d.JiraInstallationType
	)

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntGetHomeDir, err.Error())
	}

	rootCmd := &cobra.Command{
		Use:           "punchout",
		Short:         "punchout takes the suck out of logging time on JIRA.",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if flagConfigFilePath == "" {
				return errConfigFilePathEmpty
			}

			if flagDBPath == "" {
				return errDBPathEmpty
			}

			dbPathFull = expandTilde(flagDBPath, userHomeDir)

			var jiraTimeDeltaMins int
			if flagJiraTimeDeltaMinsStr != "" {
				jiraTimeDeltaMins, err = strconv.Atoi(flagJiraTimeDeltaMinsStr)
				if err != nil {
					return fmt.Errorf("%w: %s", errTimeDeltaIncorrect, err.Error())
				}
			}

			configPathFull = expandTilde(flagConfigFilePath, userHomeDir)

			var err error
			cfg, err = getConfig(configPathFull)
			if err != nil {
				return fmt.Errorf("%w: %s", errCouldntParseConfigFile, err.Error())
			}

			if flagJiraInstallationType != "" {
				cfg.Jira.InstallationType = flagJiraInstallationType
			}

			if flagJiraURL != "" {
				cfg.Jira.JiraURL = &flagJiraURL
			}

			if flagJiraToken != "" {
				cfg.Jira.JiraToken = &flagJiraToken
			}

			if flagJiraUsername != "" {
				cfg.Jira.JiraUsername = &flagJiraUsername
			}

			if flagJQL != "" {
				cfg.Jira.JQL = &flagJQL
			}

			if flagJiraTimeDeltaMinsStr != "" {
				cfg.Jira.JiraTimeDeltaMins = jiraTimeDeltaMins
			}

			if flagFallbackComment != "" {
				cfg.Jira.FallbackComment = &flagFallbackComment
			}

			// validations
			switch cfg.Jira.InstallationType {
			case "", jiraInstallationTypeOnPremise: // "" to maintain backwards compatibility
				installationType = d.OnPremiseInstallation
				cfg.Jira.InstallationType = jiraInstallationTypeOnPremise
			case jiraInstallationTypeCloud:
				installationType = d.CloudInstallation
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

			if installationType == d.CloudInstallation && (cfg.Jira.JiraUsername == nil || *cfg.Jira.JiraUsername == "") {
				return fmt.Errorf("jira-username cannot be empty for cloud installation")
			}

			if cfg.Jira.FallbackComment != nil && strings.TrimSpace(*cfg.Jira.FallbackComment) == "" {
				return fmt.Errorf("fallback-comment cannot be empty")
			}

			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			if flagListConfig {
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

				if installationType == d.CloudInstallation {
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

			var jiraSvc svc.JiraSvc
			var svcErr error
			switch installationType {
			case d.OnPremiseInstallation:
				jiraSvc, svcErr = svc.NewOnPremJiraSvc(*cfg.Jira.JiraURL, *cfg.Jira.JiraToken)
			case d.CloudInstallation:
				jiraSvc, svcErr = svc.NewCloudJiraSvc(*cfg.Jira.JiraURL, *cfg.Jira.JiraUsername, *cfg.Jira.JiraToken)
			}

			if svcErr != nil {
				return svcErr
			}

			jiraCfg := d.JiraConfig{
				InstallationType: installationType,
				JQL:              *cfg.Jira.JQL,
				TimeDeltaMins:    cfg.Jira.JiraTimeDeltaMins,
				FallbackComment:  cfg.Jira.FallbackComment,
			}

			return ui.RenderUI(db, jiraSvc, jiraCfg)
		},
	}

	mcpCmd := &cobra.Command{
		Use:   "mcp <COMMAND>",
		Short: "Interact with punchout's MCP server",
	}

	mcpServerCmd := &cobra.Command{
		Use:   "server",
		Short: "Run punchout's MCP server",
		RunE: func(_ *cobra.Command, _ []string) error {
			db, err := pers.GetDB(dbPathFull)
			if err != nil {
				return fmt.Errorf("%w: %s", errCouldntCreateDB, err.Error())
			}

			err = pers.InitDB(db)
			if err != nil {
				return fmt.Errorf("%w: %s", errCouldntInitializeDB, err.Error())
			}

			var jiraSvc svc.JiraSvc
			var svcErr error
			switch installationType {
			case d.OnPremiseInstallation:
				jiraSvc, svcErr = svc.NewOnPremJiraSvc(*cfg.Jira.JiraURL, *cfg.Jira.JiraToken)
			case d.CloudInstallation:
				jiraSvc, svcErr = svc.NewCloudJiraSvc(*cfg.Jira.JiraURL, *cfg.Jira.JiraUsername, *cfg.Jira.JiraToken)
			}

			if svcErr != nil {
				return svcErr
			}

			jiraCfg := d.JiraConfig{
				InstallationType: installationType,
				JQL:              *cfg.Jira.JQL,
				TimeDeltaMins:    cfg.Jira.JiraTimeDeltaMins,
				FallbackComment:  cfg.Jira.FallbackComment,
			}

			return mcp.Serve(db, jiraSvc, jiraCfg)
		},
	}

	ros := runtime.GOOS
	var defaultConfigFilePath string

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntGetConfigDir, err.Error())
	}

	switch ros {
	case "darwin":
		// This is to maintain backwards compatibility with a decision made in the first release of punchout
		defaultConfigFilePath = filepath.Join(userHomeDir, ".config", configFileName)
	default:
		defaultConfigFilePath = filepath.Join(userConfigDir, configFileName)
	}

	dbFileName := fmt.Sprintf("punchout.v%s.db", pers.DBVersion)
	defaultDBPath := filepath.Join(userHomeDir, dbFileName)

	rootCmd.Flags().StringVarP(&flagConfigFilePath, "config-file-path", "", defaultConfigFilePath, "location of punchout's config file")
	rootCmd.Flags().StringVarP(&flagDBPath, "db-path", "", defaultDBPath, "location of punchout's local database")
	rootCmd.Flags().StringVarP(&flagJiraInstallationType, "jira-installation-type", "", "", "JIRA installation type; allowed values: [cloud, onpremise]")
	rootCmd.Flags().StringVarP(&flagJiraURL, "jira-url", "", "", "URL of the JIRA server")
	rootCmd.Flags().StringVarP(&flagJiraToken, "jira-token", "", "", "jira token (PAT for on-premise installation, API token for cloud installation)")
	rootCmd.Flags().StringVarP(&flagJiraUsername, "jira-username", "", "", "username for authentication (for cloud installation)")
	rootCmd.Flags().StringVarP(&flagJQL, "jql", "", "", "JQL to use to query issues")
	rootCmd.Flags().StringVarP(&flagFallbackComment, "fallback-comment", "", "", "fallback comment to use for worklog entries")
	rootCmd.Flags().StringVarP(&flagJiraTimeDeltaMinsStr, "jira-time-delta-mins", "", "", "time delta (in minutes) between your timezone and the timezone of the JIRA server; can be +/-")
	rootCmd.Flags().BoolVarP(&flagListConfig, "list-config", "", false, "print the config that punchout will use")

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	mcpCmd.AddCommand(mcpServerCmd)
	rootCmd.AddCommand(mcpCmd)

	return rootCmd, nil
}
