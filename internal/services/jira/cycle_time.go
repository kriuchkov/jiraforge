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

var cycleTimeReportNow = time.Now

func (s *Service) GetCycleTimeReport(ctx context.Context, req models.GetCycleTimeReportRequest) (*models.CycleTimeReport, error) {
	req = req.Normalized()
	if strings.TrimSpace(req.JQL) == "" && strings.TrimSpace(req.ProjectKey) == "" {
		return nil, invalid("service.jira.GetCycleTimeReport", "either jql or project_key is required")
	}

	now := cycleTimeReportNow().UTC()
	windowEnd, windowStart, err := resolveReportWindow(req.StartDate, req.EndDate, req.WindowDays, now)
	if err != nil {
		return nil, err
	}

	query := strings.TrimSpace(req.JQL)
	if query == "" {
		query = buildCycleTimeJQL(req, windowStart, windowEnd)
	}

	searchResult, err := s.gateway.SearchIssues(ctx, models.SearchIssuesRequest{
		JQL:        query,
		Fields:     []string{"summary", "status", "assignee", "created", "issuetype", "resolutiondate"},
		MaxResults: req.MaxResults,
	})
	if err != nil {
		return nil, errors.Wrap(err, "search issues")
	}

	report := &models.CycleTimeReport{
		Query:         query,
		ProjectKey:    req.ProjectKey,
		StartStatuses: append([]string(nil), req.StartStatuses...),
		DoneStatuses:  append([]string(nil), req.DoneStatuses...),
		WindowDays:    daysBetween(windowStart, windowEnd),
		WindowStart:   windowStart.Format(time.RFC3339),
		WindowEnd:     windowEnd.Format(time.RFC3339),
	}
	if searchResult == nil {
		return report, nil
	}
	report.Summary.AnalyzedIssues = len(searchResult.Issues)

	startSet := makeNormalizedStringSet(req.StartStatuses)
	doneSet := makeNormalizedStringSet(req.DoneStatuses)

	type accBucket struct {
		count    int
		cycleHrs []float64
		leadHrs  []float64
	}
	byAssignee := map[string]*accBucket{}
	byType := map[string]*accBucket{}

	for _, issue := range searchResult.Issues {
		history, err := s.gateway.GetIssueHistory(ctx, models.GetIssueHistoryRequest{IssueKey: issue.Key})
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("get issue history for %s", issue.Key))
		}
		startedAt, completedAt := findCycleTimes(history, startSet, doneSet)
		if completedAt.IsZero() {
			continue
		}
		if completedAt.Before(windowStart) || completedAt.After(windowEnd) {
			continue
		}

		item := models.CycleTimeReportItem{
			Key:         issue.Key,
			Summary:     issue.Summary,
			Type:        issueTypeLabel(issue.Type),
			Status:      statusLabel(issue.Status),
			Assignee:    flowEfficiencyAssigneeLabel(personLabel(issue.Assignee)),
			Created:     issue.Created,
			CompletedAt: completedAt.Format(time.RFC3339),
		}
		if !startedAt.IsZero() {
			item.StartedAt = startedAt.Format(time.RFC3339)
			item.CycleHours = roundHours(completedAt.Sub(startedAt))
		}
		if created, ok := parseAgingTime(issue.Created); ok {
			item.LeadHours = roundHours(completedAt.Sub(created))
		}
		report.Items = append(report.Items, item)

		bucket := byAssignee[item.Assignee]
		if bucket == nil {
			bucket = &accBucket{}
			byAssignee[item.Assignee] = bucket
		}
		bucket.count++
		if item.CycleHours > 0 {
			bucket.cycleHrs = append(bucket.cycleHrs, item.CycleHours)
		}
		if item.LeadHours > 0 {
			bucket.leadHrs = append(bucket.leadHrs, item.LeadHours)
		}

		typeKey := item.Type
		if typeKey == "" {
			typeKey = "Unknown"
		}
		typeBucket := byType[typeKey]
		if typeBucket == nil {
			typeBucket = &accBucket{}
			byType[typeKey] = typeBucket
		}
		typeBucket.count++
		if item.CycleHours > 0 {
			typeBucket.cycleHrs = append(typeBucket.cycleHrs, item.CycleHours)
		}
	}

	sort.SliceStable(report.Items, func(i, j int) bool {
		if report.Items[i].CompletedAt != report.Items[j].CompletedAt {
			return report.Items[i].CompletedAt > report.Items[j].CompletedAt
		}
		return report.Items[i].Key < report.Items[j].Key
	})

	allCycle := make([]float64, 0)
	allLead := make([]float64, 0)
	for _, item := range report.Items {
		if item.CycleHours > 0 {
			allCycle = append(allCycle, item.CycleHours)
		}
		if item.LeadHours > 0 {
			allLead = append(allLead, item.LeadHours)
		}
	}

	report.Summary.CompletedIssues = len(report.Items)
	report.Summary.AverageCycleHours = roundHoursValue(meanFloat(allCycle))
	report.Summary.MedianCycleHours = roundHoursValue(medianFloat(allCycle))
	report.Summary.P85CycleHours = roundHoursValue(percentileFloat(allCycle, 0.85))
	report.Summary.AverageLeadHours = roundHoursValue(meanFloat(allLead))
	report.Summary.ThroughputTotal = report.Summary.CompletedIssues

	if report.WindowDays > 0 {
		weeks := float64(report.WindowDays) / 7.0
		if weeks > 0 {
			report.Summary.ThroughputPerWeek = roundHoursValue(float64(report.Summary.CompletedIssues) / weeks)
		}
	}

	for assignee, bucket := range byAssignee {
		report.Assignees = append(report.Assignees, models.CycleTimeAssigneeSummary{
			Assignee:        assignee,
			CompletedIssues: bucket.count,
			AverageCycleHrs: roundHoursValue(meanFloat(bucket.cycleHrs)),
			MedianCycleHrs:  roundHoursValue(medianFloat(bucket.cycleHrs)),
			AverageLeadHrs:  roundHoursValue(meanFloat(bucket.leadHrs)),
		})
	}
	sort.SliceStable(report.Assignees, func(i, j int) bool {
		if report.Assignees[i].CompletedIssues != report.Assignees[j].CompletedIssues {
			return report.Assignees[i].CompletedIssues > report.Assignees[j].CompletedIssues
		}
		return report.Assignees[i].Assignee < report.Assignees[j].Assignee
	})

	for issueType, bucket := range byType {
		report.Types = append(report.Types, models.CycleTimeTypeBreakdown{
			Type:            issueType,
			CompletedIssues: bucket.count,
			AverageCycleHrs: roundHoursValue(meanFloat(bucket.cycleHrs)),
			MedianCycleHrs:  roundHoursValue(medianFloat(bucket.cycleHrs)),
		})
	}
	sort.SliceStable(report.Types, func(i, j int) bool {
		if report.Types[i].CompletedIssues != report.Types[j].CompletedIssues {
			return report.Types[i].CompletedIssues > report.Types[j].CompletedIssues
		}
		return report.Types[i].Type < report.Types[j].Type
	})

	return report, nil
}

func buildCycleTimeJQL(req models.GetCycleTimeReportRequest, start, end time.Time) string {
	parts := []string{}
	if req.ProjectKey != "" {
		parts = append(parts, fmt.Sprintf("project = %s", quoteJQLString(req.ProjectKey)))
	}
	if len(req.DoneStatuses) > 0 {
		quoted := make([]string, 0, len(req.DoneStatuses))
		for _, status := range req.DoneStatuses {
			quoted = append(quoted, quoteJQLString(status))
		}
		parts = append(parts, fmt.Sprintf("status IN (%s)", strings.Join(quoted, ", ")))
	}
	parts = append(parts,
		fmt.Sprintf("statusCategoryChangedDate >= %s", quoteJQLString(start.Format("2006-01-02"))),
		fmt.Sprintf("statusCategoryChangedDate <= %s", quoteJQLString(end.Format("2006-01-02"))),
	)
	if req.Assignee != "" {
		parts = append(parts, fmt.Sprintf("assignee = %s", quoteJQLString(req.Assignee)))
	}
	if req.IssueType != "" {
		parts = append(parts, fmt.Sprintf("issuetype = %s", quoteJQLString(req.IssueType)))
	}
	return strings.Join(parts, " AND ") + " ORDER BY statusCategoryChangedDate DESC"
}

func findCycleTimes(history *models.IssueHistory, startSet, doneSet map[string]struct{}) (started time.Time, done time.Time) {
	if history == nil {
		return time.Time{}, time.Time{}
	}
	for _, entry := range history.Entries {
		when, ok := parseAgingTime(entry.Date)
		if !ok {
			continue
		}
		for _, change := range entry.Changes {
			if !strings.EqualFold(strings.TrimSpace(change.Field), "status") {
				continue
			}
			to := strings.ToLower(strings.TrimSpace(change.To))
			if _, ok := startSet[to]; ok && started.IsZero() {
				started = when
			}
			if _, ok := doneSet[to]; ok {
				done = when
			}
		}
	}
	return started, done
}

func makeNormalizedStringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.ToLower(strings.TrimSpace(value))
		if trimmed == "" {
			continue
		}
		set[trimmed] = struct{}{}
	}
	return set
}

func meanFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func medianFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func percentileFloat(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	idx := int(float64(len(sorted)-1) * p)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func roundHours(d time.Duration) float64 {
	return roundHoursValue(d.Hours())
}

func roundHoursValue(value float64) float64 {
	return float64(int64(value*100+0.5)) / 100
}
