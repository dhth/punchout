package tools

import (
	"context"
	"fmt"
	"log/slog"

	d "github.com/dhth/punchout/internal/domain"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getIssuesOutput struct {
	Issues []d.Issue `json:"issues" jsonschema:"jira issues"`
}

func (h Handler) getIssues(_ context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, getIssuesOutput, error) {
	tErr := toolCallError[getIssuesOutput]
	tSuc := toolCallSuccess[getIssuesOutput]

	slog.Info("got request for fetching issues")

	issues, _, err := h.JiraSvc.GetIssues(h.JiraCfg.JQL)
	if err != nil {
		return tErr(err)
	}

	output := getIssuesOutput{Issues: issues}

	return tSuc(output)
}

func getIssuesTool() (mcp.Tool, error) {
	var zero mcp.Tool

	outputSch, err := jsonschema.For[getIssuesOutput](nil)
	if err != nil {
		return zero, fmt.Errorf("%w: %w", ErrCouldntConstructOutputSchema, err)
	}

	hintFalse := false

	return mcp.Tool{
		Name:         "get_jira_issues",
		Description:  "get jira issues",
		OutputSchema: outputSch,
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: &hintFalse,
			IdempotentHint:  true,
			OpenWorldHint:   &hintFalse,
			ReadOnlyHint:    true,
		},
	}, nil
}
