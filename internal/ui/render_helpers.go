package ui

import "fmt"

func (issue *Issue) setDesc() {
	// TODO: The padding here is a bit of a mess; make it more readable
	var assignee string
	var status string
	var totalSecsSpent string

	issueType := getIssueTypeStyle(issue.issueType).Render(issue.issueType)

	if issue.assignee != "" {
		assignee = assigneeStyle(issue.assignee).Render(RightPadTrim(issue.assignee, int(listWidth/4)))
	} else {
		assignee = assigneeStyle(issue.assignee).Render(RightPadTrim("", int(listWidth/4)))
	}

	status = issueStatusStyle.Render(RightPadTrim(issue.status, int(listWidth/4)))

	if issue.aggSecondsSpent > 0 {
		totalSecsSpent = aggTimeSpentStyle.Render(humanizeDuration(issue.aggSecondsSpent))
	}

	issue.desc = fmt.Sprintf("%s%s%s%s%s", RightPadTrim(issue.issueKey, int(listWidth/4)), status, assignee, issueType, totalSecsSpent)
}
