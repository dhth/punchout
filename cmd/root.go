package cmd

import (
	"flag"
	"fmt"
	"os"
	"os/user"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/dhth/punchout/ui"
)

func die(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

var (
	jiraURL           = flag.String("jira-url", "https://jira.company.com", "URL of the JIRA server")
	jiraToken         = flag.String("jira-token", "", "personal access token for the JIRA server")
	jql               = flag.String("jql", "assignee = currentUser() AND updatedDate >= -14d ORDER BY updatedDate DESC", "JQL to use to query issues at startup")
	jiraTimeDeltaMins = flag.Int("jira-time-delta-mins", 0, "Time delta (in minutes) between your timezone and the timezone of the server; can be +/-")
)

func Execute() {
	currentUser, err := user.Current()
	var defaultDBPath string
	if err == nil {
		defaultDBPath = fmt.Sprintf("%s/punchout.v%s.db", currentUser.HomeDir, PUNCHOUT_DB_VERSION)
	}
	dbPath := flag.String("db-path", defaultDBPath, "location where punchout should create its DB file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Take the suck out of logging time on JIRA.\n\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n------\n%s", ui.HelpText)
	}
	flag.Parse()

	if *dbPath == "" {
		die("db-path cannot be empty")
	}

	if *jql == "" {
		die("jql cannot be empty")
	}

	if *jiraURL == "" {
		die("jira-url cannot be empty")
	}

	if *jiraToken == "" {
		die("jira-token cannot be empty")
	}

	db, err := setupDB(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't set up punchout database. This is a fatal error")
		os.Exit(1)
	}

	tp := jira.BearerAuthTransport{
		Token: *jiraToken,
	}
	cl, err := jira.NewClient(*jiraURL, tp.Client())
	if err != nil {
		panic(err)
	}
	ui.RenderUI(db, cl, *jql, *jiraTimeDeltaMins)

}
