package jira

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kriuchkov/jiraforge/internal/core/models"
)

func TestGetFlowEfficiencyReportBuildsMetricsFromHistory(t *testing.T) {
	previousNow := agingReportNow
	agingReportNow = func() time.Time {
		return time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC)
	}
	defer func() { agingReportNow = previousNow }()

	gateway := &stubGateway{
		searchIssuesResult: &models.SearchIssuesResult{Issues: []models.Issue{
			{Key: "PROJ-1", Summary: "Feature review", Status: &models.Status{Name: "Code Review"}, Assignee: &models.Person{DisplayName: "Alice"}, Type: &models.IssueType{Name: "Story"}, Updated: "2026-04-10T10:00:00Z", Created: "2026-04-01T12:00:00Z"},
			{Key: "PROJ-2", Summary: "Waiting on ops", Status: &models.Status{Name: "Blocked"}, Assignee: &models.Person{DisplayName: "Bob"}, Type: &models.IssueType{Name: "Task"}, Updated: "2026-04-10T11:00:00Z", Created: "2026-04-05T12:00:00Z"},
			{Key: "PROJ-3", Summary: "Queued item", Status: &models.Status{Name: "To Do"}, Assignee: &models.Person{DisplayName: "Carol"}, Type: &models.IssueType{Name: "Bug"}, Updated: "2026-04-10T11:00:00Z", Created: "2026-04-09T12:00:00Z"},
		}},
		issueHistoryResults: map[string]*models.IssueHistory{
			"PROJ-1": {IssueKey: "PROJ-1", Entries: []models.HistoryEntry{
				{Date: "2026-04-02 12:00:00", Changes: []models.HistoryChange{{Field: "status", From: "To Do", To: "In Progress"}}},
				{Date: "2026-04-04 12:00:00", Changes: []models.HistoryChange{{Field: "status", From: "In Progress", To: "Blocked"}}},
				{Date: "2026-04-06 12:00:00", Changes: []models.HistoryChange{{Field: "status", From: "Blocked", To: "In Progress"}}},
				{Date: "2026-04-08 12:00:00", Changes: []models.HistoryChange{{Field: "status", From: "In Progress", To: "Code Review"}}},
			}},
			"PROJ-2": {IssueKey: "PROJ-2", Entries: []models.HistoryEntry{
				{Date: "2026-04-05 12:00:00", Changes: []models.HistoryChange{{Field: "status", From: "To Do", To: "In Progress"}}},
				{Date: "2026-04-07 12:00:00", Changes: []models.HistoryChange{{Field: "status", From: "In Progress", To: "Blocked"}}},
			}},
			"PROJ-3": {IssueKey: "PROJ-3", Entries: []models.HistoryEntry{{Date: "2026-04-09 12:00:00", Changes: []models.HistoryChange{{Field: "status", From: "Backlog", To: "To Do"}}}}},
		},
	}

	service := &Service{gateway: gateway, attachmentStore: &stubAttachmentStore{}}
	report, err := service.GetFlowEfficiencyReport(context.Background(), models.GetFlowEfficiencyReportRequest{
		ProjectKey:      "PROJ",
		Statuses:        []string{"In Progress", "Code Review", "Blocked"},
		ActiveStatuses:  []string{"In Progress", "Code Review"},
		BlockedStatuses: []string{"Blocked"},
		MinBlockedDays:  1,
		WindowDays:      7,
		MaxResults:      50,
	})
	require.NoError(t, err)
	require.NotNil(t, report)

	require.Equal(t, `project = "PROJ" AND resolution = Unresolved AND status IN ("In Progress", "Code Review", "Blocked") ORDER BY updated ASC`, report.Query)
	require.Equal(t, 7, report.WindowDays)
	require.Equal(t, "2026-04-03T12:00:00Z", report.WindowStart)
	require.Equal(t, "2026-04-10T12:00:00Z", report.WindowEnd)
	require.Equal(t, 3, report.Summary.AnalyzedIssues)
	require.Equal(t, 2, report.Summary.MatchingIssues)
	require.Equal(t, 2, report.Summary.IssuesWithBlockedTime)
	require.Equal(t, 1, report.Summary.CurrentlyBlockedIssues)
	require.Equal(t, 12, report.Summary.TotalObservedDays)
	require.Equal(t, 7, report.Summary.TotalActiveDays)
	require.Equal(t, 5, report.Summary.TotalBlockedDays)
	require.InDelta(t, 0.557142857, report.Summary.AverageFlowEfficiency, 0.0001)
	require.InDelta(t, 0.583333333, report.Summary.PortfolioFlowEfficiency, 0.0001)
	require.Len(t, report.Assignees, 2)
	require.Equal(t, "Bob", report.Assignees[0].Assignee)
	require.Equal(t, 1, report.Assignees[0].Issues)
	require.Equal(t, 1, report.Assignees[0].CurrentlyBlockedIssues)
	require.Equal(t, 5, report.Assignees[0].TotalObservedDays)
	require.Equal(t, 2, report.Assignees[0].TotalActiveDays)
	require.Equal(t, 3, report.Assignees[0].TotalBlockedDays)
	require.InDelta(t, 0.4, report.Assignees[0].AverageFlowEfficiency, 0.0001)
	require.InDelta(t, 0.4, report.Assignees[0].PortfolioFlowEfficiency, 0.0001)
	require.Equal(t, "Alice", report.Assignees[1].Assignee)
	require.InDelta(t, 0.714285714, report.Assignees[1].AverageFlowEfficiency, 0.0001)
	require.Len(t, report.Items, 2)

	require.Equal(t, "PROJ-2", report.Items[0].Key)
	require.True(t, report.Items[0].CurrentlyBlocked)
	require.Equal(t, 5, report.Items[0].ObservedDays)
	require.Equal(t, 2, report.Items[0].ActiveDays)
	require.Equal(t, 3, report.Items[0].BlockedDays)
	require.Equal(t, 1, report.Items[0].BlockedTransitions)
	require.InDelta(t, 0.4, report.Items[0].FlowEfficiency, 0.0001)

	require.Equal(t, "PROJ-1", report.Items[1].Key)
	require.False(t, report.Items[1].CurrentlyBlocked)
	require.Equal(t, 7, report.Items[1].ObservedDays)
	require.Equal(t, 5, report.Items[1].ActiveDays)
	require.Equal(t, 2, report.Items[1].BlockedDays)
	require.Equal(t, 1, report.Items[1].BlockedTransitions)
	require.InDelta(t, 0.714285714, report.Items[1].FlowEfficiency, 0.0001)

	require.Equal(t, []models.GetIssueHistoryRequest{{IssueKey: "PROJ-1"}, {IssueKey: "PROJ-2"}, {IssueKey: "PROJ-3"}}, gateway.getIssueHistoryReqs)
	require.Equal(t, []string{"summary", "status", "assignee", "updated", "created", "issuetype", "statuscategorychangedate"}, gateway.searchIssuesReq.Fields)
}

func TestGetFlowEfficiencyReportBuildsExplicitDateRange(t *testing.T) {
	previousNow := agingReportNow
	agingReportNow = func() time.Time {
		return time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC)
	}
	defer func() { agingReportNow = previousNow }()

	gateway := &stubGateway{
		searchIssuesResult: &models.SearchIssuesResult{Issues: []models.Issue{{
			Key: "PROJ-1", Summary: "Feature review", Status: &models.Status{Name: "Code Review"}, Assignee: &models.Person{DisplayName: "Alice"}, Type: &models.IssueType{Name: "Story"}, Updated: "2026-04-10T10:00:00Z", Created: "2026-04-01T12:00:00Z",
		}}},
		issueHistoryResults: map[string]*models.IssueHistory{
			"PROJ-1": {IssueKey: "PROJ-1", Entries: []models.HistoryEntry{
				{Date: "2026-04-02 12:00:00", Changes: []models.HistoryChange{{Field: "status", From: "To Do", To: "In Progress"}}},
				{Date: "2026-04-04 12:00:00", Changes: []models.HistoryChange{{Field: "status", From: "In Progress", To: "Blocked"}}},
				{Date: "2026-04-06 12:00:00", Changes: []models.HistoryChange{{Field: "status", From: "Blocked", To: "In Progress"}}},
				{Date: "2026-04-08 12:00:00", Changes: []models.HistoryChange{{Field: "status", From: "In Progress", To: "Code Review"}}},
			}},
		},
	}

	service := &Service{gateway: gateway, attachmentStore: &stubAttachmentStore{}}
	report, err := service.GetFlowEfficiencyReport(context.Background(), models.GetFlowEfficiencyReportRequest{
		ProjectKey:      "PROJ",
		Statuses:        []string{"In Progress", "Code Review", "Blocked"},
		ActiveStatuses:  []string{"In Progress", "Code Review"},
		BlockedStatuses: []string{"Blocked"},
		StartDate:       "2026-04-04",
		EndDate:         "2026-04-08",
		MaxResults:      10,
	})
	require.NoError(t, err)
	require.NotNil(t, report)
	require.Equal(t, 0, report.WindowDays)
	require.Equal(t, "2026-04-04T00:00:00Z", report.WindowStart)
	require.Equal(t, "2026-04-09T00:00:00Z", report.WindowEnd)
	require.Len(t, report.Items, 1)
	require.Equal(t, 5, report.Items[0].ObservedDays)
	require.Equal(t, 3, report.Items[0].ActiveDays)
	require.Equal(t, 2, report.Items[0].BlockedDays)
	require.InDelta(t, 0.6, report.Items[0].FlowEfficiency, 0.0001)
	require.Len(t, report.Assignees, 1)
	require.Equal(t, "Alice", report.Assignees[0].Assignee)
	require.InDelta(t, 0.6, report.Assignees[0].PortfolioFlowEfficiency, 0.0001)
}

func TestGetFlowEfficiencyReportRejectsAmbiguousWindowInputs(t *testing.T) {
	service := &Service{gateway: &stubGateway{}, attachmentStore: &stubAttachmentStore{}}
	_, err := service.GetFlowEfficiencyReport(context.Background(), models.GetFlowEfficiencyReportRequest{
		ProjectKey: "PROJ",
		WindowDays: 7,
		StartDate:  "2026-04-01",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "window_days cannot be combined with start_date or end_date")
}
