package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/dhth/punchout/internal/config"
	"github.com/dhth/punchout/internal/issuecache"
	"github.com/dhth/punchout/internal/mcp"
	pers "github.com/dhth/punchout/internal/persistence"
	svc "github.com/dhth/punchout/internal/service"
	"github.com/dhth/punchout/internal/ui"
	"github.com/dhth/punchout/internal/ui/theme"
	"github.com/dhth/punchout/internal/ui/tour"
	"github.com/dhth/punchout/internal/utils"
	"github.com/spf13/cobra"
)

const (
	configFileName = "punchout/punchout.toml"
)

var (
	errCouldntGetHomeDir   = errors.New("couldn't get your home directory")
	errCouldntGetConfigDir = errors.New("couldn't get your config directory")
	errCouldntGetCacheDir  = errors.New("couldn't get your cache directory")
	errConfigFilePathEmpty = errors.New("config file path cannot be empty")
	errTimeDeltaIncorrect  = errors.New("couldn't convert time delta to a number")
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
		flagTheme                string
		flagUseCacheOnStartup    bool

		flagMcpTransportStr string
		flagMcpServerPort   uint16

		defaultConfigFilePath string
		appCfg                config.Config
		configPathFull        string
		jiraSvc               svc.Jira
		resolvedTheme         theme.Theme
	)

	addConfigFlags := func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&flagConfigFilePath, "config-file-path", "", defaultConfigFilePath, "location of punchout's config file")
		cmd.Flags().StringVarP(&flagDBPath, "db-path", "", "", "override the location of punchout's local database")
		cmd.Flags().StringVarP(&flagJiraInstallationType, "jira-installation-type", "", "", "JIRA installation type; allowed values: [cloud, onpremise]")
		cmd.Flags().StringVarP(&flagJiraURL, "jira-url", "", "", "URL of the JIRA server")
		cmd.Flags().StringVarP(&flagJiraToken, "jira-token", "", "", "jira token (PAT for on-premise installation, API token for cloud installation)")
		cmd.Flags().StringVarP(&flagJiraUsername, "jira-username", "", "", "username for authentication (for cloud installation)")
		cmd.Flags().StringVarP(&flagJQL, "jql", "", "", "JQL to use to query issues")
		cmd.Flags().StringVarP(&flagFallbackComment, "fallback-comment", "", "", "fallback comment to use for worklog entries")
		cmd.Flags().StringVarP(&flagJiraTimeDeltaMinsStr, "jira-time-delta-mins", "", "", "time delta (in minutes) between your timezone and the timezone of the JIRA server; can be +/-")
		cmd.Flags().BoolVarP(&flagListConfig, "list-config", "", false, "print the config that punchout will use")
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntGetHomeDir, err.Error())
	}
	dbFileName := fmt.Sprintf("punchout.v%s.db", pers.DBVersion)
	defaultDBPath := filepath.Join(userHomeDir, dbFileName)

	rootCmd := &cobra.Command{
		Use:           "punchout",
		Short:         "punchout takes the suck out of logging time on JIRA.",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if flagConfigFilePath == "" {
				return errConfigFilePathEmpty
			}

			var jiraTimeDeltaMins int
			if cmd.Flags().Changed("jira-time-delta-mins") {
				jiraTimeDeltaMins, err = strconv.Atoi(flagJiraTimeDeltaMinsStr)
				if err != nil {
					return fmt.Errorf("%w: %s", errTimeDeltaIncorrect, err.Error())
				}
			}

			configPathFull = utils.ExpandTilde(flagConfigFilePath, userHomeDir)

			overrides := config.Overrides{}
			if cmd.Flags().Changed("db-path") {
				overrides.DBPath = &flagDBPath
			}

			if cmd.Flags().Changed("jira-installation-type") {
				overrides.Jira.InstallationType = &flagJiraInstallationType
			}

			if cmd.Flags().Changed("jira-url") {
				overrides.Jira.URL = &flagJiraURL
			}

			if cmd.Flags().Changed("jira-token") {
				overrides.Jira.Token = &flagJiraToken
			}

			if cmd.Flags().Changed("jira-username") {
				overrides.Jira.Username = &flagJiraUsername
			}

			if cmd.Flags().Changed("jql") {
				overrides.Jira.JQL = &flagJQL
			}

			if cmd.Flags().Changed("jira-time-delta-mins") {
				overrides.Jira.TimeDeltaMins = &jiraTimeDeltaMins
			}

			if cmd.Flags().Changed("fallback-comment") {
				overrides.Jira.FallbackComment = &flagFallbackComment
			}

			if cmd.Flags().Changed("use-cache-on-startup") {
				overrides.TUI.UseCacheOnStartup = &flagUseCacheOnStartup
			}

			if cmd.Flags().Changed("theme") {
				overrides.TUI.ThemeName = &flagTheme
			}

			if cmd.Flags().Changed("transport") {
				overrides.MCP.Transport = &flagMcpTransportStr
			}

			if cmd.Flags().Changed("http-port") {
				overrides.MCP.HTTPPort = &flagMcpServerPort
			}

			appCfg, err = config.Load(
				configPathFull,
				config.LoadOptions{
					HomeDir: userHomeDir,
					Defaults: config.Defaults{
						DBPath: defaultDBPath,
					},
					Overrides: overrides,
				},
			)
			if err != nil {
				return err
			}

			// this validates the theme for all subcommands, which is intentional
			resolvedTheme, err = theme.Get(appCfg.TUI.ThemeName)
			if err != nil {
				return fmt.Errorf("%w; available themes: [%s]", err, strings.Join(theme.All(), ", "))
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if flagListConfig {
				fmt.Fprint(os.Stdout, formatTUIConfig(configPathFull, appCfg))
				return nil
			}

			userCacheDir, err := os.UserCacheDir()
			if err != nil {
				return fmt.Errorf("%w: %s", errCouldntGetCacheDir, err.Error())
			}
			issueStore, err := issuecache.NewStore(userCacheDir, appCfg.Jira.Installation, appCfg.Jira.Options.JQL)
			if err != nil {
				return fmt.Errorf("couldn't initialize issue cache: %w", err)
			}

			worklogStore, err := pers.NewSQLiteStore(appCfg.DBPath)
			if err != nil {
				return err
			}
			defer func() { _ = worklogStore.Close() }()

			jiraSvc, err = getJiraSvc(appCfg.Jira.Installation)
			if err != nil {
				return err
			}

			return ui.RenderUI(cmd.Context(), worklogStore, jiraSvc, issueStore, ui.Options{
				Jira:              appCfg.Jira.Options,
				UseCacheOnStartup: appCfg.TUI.UseCacheOnStartup,
			}, resolvedTheme)
		},
	}

	mcpCmd := &cobra.Command{
		Use:   "mcp <COMMAND>",
		Short: "Interact with punchout's MCP server",
	}

	mcpServeCmd := &cobra.Command{
		Use:   "serve",
		Short: "Run punchout's MCP server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if flagListConfig {
				fmt.Fprint(os.Stdout, formatMCPConfig(
					configPathFull,
					appCfg,
				))
				return nil
			}

			store, err := pers.NewSQLiteStore(appCfg.DBPath)
			if err != nil {
				return err
			}
			defer func() { _ = store.Close() }()

			jiraSvc, err = getJiraSvc(appCfg.Jira.Installation)
			if err != nil {
				return err
			}

			return mcp.Serve(cmd.Context(), store, jiraSvc, appCfg.Jira.Options, appCfg.MCP)
		},
	}

	tourCmd := &cobra.Command{
		Use:   "tour",
		Short: "Take a quick tour of punchout",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			thm, err := theme.Get(theme.DefaultName)
			if err != nil {
				return err
			}

			return tour.Run(defaultConfigFilePath, thm)
		},
	}

	ros := runtime.GOOS
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

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect and validate punchout configuration",
		Args:  cobra.NoArgs,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return nil
		},
	}
	configShowSampleCmd := &cobra.Command{
		Use:   "show-sample",
		Short: "Print sample configuration",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Fprint(os.Stdout, config.SampleConfig)
			return nil
		},
	}
	configValidateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a configuration file",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if flagConfigFilePath == "" {
				return errConfigFilePathEmpty
			}

			configPath := utils.ExpandTilde(flagConfigFilePath, userHomeDir)
			cfg, err := config.Load(configPath, config.LoadOptions{
				HomeDir: userHomeDir,
				Defaults: config.Defaults{
					DBPath: defaultDBPath,
				},
			})
			if err != nil {
				return err
			}

			_, err = theme.Get(cfg.TUI.ThemeName)
			if err != nil {
				return fmt.Errorf("%w; available themes: [%s]", err, strings.Join(theme.All(), ", "))
			}

			fmt.Fprintf(os.Stdout, "Config file %q is valid.\n", configPath)
			return nil
		},
	}
	configValidateCmd.Flags().StringVarP(&flagConfigFilePath, "config-file-path", "", defaultConfigFilePath, "location of punchout's config file")
	configCmd.AddCommand(configShowSampleCmd, configValidateCmd)

	addConfigFlags(rootCmd)
	addConfigFlags(mcpServeCmd)
	themeFlagUsage := fmt.Sprintf("theme to use; possible values: [%s]", strings.Join(theme.All(), ", "))
	rootCmd.Flags().StringVarP(&flagTheme, "theme", "t", theme.DefaultName, themeFlagUsage)
	rootCmd.Flags().BoolVarP(&flagUseCacheOnStartup, "use-cache-on-startup", "", false, "load JIRA issues from the local cache on startup")

	mcpServeCmd.Flags().StringVarP(&flagMcpTransportStr, "transport", "t", "stdio", "transport to use (possible values: [stdio, http])")
	mcpServeCmd.Flags().Uint16VarP(&flagMcpServerPort, "http-port", "p", config.DefaultMCPHTTPPort, "port to use (when transport is http)")

	mcpCmd.AddCommand(mcpServeCmd)
	rootCmd.AddCommand(configCmd, mcpCmd, tourCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	return rootCmd, nil
}

func getJiraSvc(installation config.JiraInstallation) (svc.Jira, error) {
	switch installation := installation.(type) {
	case config.OnPremiseInstallation:
		return svc.NewOnPremJiraSvc(installation.URL, installation.Token)
	case config.CloudInstallation:
		return svc.NewCloudJiraSvc(installation.URL, installation.Username, installation.Token)
	default:
		return nil, fmt.Errorf("unsupported JIRA installation type %T", installation)
	}
}
