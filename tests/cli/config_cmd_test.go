package cli

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestConfigCmd(t *testing.T) {
	fx, err := newFixture()
	require.NoErrorf(t, err, "error setting up fixture: %s", err)

	defer func() {
		err := fx.cleanup()
		require.NoErrorf(t, err, "error cleaning up fixture: %s", err)
	}()

	t.Run("group help works", func(t *testing.T) {
		result, err := fx.runCmd([]string{"config", "--help"})

		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("show-sample help works", func(t *testing.T) {
		result, err := fx.runCmd([]string{"config", "show-sample", "--help"})

		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("validate help works", func(t *testing.T) {
		result, err := fx.runCmd([]string{"config", "validate", "--help"})

		require.NoError(t, err)
		result = pathRegex.ReplaceAllString(result, `default "[PATH]"`)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("show-sample works", func(t *testing.T) {
		result, err := fx.runCmd([]string{"config", "show-sample"})

		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("show-sample rejects operational flags", func(t *testing.T) {
		result, err := fx.runCmd([]string{"config", "show-sample", "--db-path", "db.db"})

		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("validates a config file", func(t *testing.T) {
		result, err := fx.runCmd([]string{"config", "validate", "--config-file-path", "config/good.toml"})

		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("validation reports malformed config", func(t *testing.T) {
		result, err := fx.runCmd([]string{"config", "validate", "--config-file-path", "config/malformed.toml"})

		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("validation reports invalid theme", func(t *testing.T) {
		result, err := fx.runCmd([]string{"config", "validate", "--config-file-path", "config/theme-invalid.toml"})

		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("validate rejects operational flags", func(t *testing.T) {
		result, err := fx.runCmd([]string{"config", "validate", "--jira-url", "https://jira.example.com"})

		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})
}
