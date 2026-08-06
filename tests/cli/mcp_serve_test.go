package cli

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestMCPServeCmd(t *testing.T) {
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
			"mcp",
			"serve",
			"--help",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)

		result = pathRegex.ReplaceAllString(result, `default "[PATH]"`)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("listing config works", func(t *testing.T) {
		// GIVEN
		args := []string{
			"mcp",
			"serve",
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

	t.Run("listing config uses MCP settings from config file", func(t *testing.T) {
		// GIVEN
		args := []string{
			"mcp",
			"serve",
			"--config-file-path", "config/mcp.toml",
			"--db-path", "db.db",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("changing transport works", func(t *testing.T) {
		// GIVEN
		args := []string{
			"mcp",
			"serve",
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--transport", "http",
			"--http-port", "3000",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("flags override MCP config file settings", func(t *testing.T) {
		// GIVEN
		args := []string{
			"mcp",
			"serve",
			"--config-file-path", "config/mcp.toml",
			"--db-path", "db.db",
			"--transport", "http",
			"--http-port", "4000",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("transport flag overrides MCP config file setting", func(t *testing.T) {
		// GIVEN
		args := []string{
			"mcp",
			"serve",
			"--config-file-path", "config/mcp.toml",
			"--db-path", "db.db",
			"--transport", "stdio",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	// FAILURES
	t.Run("fails if invalid transport provided", func(t *testing.T) {
		// GIVEN
		args := []string{
			"mcp",
			"serve",
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--transport", "blah",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("fails if config file contains invalid MCP transport", func(t *testing.T) {
		// GIVEN
		args := []string{
			"mcp",
			"serve",
			"--config-file-path", "config/mcp-invalid.toml",
			"--db-path", "db.db",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("fails if invalid http port provided", func(t *testing.T) {
		// GIVEN
		args := []string{
			"mcp",
			"serve",
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--transport", "http",
			"--http-port", "blah",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("fails if HTTP port exceeds its upper boundary", func(t *testing.T) {
		// GIVEN
		args := []string{
			"mcp",
			"serve",
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--transport", "http",
			"--http-port", "65536",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})
}
