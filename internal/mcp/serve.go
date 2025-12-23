package mcp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	d "github.com/dhth/punchout/internal/domain"
	"github.com/dhth/punchout/internal/mcp/tools"
	svc "github.com/dhth/punchout/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	errCouldntRunServer    = errors.New("couldn't run MCP server")
	errCouldntListenOnAddr = errors.New("MCP server couldn't listen on address")
)

func Serve(ctx context.Context, db *sql.DB, jiraSvc svc.Jira, jiraCfg d.JiraConfig, mcpCfg d.McpConfig) error {
	opts := &mcp.ServerOptions{
		Instructions: "Use this server for creating worklogs and syncing them to JIRA. You can also use it to fetch issues from JIRA, and view unsynced worklogs.",
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "punchout"}, opts)

	toolsHandler := tools.Handler{
		DB:      db,
		JiraSvc: jiraSvc,
		JiraCfg: jiraCfg,
	}

	err := toolsHandler.AddToolsToServer(server)
	if err != nil {
		return err
	}

	if mcpCfg.Transport == d.McpTransportStdio {
		err := server.Run(ctx, &mcp.StdioTransport{})
		if err != nil {
			return fmt.Errorf("%w: %w", errCouldntRunServer, err)
		}

		return nil
	}

	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, nil)

	mux := http.NewServeMux()
	mux.Handle("/v1", handler)
	mux.HandleFunc("/health", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "HEALTHY")
	}))

	addr := fmt.Sprintf("127.0.0.1:%d", mcpCfg.HTTPPort)
	slog.Info("starting MCP HTTP server", "address", addr)
	err = http.ListenAndServe(addr, mux)
	if err != nil {
		return fmt.Errorf(`%w "%s": %w`, errCouldntListenOnAddr, addr, err)
	}

	return nil
}
