package service

import (
	"errors"
	"time"

	d "github.com/dhth/punchout/internal/domain"
)

var (
	errJIRARepliedWithEmptyWorklog = errors.New("JIRA replied with an empty worklog; something is probably wrong")
	errCouldntCreateJiraClient     = errors.New("couldn't create JIRA client")
	errCouldntFetchIssuesFromJira  = errors.New("couldn't fetch issues from JIRA")
	errCannotSyncWLWithoutEndTime  = errors.New("cannot sync worklog without an end time")
)

func getWorklogStart(entry d.WorklogEntry, timeDeltaMins int) time.Time {
	start := entry.BeginTS

	if timeDeltaMins != 0 {
		start = start.Add(time.Minute * time.Duration(timeDeltaMins))
	}

	return start
}

func getTimeSpentSeconds(beginTS, endTS time.Time) int {
	return int(endTS.Sub(beginTS).Seconds())
}
