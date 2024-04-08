# punchout

✨ Overview
---

`punchout` takes the suck out of logging time on JIRA.

<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/punchout.gif" alt="Usage" />
</p>

Install
---

**homebrew**:

```sh
brew install dhth/tap/punchout
```

**go**:

```sh
go install github.com/dhth/punchout@latest
```

⚡️ Usage
---

`punchout` can receive its configuration via command line flags, or a config
file.

### Using a config file

Create a toml file that looks like the following. The default location for this
file is `~/.config/punchout/punchout.toml`.

```toml
db_path = "/path/to/punchout/db/file.db"

[jira]
jira_url = "https://jira.company.com"
jira_token = "XXX"
# put whatever JQL you want to query for
jql = "assignee = currentUser() AND updatedDate >= -14d ORDER BY updatedDate DESC"
# I don't know how many people will find use for this.
# I need this, since the JIRA server I use runs 5 hours behind
# the actual time, for whatever reason 🤷
jira_time_delta_mins = 300
```

*Note: `punchout` only supports [on-premise] installations of JIRA for now. I
might add support for cloud installations in the future.*

### Using command line flags

Use `punchout -h` for help.

```bash
punchout \
    db-path='/path/to/punchout/db/file.db' \
    jira-url='https://jira.company.com' \
    jira-token='XXX' \
    jql='assignee = currentUser() AND updatedDate >= -14d ORDER BY updatedDate DESC' \
    jira-time-delta-mins='300' \
```

Both the config file and the command line flags can be used in conjunction, but
the latter will take precedence over the former.

```bash
punchout \
    config-file-path='/path/to/punchout/config/file.toml' \
    jira-token='XXX' \
    jql='assignee = currentUser() AND updatedDate >= -14d ORDER BY updatedDate DESC'
```

Screenshots
---

<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/punchout-1.png" alt="Usage" />
</p>
<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/punchout-2.png" alt="Usage" />
</p>

Reference Manual
---

```
punchout Reference Manual

punchout has 2 sections:
- Issues List View
- Worklog List View

Keyboard Shortcuts:

General
    1           Switch to Issues List View
    2           Switch to Worklog List View
    <tab>       Go to next view
    <shift+tab> Go to previous view

General List Controls
    h/<Up>      Move cursor up
    k/<Down>    Move cursor down
    h<Left>     Go to previous page
    l<Right>    Go to next page
    /           Start filtering

Issue List View
    s           Toggle recording time on the currently selected issue,
                will open up a form to record a comment on the second
                "s" keypress
    <ctrl+s>    Add manual worklog entry

Worklog List View
    <ctrl+s>    Update worklog entry
    <ctrl+d>    Delete worklog entry
    s           Sync all visible entries to JIRA
    <ctrl+r>    Refresh list
```

Acknowledgements
---

`punchout` is built using the awesome TUI framework [bubbletea][1].

[1]: https://github.com/charmbracelet/bubbletea
[2]: https://community.atlassian.com/t5/Atlassian-Migration-Program/Product-features-comparison-Atlassian-Cloud-vs-on-premise/ba-p/1918147
