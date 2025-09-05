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

type addMultipleWorklogsInput struct {
	Worklogs []addWorkLogInput `json:"worklogs" jsonschema:"array of worklog entries to create"`
}

type addMultipleWorklogsOutput struct {
	SuccessfulWorklogs []addWorkLogOutput `json:"successful_worklogs" jsonschema:"successfully created worklog entries"`
	FailedWorklogs     []failedWorklog    `json:"failed_worklogs" jsonschema:"failed worklog entries with error details"`
}

type failedWorklog struct {
	Input addWorkLogInput `json:"input" jsonschema:"original input that failed"`
	Error string          `json:"error" jsonschema:"error message describing the failure"`
	Index int             `json:"index" jsonschema:"index of the failed entry in the original input array"`
}

type getUnsyncedWorklogsOutput struct {
	Worklogs []d.WorklogEntry `json:"worklogs" jsonschema:"unsynced worklog entries"`
}

func (h Handler) addWorklog(_ context.Context, _ *mcp.CallToolRequest, params addWorkLogInput) (*mcp.CallToolResult, addWorkLogOutput, error) {
	tErr := toolCallError[addWorkLogOutput]
	tSuc := toolCallSuccess[addWorkLogOutput]

	slog.Info("got request for adding worklog", "issue_key", params.IssueKey, "begin_time", params.BeginTS, "end_time", params.EndTS)

	validatedWorkLog, err := h.validateWorklogInput(params)
	if err != nil {
		return tErr(err.Error())
	}

	err = pers.InsertManualWLInDB(h.DB, validatedWorkLog)
	if err != nil {
		return tErr(err.Error())
	}

	output := addWorkLogOutput{
		IssueKey: validatedWorkLog.IssueKey,
		BeginTS:  params.BeginTS,
		EndTS:    params.EndTS,
		Comment:  validatedWorkLog.Comment,
	}

	return tSuc(output)
}

func (h Handler) addMultipleWorklogs(_ context.Context, _ *mcp.CallToolRequest, params addMultipleWorklogsInput) (*mcp.CallToolResult, addMultipleWorklogsOutput, error) {
	tSuc := toolCallSuccess[addMultipleWorklogsOutput]

	slog.Info("got request for adding multiple worklogs", "count", len(params.Worklogs))

	successfulWorklogs := make([]addWorkLogOutput, 0, len(params.Worklogs))
	failedWorklogs := make([]failedWorklog, 0, len(params.Worklogs))

	for i, worklogInput := range params.Worklogs {
		validatedWorkLog, err := h.validateWorklogInput(worklogInput)
		if err != nil {
			failedWorklogs = append(failedWorklogs, failedWorklog{
				Input: worklogInput,
				Error: err.Error(),
				Index: i,
			})
			continue
		}

		err = pers.InsertManualWLInDB(h.DB, validatedWorkLog)
		if err != nil {
			failedWorklogs = append(failedWorklogs, failedWorklog{
				Input: worklogInput,
				Error: err.Error(),
				Index: i,
			})
			continue
		}

		successfulWorklogs = append(successfulWorklogs, addWorkLogOutput{
			IssueKey: validatedWorkLog.IssueKey,
			BeginTS:  worklogInput.BeginTS,
			EndTS:    worklogInput.EndTS,
			Comment:  validatedWorkLog.Comment,
		})
	}

	output := addMultipleWorklogsOutput{
		SuccessfulWorklogs: successfulWorklogs,
		FailedWorklogs:     failedWorklogs,
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

func (h Handler) validateWorklogInput(input addWorkLogInput) (d.ValidatedWorkLog, error) {
	var zero d.ValidatedWorkLog
	if !jiraIssueRegex.MatchString(input.IssueKey) {
		return zero, fmt.Errorf(`issue_key doesn't look valid; JIRA issue keys match the regex '[A-Z]{2,}-\d+'`)
	}

	beginTS, err := time.ParseInLocation(timeFormat, input.BeginTS, time.Local)
	if err != nil {
		return zero, fmt.Errorf("begin_time is not correct, expected format: 2006/01/02 15:04")
	}

	endTS, err := time.ParseInLocation(timeFormat, input.EndTS, time.Local)
	if err != nil {
		return zero, fmt.Errorf("end_time is not correct, expected format: 2006/01/02 15:04")
	}

	if endTS.Sub(beginTS).Seconds() < 60 {
		return zero, fmt.Errorf("duration is not valid, end time needs to be atleast a minute after begin time")
	}

	var comment string

	if input.Comment != nil {
		comment = *input.Comment
	} else if h.JiraCfg.FallbackComment != nil {
		comment = *h.JiraCfg.FallbackComment
	}

	return d.ValidatedWorkLog{
		IssueKey: input.IssueKey,
		BeginTS:  beginTS,
		EndTS:    endTS,
		Comment:  comment,
	}, nil
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

func addMultipleWorklogsTool() (mcp.Tool, error) {
	var zero mcp.Tool

	inputSch, err := jsonschema.For[addMultipleWorklogsInput](nil)
	if err != nil {
		return zero, fmt.Errorf("couldn't construct input jsonschema: %w", err)
	}

	outputSch, err := jsonschema.For[addMultipleWorklogsOutput](nil)
	if err != nil {
		return zero, fmt.Errorf("couldn't construct output jsonschema: %w", err)
	}

	hintFalse := false
	return mcp.Tool{
		Name:         "add_multiple_worklogs",
		Description:  "add multiple worklogs for JIRA issues in a single operation; will save the worklogs to punchout's local database (this will not sync the worklogs to JIRA yet). Returns detailed success/failure information for each worklog entry.",
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
