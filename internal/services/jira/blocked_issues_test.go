package jira

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kriuchkov/jiraforge/internal/core/models"
)

func TestGetBlockedIssuesReportBuildsBlockedSummary(t *testing.T) {
	previousNow := agingReportNow
	agingReportNow = func() time.Time {
		return time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC)
	}
	defer func() { agingReportNow = previousNow }()

	gateway := &stubGateway{
		searchIssuesResult: &models.SearchIssuesResult{Issues: []models.Issue{
			{Key: "PROJ-1", Summary: "Explicitly blocked", Status: &models.Status{Name: "Blocked"}, Assignee: &models.Person{DisplayName: "Alice"}, Type: &models.IssueType{Name: "Bug"}, Updated: "2026-04-09T12:00:00Z", Created: "2026-04-01T12:00:00Z"},
			{Key: "PROJ-2", Summary: "Waiting on another team", Status: &models.Status{Name: "In Progress"}, Assignee: &models.Person{DisplayName: "Bob"}, Type: &models.IssueType{Name: "Story"}, Updated: "2026-04-09T12:00:00Z", Created: "2026-04-05T12:00:00Z", RelatedIssues: []models.IssueRelation{{Type: "is blocked by", Direction: "inward", Issue: models.IssueRef{Key: "OPS-10", Summary: "Provision environment", Status: &models.Status{Name: "In Progress"}}}}},
			{Key: "PROJ-3", Summary: "Healthy item", Status: &models.Status{Name: "In Progress"}, Assignee: &models.Person{DisplayName: "Carol"}, Type: &models.IssueType{Name: "Task"}, Updated: "2026-04-10T11:00:00Z", Created: "2026-04-09T12:00:00Z"},
		}},
		issueHistoryResults: map[string]*models.IssueHistory{
			"PROJ-1": {IssueKey: "PROJ-1", Entries: []models.HistoryEntry{{Date: "2026-04-03 12:00:00", Changes: []models.HistoryChange{{Field: "status", From: "In Progress", To: "Blocked"}}}}},
			"PROJ-2": {IssueKey: "PROJ-2", Entries: []models.HistoryEntry{{Date: "2026-04-05 12:00:00", Changes: []models.HistoryChange{{Field: "status", From: "To Do", To: "In Progress"}}}}},
			"PROJ-3": {IssueKey: "PROJ-3", Entries: []models.HistoryEntry{{Date: "2026-04-09 12:00:00", Changes: []models.HistoryChange{{Field: "status", From: "To Do", To: "In Progress"}}}}},
		},
	}

	service := &Service{gateway: gateway, attachmentStore: &stubAttachmentStore{}}

	report, err := service.GetBlockedIssuesReport(context.Background(), models.GetBlockedIssuesReportRequest{ProjectKey: "PROJ", Statuses: []string{"In Progress", "Blocked"}, MinDaysInStatus: 2, MaxResults: 50})
	require.NoError(t, err)
	require.NotNil(t, report)

	require.Equal(t, `project = "PROJ" AND resolution = Unresolved AND status IN ("In Progress", "Blocked") ORDER BY updated ASC`, report.Query)
	require.Equal(t, []string{"Blocked"}, report.BlockedStatuses)
	require.Equal(t, 3, report.Summary.AnalyzedIssues)
	require.Equal(t, 2, report.Summary.BlockedIssues)
	require.Equal(t, 1, report.Summary.BlockedByStatusIssues)
	require.Equal(t, 1, report.Summary.BlockedByLinkIssues)
	require.Equal(t, 1, report.Summary.ExternalDependencyIssues)
	require.Equal(t, 7, report.Summary.OldestDaysInStatus)
	require.InDelta(t, 6.0, report.Summary.AverageDaysInStatus, 0.0001)
	require.Len(t, report.Items, 2)
	require.Equal(t, "PROJ-1", report.Items[0].Key)
	require.True(t, report.Items[0].BlockedByStatus)
	require.Contains(t, report.Items[0].Reasons, "blocked_status")
	require.Equal(t, "PROJ-2", report.Items[1].Key)
	require.Len(t, report.Items[1].BlockingLinks, 1)
	require.Equal(t, "OPS-10", report.Items[1].BlockingLinks[0].Key)
	require.True(t, report.Items[1].BlockingLinks[0].External)
	require.Contains(t, report.Items[1].Reasons, "blocking_dependency")
	require.Contains(t, report.Items[1].Reasons, "external_dependency")
	require.Equal(t, []models.GetIssueHistoryRequest{{IssueKey: "PROJ-1"}, {IssueKey: "PROJ-2"}}, gateway.getIssueHistoryReqs)
	require.Equal(t, []string{"summary", "status", "assignee", "updated", "created", "issuetype", "statuscategorychangedate", "issuelinks"}, gateway.searchIssuesReq.Fields)
}
