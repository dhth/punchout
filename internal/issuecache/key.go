package issuecache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/dhth/punchout/internal/config"
)

const keyFormatNamespace = "punchout-issue-cache-v1"

type cloudInstallationKeyInputs struct {
	Type     string `json:"type"`
	URL      string `json:"url"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

type onPremiseInstallationKeyInputs struct {
	Type  string `json:"type"`
	URL   string `json:"url"`
	Token string `json:"token"`
}

type installationKeyInput interface {
	cloudInstallationKeyInputs | onPremiseInstallationKeyInputs
}

type cacheKeyInputs[T installationKeyInput] struct {
	Namespace    string `json:"namespace"`
	JQL          string `json:"jql"`
	Installation T      `json:"installation"`
}

func deriveKey(installation config.JiraInstallation, jql string) (string, error) {
	switch installation := installation.(type) {
	case config.CloudInstallation:
		return deriveKeyFromInputs(jql, cloudInstallationKeyInputs{
			Type:     config.JiraInstallationTypeCloud,
			URL:      installation.URL,
			Username: installation.Username,
			Token:    installation.Token,
		})
	case config.OnPremiseInstallation:
		return deriveKeyFromInputs(jql, onPremiseInstallationKeyInputs{
			Type:  config.JiraInstallationTypeOnPremise,
			URL:   installation.URL,
			Token: installation.Token,
		})
	default:
		return "", fmt.Errorf("unsupported JIRA installation type %T", installation)
	}
}

func deriveKeyFromInputs[T installationKeyInput](jql string, installation T) (string, error) {
	encodedKeyInputs, err := json.Marshal(cacheKeyInputs[T]{
		Namespace:    keyFormatNamespace,
		JQL:          jql,
		Installation: installation,
	})
	if err != nil {
		return "", fmt.Errorf("couldn't encode issue cache key inputs: %w", err)
	}

	keyHash := sha256.Sum256(encodedKeyInputs)

	return hex.EncodeToString(keyHash[:]), nil
}
