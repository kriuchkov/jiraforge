package jira

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/go-faster/errors"

	"github.com/kriuchkov/jiraforge/internal/core/models"
)

var defaultBlockingLinkKeywords = []string{"blocked by", "depends on", "is impeded by", "requires"}

func (s *Service) GetBlockedIssuesReport(ctx context.Context, req models.GetBlockedIssuesReportRequest) (*models.BlockedIssuesReport, error) {
	req = req.Normalized()
	if strings.TrimSpace(req.JQL) == "" && strings.TrimSpace(req.ProjectKey) == "" {
		return nil, invalid("service.jira.GetBlockedIssuesReport", "either jql or project_key is required")
	}

	query := strings.TrimSpace(req.JQL)
	if query == "" {
		query = buildBlockedIssuesReportJQL(req)
	}

	searchResult, err := s.gateway.SearchIssues(ctx, models.SearchIssuesRequest{
		JQL:        query,
		Fields:     []string{"summary", "status", "assignee", "updated", "created", "issuetype", "statuscategorychangedate", "issuelinks"},
		MaxResults: req.MaxResults,
	})
	if err != nil {
		return nil, errors.Wrap(err, "search issues")
	}

	report := &models.BlockedIssuesReport{
		Query:           query,
		ProjectKey:      req.ProjectKey,
		Statuses:        append([]string(nil), req.Statuses...),
		BlockedStatuses: append([]string(nil), req.BlockedStatuses...),
		LinkTypes:       append([]string(nil), req.LinkTypes...),
		Assignee:        req.Assignee,
		MinDaysInStatus: req.MinDaysInStatus,
	}
	if searchResult == nil {
		return report, nil
	}

	blockedStatusSet := makeNormalizedSet(req.BlockedStatuses)
	linkTypeSet := makeNormalizedSet(req.LinkTypes)
	now := agingReportNow().UTC()
	report.Summary.AnalyzedIssues = len(searchResult.Issues)
	var totalDays int
	for _, issue := range searchResult.Issues {
		blockedByStatus := isBlockedByStatus(issue, blockedStatusSet)
		blockingLinks := collectBlockingLinks(issue, linkTypeSet)
		if !blockedByStatus && len(blockingLinks) == 0 {
			continue
		}

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

		reasons, hasExternalDependency := blockedIssueReasons(blockedByStatus, blockingLinks)
		item := models.BlockedIssuesReportItem{
			Key:             issue.Key,
			Summary:         issue.Summary,
			Type:            issueTypeLabel(issue.Type),
			Status:          statusLabel(issue.Status),
			Assignee:        personLabel(issue.Assignee),
			Updated:         issue.Updated,
			StatusSince:     statusSince.Format(timeLayoutRFC3339),
			DaysInStatus:    daysInStatus,
			BlockedByStatus: blockedByStatus,
			BlockingLinks:   blockingLinks,
			Reasons:         reasons,
		}
		report.Items = append(report.Items, item)
		totalDays += daysInStatus
		if daysInStatus > report.Summary.OldestDaysInStatus {
			report.Summary.OldestDaysInStatus = daysInStatus
		}
		if blockedByStatus {
			report.Summary.BlockedByStatusIssues++
		}
		if len(blockingLinks) > 0 {
			report.Summary.BlockedByLinkIssues++
		}
		if hasExternalDependency {
			report.Summary.ExternalDependencyIssues++
		}
	}

	sort.SliceStable(report.Items, func(i, j int) bool {
		if report.Items[i].DaysInStatus != report.Items[j].DaysInStatus {
			return report.Items[i].DaysInStatus > report.Items[j].DaysInStatus
		}
		return report.Items[i].Key < report.Items[j].Key
	})

	report.Summary.BlockedIssues = len(report.Items)
	if len(report.Items) > 0 {
		report.Summary.AverageDaysInStatus = float64(totalDays) / float64(len(report.Items))
	}

	return report, nil
}

const timeLayoutRFC3339 = "2006-01-02T15:04:05Z07:00"

func buildBlockedIssuesReportJQL(req models.GetBlockedIssuesReportRequest) string {
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

func isBlockedByStatus(issue models.Issue, blockedStatusSet map[string]struct{}) bool {
	status := normalizedString(statusLabel(issue.Status))
	_, ok := blockedStatusSet[status]
	return ok
}

func collectBlockingLinks(issue models.Issue, linkTypeSet map[string]struct{}) []models.BlockedIssueDependency {
	if len(issue.RelatedIssues) == 0 {
		return nil
	}

	dependencies := make([]models.BlockedIssueDependency, 0, len(issue.RelatedIssues))
	projectKey := issueProjectKey(issue.Key)
	for _, relation := range issue.RelatedIssues {
		if !isBlockingRelation(relation, linkTypeSet) {
			continue
		}
		dependency := models.BlockedIssueDependency{
			Type:      relation.Type,
			Direction: relation.Direction,
			Key:       relation.Issue.Key,
			Summary:   relation.Issue.Summary,
			Status:    issueRefStatusLabel(relation.Issue),
			External:  projectKey != "" && issueProjectKey(relation.Issue.Key) != "" && issueProjectKey(relation.Issue.Key) != projectKey,
		}
		dependencies = append(dependencies, dependency)
	}
	return dependencies
}

func blockedIssueReasons(blockedByStatus bool, dependencies []models.BlockedIssueDependency) ([]string, bool) {
	var reasons []string
	hasExternalDependency := false
	if blockedByStatus {
		reasons = append(reasons, "blocked_status")
	}
	if len(dependencies) > 0 {
		reasons = append(reasons, "blocking_dependency")
		for _, dependency := range dependencies {
			if dependency.External {
				hasExternalDependency = true
				break
			}
		}
	}
	if hasExternalDependency {
		reasons = append(reasons, "external_dependency")
	}
	return reasons, hasExternalDependency
}

func isBlockingRelation(relation models.IssueRelation, linkTypeSet map[string]struct{}) bool {
	normalizedType := normalizedString(relation.Type)
	if len(linkTypeSet) > 0 {
		_, ok := linkTypeSet[normalizedType]
		return ok
	}
	for _, keyword := range defaultBlockingLinkKeywords {
		if strings.Contains(normalizedType, keyword) {
			return true
		}
	}
	return false
}

func issueProjectKey(key string) string {
	if index := strings.IndexByte(strings.TrimSpace(key), '-'); index > 0 {
		return key[:index]
	}
	return ""
}

func issueRefStatusLabel(issue models.IssueRef) string {
	if issue.Status == nil {
		return ""
	}
	return issue.Status.Name
}

func normalizeReportStringSlice(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func normalizedString(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func makeNormalizedSet(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := normalizedString(value)
		if normalized != "" {
			result[normalized] = struct{}{}
		}
	}
	return result
}
