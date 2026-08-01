package issuecache

import (
	"testing"

	"github.com/dhth/punchout/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveKeyIsDeterministic(t *testing.T) {
	cases := []struct {
		name         string
		installation config.JiraInstallation
	}{
		{
			name: "cloud installation",
			installation: config.CloudInstallation{
				URL:      "https://cloud.jira.example.com",
				Username: "cloud-user@example.com",
				Token:    "cloud-token",
			},
		},
		{
			name: "on-premise installation",
			installation: config.OnPremiseInstallation{
				URL:   "https://jira.example.com",
				Token: "on-premise-token",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reference, err := deriveKey(tc.installation, "project = TEST")
			require.NoError(t, err)

			for i := range 10 {
				other, err := deriveKey(tc.installation, "project = TEST")
				require.NoError(t, err, "deriveKey failed in iteration #%d", i)
				assert.Equal(t, reference, other, "computed value in iteration #%d wasn't equal to the reference", i)
			}
		})
	}
}

func TestDeriveKeyChangesWithCloudInputs(t *testing.T) {
	baseInstallation := config.CloudInstallation{
		URL:      "https://cloud.jira.example.com",
		Username: "cloud-user@example.com",
		Token:    "cloud-token",
	}
	baseJQL := "project = TEST"

	baseKey, err := deriveKey(baseInstallation, baseJQL)
	require.NoError(t, err)

	cases := []struct {
		name         string
		installation config.CloudInstallation
		jql          string
	}{
		{
			name: "URL changes",
			installation: config.CloudInstallation{
				URL:      "https://other-cloud.jira.example.com",
				Username: baseInstallation.Username,
				Token:    baseInstallation.Token,
			},
			jql: baseJQL,
		},
		{
			name: "username changes",
			installation: config.CloudInstallation{
				URL:      baseInstallation.URL,
				Username: "other-cloud-user@example.com",
				Token:    baseInstallation.Token,
			},
			jql: baseJQL,
		},
		{
			name: "token changes",
			installation: config.CloudInstallation{
				URL:      baseInstallation.URL,
				Username: baseInstallation.Username,
				Token:    "other-cloud-token",
			},
			jql: baseJQL,
		},
		{
			name:         "JQL changes",
			installation: baseInstallation,
			jql:          "project = OTHER",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key, err := deriveKey(tc.installation, tc.jql)
			require.NoError(t, err)
			assert.NotEqual(t, baseKey, key)
		})
	}
}

func TestDeriveKeyChangesWithOnPremiseInputs(t *testing.T) {
	baseInstallation := config.OnPremiseInstallation{
		URL:   "https://jira.example.com",
		Token: "on-premise-token",
	}
	baseJQL := "project = TEST"

	baseKey, err := deriveKey(baseInstallation, baseJQL)
	require.NoError(t, err)

	cases := []struct {
		name         string
		installation config.OnPremiseInstallation
		jql          string
	}{
		{
			name: "URL changes",
			installation: config.OnPremiseInstallation{
				URL:   "https://other-jira.example.com",
				Token: baseInstallation.Token,
			},
			jql: baseJQL,
		},
		{
			name: "token changes",
			installation: config.OnPremiseInstallation{
				URL:   baseInstallation.URL,
				Token: "other-on-premise-token",
			},
			jql: baseJQL,
		},
		{
			name:         "JQL changes",
			installation: baseInstallation,
			jql:          "project = OTHER",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key, err := deriveKey(tc.installation, tc.jql)
			require.NoError(t, err)
			assert.NotEqual(t, baseKey, key)
		})
	}
}

func TestDeriveKeySeparatesInstallationTypes(t *testing.T) {
	jiraURL := "https://jira.example.com"
	token := "shared-token"
	jql := "project = TEST"

	cloudKey, err := deriveKey(config.CloudInstallation{
		URL:      jiraURL,
		Username: "user@example.com",
		Token:    token,
	}, jql)
	require.NoError(t, err)

	onPremiseKey, err := deriveKey(config.OnPremiseInstallation{
		URL:   jiraURL,
		Token: token,
	}, jql)
	require.NoError(t, err)

	assert.NotEqual(t, cloudKey, onPremiseKey)
}

func TestDeriveKeyReturnsLowercaseSHA256(t *testing.T) {
	cases := []struct {
		name         string
		installation config.JiraInstallation
	}{
		{
			name: "cloud installation",
			installation: config.CloudInstallation{
				URL:      "https://cloud.jira.example.com",
				Username: "cloud-user@example.com",
				Token:    "cloud-token",
			},
		},
		{
			name: "on-premise installation",
			installation: config.OnPremiseInstallation{
				URL:   "https://jira.example.com",
				Token: "on-premise-token",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key, err := deriveKey(tc.installation, "project = TEST")
			require.NoError(t, err)
			assert.Regexp(t, `^[0-9a-f]{64}$`, key)
		})
	}
}

func TestDeriveKeyRejectsUnsupportedInstallation(t *testing.T) {
	_, err := deriveKey(nil, "project = TEST")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported JIRA installation type")
}
