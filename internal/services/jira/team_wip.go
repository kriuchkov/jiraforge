package jira

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/go-faster/errors"

	"github.com/kriuchkov/jiraforge/internal/core/models"
)

func (s *Service) GetTeamWipSnapshot(ctx context.Context, req models.GetTeamWipSnapshotRequest) (*models.TeamWipSnapshot, error) {
	req = req.Normalized()
	if strings.TrimSpace(req.JQL) == "" && strings.TrimSpace(req.ProjectKey) == "" {
		return nil, invalid("service.jira.GetTeamWipSnapshot", "either jql or project_key is required")
	}

	query := strings.TrimSpace(req.JQL)
	if query == "" {
		query = buildTeamWipJQL(req)
	}

	searchResult, err := s.gateway.SearchIssues(ctx, models.SearchIssuesRequest{
		JQL:        query,
		Fields:     []string{"summary", "status", "assignee", "issuetype", "priority", "updated"},
		MaxResults: req.MaxResults,
	})
	if err != nil {
		return nil, errors.Wrap(err, "search issues")
	}

	snapshot := &models.TeamWipSnapshot{
		Query:      query,
		ProjectKey: req.ProjectKey,
		Statuses:   append([]string(nil), req.Statuses...),
		WipLimit:   req.WipLimit,
	}
	if searchResult == nil {
		return snapshot, nil
	}
	snapshot.Summary.AnalyzedIssues = len(searchResult.Issues)

	assigneeFilter := makeNormalizedStringSet(req.Assignees)
	statusTotals := map[string]int{}

	type statusBucket struct {
		count int
		items []models.TeamWipIssue
	}
	type assigneeBucket struct {
		assignee string
		statuses map[string]*statusBucket
		total    int
	}
	byAssignee := map[string]*assigneeBucket{}

	for _, issue := range searchResult.Issues {
		assignee := flowEfficiencyAssigneeLabel(personLabel(issue.Assignee))
		if len(assigneeFilter) > 0 {
			if _, ok := assigneeFilter[strings.ToLower(assignee)]; !ok {
				continue
			}
		}

		status := statusLabel(issue.Status)
		if status == "" {
			status = "Unknown"
		}
		statusTotals[status]++

		bucket, ok := byAssignee[assignee]
		if !ok {
			bucket = &assigneeBucket{assignee: assignee, statuses: map[string]*statusBucket{}}
			byAssignee[assignee] = bucket
		}
		bucket.total++
		statusEntry, ok := bucket.statuses[status]
		if !ok {
			statusEntry = &statusBucket{}
			bucket.statuses[status] = statusEntry
		}
		statusEntry.count++
		statusEntry.items = append(statusEntry.items, models.TeamWipIssue{
			Key:      issue.Key,
			Summary:  issue.Summary,
			Status:   status,
			Type:     issueTypeLabel(issue.Type),
			Priority: issue.Priority,
			Updated:  issue.Updated,
		})

		if assignee == "Unassigned" {
			snapshot.Summary.UnassignedIssues++
		}
	}

	for assignee, bucket := range byAssignee {
		summary := models.TeamWipAssignee{
			Assignee:  assignee,
			TotalWip:  bucket.total,
			OverLimit: req.WipLimit > 0 && bucket.total > req.WipLimit,
		}
		for status, entry := range bucket.statuses {
			sort.SliceStable(entry.items, func(i, j int) bool { return entry.items[i].Key < entry.items[j].Key })
			summary.Statuses = append(summary.Statuses, models.TeamWipStatusBreakdown{
				Status: status,
				Issues: entry.count,
				Items:  entry.items,
			})
		}
		sort.SliceStable(summary.Statuses, func(i, j int) bool {
			if summary.Statuses[i].Issues != summary.Statuses[j].Issues {
				return summary.Statuses[i].Issues > summary.Statuses[j].Issues
			}
			return summary.Statuses[i].Status < summary.Statuses[j].Status
		})
		_ = assignee
		snapshot.Assignees = append(snapshot.Assignees, summary)
	}

	sort.SliceStable(snapshot.Assignees, func(i, j int) bool {
		if snapshot.Assignees[i].TotalWip != snapshot.Assignees[j].TotalWip {
			return snapshot.Assignees[i].TotalWip > snapshot.Assignees[j].TotalWip
		}
		return snapshot.Assignees[i].Assignee < snapshot.Assignees[j].Assignee
	})

	snapshot.Summary.DistinctAssignees = 0
	for assignee := range byAssignee {
		if assignee != "Unassigned" {
			snapshot.Summary.DistinctAssignees++
		}
	}
	if len(statusTotals) > 0 {
		snapshot.Summary.StatusTotals = statusTotals
	}

	return snapshot, nil
}

func buildTeamWipJQL(req models.GetTeamWipSnapshotRequest) string {
	parts := []string{}
	if req.ProjectKey != "" {
		parts = append(parts, fmt.Sprintf("project = %s", quoteJQLString(req.ProjectKey)))
	}
	if len(req.Statuses) > 0 {
		quoted := make([]string, 0, len(req.Statuses))
		for _, status := range req.Statuses {
			quoted = append(quoted, quoteJQLString(status))
		}
		parts = append(parts, fmt.Sprintf("status IN (%s)", strings.Join(quoted, ", ")))
	} else {
		parts = append(parts, "statusCategory = \"In Progress\"")
	}
	if len(req.Assignees) > 0 {
		quoted := make([]string, 0, len(req.Assignees))
		for _, assignee := range req.Assignees {
			quoted = append(quoted, quoteJQLString(assignee))
		}
		parts = append(parts, fmt.Sprintf("assignee IN (%s)", strings.Join(quoted, ", ")))
	}
	return strings.Join(parts, " AND ") + " ORDER BY updated DESC"
}
