package ui

import "fmt"

func (issue *Issue) setDesc() {
	// TODO: The padding here is a bit of a mess; make it more readable
	var assignee string
	var status string
	var totalSecsSpent string

	issueType := getIssueTypeStyle(issue.issueType).Render(issue.issueType)

	if issue.assignee != "" {
		assignee = assigneeStyle(issue.assignee).Render(RightPadTrim("@"+issue.assignee, int(float64(listWidth)*0.2)))
	} else {
		assignee = assigneeStyle(issue.assignee).Render(RightPadTrim("", int(float64(listWidth)*0.2)))
	}

	status = issueStatusStyle.Render(RightPadTrim(issue.status, int(float64(listWidth)*0.2)))

	if issue.aggSecondsSpent > 0 {
		if issue.aggSecondsSpent < 3600 {
			totalSecsSpent = aggTimeSpentStyle.Render(fmt.Sprintf("%2dm", int(issue.aggSecondsSpent/60)))
		} else {
			totalSecsSpent = aggTimeSpentStyle.Render(fmt.Sprintf("%2dh", int(issue.aggSecondsSpent/3600)))
		}
	}

	issue.desc = fmt.Sprintf("%s%s%s%s%s", RightPadTrim(issue.issueKey, int(float64(listWidth)*0.3)), status, assignee, issueType, totalSecsSpent)
}
