package cmd

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"strconv"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/dhth/punchout/internal/ui"
)

func die(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

var (
	jiraURL              = flag.String("jira-url", "", "URL of the JIRA server")
	jiraToken            = flag.String("jira-token", "", "personal access token for the JIRA server")
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
			die("could't convert jira-time-delta-mins to a number")
		}
	}

	poCfg, err := readConfig(*configFilePath)
	if err != nil {
		die("error reading config at %s: %s", *configFilePath, err.Error())
	}

	if *jiraURL != "" {
		poCfg.Jira.JiraURL = jiraURL
	}

	if *jiraToken != "" {
		poCfg.Jira.JiraToken = jiraToken
	}

	if *jql != "" {
		poCfg.Jira.Jql = jql
	}
	if *jiraTimeDeltaMinsStr != "" {
		poCfg.Jira.JiraTimeDeltaMins = &jiraTimeDeltaMins
	}

	configKeyMaxLen := 40
	if *listConfig {
		fmt.Fprint(os.Stdout, "Config:\n\n")
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("Config File Path", configKeyMaxLen), *configFilePath)
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("DB File Path", configKeyMaxLen), dbPathFull)
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JIRA URL", configKeyMaxLen), *poCfg.Jira.JiraURL)
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JIRA Token", configKeyMaxLen), *poCfg.Jira.JiraToken)
		fmt.Fprintf(os.Stdout, "%s%s\n", ui.RightPadTrim("JQL", configKeyMaxLen), *poCfg.Jira.Jql)
		fmt.Fprintf(os.Stdout, "%s%d\n", ui.RightPadTrim("JIRA Time Delta Mins", configKeyMaxLen), *poCfg.Jira.JiraTimeDeltaMins)
		os.Exit(0)
	}

	// validations

	if *poCfg.Jira.JiraURL == "" {
		die("jira-url cannot be empty")
	}

	if *poCfg.Jira.JiraToken == "" {
		die("jira-token cannot be empty")
	}

	if *poCfg.Jira.Jql == "" {
		die("jql cannot be empty")
	}

	db, err := setupDB(dbPathFull)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't set up punchout database. This is a fatal error\n")
		os.Exit(1)
	}

	tp := jira.BearerAuthTransport{
		Token: *poCfg.Jira.JiraToken,
	}
	cl, err := jira.NewClient(*poCfg.Jira.JiraURL, tp.Client())
	if err != nil {
		panic(err)
	}

	ui.RenderUI(db, cl, *poCfg.Jira.Jql, *poCfg.Jira.JiraTimeDeltaMins)

}
