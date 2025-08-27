package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	d "github.com/dhth/punchout/internal/domain"
	pers "github.com/dhth/punchout/internal/persistence"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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
	Err          string `json:"error" jsonschema:"any error that occured during the sync"`
}

func (h Handler) syncWorklogsToJira(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, syncWorklogsOutput, error) {
	tErr := toolCallError[syncWorklogsOutput]
	tSuc := toolCallSuccess[syncWorklogsOutput]

	slog.Info("got request for syncing worklogs to JIRA")

	entries, err := pers.FetchUnsyncedWLsFromDB(h.DB)
	if err != nil {
		return tErr(err.Error())
	}

	if len(entries) == 0 {
		return tErr("there are no unsynced worklogs")
	}

	semaphore := make(chan struct{}, 5)
	resultChan := make(chan syncResult)
	var wg sync.WaitGroup

	for _, entry := range entries {
		wg.Add(1)
		go func(entry d.WorklogEntry) {
			defer wg.Done()
			defer func() {
				<-semaphore
			}()
			semaphore <- struct{}{}
			var comment string
			var fallbackCommentUsed bool
			if entry.NeedsComment() && h.JiraCfg.FallbackComment != nil {
				comment = *h.JiraCfg.FallbackComment
				fallbackCommentUsed = true
			} else if entry.Comment != nil {
				comment = *entry.Comment
			}

			sr := syncResult{
				EntryID:  entry.ID,
				IssueKey: entry.IssueKey,
			}

			err := h.JiraSvc.SyncWLToJIRA(ctx, entry, comment, h.JiraCfg.TimeDeltaMins)
			if err != nil {
				sr.Err = err
				resultChan <- sr
			}

			slog.Info("synced worklog to jira", "issue_key", entry.IssueKey, "worklog_id", entry.ID)
			sr.SyncedToJira = true

			if fallbackCommentUsed {
				err = pers.UpdateSyncStatusAndCommentForWLInDB(h.DB, entry.ID, comment)
			} else {
				err = pers.UpdateSyncStatusForWLInDB(h.DB, entry.ID)
			}
			if err != nil {
				sr.Err = err
				resultChan <- sr
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
	var successes []syncSuccess
	var errors []syncError
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
		return zero, fmt.Errorf("couldn't construct output jsonschema")
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
