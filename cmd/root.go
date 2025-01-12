package cmd

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	jiraCloud "github.com/andygrunwald/go-jira/v2/cloud"
	jiraOnPremise "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/dhth/punchout/internal/ui"
)

const (
	configFileName = "punchout/punchout.toml"
)

var (
	dbFileName           = fmt.Sprintf("punchout.v%s.db", PUNCHOUT_DB_VERSION)
	jiraInstallationType = flag.String("jira-installation-type", "", "JIRA installation type; allowed values: [cloud, onpremise]")
	jiraURL              = flag.String("jira-url", "", "URL of the JIRA server")
	jiraToken            = flag.String("jira-token", "", "jira token (PAT for on-premise installation, API token for cloud installation)")
	jiraUsername         = flag.String("jira-username", "", "username for authentication")
	jql                  = flag.String("jql", "", "JQL to use to query issues")
	jiraTimeDeltaMinsStr = flag.String("jira-time-delta-mins", "", "Time delta (in minutes) between your timezone and the timezone of the server; can be +/-")
	listConfig           = flag.Bool("list-config", false, "print the config that punchout will use")
)

var (
	errCouldntGetHomeDir       = errors.New("couldn't get your home directory")
	errCouldntGetConfigDir     = errors.New("couldn't get your default config directory")
	errConfigFilePathEmpty     = errors.New("config file path cannot be empty")
	errDBPathEmpty             = errors.New("db file path cannot be empty")
	errTimeDeltaIncorrect      = errors.New("couldn't convert time delta to a number")
	errCouldntParseConfigFile  = errors.New("couldn't parse config file")
	errInvalidInstallationType = fmt.Errorf("invalid value for jira installation type (allowed values: [%s, %s])", jiraInstallationTypeOnPremise, jiraInstallationTypeCloud)
	errCouldntSetUpDB          = errors.New("couldn't set up punchout database")
	errCouldntCreateJiraClient = errors.New("couldn't create JIRA client")
)

func Execute() error {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("%w: %s", errCouldntGetHomeDir, err.Error())
	}

	defaultConfigDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("%w: %s", errCouldntGetConfigDir, err.Error())
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

	configFilePath := flag.String("config-file-path", defaultConfigFilePath, "location of the punchout config file")

	defaultDBPath := filepath.Join(userHomeDir, dbFileName)
	dbPath := flag.String("db-path", defaultDBPath, "location of punchout's local database")

	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "punchout takes the suck out of logging time on JIRA.\n\nFlags:\n")
		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()
	}
	flag.Parse()

	if *configFilePath == "" {
		return errConfigFilePathEmpty
	}

	if *dbPath == "" {
		return errDBPathEmpty
	}

	dbPathFull := expandTilde(*dbPath, userHomeDir)

	var jiraTimeDeltaMins int
	if *jiraTimeDeltaMinsStr != "" {
		jiraTimeDeltaMins, err = strconv.Atoi(*jiraTimeDeltaMinsStr)
		if err != nil {
			return fmt.Errorf("%s: %s", errTimeDeltaIncorrect, err)
		}
	}

	configPathFull := expandTilde(*configFilePath, userHomeDir)

	cfg, err := getConfig(configPathFull)
	if err != nil {
		return fmt.Errorf("%s: %s", errCouldntParseConfigFile, err)
	}

	if *jiraInstallationType != "" {
		cfg.Jira.InstallationType = *jiraInstallationType
	}

	if *jiraURL != "" {
		cfg.Jira.JiraURL = jiraURL
	}

	if *jiraToken != "" {
		cfg.Jira.JiraToken = jiraToken
	}

	if *jiraUsername != "" {
		cfg.Jira.JiraUsername = jiraUsername
	}

	if *jql != "" {
		cfg.Jira.JQL = jql
	}

	if *jiraTimeDeltaMinsStr != "" {
		cfg.Jira.JiraTimeDeltaMins = jiraTimeDeltaMins
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

	configKeyMaxLen := 40
	if *listConfig {
		fmt.Fprint(os.Stdout, "Config:\n\n")
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("Config File Path", configKeyMaxLen), configPathFull)
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("DB File Path", configKeyMaxLen), dbPathFull)
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JIRA Installation Type", configKeyMaxLen), cfg.Jira.InstallationType)
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JIRA URL", configKeyMaxLen), *cfg.Jira.JiraURL)
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JIRA Token", configKeyMaxLen), *cfg.Jira.JiraToken)
		if installationType == ui.CloudInstallation {
			fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JIRA Username", configKeyMaxLen), *cfg.Jira.JiraUsername)
		}
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JQL", configKeyMaxLen), *cfg.Jira.JQL)
		fmt.Fprintf(os.Stdout, "%s%d\n", ui.RightPadTrim("JIRA Time Delta Mins", configKeyMaxLen), cfg.Jira.JiraTimeDeltaMins)
		os.Exit(0)
	}

	db, err := setupDB(dbPathFull)
	if err != nil {
		return fmt.Errorf("%w: %s", errCouldntSetUpDB, err.Error())
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

	return ui.RenderUI(db, cl, installationType, *cfg.Jira.JQL, cfg.Jira.JiraTimeDeltaMins)
}
