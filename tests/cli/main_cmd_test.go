package cli

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestMainCmd(t *testing.T) {
	fx, err := newFixture()
	require.NoErrorf(t, err, "error setting up fixture: %s", err)

	defer func() {
		err := fx.cleanup()
		if err != nil {
			fmt.Fprintf(os.Stderr, `error cleaning up: %s`, err.Error())
		}
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

		re := regexp.MustCompile(`"[^"]*"`)
		result = re.ReplaceAllString(result, `"[PATH]"`)
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

	// FAILURES
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

	t.Run("providing empty jira url fails", func(t *testing.T) {
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

	t.Run("providing empty jira token fails", func(t *testing.T) {
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

	t.Run("providing empty jira username fails", func(t *testing.T) {
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
