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

type statusChangeEvent struct {
	at   time.Time
	from string
	to   string
}

type flowEfficiencyAssigneeAccumulator struct {
	issues                 int
	issuesWithBlockedTime  int
	currentlyBlockedIssues int
	totalObservedDuration  time.Duration
	totalActiveDuration    time.Duration
	totalBlockedDuration   time.Duration
	totalEfficiency        float64
}

func (s *Service) GetFlowEfficiencyReport(ctx context.Context, req models.GetFlowEfficiencyReportRequest) (*models.FlowEfficiencyReport, error) {
	req = req.Normalized()
	if strings.TrimSpace(req.JQL) == "" && strings.TrimSpace(req.ProjectKey) == "" {
		return nil, invalid("service.jira.GetFlowEfficiencyReport", "either jql or project_key is required")
	}
	if req.WindowDays > 0 && (req.StartDate != "" || req.EndDate != "") {
		return nil, invalid("service.jira.GetFlowEfficiencyReport", "window_days cannot be combined with start_date or end_date")
	}

	query := strings.TrimSpace(req.JQL)
	if query == "" {
		query = buildFlowEfficiencyReportJQL(req)
	}

	searchResult, err := s.gateway.SearchIssues(ctx, models.SearchIssuesRequest{
		JQL:        query,
		Fields:     []string{"summary", "status", "assignee", "updated", "created", "issuetype", "statuscategorychangedate"},
		MaxResults: req.MaxResults,
	})
	if err != nil {
		return nil, errors.Wrap(err, "search issues")
	}

	report := &models.FlowEfficiencyReport{
		Query:           query,
		ProjectKey:      req.ProjectKey,
		Statuses:        append([]string(nil), req.Statuses...),
		ActiveStatuses:  append([]string(nil), req.ActiveStatuses...),
		BlockedStatuses: append([]string(nil), req.BlockedStatuses...),
		Assignee:        req.Assignee,
		MinBlockedDays:  req.MinBlockedDays,
	}
	if searchResult == nil {
		return report, nil
	}

	now := agingReportNow().UTC()
	blockedSet := makeNormalizedSet(req.BlockedStatuses)
	activeSet := makeNormalizedSet(req.ActiveStatuses)
	windowStart, windowEnd, err := resolveFlowEfficiencyWindow(req, now)
	if err != nil {
		return nil, invalid("service.jira.GetFlowEfficiencyReport", err.Error())
	}
	if !windowStart.IsZero() {
		report.WindowStart = windowStart.Format(time.RFC3339)
		report.WindowEnd = windowEnd.Format(time.RFC3339)
	}
	if req.WindowDays > 0 {
		report.WindowDays = req.WindowDays
	}
	report.Summary.AnalyzedIssues = len(searchResult.Issues)
	var totalEfficiency float64
	var totalObservedDuration time.Duration
	var totalActiveDuration time.Duration
	var totalBlockedDuration time.Duration
	assigneeAccumulators := make(map[string]*flowEfficiencyAssigneeAccumulator)

	for _, issue := range searchResult.Issues {
		history, err := s.gateway.GetIssueHistory(ctx, models.GetIssueHistoryRequest{IssueKey: issue.Key})
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("get issue history for %s", issue.Key))
		}

		item, observedDuration, activeDuration, blockedDuration, ok := buildFlowEfficiencyItem(issue, history, activeSet, blockedSet, windowStart, windowEnd)
		if !ok {
			continue
		}
		if item.BlockedDays < req.MinBlockedDays {
			continue
		}

		report.Items = append(report.Items, item)
		totalEfficiency += item.FlowEfficiency
		totalObservedDuration += observedDuration
		totalActiveDuration += activeDuration
		totalBlockedDuration += blockedDuration
		if item.BlockedDays > 0 {
			report.Summary.IssuesWithBlockedTime++
		}
		if item.CurrentlyBlocked {
			report.Summary.CurrentlyBlockedIssues++
		}

		assignee := flowEfficiencyAssigneeLabel(item.Assignee)
		accumulator := assigneeAccumulators[assignee]
		if accumulator == nil {
			accumulator = &flowEfficiencyAssigneeAccumulator{}
			assigneeAccumulators[assignee] = accumulator
		}
		accumulator.issues++
		accumulator.totalEfficiency += item.FlowEfficiency
		accumulator.totalObservedDuration += observedDuration
		accumulator.totalActiveDuration += activeDuration
		accumulator.totalBlockedDuration += blockedDuration
		if item.BlockedDays > 0 {
			accumulator.issuesWithBlockedTime++
		}
		if item.CurrentlyBlocked {
			accumulator.currentlyBlockedIssues++
		}
	}

	sort.SliceStable(report.Items, func(i, j int) bool {
		if report.Items[i].FlowEfficiency != report.Items[j].FlowEfficiency {
			return report.Items[i].FlowEfficiency < report.Items[j].FlowEfficiency
		}
		if report.Items[i].BlockedDays != report.Items[j].BlockedDays {
			return report.Items[i].BlockedDays > report.Items[j].BlockedDays
		}
		return report.Items[i].Key < report.Items[j].Key
	})

	report.Summary.MatchingIssues = len(report.Items)
	report.Summary.TotalObservedDays = durationToDays(totalObservedDuration)
	report.Summary.TotalActiveDays = durationToDays(totalActiveDuration)
	report.Summary.TotalBlockedDays = durationToDays(totalBlockedDuration)
	if len(report.Items) > 0 {
		report.Summary.AverageFlowEfficiency = totalEfficiency / float64(len(report.Items))
	}
	if totalObservedDuration > 0 {
		report.Summary.PortfolioFlowEfficiency = activeDurationRatio(totalActiveDuration, totalObservedDuration)
	}

	for assignee, accumulator := range assigneeAccumulators {
		summary := models.FlowEfficiencyAssigneeSummary{
			Assignee:               assignee,
			Issues:                 accumulator.issues,
			IssuesWithBlockedTime:  accumulator.issuesWithBlockedTime,
			CurrentlyBlockedIssues: accumulator.currentlyBlockedIssues,
			TotalObservedDays:      durationToDays(accumulator.totalObservedDuration),
			TotalActiveDays:        durationToDays(accumulator.totalActiveDuration),
			TotalBlockedDays:       durationToDays(accumulator.totalBlockedDuration),
		}
		if accumulator.issues > 0 {
			summary.AverageFlowEfficiency = accumulator.totalEfficiency / float64(accumulator.issues)
		}
		if accumulator.totalObservedDuration > 0 {
			summary.PortfolioFlowEfficiency = activeDurationRatio(accumulator.totalActiveDuration, accumulator.totalObservedDuration)
		}
		report.Assignees = append(report.Assignees, summary)
	}

	sort.SliceStable(report.Assignees, func(i, j int) bool {
		if report.Assignees[i].PortfolioFlowEfficiency != report.Assignees[j].PortfolioFlowEfficiency {
			return report.Assignees[i].PortfolioFlowEfficiency < report.Assignees[j].PortfolioFlowEfficiency
		}
		if report.Assignees[i].TotalBlockedDays != report.Assignees[j].TotalBlockedDays {
			return report.Assignees[i].TotalBlockedDays > report.Assignees[j].TotalBlockedDays
		}
		return report.Assignees[i].Assignee < report.Assignees[j].Assignee
	})

	return report, nil
}

func resolveFlowEfficiencyWindow(req models.GetFlowEfficiencyReportRequest, now time.Time) (time.Time, time.Time, error) {
	if req.WindowDays > 0 {
		start := now.Add(-time.Duration(req.WindowDays) * 24 * time.Hour)
		return start, now, nil
	}

	if req.EndDate != "" && req.StartDate == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("start_date is required when end_date is provided")
	}

	start, hasStart, err := parseFlowEfficiencyBoundary(req.StartDate, false)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date: %w", err)
	}
	end, hasEnd, err := parseFlowEfficiencyBoundary(req.EndDate, true)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date: %w", err)
	}
	if !hasStart {
		return time.Time{}, now, nil
	}
	if !hasEnd {
		end = now
	}
	if !end.After(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("end_date must be after start_date")
	}
	return start, end, nil
}

func buildFlowEfficiencyReportJQL(req models.GetFlowEfficiencyReportRequest) string {
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

func buildFlowEfficiencyItem(issue models.Issue, history *models.IssueHistory, activeSet, blockedSet map[string]struct{}, windowStart, windowEnd time.Time) (models.FlowEfficiencyReportItem, time.Duration, time.Duration, time.Duration, bool) {
	events := statusChangeEvents(history)
	currentStatusSince := latestStatusChangeTime(history)
	if currentStatusSince.IsZero() {
		currentStatusSince = fallbackStatusSince(issue)
	}

	observedDuration, activeDuration, blockedDuration, blockedTransitions := computeStatusDurations(issue, events, activeSet, blockedSet, windowStart, windowEnd)
	if observedDuration <= 0 {
		return models.FlowEfficiencyReportItem{}, 0, 0, 0, false
	}

	if currentStatusSince.IsZero() {
		currentStatusSince = windowEnd
	}

	flowEfficiency := activeDurationRatio(activeDuration, observedDuration)
	item := models.FlowEfficiencyReportItem{
		Key:                issue.Key,
		Summary:            issue.Summary,
		Type:               issueTypeLabel(issue.Type),
		Status:             statusLabel(issue.Status),
		Assignee:           personLabel(issue.Assignee),
		Updated:            issue.Updated,
		CurrentStatusSince: currentStatusSince.Format(time.RFC3339),
		CurrentStatusDays:  daysBetween(currentStatusSince, windowEnd),
		ObservedDays:       durationToDays(observedDuration),
		ActiveDays:         durationToDays(activeDuration),
		BlockedDays:        durationToDays(blockedDuration),
		FlowEfficiency:     flowEfficiency,
		BlockedTransitions: blockedTransitions,
		CurrentlyBlocked:   isBlockedStatusName(statusLabel(issue.Status), blockedSet),
	}
	return item, observedDuration, activeDuration, blockedDuration, true
}

func statusChangeEvents(history *models.IssueHistory) []statusChangeEvent {
	if history == nil {
		return nil
	}

	result := make([]statusChangeEvent, 0, len(history.Entries))
	for _, entry := range history.Entries {
		parsed, ok := parseAgingTime(entry.Date)
		if !ok {
			continue
		}
		for _, change := range entry.Changes {
			if !strings.EqualFold(strings.TrimSpace(change.Field), "status") {
				continue
			}
			result = append(result, statusChangeEvent{
				at:   parsed,
				from: strings.TrimSpace(change.From),
				to:   strings.TrimSpace(change.To),
			})
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].at.Before(result[j].at)
	})
	return result
}

func computeStatusDurations(issue models.Issue, events []statusChangeEvent, activeSet, blockedSet map[string]struct{}, windowStart, windowEnd time.Time) (time.Duration, time.Duration, time.Duration, int) {
	createdAt, _ := parseAgingTime(issue.Created)
	currentStatus := ""
	cursorTime := time.Time{}
	if len(events) > 0 {
		if !createdAt.IsZero() && events[0].from != "" {
			currentStatus = events[0].from
			cursorTime = createdAt
		} else if events[0].from != "" {
			currentStatus = events[0].from
			cursorTime = events[0].at
		}
	} else {
		currentStatus = statusLabel(issue.Status)
		if !createdAt.IsZero() {
			cursorTime = createdAt
		} else {
			cursorTime = fallbackStatusSince(issue)
		}
	}

	var observedDuration time.Duration
	var activeDuration time.Duration
	var blockedDuration time.Duration
	blockedTransitions := 0
	for _, event := range events {
		interval := clippedDuration(cursorTime, event.at, windowStart, windowEnd)
		if interval > 0 {
			observedDuration += classifyStatusDuration(currentStatus, interval, activeSet, blockedSet, &activeDuration, &blockedDuration)
		}
		if isBlockedStatusName(event.to, blockedSet) && isWithinFlowEfficiencyWindow(event.at, windowStart, windowEnd) {
			blockedTransitions++
		}
		if event.to != "" {
			currentStatus = event.to
		}
		cursorTime = event.at
	}

	if currentStatus == "" {
		currentStatus = statusLabel(issue.Status)
	}
	if cursorTime.IsZero() {
		cursorTime = fallbackStatusSince(issue)
	}
	interval := clippedDuration(cursorTime, windowEnd, windowStart, windowEnd)
	if interval > 0 {
		observedDuration += classifyStatusDuration(currentStatus, interval, activeSet, blockedSet, &activeDuration, &blockedDuration)
	}

	return observedDuration, activeDuration, blockedDuration, blockedTransitions
}

func classifyStatusDuration(status string, interval time.Duration, activeSet, blockedSet map[string]struct{}, activeDuration, blockedDuration *time.Duration) time.Duration {
	if interval <= 0 {
		return 0
	}
	if isBlockedStatusName(status, blockedSet) {
		*blockedDuration += interval
		return interval
	}
	if isActiveStatusName(status, activeSet, blockedSet) {
		*activeDuration += interval
		return interval
	}
	return 0
}

func isBlockedStatusName(status string, blockedSet map[string]struct{}) bool {
	_, ok := blockedSet[normalizedString(status)]
	return ok
}

func isActiveStatusName(status string, activeSet, blockedSet map[string]struct{}) bool {
	normalized := normalizedString(status)
	if normalized == "" {
		return false
	}
	if _, blocked := blockedSet[normalized]; blocked {
		return false
	}
	if len(activeSet) == 0 {
		return true
	}
	_, ok := activeSet[normalized]
	return ok
}

func durationToDays(value time.Duration) int {
	if value <= 0 {
		return 0
	}
	return int(value.Hours() / 24)
}

func activeDurationRatio(activeDuration, observedDuration time.Duration) float64 {
	if activeDuration <= 0 || observedDuration <= 0 {
		return 0
	}
	return float64(activeDuration) / float64(observedDuration)
}

func parseFlowEfficiencyBoundary(value string, isEnd bool) (time.Time, bool, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, false, nil
	}
	if parsed, err := time.ParseInLocation("2006-01-02", trimmed, time.UTC); err == nil {
		if isEnd {
			return parsed.Add(24 * time.Hour), true, nil
		}
		return parsed, true, nil
	}
	parsed, ok := parseAgingTime(trimmed)
	if !ok {
		return time.Time{}, false, fmt.Errorf("unsupported date format %q", trimmed)
	}
	return parsed, true, nil
}

func isWithinFlowEfficiencyWindow(at, windowStart, windowEnd time.Time) bool {
	if at.IsZero() {
		return false
	}
	if !windowStart.IsZero() && at.Before(windowStart) {
		return false
	}
	if !windowEnd.IsZero() && !at.Before(windowEnd) {
		return false
	}
	return true
}

func clippedDuration(start, end, windowStart, windowEnd time.Time) time.Duration {
	if start.IsZero() || !end.After(start) {
		return 0
	}
	if !windowStart.IsZero() && start.Before(windowStart) {
		start = windowStart
	}
	if end.After(windowEnd) {
		end = windowEnd
	}
	if !end.After(start) {
		return 0
	}
	return end.Sub(start)
}

func flowEfficiencyAssigneeLabel(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "Unassigned"
	}
	return trimmed
}
