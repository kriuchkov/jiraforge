package jira

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kriuchkov/jiraforge/internal/core/models"
)

func TestGetCycleTimeReportComputesMetrics(t *testing.T) {
	prev := cycleTimeReportNow
	cycleTimeReportNow = func() time.Time { return time.Date(2026, time.April, 30, 12, 0, 0, 0, time.UTC) }
	defer func() { cycleTimeReportNow = prev }()

	gateway := &stubGateway{
		searchIssuesResult: &models.SearchIssuesResult{Issues: []models.Issue{
			{Key: "PROJ-1", Status: &models.Status{Name: "Done"}, Assignee: &models.Person{DisplayName: "Alice"}, Type: &models.IssueType{Name: "Story"}, Created: "2026-04-01T12:00:00Z"},
			{Key: "PROJ-2", Status: &models.Status{Name: "Done"}, Assignee: &models.Person{DisplayName: "Bob"}, Type: &models.IssueType{Name: "Bug"}, Created: "2026-04-05T12:00:00Z"},
		}},
		issueHistoryResults: map[string]*models.IssueHistory{
			"PROJ-1": {Entries: []models.HistoryEntry{
				{Date: "2026-04-10T12:00:00Z", Changes: []models.HistoryChange{{Field: "status", To: "In Progress"}}},
				{Date: "2026-04-12T12:00:00Z", Changes: []models.HistoryChange{{Field: "status", To: "Done"}}},
			}},
			"PROJ-2": {Entries: []models.HistoryEntry{
				{Date: "2026-04-15T12:00:00Z", Changes: []models.HistoryChange{{Field: "status", To: "In Progress"}}},
				{Date: "2026-04-20T12:00:00Z", Changes: []models.HistoryChange{{Field: "status", To: "Done"}}},
			}},
		},
	}
	service := &Service{gateway: gateway, attachmentStore: &stubAttachmentStore{}}

	report, err := service.GetCycleTimeReport(context.Background(), models.GetCycleTimeReportRequest{
		ProjectKey: "PROJ",
		WindowDays: 30,
	})
	require.NoError(t, err)
	require.Equal(t, 2, report.Summary.AnalyzedIssues)
	require.Equal(t, 2, report.Summary.CompletedIssues)
	require.InDelta(t, 84.0, report.Summary.AverageCycleHours, 0.01) // (48+120)/2
	require.InDelta(t, 84.0, report.Summary.MedianCycleHours, 0.01)
	require.Greater(t, report.Summary.ThroughputPerWeek, 0.0)
	require.Len(t, report.Items, 2)
	require.Equal(t, "PROJ-2", report.Items[0].Key) // most recent first
}

func TestGetTeamWipSnapshotGroupsByAssignee(t *testing.T) {
	gateway := &stubGateway{
		searchIssuesResult: &models.SearchIssuesResult{Issues: []models.Issue{
			{Key: "PROJ-1", Summary: "A", Status: &models.Status{Name: "In Progress"}, Assignee: &models.Person{DisplayName: "Alice"}, Type: &models.IssueType{Name: "Story"}},
			{Key: "PROJ-2", Summary: "B", Status: &models.Status{Name: "Code Review"}, Assignee: &models.Person{DisplayName: "Alice"}, Type: &models.IssueType{Name: "Bug"}},
			{Key: "PROJ-3", Summary: "C", Status: &models.Status{Name: "In Progress"}, Assignee: &models.Person{DisplayName: "Bob"}, Type: &models.IssueType{Name: "Task"}},
			{Key: "PROJ-4", Summary: "D", Status: &models.Status{Name: "In Progress"}, Type: &models.IssueType{Name: "Task"}},
		}},
	}
	service := &Service{gateway: gateway, attachmentStore: &stubAttachmentStore{}}

	snapshot, err := service.GetTeamWipSnapshot(context.Background(), models.GetTeamWipSnapshotRequest{
		ProjectKey: "PROJ",
		WipLimit:   1,
	})
	require.NoError(t, err)
	require.Equal(t, 4, snapshot.Summary.AnalyzedIssues)
	require.Equal(t, 2, snapshot.Summary.DistinctAssignees)
	require.Equal(t, 1, snapshot.Summary.UnassignedIssues)
	require.Len(t, snapshot.Assignees, 3)

	// Alice has 2 issues, must be over WIP limit of 1
	var alice *models.TeamWipAssignee
	for i := range snapshot.Assignees {
		if snapshot.Assignees[i].Assignee == "Alice" {
			alice = &snapshot.Assignees[i]
		}
	}
	require.NotNil(t, alice)
	require.Equal(t, 2, alice.TotalWip)
	require.True(t, alice.OverLimit)
}

func TestGetSprintWorkloadReportRequiresSprintID(t *testing.T) {
	service := &Service{gateway: &stubGateway{}, attachmentStore: &stubAttachmentStore{}}
	_, err := service.GetSprintWorkloadReport(context.Background(), models.GetSprintWorkloadReportRequest{})
	require.Error(t, err)
}

func TestGetUtilizationReportRequiresProjectOrJQL(t *testing.T) {
	service := &Service{gateway: &stubGateway{}, attachmentStore: &stubAttachmentStore{}}
	_, err := service.GetUtilizationReport(context.Background(), models.GetUtilizationReportRequest{})
	require.Error(t, err)
}
