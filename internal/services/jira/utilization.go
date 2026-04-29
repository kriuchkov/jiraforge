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

var utilizationReportNow = time.Now

func (s *Service) GetUtilizationReport(ctx context.Context, req models.GetUtilizationReportRequest) (*models.UtilizationReport, error) {
	req = req.Normalized()
	if strings.TrimSpace(req.JQL) == "" && strings.TrimSpace(req.ProjectKey) == "" {
		return nil, invalid("service.jira.GetUtilizationReport", "either jql or project_key is required")
	}

	now := utilizationReportNow().UTC()
	windowEnd, windowStart, err := resolveReportWindow(req.StartDate, req.EndDate, req.WindowDays, now)
	if err != nil {
		return nil, err
	}

	query := strings.TrimSpace(req.JQL)
	if query == "" {
		query = buildUtilizationJQL(req, windowStart, windowEnd)
	}

	searchResult, err := s.gateway.SearchIssues(ctx, models.SearchIssuesRequest{
		JQL:        query,
		Fields:     []string{"summary", "status", "assignee", "issuetype", "updated"},
		MaxResults: req.MaxResults,
	})
	if err != nil {
		return nil, errors.Wrap(err, "search issues")
	}

	report := &models.UtilizationReport{
		Query:       query,
		ProjectKey:  req.ProjectKey,
		WindowDays:  daysBetween(windowStart, windowEnd),
		WindowStart: windowStart.Format(time.RFC3339),
		WindowEnd:   windowEnd.Format(time.RFC3339),
	}
	if searchResult == nil {
		return report, nil
	}
	report.Summary.AnalyzedIssues = len(searchResult.Issues)
	report.Summary.WindowDays = report.WindowDays

	type assigneeAccumulator struct {
		assignee       string
		issueDetails   map[string]*models.UtilizationIssueDetail
		seconds        int
		entries        int
	}
	accumulators := map[string]*assigneeAccumulator{}
	getAcc := func(name string) *assigneeAccumulator {
		key := flowEfficiencyAssigneeLabel(name)
		acc, ok := accumulators[key]
		if !ok {
			acc = &assigneeAccumulator{assignee: key, issueDetails: map[string]*models.UtilizationIssueDetail{}}
			accumulators[key] = acc
		}
		return acc
	}

	for _, issue := range searchResult.Issues {
		worklogs, err := s.gateway.GetWorklogs(ctx, models.GetWorklogsRequest{
			IssueKey:     issue.Key,
			StartedAfter: windowStart.Format(time.RFC3339),
		})
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("get worklogs for %s", issue.Key))
		}
		if worklogs == nil {
			continue
		}
		for _, entry := range worklogs.Worklogs {
			started, ok := parseAgingTime(entry.Started)
			if !ok || started.Before(windowStart) || started.After(windowEnd) {
				continue
			}
			author := personLabel(entry.Author)
			acc := getAcc(author)
			acc.seconds += entry.TimeSpentSeconds
			acc.entries++
			detail, ok := acc.issueDetails[issue.Key]
			if !ok {
				detail = &models.UtilizationIssueDetail{
					Key:     issue.Key,
					Summary: issue.Summary,
					Status:  statusLabel(issue.Status),
					Type:    issueTypeLabel(issue.Type),
				}
				acc.issueDetails[issue.Key] = detail
			}
			detail.SecondsLogged += entry.TimeSpentSeconds
			detail.WorklogEntries++
			report.Summary.SecondsLogged += entry.TimeSpentSeconds
			report.Summary.WorklogEntries++
		}
	}

	report.Summary.HoursLogged = secondsToHours(report.Summary.SecondsLogged)
	report.Summary.DistinctAuthors = len(accumulators)

	windowDays := report.WindowDays
	if windowDays <= 0 {
		windowDays = 1
	}

	for _, acc := range accumulators {
		summary := models.UtilizationAssigneeSummary{
			Assignee:        acc.assignee,
			SecondsLogged:   acc.seconds,
			HoursLogged:     secondsToHours(acc.seconds),
			WorklogEntries:  acc.entries,
			IssuesTouched:   len(acc.issueDetails),
			DailyAverageHrs: secondsToHours(acc.seconds) / float64(windowDays),
		}
		details := make([]models.UtilizationIssueDetail, 0, len(acc.issueDetails))
		for _, detail := range acc.issueDetails {
			detail.HoursLogged = secondsToHours(detail.SecondsLogged)
			details = append(details, *detail)
		}
		sort.SliceStable(details, func(i, j int) bool {
			if details[i].SecondsLogged != details[j].SecondsLogged {
				return details[i].SecondsLogged > details[j].SecondsLogged
			}
			return details[i].Key < details[j].Key
		})
		if req.TopIssuesPerUser > 0 && len(details) > req.TopIssuesPerUser {
			details = details[:req.TopIssuesPerUser]
		}
		summary.TopIssues = details
		report.Assignees = append(report.Assignees, summary)
	}

	sort.SliceStable(report.Assignees, func(i, j int) bool {
		if report.Assignees[i].SecondsLogged != report.Assignees[j].SecondsLogged {
			return report.Assignees[i].SecondsLogged > report.Assignees[j].SecondsLogged
		}
		return report.Assignees[i].Assignee < report.Assignees[j].Assignee
	})

	return report, nil
}

func resolveReportWindow(startDate, endDate string, windowDays int, now time.Time) (windowEnd time.Time, windowStart time.Time, err error) {
	windowEnd = now
	if endDate != "" {
		parsed, ok := parseReportDate(endDate)
		if !ok {
			return time.Time{}, time.Time{}, invalid("service.jira", fmt.Sprintf("invalid end_date: %q", endDate))
		}
		windowEnd = parsed
	}
	if startDate != "" {
		parsed, ok := parseReportDate(startDate)
		if !ok {
			return time.Time{}, time.Time{}, invalid("service.jira", fmt.Sprintf("invalid start_date: %q", startDate))
		}
		windowStart = parsed
	} else {
		days := windowDays
		if days <= 0 {
			days = 7
		}
		windowStart = windowEnd.AddDate(0, 0, -days)
	}
	if !windowEnd.After(windowStart) {
		return time.Time{}, time.Time{}, invalid("service.jira", "end_date must be after start_date")
	}
	return windowEnd, windowStart, nil
}

func parseReportDate(value string) (time.Time, bool) {
	if parsed, ok := parseAgingTime(value); ok {
		return parsed, true
	}
	if parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(value), time.UTC); err == nil {
		return parsed, true
	}
	return time.Time{}, false
}

func buildUtilizationJQL(req models.GetUtilizationReportRequest, start, end time.Time) string {
	parts := []string{}
	if req.ProjectKey != "" {
		parts = append(parts, fmt.Sprintf("project = %s", quoteJQLString(req.ProjectKey)))
	}
	parts = append(parts,
		fmt.Sprintf("worklogDate >= %s", quoteJQLString(start.Format("2006-01-02"))),
		fmt.Sprintf("worklogDate <= %s", quoteJQLString(end.Format("2006-01-02"))),
	)
	if req.Assignee != "" {
		parts = append(parts, fmt.Sprintf("worklogAuthor = %s", quoteJQLString(req.Assignee)))
	}
	return strings.Join(parts, " AND ") + " ORDER BY updated DESC"
}

func secondsToHours(seconds int) float64 {
	return float64(seconds) / 3600.0
}
