package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/dhth/punchout/internal/ui/theme"
	"github.com/dhth/punchout/internal/utils"
)

const (
	JiraInstallationTypeOnPremise = "onpremise"
	JiraInstallationTypeCloud     = "cloud"
	DefaultMCPHTTPPort            = 18899
	mcpTransportNameStdio         = "stdio"
	mcpTransportNameHTTP          = "http"
)

var (
	ErrOpenConfigFile          = errors.New("couldn't open config file")
	ErrParseConfigFile         = errors.New("couldn't parse config file")
	ErrInvalidConfig           = errors.New("invalid configuration")
	errInvalidInstallationType = fmt.Errorf(
		"invalid value for jira installation type (allowed values: [%s, %s])",
		JiraInstallationTypeOnPremise,
		JiraInstallationTypeCloud,
	)
)

type Config struct {
	DBPath string
	Jira   JiraConfig
	TUI    TUIConfig
	MCP    MCPConfig
}

type Defaults struct {
	DBPath string
}

type LoadOptions struct {
	HomeDir   string
	Defaults  Defaults
	Overrides Overrides
}

type MCPConfig struct {
	Transport MCPTransport
	HTTPPort  uint16
}

type MCPTransport uint8

const (
	MCPTransportStdio MCPTransport = iota
	MCPTransportHTTP
)

func (t MCPTransport) String() string {
	switch t {
	case MCPTransportStdio:
		return mcpTransportNameStdio
	case MCPTransportHTTP:
		return mcpTransportNameHTTP
	default:
		return "unknown"
	}
}

type TUIConfig struct {
	UseCacheOnStartup bool
	ThemeName         string
}

type JiraConfig struct {
	Options      JiraOptions
	Installation JiraInstallation
}

type JiraOptions struct {
	JQL             string
	TimeDeltaMins   int
	FallbackComment *string
}

type JiraInstallation interface {
	isJiraInstallation()
}

type CloudInstallation struct {
	URL      string
	Username string
	Token    string
}

func (CloudInstallation) isJiraInstallation() {}

type OnPremiseInstallation struct {
	URL   string
	Token string
}

func (OnPremiseInstallation) isJiraInstallation() {}

type Overrides struct {
	DBPath *string
	Jira   JiraOverrides
	TUI    TUIOverrides
	MCP    MCPOverrides
}

type MCPOverrides struct {
	Transport *string
	HTTPPort  *uint16
}

type TUIOverrides struct {
	UseCacheOnStartup *bool
	ThemeName         *string
}

type JiraOverrides struct {
	InstallationType *string
	URL              *string
	JQL              *string
	TimeDeltaMins    *int
	Token            *string
	Username         *string
	FallbackComment  *string
}

type fileConfig struct {
	DBPath *string        `toml:"db_path"`
	Jira   fileJiraConfig `toml:"jira"`
	TUI    fileTUIConfig  `toml:"tui"`
	MCP    fileMCPConfig  `toml:"mcp"`
}

type fileMCPConfig struct {
	Transport *string `toml:"transport"`
	HTTPPort  *uint16 `toml:"http_port"`
}

type fileTUIConfig struct {
	UseCacheOnStartup bool    `toml:"use_cache_on_startup"`
	ThemeName         *string `toml:"theme"`
}

type fileJiraConfig struct {
	InstallationType *string `toml:"installation_type"`
	URL              *string `toml:"jira_url"`
	JQL              *string
	TimeDeltaMins    int     `toml:"jira_time_delta_mins"`
	Token            *string `toml:"jira_token"`
	Username         *string `toml:"jira_username"`
	FallbackComment  *string `toml:"fallback_comment"`
}

func Load(filePath string, options LoadOptions) (Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Config{}, fmt.Errorf("%w: %s", ErrOpenConfigFile, err.Error())
	}
	defer file.Close()

	return decodeAndResolve(file, options)
}

func decodeAndResolve(reader io.Reader, options LoadOptions) (Config, error) {
	var fileCfg fileConfig
	_, err := toml.NewDecoder(reader).Decode(&fileCfg)
	if err != nil {
		return Config{}, fmt.Errorf("%w: %s", ErrParseConfigFile, err.Error())
	}

	if fileCfg.DBPath != nil && options.Overrides.DBPath == nil {
		expandedDBPath, err := expandEnv(*fileCfg.DBPath)
		if err != nil {
			return Config{}, fmt.Errorf("%w: couldn't expand db_path: %s", ErrInvalidConfig, err.Error())
		}
		fileCfg.DBPath = &expandedDBPath
	}
	effectiveCfg := withOverrides(fileCfg, options.Overrides)
	dbPath, err := resolveDBPath(
		effectiveCfg.DBPath,
		options.Defaults.DBPath,
		options.HomeDir,
	)
	if err != nil {
		return Config{}, fmt.Errorf("%w: %s", ErrInvalidConfig, err.Error())
	}

	jiraCfg, err := resolveJiraConfig(effectiveCfg.Jira)
	if err != nil {
		return Config{}, fmt.Errorf("%w: %s", ErrInvalidConfig, err.Error())
	}
	mcpCfg, err := resolveMCPConfig(effectiveCfg.MCP)
	if err != nil {
		return Config{}, fmt.Errorf("%w: %s", ErrInvalidConfig, err.Error())
	}

	return Config{
		DBPath: dbPath,
		Jira:   jiraCfg,
		TUI:    resolveTUIConfig(effectiveCfg.TUI),
		MCP:    mcpCfg,
	}, nil
}

func withOverrides(cfg fileConfig, overrides Overrides) fileConfig {
	result := cfg

	if overrides.DBPath != nil {
		result.DBPath = overrides.DBPath
	}
	if overrides.Jira.InstallationType != nil {
		result.Jira.InstallationType = overrides.Jira.InstallationType
	}
	if overrides.Jira.URL != nil {
		result.Jira.URL = overrides.Jira.URL
	}
	if overrides.Jira.JQL != nil {
		result.Jira.JQL = overrides.Jira.JQL
	}
	if overrides.Jira.TimeDeltaMins != nil {
		result.Jira.TimeDeltaMins = *overrides.Jira.TimeDeltaMins
	}
	if overrides.Jira.Token != nil {
		result.Jira.Token = overrides.Jira.Token
	}
	if overrides.Jira.Username != nil {
		result.Jira.Username = overrides.Jira.Username
	}
	if overrides.Jira.FallbackComment != nil {
		result.Jira.FallbackComment = overrides.Jira.FallbackComment
	}
	if overrides.TUI.UseCacheOnStartup != nil {
		result.TUI.UseCacheOnStartup = *overrides.TUI.UseCacheOnStartup
	}
	if overrides.TUI.ThemeName != nil {
		result.TUI.ThemeName = overrides.TUI.ThemeName
	}
	if overrides.MCP.Transport != nil {
		result.MCP.Transport = overrides.MCP.Transport
	}
	if overrides.MCP.HTTPPort != nil {
		result.MCP.HTTPPort = overrides.MCP.HTTPPort
	}

	return result
}

func expandEnv(value string) (string, error) {
	var unset []string
	seen := make(map[string]struct{})

	expanded := os.Expand(value, func(name string) string {
		envValue, ok := os.LookupEnv(name)
		if ok {
			return envValue
		}

		if _, exists := seen[name]; !exists {
			seen[name] = struct{}{}
			unset = append(unset, name)
		}

		return ""
	})

	switch len(unset) {
	case 0:
		return expanded, nil
	case 1:
		return "", fmt.Errorf("environment variable %q is not set", unset[0])
	default:
		return "", fmt.Errorf("environment variables %q are not set", unset)
	}
}

func resolveDBPath(dbPath *string, defaultDBPath, homeDir string) (string, error) {
	result := defaultDBPath
	if dbPath != nil {
		result = *dbPath
	}
	result = utils.ExpandTilde(result, homeDir)
	if strings.TrimSpace(result) == "" {
		return "", fmt.Errorf("db file path cannot be empty")
	}

	return result, nil
}

func resolveMCPConfig(cfg fileMCPConfig) (MCPConfig, error) {
	transport := MCPTransportStdio
	if cfg.Transport != nil {
		switch *cfg.Transport {
		case mcpTransportNameStdio:
		case mcpTransportNameHTTP:
			transport = MCPTransportHTTP
		default:
			return MCPConfig{}, fmt.Errorf("invalid value for MCP transport: %q", *cfg.Transport)
		}
	}

	httpPort := uint16(DefaultMCPHTTPPort)
	if cfg.HTTPPort != nil {
		httpPort = *cfg.HTTPPort
	}
	if httpPort == 0 {
		return MCPConfig{}, fmt.Errorf("mcp http port must be greater than zero")
	}

	return MCPConfig{
		Transport: transport,
		HTTPPort:  httpPort,
	}, nil
}

func resolveTUIConfig(cfg fileTUIConfig) TUIConfig {
	themeName := theme.DefaultName
	if cfg.ThemeName != nil {
		themeName = *cfg.ThemeName
	}

	return TUIConfig{
		UseCacheOnStartup: cfg.UseCacheOnStartup,
		ThemeName:         themeName,
	}
}

func resolveJiraConfig(cfg fileJiraConfig) (JiraConfig, error) {
	var isCloud bool
	if cfg.InstallationType != nil {
		switch *cfg.InstallationType {
		case JiraInstallationTypeOnPremise:
		case JiraInstallationTypeCloud:
			isCloud = true
		default:
			return JiraConfig{}, errInvalidInstallationType
		}
	}

	if cfg.URL == nil || strings.TrimSpace(*cfg.URL) == "" {
		return JiraConfig{}, fmt.Errorf("jira-url cannot be empty")
	}
	if cfg.JQL == nil || strings.TrimSpace(*cfg.JQL) == "" {
		return JiraConfig{}, fmt.Errorf("jql cannot be empty")
	}
	if cfg.Token == nil || strings.TrimSpace(*cfg.Token) == "" {
		return JiraConfig{}, fmt.Errorf("jira-token cannot be empty")
	}
	if isCloud && (cfg.Username == nil || strings.TrimSpace(*cfg.Username) == "") {
		return JiraConfig{}, fmt.Errorf("jira-username cannot be empty for cloud installation")
	}
	if cfg.FallbackComment != nil && strings.TrimSpace(*cfg.FallbackComment) == "" {
		return JiraConfig{}, fmt.Errorf("fallback-comment cannot be empty")
	}

	options := JiraOptions{
		JQL:             *cfg.JQL,
		TimeDeltaMins:   cfg.TimeDeltaMins,
		FallbackComment: cfg.FallbackComment,
	}

	var installation JiraInstallation
	if isCloud {
		installation = CloudInstallation{
			URL:      *cfg.URL,
			Username: *cfg.Username,
			Token:    *cfg.Token,
		}
	} else {
		installation = OnPremiseInstallation{
			URL:   *cfg.URL,
			Token: *cfg.Token,
		}
	}

	return JiraConfig{
		Options:      options,
		Installation: installation,
	}, nil
}
