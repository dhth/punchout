package ui

var (
	HelpText = `
TUI Reference Manual

punchout has 2 sections:
- Issues List View
- Worklog List View

Keyboard Shortcuts:

General
    1          Switch to Issues List View
    2          Switch to Worklog List View

General List Controls
    h/<Up>      Move cursor up
    k/<Down>    Move cursor down
    h<Left>     Go to previous page
    l<Right>    Go to next page
    /           Start filtering

Worklog List View
    d           Delete worklog entry
    s           Sync all visible entries to JIRA
    <ctrl-r>    Refresh list
`
)
