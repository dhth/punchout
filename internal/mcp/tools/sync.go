package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	d "github.com/dhth/punchout/internal/domain"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const markWorklogSyncedTimeout = 5 * time.Second

type syncWorklogsOutput struct {
	Successes []syncSuccess `json:"successes" jsonschema:"worklog entries that were successfully synced"`
	Errors    []syncError   `json:"errors" jsonschema:"worklog entries for which syncing failed"`
}

type syncResult struct {
	EntryID      int
	IssueKey     string
	SyncedToJira bool
	UpdatedInDB  bool
	Err          error
}

type syncSuccess struct {
	EntryID  int    `json:"worklog_id" jsonschema:"ID of the worklog entry"`
	IssueKey string `json:"issue_key" jsonschema:"jira issue key"`
}

type syncError struct {
	EntryID      int    `json:"worklog_id" jsonschema:"ID of the worklog entry"`
	IssueKey     string `json:"issue_key" jsonschema:"jira issue key"`
	SyncedToJira bool   `json:"synced_to_jira" jsonschema:"whether the worklog was synced to jira"`
	UpdatedInDB  bool   `json:"updated_in_db" jsonschema:"whether the worklog was updated in punchout's local db"`
	Err          string `json:"error" jsonschema:"any error that occurred during the sync"`
}

func (h Handler) syncWorklogsToJira(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, syncWorklogsOutput, error) {
	tErr := toolCallError[syncWorklogsOutput]
	tSuc := toolCallSuccess[syncWorklogsOutput]

	slog.Info("got request for syncing worklogs to JIRA")

	entries, err := h.store.UnsyncedWorklogs(ctx)
	if err != nil {
		return tErr(err)
	}

	if len(entries) == 0 {
		return tErr(fmt.Errorf("there are no unsynced worklogs"))
	}

	semaphore := make(chan struct{}, 5)
	resultChan := make(chan syncResult)
	var wg sync.WaitGroup

	for _, entry := range entries {
		wg.Add(1)
		go func(entry d.StoredWorklog) {
			defer wg.Done()
			defer func() {
				<-semaphore
			}()
			semaphore <- struct{}{}
			var fallbackCommentUsed bool
			worklog := entry.Worklog
			if worklog.NeedsComment() && h.jiraOpts.FallbackComment != nil {
				worklog.Comment = *h.jiraOpts.FallbackComment
				fallbackCommentUsed = true
			}

			sr := syncResult{
				EntryID:  entry.ID,
				IssueKey: entry.IssueKey,
			}

			err := h.jiraSvc.SyncWLToJIRA(ctx, worklog, h.jiraOpts.TimeDeltaMins)
			if err != nil {
				sr.Err = err
				resultChan <- sr
				return
			}

			slog.Info("synced worklog to jira", "issue_key", entry.IssueKey, "worklog_id", entry.ID)
			sr.SyncedToJira = true

			var comment *string
			if fallbackCommentUsed {
				comment = &worklog.Comment
			}
			// Jira has accepted a non-idempotent write, so request cancellation must not prevent us from recording it locally.
			persistCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), markWorklogSyncedTimeout)
			err = h.store.MarkWorklogSynced(persistCtx, entry.ID, comment)
			cancel()
			if err != nil {
				sr.Err = err
				resultChan <- sr
				return
			}

			slog.Info("updated worklog in db", "issue_key", entry.IssueKey, "worklog_id", entry.ID)
			sr.UpdatedInDB = true
			resultChan <- sr
		}(entry)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	//nolint:prealloc
	successes := make([]syncSuccess, 0)
	errors := make([]syncError, 0)
	for sr := range resultChan {
		if sr.Err != nil {
			errors = append(errors, syncError{
				EntryID:      sr.EntryID,
				IssueKey:     sr.IssueKey,
				SyncedToJira: sr.SyncedToJira,
				UpdatedInDB:  sr.UpdatedInDB,
				Err:          sr.Err.Error(),
			})
		} else {
			successes = append(successes, syncSuccess{
				EntryID:  sr.EntryID,
				IssueKey: sr.IssueKey,
			})
		}
	}

	output := syncWorklogsOutput{Successes: successes, Errors: errors}

	if len(output.Errors) > 0 {
		jsonBytes, err := json.Marshal(&output)
		if err != nil {
			slog.Error("failed to marshal results to json")
			return handleErr[syncWorklogsOutput](fmt.Errorf("%w: %w", errFailedtoMarshalToJSON, err))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: string(jsonBytes),
				},
			},
			IsError: true,
		}, output, nil
	}

	return tSuc(output)
}

func syncWorklogsTool() (mcp.Tool, error) {
	var zero mcp.Tool
	outputSch, err := jsonschema.For[syncWorklogsOutput](nil)
	if err != nil {
		return zero, fmt.Errorf("%w: %w", ErrCouldntConstructOutputSchema, err)
	}

	hintFalse := false
	return mcp.Tool{
		Name:         "sync_worklogs_to_jira",
		Description:  "syncs all unsynced worklogs to JIRA and updates punchout's local database",
		OutputSchema: outputSch,
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: &hintFalse,
			IdempotentHint:  false,
			OpenWorldHint:   &hintFalse,
			ReadOnlyHint:    false,
		},
	}, nil
}
