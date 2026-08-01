package cmd

import (
	"fmt"
	"strings"

	"github.com/dhth/punchout/internal/config"
	"github.com/dhth/punchout/internal/mcp"
)

func formatTUIConfig(configPath, dbPath string, cfg config.Config) string {
	var result strings.Builder
	writePathsSection(&result, configPath, dbPath)
	writeTUISection(&result, cfg.TUI)
	writeJiraSection(&result, cfg.Jira)
	return result.String()
}

func formatMCPConfig(configPath, dbPath string, cfg config.Config, mcpCfg mcp.Config) string {
	var result strings.Builder
	writePathsSection(&result, configPath, dbPath)
	writeJiraSection(&result, cfg.Jira)
	writeMCPSection(&result, mcpCfg)
	return result.String()
}

func writePathsSection(result *strings.Builder, configPath, dbPath string) {
	fmt.Fprint(result, "[Paths]\n")
	writeField(result, "Config File Path", configPath)
	writeField(result, "DB File Path", dbPath)
}

func writeTUISection(result *strings.Builder, cfg config.TUIConfig) {
	fmt.Fprint(result, "\n[TUI]\n")
	writeField(result, "Use Cache On Startup", cfg.UseCacheOnStartup)
}

func writeJiraSection(result *strings.Builder, cfg config.JiraConfig) {
	var installationType string
	var jiraURL string
	var jiraUsername string

	switch installation := cfg.Installation.(type) {
	case config.OnPremiseInstallation:
		installationType = config.JiraInstallationTypeOnPremise
		jiraURL = installation.URL
	case config.CloudInstallation:
		installationType = config.JiraInstallationTypeCloud
		jiraURL = installation.URL
		jiraUsername = installation.Username
	}

	fmt.Fprint(result, "\n[JIRA]\n")
	writeField(result, "JIRA Installation Type", installationType)
	writeField(result, "JIRA URL", jiraURL)
	writeField(result, "JIRA Token", "[REDACTED]")
	writeField(result, "JQL", cfg.Options.JQL)
	writeField(result, "JIRA Time Delta Mins", cfg.Options.TimeDeltaMins)

	if jiraUsername != "" {
		writeField(result, "JIRA Username", jiraUsername)
	}

	if cfg.Options.FallbackComment != nil {
		writeField(result, "Fallback Comment", *cfg.Options.FallbackComment)
	}
}

func writeMCPSection(result *strings.Builder, cfg mcp.Config) {
	fmt.Fprint(result, "\n[MCP]\n")
	writeField(result, "Transport", cfg.Transport)
	if cfg.Transport == mcp.TransportHTTP {
		writeField(result, "Port", cfg.HTTPPort)
	}
}

func writeField(result *strings.Builder, label string, value any) {
	fmt.Fprintf(result, "%-40s%v\n", label, value)
}
