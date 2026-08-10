package config

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/dhth/punchout/internal/ui/theme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeAndResolve(t *testing.T) {
	defaultMCPConfig := MCPConfig{
		Transport: MCPTransportStdio,
		HTTPPort:  DefaultMCPHTTPPort,
	}

	t.Run("works with valid configurations", func(t *testing.T) {
		cloudFallbackComment := "cloud work"
		onPremiseFallbackComment := "on-premise work"
		defaultInstallationFallbackComment := "default installation work"

		tests := []struct {
			name     string
			input    string
			expected Config
		}{
			{
				name: "when all values are set for a cloud installation",
				input: `
[jira]
installation_type = "cloud"
jira_url = "https://cloud.jira.company.com"
jql = "project = CLOUD"
jira_time_delta_mins = 300
jira_token = "cloud-token"
jira_username = "user@example.com"
fallback_comment = "cloud work"
`,
				expected: Config{
					MCP: defaultMCPConfig,
					Jira: JiraConfig{
						Options: JiraOptions{
							JQL:             "project = CLOUD",
							TimeDeltaMins:   300,
							FallbackComment: &cloudFallbackComment,
						},
						Installation: CloudInstallation{
							URL:      "https://cloud.jira.company.com",
							Username: "user@example.com",
							Token:    "cloud-token",
						},
					},
					TUI: TUIConfig{ThemeName: theme.DefaultName},
				},
			},
			{
				name: "when all values are set for an on-premise installation",
				input: `
[jira]
installation_type = "onpremise"
jira_url = "https://on-premise.jira.company.com"
jql = "project = ONPREM"
jira_time_delta_mins = -120
jira_token = "on-premise-token"
fallback_comment = "on-premise work"
`,
				expected: Config{
					MCP: defaultMCPConfig,
					Jira: JiraConfig{
						Options: JiraOptions{
							JQL:             "project = ONPREM",
							TimeDeltaMins:   -120,
							FallbackComment: &onPremiseFallbackComment,
						},
						Installation: OnPremiseInstallation{
							URL:   "https://on-premise.jira.company.com",
							Token: "on-premise-token",
						},
					},
					TUI: TUIConfig{ThemeName: theme.DefaultName},
				},
			},
			{
				name: "when installation type is omitted",
				input: `
[jira]
jira_url = "https://default.jira.company.com"
jql = "project = DEFAULT"
jira_time_delta_mins = 60
jira_token = "default-token"
fallback_comment = "default installation work"
`,
				expected: Config{
					MCP: defaultMCPConfig,
					Jira: JiraConfig{
						Options: JiraOptions{
							JQL:             "project = DEFAULT",
							TimeDeltaMins:   60,
							FallbackComment: &defaultInstallationFallbackComment,
						},
						Installation: OnPremiseInstallation{
							URL:   "https://default.jira.company.com",
							Token: "default-token",
						},
					},
					TUI: TUIConfig{ThemeName: theme.DefaultName},
				},
			},
			{
				name: "when optional values are omitted",
				input: `
[jira]
installation_type = "onpremise"
jira_url = "https://minimal.jira.company.com"
jql = "project = MINIMAL"
jira_token = "minimal-token"
`,
				expected: Config{
					MCP: defaultMCPConfig,
					Jira: JiraConfig{
						Options: JiraOptions{
							JQL: "project = MINIMAL",
						},
						Installation: OnPremiseInstallation{
							URL:   "https://minimal.jira.company.com",
							Token: "minimal-token",
						},
					},
					TUI: TUIConfig{ThemeName: theme.DefaultName},
				},
			},
			{
				name: "when TUI cache startup is enabled",
				input: `
[jira]
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"

[tui]
use_cache_on_startup = true
`,
				expected: Config{
					MCP: defaultMCPConfig,
					Jira: JiraConfig{
						Options: JiraOptions{
							JQL: "project = PUNCH",
						},
						Installation: OnPremiseInstallation{
							URL:   "https://jira.company.com",
							Token: "token",
						},
					},
					TUI: TUIConfig{
						UseCacheOnStartup: true,
						ThemeName:         theme.DefaultName,
					},
				},
			},
			{
				name: "when a TUI theme is selected",
				input: `
[jira]
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"

[tui]
theme = "catppuccin-mocha"
`,
				expected: Config{
					MCP: defaultMCPConfig,
					Jira: JiraConfig{
						Options: JiraOptions{
							JQL: "project = PUNCH",
						},
						Installation: OnPremiseInstallation{
							URL:   "https://jira.company.com",
							Token: "token",
						},
					},
					TUI: TUIConfig{
						ThemeName: "catppuccin-mocha",
					},
				},
			},
			{
				name: "when MCP settings are provided",
				input: `
[jira]
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"

[mcp]
transport = "http"
http_port = 3000
`,
				expected: Config{
					Jira: JiraConfig{
						Options: JiraOptions{
							JQL: "project = PUNCH",
						},
						Installation: OnPremiseInstallation{
							URL:   "https://jira.company.com",
							Token: "token",
						},
					},
					TUI: TUIConfig{ThemeName: theme.DefaultName},
					MCP: MCPConfig{
						Transport: MCPTransportHTTP,
						HTTPPort:  3000,
					},
				},
			},
			{
				name: "when MCP HTTP port is at its upper boundary",
				input: `
[jira]
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"

[mcp]
transport = "http"
http_port = 65535
`,
				expected: Config{
					Jira: JiraConfig{
						Options: JiraOptions{
							JQL: "project = PUNCH",
						},
						Installation: OnPremiseInstallation{
							URL:   "https://jira.company.com",
							Token: "token",
						},
					},
					TUI: TUIConfig{ThemeName: theme.DefaultName},
					MCP: MCPConfig{
						Transport: MCPTransportHTTP,
						HTTPPort:  65535,
					},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				actual, err := decodeAndResolve(strings.NewReader(tt.input), Overrides{})
				require.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			})
		}
	})

	t.Run("applies overrides", func(t *testing.T) {
		jqlOverride := "project = OVERRIDDEN"
		timeDeltaMinsOverride := 0
		fallbackCommentOverride := "overridden work"
		themeNameOverride := "catppuccin-mocha"
		useCacheOnStartupOverride := true
		dontUseCacheOnStartupOverride := false
		cloudInstallationType := JiraInstallationTypeCloud
		jiraURLOverride := "https://overridden.jira.company.com"
		jiraTokenOverride := "overridden-token"
		jiraUsernameOverride := "overridden-user@example.com"
		installationOverrideFallbackComment := "original work"
		mcpTransportOverride := "stdio"
		mcpHTTPPortOverride := uint16(4000)

		tests := []struct {
			name      string
			input     string
			overrides Overrides
			expected  Config
		}{
			{
				name: "when Jira options are provided",
				input: `
[jira]
installation_type = "onpremise"
jira_url = "https://jira.company.com"
jql = "project = ORIGINAL"
jira_time_delta_mins = 300
jira_token = "token"
fallback_comment = "original work"
`,
				overrides: Overrides{
					Jira: JiraOverrides{
						JQL:             &jqlOverride,
						TimeDeltaMins:   &timeDeltaMinsOverride,
						FallbackComment: &fallbackCommentOverride,
					},
				},
				expected: Config{
					MCP: defaultMCPConfig,
					Jira: JiraConfig{
						Options: JiraOptions{
							JQL:             "project = OVERRIDDEN",
							TimeDeltaMins:   0,
							FallbackComment: &fallbackCommentOverride,
						},
						Installation: OnPremiseInstallation{
							URL:   "https://jira.company.com",
							Token: "token",
						},
					},
					TUI: TUIConfig{ThemeName: theme.DefaultName},
				},
			},
			{
				name: "when installation and connection values are provided",
				input: `
[jira]
installation_type = "onpremise"
jira_url = "https://original.jira.company.com"
jql = "project = ORIGINAL"
jira_time_delta_mins = 300
jira_token = "original-token"
fallback_comment = "original work"
`,
				overrides: Overrides{
					Jira: JiraOverrides{
						InstallationType: &cloudInstallationType,
						URL:              &jiraURLOverride,
						Token:            &jiraTokenOverride,
						Username:         &jiraUsernameOverride,
					},
				},
				expected: Config{
					MCP: defaultMCPConfig,
					Jira: JiraConfig{
						Options: JiraOptions{
							JQL:             "project = ORIGINAL",
							TimeDeltaMins:   300,
							FallbackComment: &installationOverrideFallbackComment,
						},
						Installation: CloudInstallation{
							URL:      "https://overridden.jira.company.com",
							Username: "overridden-user@example.com",
							Token:    "overridden-token",
						},
					},
					TUI: TUIConfig{ThemeName: theme.DefaultName},
				},
			},
			{
				name: "when TUI cache startup is enabled by an override",
				input: `
[jira]
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"

[tui]
use_cache_on_startup = false
`,
				overrides: Overrides{
					TUI: TUIOverrides{
						UseCacheOnStartup: &useCacheOnStartupOverride,
					},
				},
				expected: Config{
					MCP: defaultMCPConfig,
					Jira: JiraConfig{
						Options: JiraOptions{
							JQL: "project = PUNCH",
						},
						Installation: OnPremiseInstallation{
							URL:   "https://jira.company.com",
							Token: "token",
						},
					},
					TUI: TUIConfig{
						UseCacheOnStartup: true,
						ThemeName:         theme.DefaultName,
					},
				},
			},
			{
				name: "when TUI cache startup is explicitly disabled by an override",
				input: `
[jira]
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"

[tui]
use_cache_on_startup = true
`,
				overrides: Overrides{
					TUI: TUIOverrides{
						UseCacheOnStartup: &dontUseCacheOnStartupOverride,
					},
				},
				expected: Config{
					MCP: defaultMCPConfig,
					Jira: JiraConfig{
						Options: JiraOptions{
							JQL: "project = PUNCH",
						},
						Installation: OnPremiseInstallation{
							URL:   "https://jira.company.com",
							Token: "token",
						},
					},
					TUI: TUIConfig{ThemeName: theme.DefaultName},
				},
			},
			{
				name: "when a TUI theme is provided by an override",
				input: `
[jira]
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"

[tui]
theme = "gruvbox-light"
`,
				overrides: Overrides{
					TUI: TUIOverrides{
						ThemeName: &themeNameOverride,
					},
				},
				expected: Config{
					MCP: defaultMCPConfig,
					Jira: JiraConfig{
						Options: JiraOptions{
							JQL: "project = PUNCH",
						},
						Installation: OnPremiseInstallation{
							URL:   "https://jira.company.com",
							Token: "token",
						},
					},
					TUI: TUIConfig{ThemeName: themeNameOverride},
				},
			},
			{
				name: "when MCP settings are provided by overrides",
				input: `
[jira]
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"

[mcp]
transport = "http"
http_port = 3000
`,
				overrides: Overrides{
					MCP: MCPOverrides{
						Transport: &mcpTransportOverride,
						HTTPPort:  &mcpHTTPPortOverride,
					},
				},
				expected: Config{
					Jira: JiraConfig{
						Options: JiraOptions{
							JQL: "project = PUNCH",
						},
						Installation: OnPremiseInstallation{
							URL:   "https://jira.company.com",
							Token: "token",
						},
					},
					TUI: TUIConfig{ThemeName: theme.DefaultName},
					MCP: MCPConfig{
						Transport: MCPTransportStdio,
						HTTPPort:  mcpHTTPPortOverride,
					},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				actual, err := decodeAndResolve(strings.NewReader(tt.input), tt.overrides)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			})
		}
	})

	t.Run("rejects invalid overrides", func(t *testing.T) {
		emptyOverride := ""
		invalidInstallationType := "serverless"
		invalidMCPTransport := "websocket"
		invalidMCPHTTPPort := uint16(0)
		cloudInstallationType := JiraInstallationTypeCloud
		cloudInput := `
[jira]
installation_type = "cloud"
jira_url = "https://jira.company.com"
jql = "project = ORIGINAL"
jira_time_delta_mins = 300
jira_token = "original-token"
jira_username = "original-user@example.com"
fallback_comment = "original work"
`
		onPremiseInput := `
[jira]
installation_type = "onpremise"
jira_url = "https://jira.company.com"
jql = "project = ORIGINAL"
jira_token = "original-token"
`

		tests := []struct {
			name      string
			input     string
			overrides Overrides
		}{
			{
				name:  "when installation type is empty",
				input: cloudInput,
				overrides: Overrides{Jira: JiraOverrides{
					InstallationType: &emptyOverride,
				}},
			},
			{
				name:  "when installation type is invalid",
				input: cloudInput,
				overrides: Overrides{Jira: JiraOverrides{
					InstallationType: &invalidInstallationType,
				}},
			},
			{
				name:  "when Jira URL is empty",
				input: cloudInput,
				overrides: Overrides{Jira: JiraOverrides{
					URL: &emptyOverride,
				}},
			},
			{
				name:  "when JQL is empty",
				input: cloudInput,
				overrides: Overrides{Jira: JiraOverrides{
					JQL: &emptyOverride,
				}},
			},
			{
				name:  "when Jira token is empty",
				input: cloudInput,
				overrides: Overrides{Jira: JiraOverrides{
					Token: &emptyOverride,
				}},
			},
			{
				name:  "when Jira username is empty",
				input: cloudInput,
				overrides: Overrides{Jira: JiraOverrides{
					Username: &emptyOverride,
				}},
			},
			{
				name:  "when fallback comment is empty",
				input: cloudInput,
				overrides: Overrides{Jira: JiraOverrides{
					FallbackComment: &emptyOverride,
				}},
			},
			{
				name:  "when switching to cloud without a username",
				input: onPremiseInput,
				overrides: Overrides{Jira: JiraOverrides{
					InstallationType: &cloudInstallationType,
				}},
			},
			{
				name:  "when MCP transport is invalid",
				input: onPremiseInput,
				overrides: Overrides{MCP: MCPOverrides{
					Transport: &invalidMCPTransport,
				}},
			},
			{
				name:  "when MCP HTTP port is zero",
				input: onPremiseInput,
				overrides: Overrides{MCP: MCPOverrides{
					HTTPPort: &invalidMCPHTTPPort,
				}},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := decodeAndResolve(strings.NewReader(tt.input), tt.overrides)
				require.ErrorIs(t, err, ErrInvalidConfig)
			})
		}
	})

	t.Run("rejects invalid file configurations", func(t *testing.T) {
		tests := []struct {
			name          string
			input         string
			expectedError error
		}{
			{
				name:          "when TOML is malformed",
				input:         "[jira",
				expectedError: ErrParseConfigFile,
			},
			{
				name: "when installation type is invalid",
				input: `
[jira]
installation_type = "serverless"
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"
`,
				expectedError: ErrInvalidConfig,
			},
			{
				name: "when installation type is empty",
				input: `
[jira]
installation_type = ""
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"
`,
				expectedError: ErrInvalidConfig,
			},
			{
				name: "when Jira URL is missing",
				input: `
[jira]
installation_type = "onpremise"
jql = "project = PUNCH"
jira_token = "token"
`,
				expectedError: ErrInvalidConfig,
			},
			{
				name: "when Jira URL contains only whitespace",
				input: `
[jira]
installation_type = "onpremise"
jira_url = "   "
jql = "project = PUNCH"
jira_token = "token"
`,
				expectedError: ErrInvalidConfig,
			},
			{
				name: "when JQL is missing",
				input: `
[jira]
installation_type = "onpremise"
jira_url = "https://jira.company.com"
jira_token = "token"
`,
				expectedError: ErrInvalidConfig,
			},
			{
				name: "when JQL contains only whitespace",
				input: `
[jira]
installation_type = "onpremise"
jira_url = "https://jira.company.com"
jql = "   "
jira_token = "token"
`,
				expectedError: ErrInvalidConfig,
			},
			{
				name: "when Jira token is missing",
				input: `
[jira]
installation_type = "onpremise"
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
`,
				expectedError: ErrInvalidConfig,
			},
			{
				name: "when Jira token contains only whitespace",
				input: `
[jira]
installation_type = "onpremise"
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "   "
`,
				expectedError: ErrInvalidConfig,
			},
			{
				name: "when Jira username is missing for a cloud installation",
				input: `
[jira]
installation_type = "cloud"
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"
`,
				expectedError: ErrInvalidConfig,
			},
			{
				name: "when Jira username contains only whitespace",
				input: `
[jira]
installation_type = "cloud"
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"
jira_username = "   "
`,
				expectedError: ErrInvalidConfig,
			},
			{
				name: "when fallback comment is empty",
				input: `
[jira]
installation_type = "onpremise"
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"
fallback_comment = ""
`,
				expectedError: ErrInvalidConfig,
			},
			{
				name: "when fallback comment contains only whitespace",
				input: `
[jira]
installation_type = "onpremise"
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"
fallback_comment = "   "
`,
				expectedError: ErrInvalidConfig,
			},
			{
				name: "when MCP transport is invalid",
				input: `
[jira]
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"

[mcp]
transport = "websocket"
`,
				expectedError: ErrInvalidConfig,
			},
			{
				name: "when MCP HTTP port is zero",
				input: `
[jira]
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"

[mcp]
http_port = 0
`,
				expectedError: ErrInvalidConfig,
			},
			{
				name: "when MCP HTTP port exceeds its upper boundary",
				input: `
[jira]
jira_url = "https://jira.company.com"
jql = "project = PUNCH"
jira_token = "token"

[mcp]
http_port = 65536
`,
				expectedError: ErrParseConfigFile,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := decodeAndResolve(strings.NewReader(tt.input), Overrides{})
				require.ErrorIs(t, err, tt.expectedError)
			})
		}
	})
}

func TestLoadReturnsErrorWhenFileDoesNotExist(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "does-not-exist.toml")

	_, err := Load(filePath, Overrides{})
	require.ErrorIs(t, err, ErrOpenConfigFile)
}
