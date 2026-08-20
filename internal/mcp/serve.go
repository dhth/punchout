package mcp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dhth/punchout/internal/config"
	"github.com/dhth/punchout/internal/mcp/tools"
	svc "github.com/dhth/punchout/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	errCouldntRunServer    = errors.New("couldn't run MCP server")
	errCouldntListenOnAddr = errors.New("MCP server couldn't listen on address")
)

func Serve(ctx context.Context, db *sql.DB, jiraSvc svc.Jira, jiraOpts config.JiraOptions, mcpCfg config.MCPConfig) error {
	server, err := newServer(db, jiraSvc, jiraOpts)
	if err != nil {
		return err
	}

	if mcpCfg.Transport == config.MCPTransportStdio {
		err := server.Run(ctx, &mcp.StdioTransport{})
		if err != nil {
			return fmt.Errorf("%w: %w", errCouldntRunServer, err)
		}

		return nil
	}

	addr := fmt.Sprintf("127.0.0.1:%d", mcpCfg.HTTPPort)
	slog.Info("starting MCP HTTP server", "address", addr)
	err = http.ListenAndServe(addr, newHTTPHandler(server))
	if err != nil {
		return fmt.Errorf(`%w "%s": %w`, errCouldntListenOnAddr, addr, err)
	}

	return nil
}

func newServer(db *sql.DB, jiraSvc svc.Jira, jiraOpts config.JiraOptions) (*mcp.Server, error) {
	opts := &mcp.ServerOptions{
		Instructions: "Use this server for creating worklogs and syncing them to JIRA. You can also use it to fetch issues from JIRA, and view unsynced worklogs.",
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "punchout"}, opts)

	toolsHandler := tools.Handler{
		DB:       db,
		JiraSvc:  jiraSvc,
		JiraOpts: jiraOpts,
	}

	err := toolsHandler.AddToolsToServer(server)
	if err != nil {
		return nil, err
	}

	return server, nil
}

func newHTTPHandler(server *mcp.Server) http.Handler {
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{Stateless: true})

	mux := http.NewServeMux()
	mux.Handle("/v1", handler)
	mux.HandleFunc("/health", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "HEALTHY")
	}))

	return mux
}
