# punchout

✨ Overview
---

`punchout` takes the suck out of logging time on JIRA.

<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/punchout.gif" alt="Usage" />
</p>

💾 Installation
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
file is `~/.config/punchout/punchout.toml`. The configuration needed for
authenticating against your JIRA installation (on-premise or cloud) will depend
on the kind of the installation.

```toml
[jira]
jira_url = "https://jira.company.com"

# for on-premise installations
installation_type = "onpremise"
jira_token = "your personal access token"

# for cloud installations
installation_type = "cloud"
jira_token = "your API token"
jira_username = "example@example.com"

# put whatever JQL you want to query for
jql = "assignee = currentUser() AND updatedDate >= -14d ORDER BY updatedDate DESC"

# I don't know how many people will find use for this.
# I need this, since the JIRA on-premise server I use runs 5 hours behind
# the actual time, for whatever reason 🤷
jira_time_delta_mins = 300
```

### Basic usage

Use `punchout -h` for help.

```bash
punchout \
    -db-path='/path/to/punchout/db/file.db' \
    -jira-url='https://jira.company.com' \
    -jira-installation-type 'onpremise' \
    -jira-token='XXX' \
    -jql='assignee = currentUser() AND updatedDate >= -14d ORDER BY updatedDate DESC'

punchout \
    -db-path='/path/to/punchout/db/file.db' \
    -jira-url='https://jira.company.com' \
    -jira-installation-type 'cloud' \
    -jira-token='XXX' \
    -jira-username='example@example.com' \
    -jql='assignee = currentUser() AND updatedDate >= -14d ORDER BY updatedDate DESC'
```

Both the config file and the command line flags can be used in conjunction, but
the latter will take precedence over the former.

```bash
punchout \
    -config-file-path='/path/to/punchout/config/file.toml' \
    -jira-token='XXX' \
    -jql='assignee = currentUser() AND updatedDate >= -14d ORDER BY updatedDate DESC'
```

🖥️ Screenshots
---

<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/punchout-1.png" alt="Usage" />
</p>
<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/punchout-2.png" alt="Usage" />
</p>
<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/punchout-3.png" alt="Usage" />
</p>

📋 Reference Manual
---

```
punchout Reference Manual

punchout has 5 panes:
  - Issues List View                      Shows you issues matching your JQL query
  - Worklog List View                     Shows you your worklog entries; you sync these entries
                                          to JIRA from here
  - Worklog Entry View                    You enter/update a worklog entry from here
  - Synced Worklog Entry View             You view the worklog entries synced to JIRA

  - Help View (this one)

Keyboard Shortcuts

General

  1                                       Switch to Issues List View
  2                                       Switch to Worklog List View
  3                                       Switch to Synced Worklog List View
  <tab>                                   Go to next view/form entry
  <shift+tab>                             Go to previous view/form entry
    ?                                     Show help view

General List Controls

  k/<Up>                                  Move cursor up
  j/<Down>                                Move cursor down
  h<Left>                                 Go to previous page
  l<Right>                                Go to next page
  /                                       Start filtering

Issue List View

  s                                       Toggle recording time on the currently selected issue,
                                              will open up a form to record a comment on the second
                                              "s" keypress
  S                                       Quick switch recording; will save a worklog entry without
                                              a comment for the currently active issue, and start
                                              recording time for another issue
  <ctrl+s>                                Add manual worklog entry
  <ctrl+t>                                Go to currently tracked item
  <ctrl+x>                                Discard currently active recording
  <ctrl+b>                                Open issue in browser

Worklog List View

  <ctrl+s>/u                              Update worklog entry
  <ctrl+d>                                Delete worklog entry
  s                                       Sync all visible entries to JIRA
  <ctrl+r>                                Refresh list

Worklog Entry View

  enter                                   Save worklog entry
  k                                       Move timestamp backwards by one minute
  j                                       Move timestamp forwards by one minute
  K                                       Move timestamp backwards by five minutes
  J                                       Move timestamp forwards by five minutes
  h                                       Move timestamp backwards by a day
  l                                       Move timestamp forwards by a day

Synced Worklog Entry View

  <ctrl+r>                                Refresh list

```

Acknowledgements
---

`punchout` is built using the awesome TUI framework [bubbletea][1].

[1]: https://github.com/charmbracelet/bubbletea
[2]: https://community.atlassian.com/t5/Atlassian-Migration-Program/Product-features-comparison-Atlassian-Cloud-vs-on-premise/ba-p/1918147
