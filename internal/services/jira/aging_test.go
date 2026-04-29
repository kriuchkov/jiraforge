package jira

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kriuchkov/jiraforge/internal/core/models"
)

func TestGetAgingReportBuildsAgingSummary(t *testing.T) {
	previousNow := agingReportNow
	agingReportNow = func() time.Time {
		return time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC)
	}
	defer func() { agingReportNow = previousNow }()

	gateway := &stubGateway{
		searchIssuesResult: &models.SearchIssuesResult{Issues: []models.Issue{
			{Key: "PROJ-1", Summary: "Stuck in progress", Status: &models.Status{Name: "In Progress"}, Assignee: &models.Person{DisplayName: "Alice"}, Type: &models.IssueType{Name: "Bug"}, Updated: "2026-04-09T12:00:00Z", Created: "2026-04-01T12:00:00Z"},
			{Key: "PROJ-2", Summary: "Needs review", Status: &models.Status{Name: "Code Review"}, Assignee: &models.Person{DisplayName: "Bob"}, Type: &models.IssueType{Name: "Story"}, Updated: "2026-04-09T12:00:00Z", Created: "2026-04-08T12:00:00Z"},
			{Key: "PROJ-3", Summary: "Fresh item", Status: &models.Status{Name: "In Progress"}, Assignee: &models.Person{DisplayName: "Carol"}, Type: &models.IssueType{Name: "Task"}, Updated: "2026-04-10T11:00:00Z", Created: "2026-04-09T12:00:00Z"},
		}},
		issueHistoryResults: map[string]*models.IssueHistory{
			"PROJ-1": {IssueKey: "PROJ-1", Entries: []models.HistoryEntry{{Date: "2026-04-04 12:00:00", Changes: []models.HistoryChange{{Field: "status", From: "To Do", To: "In Progress"}}}}},
			"PROJ-2": {IssueKey: "PROJ-2"},
			"PROJ-3": {IssueKey: "PROJ-3", Entries: []models.HistoryEntry{{Date: "2026-04-09 12:00:00", Changes: []models.HistoryChange{{Field: "status", From: "To Do", To: "In Progress"}}}}},
		},
	}

	service := &Service{
		gateway:         gateway,
		attachmentStore: &stubAttachmentStore{},
	}

	report, err := service.GetAgingReport(context.Background(), models.GetAgingReportRequest{ProjectKey: "PROJ", Statuses: []string{"In Progress", "Code Review"}, MinDaysInStatus: 2, MaxResults: 50})
	require.NoError(t, err)
	require.NotNil(t, report)

	require.Equal(t, `project = "PROJ" AND resolution = Unresolved AND status IN ("In Progress", "Code Review") ORDER BY updated ASC`, report.Query)
	require.Equal(t, []string{"In Progress", "Code Review"}, report.Statuses)
	require.Equal(t, 2, report.MinDaysInStatus)
	require.Equal(t, 3, report.Summary.AnalyzedIssues)
	require.Equal(t, 2, report.Summary.MatchingIssues)
	require.Equal(t, 6, report.Summary.OldestDaysInStatus)
	require.InDelta(t, 4.0, report.Summary.AverageDaysInStatus, 0.0001)
	require.Len(t, report.Items, 2)
	require.Equal(t, "PROJ-1", report.Items[0].Key)
	require.Equal(t, 6, report.Items[0].DaysInStatus)
	require.Equal(t, "Alice", report.Items[0].Assignee)
	require.Equal(t, "PROJ-2", report.Items[1].Key)
	require.Equal(t, 2, report.Items[1].DaysInStatus)
	require.Equal(t, []models.GetIssueHistoryRequest{{IssueKey: "PROJ-1"}, {IssueKey: "PROJ-2"}, {IssueKey: "PROJ-3"}}, gateway.getIssueHistoryReqs)
	require.Equal(t, []string{"summary", "status", "assignee", "updated", "created", "issuetype", "statuscategorychangedate"}, gateway.searchIssuesReq.Fields)
	require.Equal(t, 50, gateway.searchIssuesReq.MaxResults)
}
