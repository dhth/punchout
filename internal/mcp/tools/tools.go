package tools

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	errFailedtoMarshalToJSON  = errors.New("failed to marshal output to JSON")
	ErrCouldntAddToolToServer = errors.New("couldn't add tool to server")
)

func (h Handler) AddToolsToServer(server *mcp.Server) error {
	var tools []string

	getIssuesTool, err := getIssuesTool()
	if err != nil {
		return fmt.Errorf("%w: get_jira_issues: %s", ErrCouldntAddToolToServer, err.Error())
	}
	tools = append(tools, getIssuesTool.Name)

	addWorkLogTool, err := addWorkLogTool()
	if err != nil {
		return fmt.Errorf("%w: add_worklog: %s", ErrCouldntAddToolToServer, err.Error())
	}
	tools = append(tools, addWorkLogTool.Name)

	addMultipleWorklogsTool, err := addMultipleWorklogsTool()
	if err != nil {
		return fmt.Errorf("%w: add_multiple_worklogs: %s", ErrCouldntAddToolToServer, err.Error())
	}
	tools = append(tools, addMultipleWorklogsTool.Name)

	getUnsyncedWorklogsTool, err := getUnsyncedWorklogsTool()
	if err != nil {
		return fmt.Errorf("%w: get_unsynced_worklogs: %s", ErrCouldntAddToolToServer, err.Error())
	}
	tools = append(tools, getUnsyncedWorklogsTool.Name)

	syncWorklogsTool, err := syncWorklogsTool()
	if err != nil {
		return fmt.Errorf("%w: sync_worklogs_to_jira: %s", ErrCouldntAddToolToServer, err.Error())
	}
	tools = append(tools, syncWorklogsTool.Name)

	slog.Info("set up tools", "list", tools)

	mcp.AddTool(server, &getIssuesTool, h.getIssues)
	mcp.AddTool(server, &addWorkLogTool, h.addWorklog)
	mcp.AddTool(server, &addMultipleWorklogsTool, h.addMultipleWorklogs)
	mcp.AddTool(server, &getUnsyncedWorklogsTool, h.getUnsyncedWorklogs)
	mcp.AddTool(server, &syncWorklogsTool, h.syncWorklogsToJira)

	return nil
}

func toolCallError[T any](err error) (*mcp.CallToolResult, T, error) {
	var zero T
	return nil, zero, err
}

func toolCallSuccess[T any](output T) (*mcp.CallToolResult, T, error) {
	return nil, output, nil
}

func handleErr[T any](err error) (*mcp.CallToolResult, T, error) {
	var zero T
	return nil, zero, err
}
