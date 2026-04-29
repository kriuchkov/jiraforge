package text

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kriuchkov/jiraforge/internal/core/models"
)

func FormatIssue(issue *models.Issue) string {
	if issue == nil {
		return ""
	}

	var builder strings.Builder
	writeLine(&builder, fmt.Sprintf("Key: %s", issue.Key))
	writeIf(&builder, issue.ID, "ID: %s")
	writeIf(&builder, issue.URL, "URL: %s")
	writeIf(&builder, issue.ProjectName, "Project: %s")
	writeIf(&builder, issue.ProjectKey, "Project Key: %s")
	writeIf(&builder, issue.Summary, "Summary: %s")
	if issue.Type != nil {
		writeIf(&builder, issue.Type.Name, "Type: %s")
		writeIf(&builder, issue.Type.Description, "Type Description: %s")
	}
	if issue.Status != nil {
		writeIf(&builder, issue.Status.Name, "Status: %s")
		writeIf(&builder, issue.Status.Description, "Status Description: %s")
	}
	writeIf(&builder, issue.Priority, "Priority: %s")
	writeIf(&builder, issue.Resolution, "Resolution: %s")
	writeIf(&builder, issue.ResolutionDescription, "Resolution Description: %s")
	writeIf(&builder, issue.ResolutionDate, "Resolution Date: %s")
	writeIf(&builder, personLabel(issue.Reporter), "Reporter: %s")
	writeIf(&builder, personLabel(issue.Assignee), "Assignee: %s")
	writeIf(&builder, personLabel(issue.Creator), "Creator: %s")
	writeIf(&builder, issue.Created, "Created: %s")
	writeIf(&builder, issue.Updated, "Updated: %s")
	writeIf(&builder, issue.LastViewed, "Last Viewed: %s")
	writeIf(&builder, issue.StatusCategoryChange, "Status Category Change: %s")
	if issue.Parent != nil {
		writeLine(&builder, fmt.Sprintf("Parent: %s", issueRefLabel(*issue.Parent)))
	}
	writeIf(&builder, strings.Join(issue.Labels, ", "), "Labels: %s")
	writeIf(&builder, strings.Join(issue.Components, ", "), "Components: %s")
	writeIf(&builder, strings.Join(issue.FixVersions, ", "), "Fix Versions: %s")
	writeIf(&builder, strings.Join(issue.AffectedVersions, ", "), "Affected Versions: %s")
	writeIf(&builder, issue.SecurityLevel, "Security Level: %s")
	if issue.Description != "" {
		writeLine(&builder, "")
		writeLine(&builder, "Description:")
		writeLine(&builder, issue.Description)
	}
	if len(issue.Subtasks) > 0 {
		writeSection(&builder, "Subtasks")
		for _, subtask := range issue.Subtasks {
			writeLine(&builder, "- "+issueRefLabel(subtask))
		}
	}
	if len(issue.RelatedIssues) > 0 {
		writeSection(&builder, "Related Issues")
		for _, related := range issue.RelatedIssues {
			writeLine(&builder, fmt.Sprintf("- %s %s", related.Type, issueRefLabel(related.Issue)))
		}
	}
	if len(issue.Attachments) > 0 {
		writeSection(&builder, "Attachments")
		for _, attachment := range issue.Attachments {
			writeLine(&builder, fmt.Sprintf("- %s (ID: %s, Type: %s, Size: %d bytes)", attachment.Filename, attachment.ID, attachment.MimeType, attachment.Size))
		}
	}
	if issue.CommentCount > 0 {
		writeLine(&builder, fmt.Sprintf("Comments: %d total", issue.CommentCount))
	}
	if issue.WorklogCount > 0 {
		writeLine(&builder, fmt.Sprintf("Worklogs: %d entries", issue.WorklogCount))
	}
	if issue.Watchers > 0 {
		writeLine(&builder, fmt.Sprintf("Watchers: %d", issue.Watchers))
	}
	if issue.Votes > 0 {
		writeLine(&builder, fmt.Sprintf("Votes: %d", issue.Votes))
	}
	if issue.StoryPointEstimate != "" {
		writeLine(&builder, fmt.Sprintf("Story Point Estimate: %s", issue.StoryPointEstimate))
	}
	if len(issue.Transitions) > 0 {
		writeSection(&builder, "Available Transitions")
		for _, transition := range issue.Transitions {
			writeLine(&builder, fmt.Sprintf("- %s (ID: %s)", transition.Name, transition.ID))
		}
	}

	return strings.TrimSpace(builder.String())
}

func FormatSearchIssues(result *models.SearchIssuesResult) string {
	if result == nil || len(result.Issues) == 0 {
		return "No issues found matching the search criteria."
	}

	var parts []string
	for _, issue := range result.Issues {
		parts = append(parts, FormatIssue(&issue))
	}
	return strings.Join(parts, "\n===\n")
}

func FormatIssueTypes(issueTypes []models.IssueType) string {
	if len(issueTypes) == 0 {
		return "No issue types found."
	}

	var builder strings.Builder
	writeLine(&builder, "Available Issue Types:")
	for _, issueType := range issueTypes {
		writeLine(&builder, "")
		writeLine(&builder, fmt.Sprintf("ID: %s", issueType.ID))
		name := issueType.Name
		if issueType.Subtask {
			name += " (Subtask Type)"
		}
		writeLine(&builder, fmt.Sprintf("Name: %s", name))
		writeIf(&builder, issueType.Description, "Description: %s")
		writeIf(&builder, issueType.IconURL, "Icon URL: %s")
		writeIf(&builder, issueType.Scope, "Scope: %s")
	}
	return strings.TrimSpace(builder.String())
}

func FormatMutation(result *models.IssueMutationResult) string {
	if result == nil {
		return ""
	}
	var builder strings.Builder
	writeIf(&builder, result.Message, "%s")
	writeIf(&builder, result.Key, "Key: %s")
	writeIf(&builder, result.ID, "ID: %s")
	writeIf(&builder, result.URL, "URL: %s")
	writeIf(&builder, result.ParentKey, "Parent: %s")
	return strings.TrimSpace(builder.String())
}

func FormatSprint(sprint *models.Sprint) string {
	if sprint == nil {
		return ""
	}
	var builder strings.Builder
	writeLine(&builder, fmt.Sprintf("ID: %d", sprint.ID))
	writeIf(&builder, sprint.Name, "Name: %s")
	writeIf(&builder, sprint.State, "State: %s")
	writeIf(&builder, sprint.StartDate, "Start Date: %s")
	writeIf(&builder, sprint.EndDate, "End Date: %s")
	writeIf(&builder, sprint.CompleteDate, "Complete Date: %s")
	if sprint.BoardID > 0 {
		writeLine(&builder, fmt.Sprintf("Board ID: %d", sprint.BoardID))
	}
	writeIf(&builder, sprint.Goal, "Goal: %s")
	return strings.TrimSpace(builder.String())
}

func FormatSprintReport(report *models.SprintReport) string {
	if report == nil {
		return ""
	}

	var builder strings.Builder
	writeLine(&builder, "Sprint Report")
	writeLine(&builder, "")
	writeLine(&builder, FormatSprint(&report.Sprint))

	if hasSprintReportEstimate(report.AllIssuesEstimate) || hasSprintReportEstimate(report.CompletedIssuesEstimate) || hasSprintReportEstimate(report.IncompleteIssuesEstimate) || hasSprintReportEstimate(report.RemovedIssuesEstimate) {
		writeSection(&builder, "Estimate Totals")
		if hasSprintReportEstimate(report.AllIssuesEstimate) {
			writeLine(&builder, fmt.Sprintf("All Issues: %s", formatSprintReportEstimate(report.AllIssuesEstimate)))
		}
		if hasSprintReportEstimate(report.CompletedIssuesEstimate) {
			writeLine(&builder, fmt.Sprintf("Completed: %s", formatSprintReportEstimate(report.CompletedIssuesEstimate)))
		}
		if hasSprintReportEstimate(report.IncompleteIssuesEstimate) {
			writeLine(&builder, fmt.Sprintf("Not Completed: %s", formatSprintReportEstimate(report.IncompleteIssuesEstimate)))
		}
		if hasSprintReportEstimate(report.RemovedIssuesEstimate) {
			writeLine(&builder, fmt.Sprintf("Removed: %s", formatSprintReportEstimate(report.RemovedIssuesEstimate)))
		}
	}

	if len(report.AddedIssueKeys) > 0 {
		writeSection(&builder, fmt.Sprintf("Added During Sprint (%d)", len(report.AddedIssueKeys)))
		for _, key := range report.AddedIssueKeys {
			writeLine(&builder, "- "+key)
		}
	}

	writeSprintReportIssues(&builder, "Completed Issues", report.CompletedIssues)
	writeSprintReportIssues(&builder, "Incomplete Issues", report.IncompleteIssues)
	writeSprintReportIssues(&builder, "Removed Issues", report.RemovedIssues)
	writeSprintReportIssues(&builder, "Completed In Another Sprint", report.CompletedInAnotherSprint)

	return strings.TrimSpace(builder.String())
}

func FormatSprintHealthReport(report *models.SprintHealthReport) string {
	if report == nil {
		return ""
	}

	var builder strings.Builder
	writeLine(&builder, "Sprint Health Report")
	writeLine(&builder, "")
	writeLine(&builder, FormatSprint(&report.Sprint))

	writeSection(&builder, "Summary")
	writeLine(&builder, fmt.Sprintf("Committed Issues: %d", report.Summary.CommittedIssues))
	writeLine(&builder, fmt.Sprintf("Completed Commitment: %d", report.Summary.CompletedCommittedIssues))
	writeLine(&builder, fmt.Sprintf("Completed Issues: %d", report.Summary.CompletedIssues))
	writeLine(&builder, fmt.Sprintf("Incomplete Issues: %d", report.Summary.IncompleteIssues))
	writeLine(&builder, fmt.Sprintf("Completed In Another Sprint: %d", report.Summary.CompletedElsewhereIssues))
	writeLine(&builder, fmt.Sprintf("Carry Over: %d", report.Summary.CarryOverIssues))
	writeLine(&builder, fmt.Sprintf("Removed Issues: %d", report.Summary.RemovedIssues))
	writeLine(&builder, fmt.Sprintf("Added During Sprint: %d", report.Summary.AddedDuringSprint))
	writeLine(&builder, fmt.Sprintf("Unplanned Completed Issues: %d", report.Summary.UnplannedCompletedIssues))
	writeLine(&builder, fmt.Sprintf("Commitment Completion: %s", formatRatio(report.Summary.CompletionRatio)))
	if report.Summary.CommittedEstimate > 0 || report.Summary.CompletedCommittedEstimate > 0 {
		writeLine(&builder, fmt.Sprintf("Committed Estimate: %.2f", report.Summary.CommittedEstimate))
		writeLine(&builder, fmt.Sprintf("Completed Committed Estimate: %.2f", report.Summary.CompletedCommittedEstimate))
		writeLine(&builder, fmt.Sprintf("Estimate Completion: %s", formatRatio(report.Summary.EstimateCompletionRatio)))
	}

	if len(report.Risks) > 0 {
		writeSection(&builder, fmt.Sprintf("Risks (%d)", len(report.Risks)))
		for _, risk := range report.Risks {
			writeLine(&builder, "- "+risk)
		}
	}

	writeSprintReportIssues(&builder, "Top Incomplete Issues", report.TopIncompleteIssues)

	return strings.TrimSpace(builder.String())
}

func FormatAgingReport(report *models.AgingReport) string {
	if report == nil {
		return ""
	}
	if len(report.Items) == 0 {
		return "No issues matched the aging report criteria."
	}

	var builder strings.Builder
	writeLine(&builder, "Aging Report")
	writeLine(&builder, "")
	writeIf(&builder, report.ProjectKey, "Project: %s")
	writeIf(&builder, report.Query, "Query: %s")
	writeIf(&builder, strings.Join(report.Statuses, ", "), "Statuses: %s")
	writeIf(&builder, report.Assignee, "Assignee: %s")
	writeLine(&builder, fmt.Sprintf("Minimum Days In Status: %d", report.MinDaysInStatus))

	writeSection(&builder, "Summary")
	writeLine(&builder, fmt.Sprintf("Analyzed Issues: %d", report.Summary.AnalyzedIssues))
	writeLine(&builder, fmt.Sprintf("Matching Issues: %d", report.Summary.MatchingIssues))
	writeLine(&builder, fmt.Sprintf("Oldest Days In Status: %d", report.Summary.OldestDaysInStatus))
	writeLine(&builder, fmt.Sprintf("Average Days In Status: %.1f", report.Summary.AverageDaysInStatus))

	writeSection(&builder, fmt.Sprintf("Issues (%d)", len(report.Items)))
	for _, item := range report.Items {
		line := fmt.Sprintf("- %s", item.Key)
		if item.Summary != "" {
			line += ": " + item.Summary
		}
		if item.Status != "" {
			line += " [" + item.Status + "]"
		}
		writeLine(&builder, line)
		writeIf(&builder, item.Assignee, "  Assignee: %s")
		writeIf(&builder, item.Type, "  Type: %s")
		writeLine(&builder, fmt.Sprintf("  Days In Status: %d", item.DaysInStatus))
		writeIf(&builder, item.StatusSince, "  Status Since: %s")
		writeIf(&builder, item.Updated, "  Updated: %s")
	}

	return strings.TrimSpace(builder.String())
}

func FormatBlockedIssuesReport(report *models.BlockedIssuesReport) string {
	if report == nil {
		return ""
	}
	if len(report.Items) == 0 {
		return "No blocked issues matched the report criteria."
	}

	var builder strings.Builder
	writeLine(&builder, "Blocked Issues Report")
	writeLine(&builder, "")
	writeIf(&builder, report.ProjectKey, "Project: %s")
	writeIf(&builder, report.Query, "Query: %s")
	writeIf(&builder, strings.Join(report.Statuses, ", "), "Statuses: %s")
	writeIf(&builder, strings.Join(report.BlockedStatuses, ", "), "Blocked Statuses: %s")
	writeIf(&builder, strings.Join(report.LinkTypes, ", "), "Link Types: %s")
	writeIf(&builder, report.Assignee, "Assignee: %s")
	writeLine(&builder, fmt.Sprintf("Minimum Days In Status: %d", report.MinDaysInStatus))

	writeSection(&builder, "Summary")
	writeLine(&builder, fmt.Sprintf("Analyzed Issues: %d", report.Summary.AnalyzedIssues))
	writeLine(&builder, fmt.Sprintf("Blocked Issues: %d", report.Summary.BlockedIssues))
	writeLine(&builder, fmt.Sprintf("Blocked By Status: %d", report.Summary.BlockedByStatusIssues))
	writeLine(&builder, fmt.Sprintf("Blocked By Links: %d", report.Summary.BlockedByLinkIssues))
	writeLine(&builder, fmt.Sprintf("External Dependency Issues: %d", report.Summary.ExternalDependencyIssues))
	writeLine(&builder, fmt.Sprintf("Oldest Days In Status: %d", report.Summary.OldestDaysInStatus))
	writeLine(&builder, fmt.Sprintf("Average Days In Status: %.1f", report.Summary.AverageDaysInStatus))

	writeSection(&builder, fmt.Sprintf("Issues (%d)", len(report.Items)))
	for _, item := range report.Items {
		line := fmt.Sprintf("- %s", item.Key)
		if item.Summary != "" {
			line += ": " + item.Summary
		}
		if item.Status != "" {
			line += " [" + item.Status + "]"
		}
		writeLine(&builder, line)
		writeIf(&builder, item.Assignee, "  Assignee: %s")
		writeIf(&builder, item.Type, "  Type: %s")
		writeLine(&builder, fmt.Sprintf("  Days In Status: %d", item.DaysInStatus))
		writeIf(&builder, item.StatusSince, "  Status Since: %s")
		writeIf(&builder, item.Updated, "  Updated: %s")
		writeLine(&builder, fmt.Sprintf("  Blocked By Status: %t", item.BlockedByStatus))
		if len(item.Reasons) > 0 {
			writeLine(&builder, fmt.Sprintf("  Reasons: %s", strings.Join(item.Reasons, ", ")))
		}
		if len(item.BlockingLinks) > 0 {
			writeLine(&builder, "  Blocking Links:")
			for _, dependency := range item.BlockingLinks {
				dependencyLine := fmt.Sprintf("    - %s", dependency.Key)
				if dependency.Summary != "" {
					dependencyLine += ": " + dependency.Summary
				}
				if dependency.Status != "" {
					dependencyLine += " [" + dependency.Status + "]"
				}
				writeLine(&builder, dependencyLine)
				writeIf(&builder, dependency.Type, "      Type: %s")
				writeIf(&builder, dependency.Direction, "      Direction: %s")
				writeLine(&builder, fmt.Sprintf("      External: %t", dependency.External))
			}
		}
	}

	return strings.TrimSpace(builder.String())
}

func FormatFlowEfficiencyReport(report *models.FlowEfficiencyReport) string {
	if report == nil {
		return ""
	}
	if len(report.Items) == 0 {
		return "No issues matched the flow efficiency report criteria."
	}

	var builder strings.Builder
	writeLine(&builder, "Flow Efficiency Report")
	writeLine(&builder, "")
	writeIf(&builder, report.ProjectKey, "Project: %s")
	writeIf(&builder, report.Query, "Query: %s")
	writeIf(&builder, strings.Join(report.Statuses, ", "), "Statuses: %s")
	writeIf(&builder, strings.Join(report.ActiveStatuses, ", "), "Active Statuses: %s")
	writeIf(&builder, strings.Join(report.BlockedStatuses, ", "), "Blocked Statuses: %s")
	writeIf(&builder, report.Assignee, "Assignee: %s")
	writeLine(&builder, fmt.Sprintf("Minimum Blocked Days: %d", report.MinBlockedDays))
	if report.WindowDays > 0 {
		writeLine(&builder, fmt.Sprintf("Window Days: %d", report.WindowDays))
	}
	writeIf(&builder, report.WindowStart, "Window Start: %s")
	writeIf(&builder, report.WindowEnd, "Window End: %s")

	writeSection(&builder, "Summary")
	writeLine(&builder, fmt.Sprintf("Analyzed Issues: %d", report.Summary.AnalyzedIssues))
	writeLine(&builder, fmt.Sprintf("Matching Issues: %d", report.Summary.MatchingIssues))
	writeLine(&builder, fmt.Sprintf("Issues With Blocked Time: %d", report.Summary.IssuesWithBlockedTime))
	writeLine(&builder, fmt.Sprintf("Currently Blocked Issues: %d", report.Summary.CurrentlyBlockedIssues))
	writeLine(&builder, fmt.Sprintf("Total Observed Days: %d", report.Summary.TotalObservedDays))
	writeLine(&builder, fmt.Sprintf("Total Active Days: %d", report.Summary.TotalActiveDays))
	writeLine(&builder, fmt.Sprintf("Total Blocked Days: %d", report.Summary.TotalBlockedDays))
	writeLine(&builder, fmt.Sprintf("Average Flow Efficiency: %s", formatRatio(report.Summary.AverageFlowEfficiency)))
	writeLine(&builder, fmt.Sprintf("Portfolio Flow Efficiency: %s", formatRatio(report.Summary.PortfolioFlowEfficiency)))

	if len(report.Assignees) > 0 {
		writeSection(&builder, fmt.Sprintf("Assignees (%d)", len(report.Assignees)))
		for _, assignee := range report.Assignees {
			writeLine(&builder, fmt.Sprintf("- %s", assignee.Assignee))
			writeLine(&builder, fmt.Sprintf("  Issues: %d", assignee.Issues))
			writeLine(&builder, fmt.Sprintf("  Issues With Blocked Time: %d", assignee.IssuesWithBlockedTime))
			writeLine(&builder, fmt.Sprintf("  Currently Blocked Issues: %d", assignee.CurrentlyBlockedIssues))
			writeLine(&builder, fmt.Sprintf("  Total Observed Days: %d", assignee.TotalObservedDays))
			writeLine(&builder, fmt.Sprintf("  Total Active Days: %d", assignee.TotalActiveDays))
			writeLine(&builder, fmt.Sprintf("  Total Blocked Days: %d", assignee.TotalBlockedDays))
			writeLine(&builder, fmt.Sprintf("  Average Flow Efficiency: %s", formatRatio(assignee.AverageFlowEfficiency)))
			writeLine(&builder, fmt.Sprintf("  Portfolio Flow Efficiency: %s", formatRatio(assignee.PortfolioFlowEfficiency)))
		}
	}

	writeSection(&builder, fmt.Sprintf("Issues (%d)", len(report.Items)))
	for _, item := range report.Items {
		line := fmt.Sprintf("- %s", item.Key)
		if item.Summary != "" {
			line += ": " + item.Summary
		}
		if item.Status != "" {
			line += " [" + item.Status + "]"
		}
		writeLine(&builder, line)
		writeIf(&builder, item.Assignee, "  Assignee: %s")
		writeIf(&builder, item.Type, "  Type: %s")
		writeLine(&builder, fmt.Sprintf("  Flow Efficiency: %s", formatRatio(item.FlowEfficiency)))
		writeLine(&builder, fmt.Sprintf("  Observed Days: %d", item.ObservedDays))
		writeLine(&builder, fmt.Sprintf("  Active Days: %d", item.ActiveDays))
		writeLine(&builder, fmt.Sprintf("  Blocked Days: %d", item.BlockedDays))
		writeLine(&builder, fmt.Sprintf("  Blocked Transitions: %d", item.BlockedTransitions))
		writeLine(&builder, fmt.Sprintf("  Currently Blocked: %t", item.CurrentlyBlocked))
		writeLine(&builder, fmt.Sprintf("  Current Status Days: %d", item.CurrentStatusDays))
		writeIf(&builder, item.CurrentStatusSince, "  Current Status Since: %s")
		writeIf(&builder, item.Updated, "  Updated: %s")
	}

	return strings.TrimSpace(builder.String())
}

func FormatSprints(collection *models.SprintCollection) string {
	if collection == nil || len(collection.Sprints) == 0 {
		return "No sprints found."
	}
	var parts []string
	for _, sprint := range collection.Sprints {
		parts = append(parts, FormatSprint(&sprint))
	}
	return strings.Join(parts, "\n\n")
}

func FormatComment(comment *models.Comment) string {
	if comment == nil {
		return ""
	}
	var builder strings.Builder
	writeLine(&builder, fmt.Sprintf("ID: %s", comment.ID))
	writeIf(&builder, personLabel(comment.Author), "Author: %s")
	writeIf(&builder, comment.Created, "Created: %s")
	writeIf(&builder, comment.Updated, "Updated: %s")
	if comment.Body != "" {
		writeLine(&builder, "Body:")
		writeLine(&builder, comment.Body)
	}
	return strings.TrimSpace(builder.String())
}

func FormatComments(list *models.CommentList) string {
	if list == nil || len(list.Comments) == 0 {
		return "No comments found for this issue."
	}
	var parts []string
	for _, comment := range list.Comments {
		parts = append(parts, FormatComment(&comment))
	}
	return strings.Join(parts, "\n\n")
}

func FormatWorklog(worklog *models.Worklog) string {
	if worklog == nil {
		return ""
	}
	var builder strings.Builder
	writeLine(&builder, "Worklog added successfully!")
	writeIf(&builder, worklog.IssueKey, "Issue: %s")
	writeIf(&builder, worklog.ID, "Worklog ID: %s")
	if worklog.TimeSpent != "" {
		writeLine(&builder, fmt.Sprintf("Time Spent: %s (%d seconds)", worklog.TimeSpent, worklog.TimeSpentSeconds))
	}
	writeIf(&builder, worklog.Started, "Date Started: %s")
	writeIf(&builder, personLabel(worklog.Author), "Author: %s")
	writeIf(&builder, worklog.Comment, "Comment: %s")
	return strings.TrimSpace(builder.String())
}

func FormatTransitions(transitions []models.Transition) string {
	if len(transitions) == 0 {
		return "No transitions available."
	}
	var builder strings.Builder
	for _, transition := range transitions {
		writeLine(&builder, fmt.Sprintf("ID: %s  Name: %s", transition.ID, transition.Name))
	}
	return strings.TrimSpace(builder.String())
}

func FormatStatuses(catalog *models.StatusCatalog) string {
	if catalog == nil || len(catalog.Groups) == 0 {
		return "No statuses found for this project."
	}
	var builder strings.Builder
	writeLine(&builder, "Available Statuses:")
	for _, group := range catalog.Groups {
		writeLine(&builder, "")
		writeLine(&builder, fmt.Sprintf("Issue Type: %s", group.IssueType.Name))
		for _, status := range group.Statuses {
			writeLine(&builder, fmt.Sprintf("  - %s: %s", status.Name, status.ID))
		}
	}
	return strings.TrimSpace(builder.String())
}

func FormatHistory(history *models.IssueHistory) string {
	if history == nil {
		return "No history found."
	}
	if len(history.Entries) == 0 {
		return fmt.Sprintf("No history found for issue %s.", history.IssueKey)
	}
	var builder strings.Builder
	for _, entry := range history.Entries {
		writeLine(&builder, fmt.Sprintf("Date: %s  Author: %s", entry.Date, entry.Author))
		for _, change := range entry.Changes {
			writeLine(&builder, fmt.Sprintf("  Field: %s  From: %s  To: %s", change.Field, change.From, change.To))
		}
	}
	return strings.TrimSpace(builder.String())
}

func FormatRelations(issueKey string, relations []models.IssueRelation) string {
	if len(relations) == 0 {
		return fmt.Sprintf("Issue %s has no linked issues.", issueKey)
	}
	var builder strings.Builder
	writeLine(&builder, fmt.Sprintf("Related issues for %s:", issueKey))
	writeLine(&builder, "")
	for _, relation := range relations {
		writeLine(&builder, fmt.Sprintf("Relationship: %s", relation.Type))
		writeLine(&builder, fmt.Sprintf("Issue: %s", relation.Issue.Key))
		writeIf(&builder, relation.Issue.Summary, "Summary: %s")
		if relation.Issue.Status != nil {
			writeIf(&builder, relation.Issue.Status.Name, "Status: %s")
		}
		writeLine(&builder, "")
	}
	return strings.TrimSpace(builder.String())
}

func FormatVersion(version *models.Version) string {
	if version == nil {
		return ""
	}
	var builder strings.Builder
	writeLine(&builder, "Version Details:")
	writeLine(&builder, "")
	writeIf(&builder, version.ID, "ID: %s")
	writeIf(&builder, version.Name, "Name: %s")
	writeIf(&builder, version.Description, "Description: %s")
	if version.ProjectID > 0 {
		writeLine(&builder, fmt.Sprintf("Project ID: %d", version.ProjectID))
	}
	writeLine(&builder, fmt.Sprintf("Released: %t", version.Released))
	writeLine(&builder, fmt.Sprintf("Archived: %t", version.Archived))
	writeIf(&builder, version.ReleaseDate, "Release Date: %s")
	writeIf(&builder, version.URL, "URL: %s")
	return strings.TrimSpace(builder.String())
}

func FormatVersions(collection *models.VersionCollection) string {
	if collection == nil || len(collection.Versions) == 0 {
		return fmt.Sprintf("No versions found for project %s.", collection.ProjectKey)
	}
	var builder strings.Builder
	writeLine(&builder, fmt.Sprintf("Project %s Versions:", collection.ProjectKey))
	writeLine(&builder, "")
	for index, version := range collection.Versions {
		if index > 0 {
			writeLine(&builder, "")
		}
		writeIf(&builder, version.ID, "ID: %s")
		writeIf(&builder, version.Name, "Name: %s")
		writeIf(&builder, version.Description, "Description: %s")
		writeIf(&builder, version.Status, "Status: %s")
		writeIf(&builder, version.ReleaseDate, "Release Date: %s")
	}
	return strings.TrimSpace(builder.String())
}

func FormatDevelopmentInfo(info *models.DevelopmentInfo) string {
	if info == nil {
		return ""
	}
	var builder strings.Builder
	writeLine(&builder, fmt.Sprintf("Development Information for %s:", info.IssueKey))

	if len(info.Branches) > 0 {
		writeSection(&builder, fmt.Sprintf("Branches (%d)", len(info.Branches)))
		for _, branch := range info.Branches {
			writeLine(&builder, fmt.Sprintf("- %s [%s]", branch.Name, branch.Repository.Name))
			writeIf(&builder, branch.URL, "  URL: %s")
			if branch.LastCommit.Message != "" {
				writeLine(&builder, fmt.Sprintf("  Last Commit: %s", branch.LastCommit.Message))
			}
		}
	}

	if len(info.PullRequests) > 0 {
		writeSection(&builder, fmt.Sprintf("Pull Requests (%d)", len(info.PullRequests)))
		for _, pullRequest := range info.PullRequests {
			writeLine(&builder, fmt.Sprintf("- %s [%s]", pullRequest.Title, pullRequest.Status))
			writeIf(&builder, personLabel(pullRequest.Author), "  Author: %s")
			writeIf(&builder, pullRequest.URL, "  URL: %s")
		}
	}

	commitCount := 0
	for _, repository := range info.Repositories {
		commitCount += len(repository.Commits)
	}
	if commitCount > 0 {
		writeSection(&builder, fmt.Sprintf("Commits (%d)", commitCount))
		for _, repository := range info.Repositories {
			if len(repository.Commits) == 0 {
				continue
			}
			writeLine(&builder, fmt.Sprintf("Repository: %s", repository.Name))
			for _, commit := range repository.Commits {
				identifier := commit.DisplayID
				if identifier == "" {
					identifier = commit.ID
				}
				writeLine(&builder, fmt.Sprintf("  - %s %s", identifier, commit.Message))
				writeIf(&builder, personLabel(commit.Author), "    Author: %s")
				writeIf(&builder, commit.AuthorTimestamp, "    Date: %s")
				writeIf(&builder, commit.URL, "    URL: %s")
			}
		}
	}

	if len(info.Builds) > 0 {
		writeSection(&builder, fmt.Sprintf("Builds (%d)", len(info.Builds)))
		for _, build := range info.Builds {
			label := build.DisplayName
			if label == "" {
				label = build.Name
			}
			writeLine(&builder, fmt.Sprintf("- %s [%s]", label, build.State))
			writeIf(&builder, build.URL, "  URL: %s")
			writeIf(&builder, build.LastUpdated, "  Updated: %s")
		}
	}

	if len(info.Branches) == 0 && len(info.PullRequests) == 0 && commitCount == 0 && len(info.Builds) == 0 {
		writeLine(&builder, "\nNo development information linked to this issue.")
	}

	return strings.TrimSpace(builder.String())
}

func FormatStoredAttachment(attachment *models.StoredAttachment) string {
	if attachment == nil {
		return ""
	}
	var builder strings.Builder
	writeLine(&builder, "Attachment downloaded successfully!")
	writeIf(&builder, attachment.Path, "File: %s")
	writeIf(&builder, attachment.Filename, "Filename: %s")
	if attachment.Size > 0 {
		writeLine(&builder, fmt.Sprintf("Size: %d bytes", attachment.Size))
	}
	writeIf(&builder, attachment.MimeType, "MIME Type: %s")
	return strings.TrimSpace(builder.String())
}

func personLabel(person *models.Person) string {
	if person == nil {
		return ""
	}
	if person.DisplayName != "" && person.Email != "" {
		return fmt.Sprintf("%s (%s)", person.DisplayName, person.Email)
	}
	if person.DisplayName != "" {
		return person.DisplayName
	}
	return person.Email
}

func issueRefLabel(issue models.IssueRef) string {
	label := issue.Key
	if issue.Summary != "" {
		label += ": " + issue.Summary
	}
	if issue.Status != nil && issue.Status.Name != "" {
		label += " [" + issue.Status.Name + "]"
	}
	return label
}

func writeSection(builder *strings.Builder, title string) {
	builder.WriteString("\n=== ")
	builder.WriteString(title)
	builder.WriteString(" ===\n")
}

func writeLine(builder *strings.Builder, value string) {
	builder.WriteString(value)
	builder.WriteByte('\n')
}

func writeIf(builder *strings.Builder, value, pattern string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	writeLine(builder, fmt.Sprintf(pattern, value))
}

func writeSprintReportIssues(builder *strings.Builder, title string, issues []models.SprintReportIssue) {
	if len(issues) == 0 {
		return
	}

	writeSection(builder, fmt.Sprintf("%s (%d)", title, len(issues)))
	for _, issue := range issues {
		line := fmt.Sprintf("- %s", issue.Key)
		if issue.Summary != "" {
			line += ": " + issue.Summary
		}
		if issue.Status != "" {
			line += " [" + issue.Status + "]"
		}
		writeLine(builder, line)
		writeIf(builder, issue.Type, "  Type: %s")
		if hasSprintReportEstimate(issue.Estimate) {
			writeLine(builder, fmt.Sprintf("  Estimated: %s", formatSprintReportEstimate(issue.Estimate)))
		}
		if hasSprintReportEstimate(issue.CurrentEstimate) {
			writeLine(builder, fmt.Sprintf("  Current Estimate: %s", formatSprintReportEstimate(issue.CurrentEstimate)))
		}
		if issue.AddedDuringSprint {
			writeLine(builder, "  Added During Sprint: yes")
		}
	}
}

func hasSprintReportEstimate(value models.SprintReportEstimate) bool {
	return strings.TrimSpace(value.Text) != "" || value.Value != 0
}

func formatSprintReportEstimate(value models.SprintReportEstimate) string {
	text := strings.TrimSpace(value.Text)
	if text != "" && value.Value != 0 {
		return fmt.Sprintf("%s (%.2f)", text, value.Value)
	}
	if text != "" {
		return text
	}
	return fmt.Sprintf("%.2f", value.Value)
}

func formatRatio(value float64) string {
	return fmt.Sprintf("%.1f%%", value*100)
}

func FormatSprintWorkloadReport(report *models.SprintWorkloadReport) string {
	if report == nil {
		return "No sprint workload report available."
	}
	var b strings.Builder
	writeSection(&b, "Sprint Workload Report")
	if report.Sprint.ID != 0 {
		writeLine(&b, fmt.Sprintf("Sprint ID: %d", report.Sprint.ID))
	}
	writeIf(&b, report.Sprint.Name, "Sprint: %s")
	writeIf(&b, report.Sprint.State, "State: %s")
	if len(report.Assignees) == 0 {
		writeLine(&b, "No assignees found in this sprint.")
		return strings.TrimSpace(b.String())
	}
	for _, assignee := range report.Assignees {
		writeSection(&b, fmt.Sprintf("%s (%d issues)", assignee.Assignee, assignee.TotalIssues))
		writeLine(&b, fmt.Sprintf("Committed: %d   Completed: %d   Incomplete: %d   Added: %d   Removed: %d",
			assignee.CommittedIssues, assignee.CompletedIssues, assignee.IncompleteIssues,
			assignee.AddedDuringSprintIssues, assignee.RemovedIssues))
		if assignee.CommittedEstimate > 0 || assignee.CompletedEstimate > 0 {
			writeLine(&b, fmt.Sprintf("Estimate: committed %.2f / completed %.2f", assignee.CommittedEstimate, assignee.CompletedEstimate))
		}
		if assignee.CommittedIssues > 0 {
			writeLine(&b, fmt.Sprintf("Commitment completion: %s", formatRatio(assignee.CommitmentCompletionRatio)))
		}
		if len(assignee.StatusBreakdown) > 0 {
			parts := make([]string, 0, len(assignee.StatusBreakdown))
			for _, status := range assignee.StatusBreakdown {
				parts = append(parts, fmt.Sprintf("%s=%d", status.Status, status.Issues))
			}
			writeLine(&b, "Statuses: "+strings.Join(parts, ", "))
		}
		if len(assignee.IncompleteIssueKeys) > 0 {
			writeLine(&b, "Incomplete: "+strings.Join(assignee.IncompleteIssueKeys, ", "))
		}
		if len(assignee.AddedDuringSprintIssueKeys) > 0 {
			writeLine(&b, "Added during sprint: "+strings.Join(assignee.AddedDuringSprintIssueKeys, ", "))
		}
	}
	return strings.TrimSpace(b.String())
}

func FormatUtilizationReport(report *models.UtilizationReport) string {
	if report == nil {
		return "No utilization report available."
	}
	var b strings.Builder
	writeSection(&b, "Utilization Report")
	writeIf(&b, report.Query, "Query: %s")
	writeIf(&b, report.WindowStart, "Window Start: %s")
	writeIf(&b, report.WindowEnd, "Window End: %s")
	writeLine(&b, fmt.Sprintf("Window Days: %d", report.WindowDays))
	writeLine(&b, fmt.Sprintf("Issues Analyzed: %d   Worklog Entries: %d   Hours Logged: %.2f   Authors: %d",
		report.Summary.AnalyzedIssues, report.Summary.WorklogEntries, report.Summary.HoursLogged, report.Summary.DistinctAuthors))
	if len(report.Assignees) == 0 {
		writeLine(&b, "No worklogs in window.")
		return strings.TrimSpace(b.String())
	}
	for _, assignee := range report.Assignees {
		writeSection(&b, fmt.Sprintf("%s — %.2fh", assignee.Assignee, assignee.HoursLogged))
		writeLine(&b, fmt.Sprintf("Entries: %d   Issues touched: %d   Daily avg: %.2fh",
			assignee.WorklogEntries, assignee.IssuesTouched, assignee.DailyAverageHrs))
		for _, issue := range assignee.TopIssues {
			line := fmt.Sprintf("- %s [%s] %.2fh (%d entries)", issue.Key, issue.Status, issue.HoursLogged, issue.WorklogEntries)
			if issue.Summary != "" {
				line += " — " + issue.Summary
			}
			writeLine(&b, line)
		}
	}
	return strings.TrimSpace(b.String())
}

func FormatCycleTimeReport(report *models.CycleTimeReport) string {
	if report == nil {
		return "No cycle time report available."
	}
	var b strings.Builder
	writeSection(&b, "Cycle Time Report")
	writeIf(&b, report.Query, "Query: %s")
	writeIf(&b, report.WindowStart, "Window Start: %s")
	writeIf(&b, report.WindowEnd, "Window End: %s")
	if len(report.StartStatuses) > 0 {
		writeLine(&b, "Start statuses: "+strings.Join(report.StartStatuses, ", "))
	}
	if len(report.DoneStatuses) > 0 {
		writeLine(&b, "Done statuses: "+strings.Join(report.DoneStatuses, ", "))
	}
	writeLine(&b, fmt.Sprintf("Analyzed: %d   Completed: %d   Throughput/week: %.2f",
		report.Summary.AnalyzedIssues, report.Summary.CompletedIssues, report.Summary.ThroughputPerWeek))
	writeLine(&b, fmt.Sprintf("Cycle hours — avg %.2f / median %.2f / p85 %.2f",
		report.Summary.AverageCycleHours, report.Summary.MedianCycleHours, report.Summary.P85CycleHours))
	writeLine(&b, fmt.Sprintf("Lead hours — avg %.2f", report.Summary.AverageLeadHours))

	if len(report.Assignees) > 0 {
		writeSection(&b, "By Assignee")
		for _, a := range report.Assignees {
			writeLine(&b, fmt.Sprintf("- %s: completed=%d avgCycle=%.2fh medianCycle=%.2fh avgLead=%.2fh",
				a.Assignee, a.CompletedIssues, a.AverageCycleHrs, a.MedianCycleHrs, a.AverageLeadHrs))
		}
	}
	if len(report.Types) > 0 {
		writeSection(&b, "By Type")
		for _, t := range report.Types {
			writeLine(&b, fmt.Sprintf("- %s: completed=%d avgCycle=%.2fh medianCycle=%.2fh",
				t.Type, t.CompletedIssues, t.AverageCycleHrs, t.MedianCycleHrs))
		}
	}
	if len(report.Items) > 0 {
		writeSection(&b, "Items")
		for _, item := range report.Items {
			line := fmt.Sprintf("- %s [%s] %s — cycle=%.2fh lead=%.2fh", item.Key, item.Status, item.Assignee, item.CycleHours, item.LeadHours)
			if item.Summary != "" {
				line += " — " + item.Summary
			}
			writeLine(&b, line)
		}
	}
	return strings.TrimSpace(b.String())
}

func FormatTeamWipSnapshot(snapshot *models.TeamWipSnapshot) string {
	if snapshot == nil {
		return "No team WIP snapshot available."
	}
	var b strings.Builder
	writeSection(&b, "Team WIP Snapshot")
	writeIf(&b, snapshot.Query, "Query: %s")
	if len(snapshot.Statuses) > 0 {
		writeLine(&b, "Statuses: "+strings.Join(snapshot.Statuses, ", "))
	}
	if snapshot.WipLimit > 0 {
		writeLine(&b, fmt.Sprintf("WIP Limit: %d", snapshot.WipLimit))
	}
	writeLine(&b, fmt.Sprintf("Analyzed: %d   Assignees: %d   Unassigned: %d",
		snapshot.Summary.AnalyzedIssues, snapshot.Summary.DistinctAssignees, snapshot.Summary.UnassignedIssues))
	if len(snapshot.Summary.StatusTotals) > 0 {
		keys := make([]string, 0, len(snapshot.Summary.StatusTotals))
		for k := range snapshot.Summary.StatusTotals {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%d", k, snapshot.Summary.StatusTotals[k]))
		}
		writeLine(&b, "Totals: "+strings.Join(parts, ", "))
	}
	for _, assignee := range snapshot.Assignees {
		header := fmt.Sprintf("%s (%d in flight)", assignee.Assignee, assignee.TotalWip)
		if assignee.OverLimit {
			header += " ⚠ over limit"
		}
		writeSection(&b, header)
		for _, status := range assignee.Statuses {
			writeLine(&b, fmt.Sprintf("[%s] %d", status.Status, status.Issues))
			for _, item := range status.Items {
				line := fmt.Sprintf("  - %s", item.Key)
				if item.Summary != "" {
					line += " — " + item.Summary
				}
				if item.Priority != "" {
					line += " (" + item.Priority + ")"
				}
				writeLine(&b, line)
			}
		}
	}
	return strings.TrimSpace(b.String())
}

func FormatWorklogList(list *models.WorklogList) string {
	if list == nil || len(list.Worklogs) == 0 {
		return "No worklogs found."
	}
	var b strings.Builder
	writeSection(&b, fmt.Sprintf("Worklogs for %s", list.IssueKey))
	writeLine(&b, fmt.Sprintf("Total: %d", list.Total))
	for _, w := range list.Worklogs {
		line := fmt.Sprintf("- %s by %s — %s (%ds)", w.Started, personLabel(w.Author), w.TimeSpent, w.TimeSpentSeconds)
		writeLine(&b, line)
	}
	return strings.TrimSpace(b.String())
}
