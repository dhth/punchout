<p align="center">
  <h1 align="center">punchout</h1>
  <p align="center">
    <a href="https://github.com/dhth/punchout/actions/workflows/main.yml"><img alt="Build Status" src="https://img.shields.io/github/actions/workflow/status/dhth/punchout/main.yml?style=flat-square"></a>
    <a href="https://github.com/dhth/punchout/actions/workflows/vulncheck.yml"><img alt="Vulnerability Check" src="https://img.shields.io/github/actions/workflow/status/dhth/punchout/vulncheck.yml?style=flat-square&label=vulncheck"></a>
    <a href="https://github.com/dhth/punchout/releases/latest"><img alt="Latest release" src="https://img.shields.io/github/release/dhth/punchout.svg?style=flat-square"></a>
    <a href="https://github.com/dhth/punchout/releases/latest"><img alt="Commits since latest release" src="https://img.shields.io/github/commits-since/dhth/punchout/latest?style=flat-square"></a>
  </p>
</p>

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

```text
punchout takes the suck out of logging time on JIRA.

Usage:
  punchout [flags]
  punchout [command]

Available Commands:
  help        Help about any command
  mcp         Interact with punchout's MCP server

Flags:
      --config-file-path string         location of punchout's config file (default "/Users/user/.config/punchout/punchout.toml")
      --db-path string                  location of punchout's local database (default "/Users/user/punchout.v1.db")
      --fallback-comment string         fallback comment to use for worklog entries
  -h, --help                            help for punchout
      --jira-installation-type string   JIRA installation type; allowed values: [cloud, onpremise]
      --jira-time-delta-mins string     time delta (in minutes) between your timezone and the timezone of the JIRA server; can be +/-
      --jira-token string               jira token (PAT for on-premise installation, API token for cloud installation)
      --jira-url string                 URL of the JIRA server
      --jira-username string            username for authentication (for cloud installation)
      --jql string                      JQL to use to query issues
      --list-config                     print the config that punchout will use
```

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

# this comment will be used for worklogs when you don't provide one; optional"
fallback_comment = "comment"
```

Both the config file and the command line flags can be used in conjunction, but
the latter will take precedence over the former.

Workflow
---

`punchout` lets you add worklogs on JIRA in a two step approach.

1. Your record one or more worklogs locally
2. You push all unsynced worklogs to your JIRA server

This can be done either via `punchout`'s TUI or its MCP server.

> **Historical context:**
>
> punchout's TUI came first. It was faster to track time using it when compared
> to JIRA's website. When AI agents became a thing, I saw an opportunity to
> offload the tedious work of creating worklogs to them by the means of the
> Model Context Protocol.

MCP Server
---

`punchout` comes with an MCP server which can allow you to automate the process
of recording worklogs and syncing them to your JIRA server. The server allows
access to 5 tools:

| Tool                    | What it does                                         |
|-------------------------|------------------------------------------------------|
| `get_jira_issues`       | Return JIRA issues based on JQL configured           |
| `add_worklog`           | Record a worklog for an issue in punchout's database |
| `add_multiple_worklogs` | Record multiple worklogs in punchout's database      |
| `get_unsynced_worklogs` | Get unsynced worklogs from punchout's database       |
| `sync_worklogs_to_jira` | Sync all unsynced worklogs to JIRA                   |

You can leverage this MCP in any way you want. I use it as follows:

[![mcp-server-usage](https://tools.dhruvs.space/images/punchout/v1-3-0/mcp-server-usage-yt.jpg)](https://youtu.be/DNA6L3Vrwrk?si=G-r9MVlI72BpXU5Y)

TUI
---

Before MCP was a thing, the primary way to interact with `punchout` was through
its TUI. It's still an option for those who prefer it.

<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/punchout-1.png" alt="Usage" />
</p>
<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/punchout-2.png" alt="Usage" />
</p>
<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/punchout-3.png" alt="Usage" />
</p>

### 📋 TUI Reference Manual

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
  q/<ctrl+c>                              Go back/reset filtering/quit
  <esc>                                   Cancel form/quit
  ?                                       Show help view

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
  <ctrl+s>                                Update active worklog entry (when tracking active), or add
                                              manual worklog entry (when not tracking)
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

`punchout`'s TUI is built using [bubbletea][1].

[1]: https://github.com/charmbracelet/bubbletea
[2]: https://community.atlassian.com/t5/Atlassian-Migration-Program/Product-features-comparison-Atlassian-Cloud-vs-on-premise/ba-p/1918147
