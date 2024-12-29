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
	jiraURL              = flag.String("jira-url", "", "URL of the JIRA server")
	jiraToken            = flag.String("jira-token", "", "personal access token for the JIRA server")
	jiraCloudToken       = flag.String("jira-cloud-token", "", "API token for the JIRA cloud")
	jiraCloudUsername    = flag.String("jira-cloud-username", "", "username for the JIRA cloud")
	jql                  = flag.String("jql", "", "JQL to use to query issues at startup")
	jiraTimeDeltaMinsStr = flag.String("jira-time-delta-mins", "", "Time delta (in minutes) between your timezone and the timezone of the server; can be +/-")
	listConfig           = flag.Bool("list-config", false, "Whether to only print out the config that punchout will use or not")
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

	poCfg, err := readConfig(*configFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading config at %s: %s.\n"+
			"continue with command line args only\n", *configFilePath, err.Error())
	}

	if *jiraURL != "" {
		poCfg.Jira.JiraURL = jiraURL
	}

	if *jiraToken != "" {
		poCfg.Jira.JiraToken = jiraToken
	}

	if *jiraCloudToken != "" {
		poCfg.Jira.JiraCloudToken = jiraCloudToken
	}

	if *jiraCloudUsername != "" {
		poCfg.Jira.JiraCloudUsername = jiraCloudUsername
	}

	if *jql != "" {
		poCfg.Jira.Jql = jql
	}
	if *jiraTimeDeltaMinsStr != "" {
		poCfg.Jira.JiraTimeDeltaMins = jiraTimeDeltaMins
	}

	// validations
	if poCfg.Jira.JiraURL == nil || *poCfg.Jira.JiraURL == "" {
		die("jira-url cannot be empty")
	}

	if poCfg.Jira.Jql == nil || *poCfg.Jira.Jql == "" {
		die("jql cannot be empty")
	}

	if (poCfg.Jira.JiraToken == nil) == (poCfg.Jira.JiraCloudToken == nil) {
		die("only one of on-premise or cloud auth method must be provided")
	}

	if poCfg.Jira.JiraToken != nil && *poCfg.Jira.JiraToken == "" {
		die("jira-token cannot be empty for on premise auth")
	}

	if poCfg.Jira.JiraCloudToken != nil && *poCfg.Jira.JiraCloudToken == "" {
		die("jira-token cannot be empty for cloud auth")
	}

	if poCfg.Jira.JiraCloudToken != nil && (poCfg.Jira.JiraCloudUsername == nil || *poCfg.Jira.JiraCloudUsername == "") {
		die("jira-username cannot be empty for cloud auth")
	}

	configKeyMaxLen := 40
	if *listConfig {
		fmt.Fprint(os.Stdout, "Config:\n\n")
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("Config File Path", configKeyMaxLen), *configFilePath)
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("DB File Path", configKeyMaxLen), dbPathFull)
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JIRA URL", configKeyMaxLen), *poCfg.Jira.JiraURL)
		if poCfg.Jira.JiraToken != nil {
			fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JIRA Token", configKeyMaxLen), *poCfg.Jira.JiraToken)
		}
		if poCfg.Jira.JiraCloudToken != nil {
			fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JIRA API Token", configKeyMaxLen), *poCfg.Jira.JiraCloudToken)
			fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JIRA Username", configKeyMaxLen), *poCfg.Jira.JiraCloudUsername)
		}
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JQL", configKeyMaxLen), *poCfg.Jira.Jql)
		fmt.Fprintf(os.Stdout, "%s%d\n", ui.RightPadTrim("JIRA Time Delta Mins", configKeyMaxLen), poCfg.Jira.JiraTimeDeltaMins)
		os.Exit(0)
	}

	db, err := setupDB(dbPathFull)
	if err != nil {
		die("couldn't set up punchout database. This is a fatal error\n")
	}

	// setup jira client with one of available auth methods
	var client *http.Client
	if poCfg.Jira.JiraToken != nil {
		tp := jiraOnPremise.BearerAuthTransport{
			Token: *poCfg.Jira.JiraToken,
		}
		client = tp.Client()
	} else {
		tp := jiraCloud.BasicAuthTransport{
			Username: *poCfg.Jira.JiraCloudUsername,
			APIToken: *poCfg.Jira.JiraCloudToken,
		}
		client = tp.Client()
	}

	cl, err := jiraOnPremise.NewClient(*poCfg.Jira.JiraURL, client)
	if err != nil {
		panic(err)
	}

	ui.RenderUI(db, cl, *poCfg.Jira.Jql, poCfg.Jira.JiraTimeDeltaMins)
}
