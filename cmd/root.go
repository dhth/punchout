package cmd

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"strconv"

	jiraCloud "github.com/andygrunwald/go-jira/v2/cloud"
	jiraOnPremise "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/dhth/punchout/internal/ui"
)

func die(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

var (
	jiraInstallationType = flag.String("jira-installation-type", "", "JIRA installation type; allowed values: [cloud, onpremise]")
	jiraURL              = flag.String("jira-url", "", "URL of the JIRA server")
	jiraToken            = flag.String("jira-token", "", "jira token (PAT for on-premise installation, API token for cloud installation)")
	jiraUsername         = flag.String("jira-username", "", "username for authentication")
	jql                  = flag.String("jql", "", "JQL to use to query issues")
	jiraTimeDeltaMinsStr = flag.String("jira-time-delta-mins", "", "Time delta (in minutes) between your timezone and the timezone of the server; can be +/-")
	listConfig           = flag.Bool("list-config", false, "print the config that punchout will use")
)

func Execute() {
	currentUser, err := user.Current()
	if err != nil {
		die("Error getting your home directory, explicitly specify the path for the config file using -config-file-path")
	}

	defaultConfigFP := fmt.Sprintf("%s/.config/punchout/punchout.toml", currentUser.HomeDir)
	configFilePath := flag.String("config-file-path", defaultConfigFP, "location of the punchout config file")

	defaultDBPath := fmt.Sprintf("%s/punchout.v%s.db", currentUser.HomeDir, PUNCHOUT_DB_VERSION)
	dbPath := flag.String("db-path", defaultDBPath, "location where punchout should create its DB file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Take the suck out of logging time on JIRA.\n\nFlags:\n")
		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()
	}
	flag.Parse()

	if *configFilePath == "" {
		die("config-file-path cannot be empty")
	}

	if *dbPath == "" {
		die("db-path cannot be empty")
	}

	dbPathFull := expandTilde(*dbPath)

	var jiraTimeDeltaMins int
	if *jiraTimeDeltaMinsStr != "" {
		jiraTimeDeltaMins, err = strconv.Atoi(*jiraTimeDeltaMinsStr)
		if err != nil {
			die("couldn't convert jira-time-delta-mins to a number")
		}
	}

	cfg, err := readConfig(*configFilePath)
	if err != nil {
		die("error reading config: %s.\n", err.Error())
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
		cfg.Jira.Jql = jql
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
		die("invalid value for jira installation type (allowed values: [%s, %s]): %q", jiraInstallationTypeOnPremise, jiraInstallationTypeCloud, cfg.Jira.InstallationType)
	}

	if cfg.Jira.JiraURL == nil || *cfg.Jira.JiraURL == "" {
		die("jira-url cannot be empty")
	}

	if cfg.Jira.Jql == nil || *cfg.Jira.Jql == "" {
		die("jql cannot be empty")
	}

	if cfg.Jira.JiraToken == nil || *cfg.Jira.JiraToken == "" {
		die("jira-token cannot be empty")
	}

	if installationType == ui.CloudInstallation && (cfg.Jira.JiraUsername == nil || *cfg.Jira.JiraUsername == "") {
		die("jira-username cannot be empty for installation type \"cloud\"")
	}

	configKeyMaxLen := 40
	if *listConfig {
		fmt.Fprint(os.Stdout, "Config:\n\n")
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("Config File Path", configKeyMaxLen), *configFilePath)
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("DB File Path", configKeyMaxLen), dbPathFull)
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JIRA Installation Type", configKeyMaxLen), cfg.Jira.InstallationType)
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JIRA URL", configKeyMaxLen), *cfg.Jira.JiraURL)
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JIRA Token", configKeyMaxLen), *cfg.Jira.JiraToken)
		if installationType == ui.CloudInstallation {
			fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JIRA Username", configKeyMaxLen), *cfg.Jira.JiraUsername)
		}
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JQL", configKeyMaxLen), *cfg.Jira.Jql)
		fmt.Fprintf(os.Stdout, "%s%d\n", ui.RightPadTrim("JIRA Time Delta Mins", configKeyMaxLen), cfg.Jira.JiraTimeDeltaMins)
		os.Exit(0)
	}

	db, err := setupDB(dbPathFull)
	if err != nil {
		die("couldn't set up punchout database. This is a fatal error\n")
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
		die("couldn't create JIRA client: %s", err)
	}

	ui.RenderUI(db, cl, installationType, *cfg.Jira.Jql, cfg.Jira.JiraTimeDeltaMins)
}
