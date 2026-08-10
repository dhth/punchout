package cli

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestMainCmd(t *testing.T) {
	fx, err := newFixture()
	require.NoErrorf(t, err, "error setting up fixture: %s", err)

	defer func() {
		err := fx.cleanup()
		require.NoErrorf(t, err, "error cleaning up fixture: %s", err)
	}()

	// SUCCESSES
	t.Run("help flag works", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--help",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)

		result = pathRegex.ReplaceAllString(result, `default "[PATH]"`)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("cloud setup works", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--jira-username", "example@example.com",
			"--jira-installation-type", "cloud",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("onpremise setup works", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--jira-installation-type", "onpremise",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("installation type falls back to onpremise", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("fallback comment works", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--fallback-comment", "test",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("flags override config", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--jira-url", "https://overridden.company.com",
			"--jira-token", "overridden",
			"--jql", "project = overridden AND sprint in openSprints ()",
			"--jira-time-delta-mins", "60",
			"--fallback-comment", "overridden",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("db path can be set from config", func(t *testing.T) {
		// GIVEN
		t.Setenv("HOME", "/home/user")
		t.Setenv("PUNCHOUT_DB_DIR", "data-dir")
		args := []string{
			"--config-file-path", "config/db-path.toml",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("db path flag overrides config", func(t *testing.T) {
		// GIVEN
		t.Setenv("HOME", "/home/user")
		t.Setenv("PUNCHOUT_DB_DIR", "data-dir")
		args := []string{
			"--config-file-path", "config/db-path.toml",
			"--db-path", "override.db",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("db path falls back to default", func(t *testing.T) {
		// GIVEN
		t.Setenv("HOME", "/home/user")
		args := []string{
			"--config-file-path", "config/good.toml",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("theme can be selected from config", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/theme-enabled.toml",
			"--db-path", "db.db",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("theme flag overrides config", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/theme-enabled.toml",
			"--db-path", "db.db",
			"--theme", "catppuccin-mocha",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("cache startup can be enabled from config", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/cache-enabled.toml",
			"--db-path", "db.db",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("cache startup can be enabled by flag", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--use-cache-on-startup",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("cache startup can be explicitly disabled by flag", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/cache-enabled.toml",
			"--db-path", "db.db",
			"--use-cache-on-startup=false",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	// FAILURES
	t.Run("empty db path flag fails validation", func(t *testing.T) {
		// GIVEN
		t.Setenv("HOME", "/home/user")
		t.Setenv("PUNCHOUT_DB_DIR", "data-dir")
		args := []string{
			"--config-file-path", "config/db-path.toml",
			"--db-path", "",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("providing an invalid theme fails", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--theme", "invalid",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("providing incorrect installation type fails", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--jira-installation-type", "blah",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("providing no token fails", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/empty.toml",
			"--db-path", "db.db",
			"--jira-url", "https://jira.company.com",
			"--jira-installation-type", "onpremise",
			"--jql", "project = BLAH AND sprint in openSprints ()",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("providing no username for cloud installation fails", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--jira-installation-type", "cloud",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("providing absent config file path fails", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/absent.toml",
			"--db-path", "db.db",
			"--jira-url", "https://jira.company.com",
			"--jira-token", "XXX",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("providing malformed config file fails", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/malformed.toml",
			"--db-path", "db.db",
			"--jira-url", "https://jira.company.com",
			"--jira-token", "XXX",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("empty jira url flag fails validation", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/empty.toml",
			"--db-path", "db.db",
			"--jira-url", "",
			"--jira-token", "XXX",
			"--jql", "project = BLAH AND sprint in openSprints ()",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("empty jira url override does not fall back to config", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--jira-url", "",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("empty jira token flag fails validation", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/empty.toml",
			"--db-path", "db.db",
			"--jira-url", "https://jira.company.com",
			"--jira-token", "",
			"--jql", "project = BLAH AND sprint in openSprints ()",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("empty jira username flag fails validation", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--jira-installation-type", "cloud",
			"--jira-username", "",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("providing incorrect value for time delta fails", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--jira-time-delta-mins", "blah",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("providing incorrect fallback comment fails", func(t *testing.T) {
		// GIVEN
		args := []string{
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--fallback-comment", "  ",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})
}
