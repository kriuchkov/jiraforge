package jira

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-faster/errors"

	"github.com/kriuchkov/jiraforge/internal/core/models"
)

var agingReportNow = time.Now

func (s *Service) GetAgingReport(ctx context.Context, req models.GetAgingReportRequest) (*models.AgingReport, error) {
	req = req.Normalized()
	if strings.TrimSpace(req.JQL) == "" && strings.TrimSpace(req.ProjectKey) == "" {
		return nil, invalid("service.jira.GetAgingReport", "either jql or project_key is required")
	}

	query := strings.TrimSpace(req.JQL)
	if query == "" {
		query = buildAgingReportJQL(req)
	}

	searchResult, err := s.gateway.SearchIssues(ctx, models.SearchIssuesRequest{
		JQL:        query,
		Fields:     []string{"summary", "status", "assignee", "updated", "created", "issuetype", "statuscategorychangedate"},
		MaxResults: req.MaxResults,
	})
	if err != nil {
		return nil, errors.Wrap(err, "search issues")
	}

	report := &models.AgingReport{
		Query:           query,
		ProjectKey:      req.ProjectKey,
		Statuses:        append([]string(nil), req.Statuses...),
		Assignee:        req.Assignee,
		MinDaysInStatus: req.MinDaysInStatus,
	}
	if searchResult == nil {
		return report, nil
	}

	now := agingReportNow().UTC()
	report.Summary.AnalyzedIssues = len(searchResult.Issues)
	var totalDays int
	for _, issue := range searchResult.Issues {
		history, err := s.gateway.GetIssueHistory(ctx, models.GetIssueHistoryRequest{IssueKey: issue.Key})
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("get issue history for %s", issue.Key))
		}

		statusSince := latestStatusChangeTime(history)
		if statusSince.IsZero() {
			statusSince = fallbackStatusSince(issue)
		}
		if statusSince.IsZero() {
			statusSince = now
		}

		daysInStatus := daysBetween(statusSince, now)
		if daysInStatus < req.MinDaysInStatus {
			continue
		}

		item := models.AgingReportItem{
			Key:          issue.Key,
			Summary:      issue.Summary,
			Type:         issueTypeLabel(issue.Type),
			Status:       statusLabel(issue.Status),
			Assignee:     personLabel(issue.Assignee),
			Updated:      issue.Updated,
			StatusSince:  statusSince.Format(time.RFC3339),
			DaysInStatus: daysInStatus,
		}
		report.Items = append(report.Items, item)
		totalDays += daysInStatus
		if daysInStatus > report.Summary.OldestDaysInStatus {
			report.Summary.OldestDaysInStatus = daysInStatus
		}
	}

	sort.SliceStable(report.Items, func(i, j int) bool {
		if report.Items[i].DaysInStatus != report.Items[j].DaysInStatus {
			return report.Items[i].DaysInStatus > report.Items[j].DaysInStatus
		}
		return report.Items[i].Key < report.Items[j].Key
	})

	report.Summary.MatchingIssues = len(report.Items)
	if len(report.Items) > 0 {
		report.Summary.AverageDaysInStatus = float64(totalDays) / float64(len(report.Items))
	}

	return report, nil
}

func buildAgingReportJQL(req models.GetAgingReportRequest) string {
	parts := []string{fmt.Sprintf("project = %s", quoteJQLString(req.ProjectKey)), "resolution = Unresolved"}
	if len(req.Statuses) > 0 {
		quoted := make([]string, 0, len(req.Statuses))
		for _, status := range req.Statuses {
			quoted = append(quoted, quoteJQLString(status))
		}
		parts = append(parts, fmt.Sprintf("status IN (%s)", strings.Join(quoted, ", ")))
	}
	if strings.TrimSpace(req.Assignee) != "" {
		parts = append(parts, fmt.Sprintf("assignee = %s", quoteJQLString(req.Assignee)))
	}
	return strings.Join(parts, " AND ") + " ORDER BY updated ASC"
}

func issueTypeLabel(issueType *models.IssueType) string {
	if issueType == nil {
		return ""
	}
	return issueType.Name
}

func statusLabel(status *models.Status) string {
	if status == nil {
		return ""
	}
	return status.Name
}

func personLabel(person *models.Person) string {
	if person == nil {
		return ""
	}
	if strings.TrimSpace(person.DisplayName) != "" {
		return person.DisplayName
	}
	return person.Email
}

func latestStatusChangeTime(history *models.IssueHistory) time.Time {
	if history == nil {
		return time.Time{}
	}

	var latest time.Time
	for _, entry := range history.Entries {
		if !historyEntryHasStatusChange(entry) {
			continue
		}
		parsed, ok := parseAgingTime(entry.Date)
		if ok && parsed.After(latest) {
			latest = parsed
		}
	}
	return latest
}

func historyEntryHasStatusChange(entry models.HistoryEntry) bool {
	for _, change := range entry.Changes {
		if strings.EqualFold(strings.TrimSpace(change.Field), "status") {
			return true
		}
	}
	return false
}

func fallbackStatusSince(issue models.Issue) time.Time {
	for _, value := range []string{issue.Created, issue.StatusCategoryChange, issue.Updated} {
		if parsed, ok := parseAgingTime(value); ok {
			return parsed
		}
	}
	return time.Time{}
}

func parseAgingTime(value string) (time.Time, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, false
	}

	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.999-0700",
	} {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return parsed.UTC(), true
		}
	}
	if parsed, err := time.ParseInLocation("2006-01-02 15:04:05", trimmed, time.UTC); err == nil {
		return parsed.UTC(), true
	}
	return time.Time{}, false
}

func daysBetween(start, end time.Time) int {
	if start.IsZero() || end.Before(start) {
		return 0
	}
	return int(end.Sub(start).Hours() / 24)
}

func quoteJQLString(value string) string {
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(strings.TrimSpace(value), "\"", `\"`))
}
