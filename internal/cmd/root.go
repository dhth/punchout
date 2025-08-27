package cmd

import (
	"database/sql"
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
	errTimeDeltaIncorrect      = errors.New("couldn't convert time delta to a number")
	errCouldntParseConfigFile  = errors.New("couldn't parse config file")
	errInvalidInstallationType = fmt.Errorf("invalid value for jira installation type (allowed values: [%s, %s])", jiraInstallationTypeOnPremise, jiraInstallationTypeCloud)
	errInvalidMCPTransport     = fmt.Errorf("invalid value provided for MCP transport")
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

		flagMcpTransportStr string
		flagMcpServerPort   uint

		userCfg        userConfig
		jiraCfg        d.JiraConfig
		configPathFull string
		dbPathFull     string
		db             *sql.DB
		jiraSvc        svc.Jira
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
			userCfg, err = getConfig(configPathFull)
			if err != nil {
				return fmt.Errorf("%w: %s", errCouldntParseConfigFile, err.Error())
			}

			if flagJiraInstallationType != "" {
				userCfg.Jira.InstallationType = flagJiraInstallationType
			}

			if flagJiraURL != "" {
				userCfg.Jira.JiraURL = &flagJiraURL
			}

			if flagJiraToken != "" {
				userCfg.Jira.JiraToken = &flagJiraToken
			}

			if flagJiraUsername != "" {
				userCfg.Jira.JiraUsername = &flagJiraUsername
			}

			if flagJQL != "" {
				userCfg.Jira.JQL = &flagJQL
			}

			if flagJiraTimeDeltaMinsStr != "" {
				userCfg.Jira.JiraTimeDeltaMins = jiraTimeDeltaMins
			}

			if flagFallbackComment != "" {
				userCfg.Jira.FallbackComment = &flagFallbackComment
			}

			// validations
			var installationType d.JiraInstallationType

			switch userCfg.Jira.InstallationType {
			case "", jiraInstallationTypeOnPremise: // "" to maintain backwards compatibility
				installationType = d.OnPremiseInstallation
				userCfg.Jira.InstallationType = jiraInstallationTypeOnPremise
			case jiraInstallationTypeCloud:
				installationType = d.CloudInstallation
			default:
				return errInvalidInstallationType
			}

			if userCfg.Jira.JiraURL == nil || *userCfg.Jira.JiraURL == "" {
				return fmt.Errorf("jira-url cannot be empty")
			}

			if userCfg.Jira.JQL == nil || *userCfg.Jira.JQL == "" {
				return fmt.Errorf("jql cannot be empty")
			}

			if userCfg.Jira.JiraToken == nil || *userCfg.Jira.JiraToken == "" {
				return fmt.Errorf("jira-token cannot be empty")
			}

			if installationType == d.CloudInstallation && (userCfg.Jira.JiraUsername == nil || *userCfg.Jira.JiraUsername == "") {
				return fmt.Errorf("jira-username cannot be empty for cloud installation")
			}

			if userCfg.Jira.FallbackComment != nil && strings.TrimSpace(*userCfg.Jira.FallbackComment) == "" {
				return fmt.Errorf("fallback-comment cannot be empty")
			}

			jiraCfg = d.JiraConfig{
				InstallationType: installationType,
				JQL:              *userCfg.Jira.JQL,
				TimeDeltaMins:    userCfg.Jira.JiraTimeDeltaMins,
				FallbackComment:  userCfg.Jira.FallbackComment,
			}

			db, err = pers.GetDB(dbPathFull)
			if err != nil {
				return err
			}

			jiraSvc, err = getJiraSvc(jiraCfg.InstallationType, userCfg)
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			if flagListConfig {
				printConfig(configPathFull, dbPathFull, userCfg)
				return nil
			}

			return ui.RenderUI(db, jiraSvc, jiraCfg)
		},
	}

	mcpCmd := &cobra.Command{
		Use:   "mcp <COMMAND>",
		Short: "Interact with punchout's MCP server",
	}

	mcpServeCmd := &cobra.Command{
		Use:   "serve",
		Short: "Run punchout's MCP server",
		RunE: func(_ *cobra.Command, _ []string) error {
			transport, ok := d.ParseMCPTransport(flagMcpTransportStr)
			if !ok {
				return fmt.Errorf("%w: %q", errInvalidMCPTransport, flagMcpTransportStr)
			}

			if flagListConfig {
				printConfig(configPathFull, dbPathFull, userCfg)
				fmt.Fprintf(os.Stdout, "Transport                               %s\n", flagMcpTransportStr)
				if transport == d.McpTransportHTTP {
					fmt.Fprintf(os.Stdout, "Port                                    %d\n", flagMcpServerPort)
				}
				return nil
			}

			mcpCfg := d.McpConfig{
				Transport: transport,
				HTTPPort:  flagMcpServerPort,
			}
			return mcp.Serve(db, jiraSvc, jiraCfg, mcpCfg)
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

	rootCmd.PersistentFlags().StringVarP(&flagConfigFilePath, "config-file-path", "", defaultConfigFilePath, "location of punchout's config file")
	rootCmd.PersistentFlags().StringVarP(&flagDBPath, "db-path", "", defaultDBPath, "location of punchout's local database")
	rootCmd.PersistentFlags().StringVarP(&flagJiraInstallationType, "jira-installation-type", "", "", "JIRA installation type; allowed values: [cloud, onpremise]")
	rootCmd.PersistentFlags().StringVarP(&flagJiraURL, "jira-url", "", "", "URL of the JIRA server")
	rootCmd.PersistentFlags().StringVarP(&flagJiraToken, "jira-token", "", "", "jira token (PAT for on-premise installation, API token for cloud installation)")
	rootCmd.PersistentFlags().StringVarP(&flagJiraUsername, "jira-username", "", "", "username for authentication (for cloud installation)")
	rootCmd.PersistentFlags().StringVarP(&flagJQL, "jql", "", "", "JQL to use to query issues")
	rootCmd.PersistentFlags().StringVarP(&flagFallbackComment, "fallback-comment", "", "", "fallback comment to use for worklog entries")
	rootCmd.PersistentFlags().StringVarP(&flagJiraTimeDeltaMinsStr, "jira-time-delta-mins", "", "", "time delta (in minutes) between your timezone and the timezone of the JIRA server; can be +/-")
	rootCmd.PersistentFlags().BoolVarP(&flagListConfig, "list-config", "", false, "print the config that punchout will use")

	mcpServeCmd.Flags().StringVarP(&flagMcpTransportStr, "transport", "t", "stdio", "transport to use (possible values: [stdio, http])")
	mcpServeCmd.Flags().UintVarP(&flagMcpServerPort, "http-port", "p", 18899, "port to use (when transport is http)")

	mcpCmd.AddCommand(mcpServeCmd)
	rootCmd.AddCommand(mcpCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	return rootCmd, nil
}

func printConfig(configPath, dbPath string, userCfg userConfig) {
	fmt.Fprintf(os.Stdout, `Config:

Config File Path                        %s
DB File Path                            %s
JIRA Installation Type                  %s
JIRA URL                                %s
JIRA Token                              %s
JQL                                     %s
JIRA Time Delta Mins                    %d
`,
		configPath,
		dbPath,
		userCfg.Jira.InstallationType,
		*userCfg.Jira.JiraURL,
		*userCfg.Jira.JiraToken,
		*userCfg.Jira.JQL,
		userCfg.Jira.JiraTimeDeltaMins)

	if userCfg.Jira.InstallationType == jiraInstallationTypeCloud {
		fmt.Fprintf(os.Stdout, "JIRA Username                           %s\n", *userCfg.Jira.JiraUsername)
	}

	if userCfg.Jira.FallbackComment != nil {
		fmt.Fprintf(os.Stdout, "Fallback Comment                        %s\n", *userCfg.Jira.FallbackComment)
	}
}

func getJiraSvc(installationType d.JiraInstallationType, cfg userConfig) (svc.Jira, error) {
	switch installationType {
	case d.OnPremiseInstallation:
		return svc.NewOnPremJiraSvc(*cfg.Jira.JiraURL, *cfg.Jira.JiraToken)
	default:
		return svc.NewCloudJiraSvc(*cfg.Jira.JiraURL, *cfg.Jira.JiraUsername, *cfg.Jira.JiraToken)
	}
}
