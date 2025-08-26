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
