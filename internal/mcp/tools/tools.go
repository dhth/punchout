package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	errFailedtoMarshalToJSON  = errors.New("failed to marshal output to JSON")
	errCouldntAddToolToServer = errors.New("couldn't add tool to server")
)

func (h Handler) AddToolsToServer(server *mcp.Server) error {
	var tools []string

	getIssuesTool, err := getIssuesTool()
	if err != nil {
		return fmt.Errorf("%w: get_jira_issues: %s", errCouldntAddToolToServer, err.Error())
	}
	tools = append(tools, getIssuesTool.Name)

	addWorkLogTool, err := addWorkLogTool()
	if err != nil {
		return fmt.Errorf("%w: add_worklog: %s", errCouldntAddToolToServer, err.Error())
	}
	tools = append(tools, addWorkLogTool.Name)

	getUnsyncedWorklogsTool, err := getUnsyncedWorklogsTool()
	if err != nil {
		return fmt.Errorf("%w: get_unsynced_worklogs: %s", errCouldntAddToolToServer, err.Error())
	}
	tools = append(tools, getUnsyncedWorklogsTool.Name)

	syncWorklogsTool, err := syncWorklogsTool()
	if err != nil {
		return fmt.Errorf("%w: sync_worklogs_to_jira: %s", errCouldntAddToolToServer, err.Error())
	}
	tools = append(tools, syncWorklogsTool.Name)

	slog.Info("set up tools", "list", tools)

	mcp.AddTool(server, &getIssuesTool, h.getIssues)
	mcp.AddTool(server, &addWorkLogTool, h.addWorklog)
	mcp.AddTool(server, &getUnsyncedWorklogsTool, h.getUnsyncedWorklogs)
	mcp.AddTool(server, &syncWorklogsTool, h.syncWorklogsToJira)

	return nil
}

func toolCallError[T any](content string) (*mcp.CallToolResult, T, error) {
	var zero T
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: content,
			},
		},
		IsError: true,
	}, zero, nil
}

func toolCallSuccess[T any](output T) (*mcp.CallToolResult, T, error) {
	var zero T
	jsonBytes, err := json.Marshal(&output)
	if err != nil {
		slog.Error("failed to marshal results to json")
		return nil, zero, fmt.Errorf("%w: %w", errFailedtoMarshalToJSON, err)
	}

	return &mcp.CallToolResult{
		// It appears the Claude Code doesn't see the tool output if only StructuredContent is sent
		// https://github.com/anthropics/claude-code/issues/4427
		// Sending it as both Content as well as StructuredContent for now
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonBytes),
			},
		},
		IsError:           false,
		StructuredContent: output,
	}, output, nil
}

func handleErr[T any](err error) (*mcp.CallToolResult, T, error) {
	var zero T
	return nil, zero, err
}
