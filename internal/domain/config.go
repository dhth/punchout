package domain

type JiraConfig struct {
	InstallationType JiraInstallationType
	JQL              string
	TimeDeltaMins    int
	FallbackComment  *string
}

type JiraInstallationType uint

const (
	OnPremiseInstallation JiraInstallationType = iota
	CloudInstallation
)

type McpConfig struct {
	Transport MCPTransport
	HTTPPort  uint
}

type MCPTransport uint8

const (
	McpTransportStdio MCPTransport = iota
	McpTransportHTTP
)

func ParseMCPTransport(value string) (MCPTransport, bool) {
	switch value {
	case "stdio":
		return McpTransportStdio, true
	case "http":
		return McpTransportHTTP, true
	default:
		return McpTransportStdio, false
	}
}
