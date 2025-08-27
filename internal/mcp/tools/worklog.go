package tools

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	d "github.com/dhth/punchout/internal/domain"
	pers "github.com/dhth/punchout/internal/persistence"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	timeFormat = "2006/01/02 15:04"
)

var jiraIssueRegex = regexp.MustCompile(`[A-Z]{2,}-\d+`)

type addWorkLogInput struct {
	IssueKey string  `json:"issue_key" jsonschema:"issue key"`
	BeginTS  string  `json:"begin_time" jsonschema:"begin time for issue (in the format 2006/01/02 15:04 in local time)"`
	EndTS    string  `json:"end_time" jsonschema:"end time for issue (in the format 2006/01/02 15:04 in local time, must be after begin_time)"`
	Comment  *string `json:"comment,omitempty" jsonschema:"optional comment for the worklog"`
}

type addWorkLogOutput struct {
	IssueKey string `json:"issue_key" jsonschema:"jira issue key"`
	BeginTS  string `json:"begin_time" jsonschema:"begin time for issue (in the format 2006/01/02 15:04 in local time)"`
	EndTS    string `json:"end_time" jsonschema:"end time for issue (in the format 2006/01/02 15:04 in local time, must be after begin_time)"`
	Comment  string `json:"comment" jsonschema:"comment for the worklog"`
}

type getUnsyncedWorklogsOutput struct {
	Worklogs []d.WorklogEntry `json:"worklogs" jsonschema:"unsynced worklog entries"`
}

func (h Handler) addWorklog(_ context.Context, _ *mcp.CallToolRequest, params addWorkLogInput) (*mcp.CallToolResult, addWorkLogOutput, error) {
	tErr := toolCallError[addWorkLogOutput]
	tSuc := toolCallSuccess[addWorkLogOutput]

	issueKey := params.IssueKey
	beginTSStr := params.BeginTS
	endTSStr := params.EndTS

	slog.Info("got request for adding worklog", "issue_key", issueKey, "begin_time", beginTSStr, "end_time", endTSStr)

	if !jiraIssueRegex.MatchString(issueKey) {
		return tErr(`issue_key doesn't look valid; JIRA issue keys match the regex '[A-Z]{2,}-\d+'`)
	}

	beginTS, err := time.ParseInLocation(timeFormat, beginTSStr, time.Local)
	if err != nil {
		return tErr("begin_time is not correct, expected format: 2006/01/02 15:04")
	}

	endTS, err := time.ParseInLocation(timeFormat, endTSStr, time.Local)
	if err != nil {
		return tErr("end_time is not correct, expected format: 2006/01/02 15:04")
	}

	if endTS.Sub(beginTS).Seconds() < 60 {
		return tErr("duration is not valid, end time needs to be atleast a minute after begin time")
	}

	var comment string

	if params.Comment != nil {
		comment = *params.Comment
	} else if h.JiraCfg.FallbackComment != nil {
		comment = *h.JiraCfg.FallbackComment
	}

	err = pers.InsertManualWLInDB(h.DB, issueKey, beginTS, endTS, comment)
	if err != nil {
		return tErr(err.Error())
	}

	output := addWorkLogOutput{
		IssueKey: issueKey,
		BeginTS:  beginTSStr,
		EndTS:    endTSStr,
		Comment:  comment,
	}

	return tSuc(output)
}

func (h Handler) getUnsyncedWorklogs(_ context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, getUnsyncedWorklogsOutput, error) {
	tErr := toolCallError[getUnsyncedWorklogsOutput]
	tSuc := toolCallSuccess[getUnsyncedWorklogsOutput]

	slog.Info("got request for fetching unsynced worklogs")

	entries, err := pers.FetchUnsyncedWLsFromDB(h.DB)
	if err != nil {
		return tErr(err.Error())
	}

	output := getUnsyncedWorklogsOutput{Worklogs: entries}

	return tSuc(output)
}

func addWorkLogTool() (mcp.Tool, error) {
	var zero mcp.Tool

	inputSch, err := jsonschema.For[addWorkLogInput](nil)
	if err != nil {
		return zero, fmt.Errorf("couldn't construct input jsonschema: %w", err)
	}

	outputSch, err := jsonschema.For[addWorkLogOutput](nil)
	if err != nil {
		return zero, fmt.Errorf("couldn't construct output jsonschema: %w", err)
	}

	hintFalse := false
	return mcp.Tool{
		Name:         "add_worklog",
		Description:  "add worklog for a JIRA issue; will save the worklog to punchout's local database (this will not sync the worklog to JIRA yet)",
		InputSchema:  inputSch,
		OutputSchema: outputSch,
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: &hintFalse,
			IdempotentHint:  false,
			OpenWorldHint:   &hintFalse,
			ReadOnlyHint:    false,
		},
	}, nil
}

func getUnsyncedWorklogsTool() (mcp.Tool, error) {
	var zero mcp.Tool

	outputSch, err := jsonschema.For[getUnsyncedWorklogsOutput](nil)
	if err != nil {
		return zero, fmt.Errorf("couldn't construct output jsonschema")
	}

	hintFalse := false
	return mcp.Tool{
		Name:         "get_unsynced_worklogs",
		Description:  "get all worklog entries that haven't been yet synced to JIRA",
		OutputSchema: outputSch,
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: &hintFalse,
			IdempotentHint:  true,
			OpenWorldHint:   &hintFalse,
			ReadOnlyHint:    true,
		},
	}, nil
}
