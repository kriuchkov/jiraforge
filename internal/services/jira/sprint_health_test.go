package jira

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kriuchkov/jiraforge/internal/core/models"
)

func TestGetSprintHealthReportBuildsSummaryFromSprintReport(t *testing.T) {
	t.Parallel()

	service := &Service{
		gateway: &stubGateway{
			sprintReportResult: &models.SprintReport{
				Sprint: models.Sprint{ID: 7, Name: "Sprint 7"},
				CompletedIssues: []models.SprintReportIssue{
					{Key: "PROJ-1", Summary: "Committed and done", Estimate: models.SprintReportEstimate{Value: 3}, CurrentEstimate: models.SprintReportEstimate{Value: 3}},
					{Key: "PROJ-5", Summary: "Unplanned and done", Estimate: models.SprintReportEstimate{Value: 1}, CurrentEstimate: models.SprintReportEstimate{Value: 1}, AddedDuringSprint: true},
				},
				IncompleteIssues: []models.SprintReportIssue{
					{Key: "PROJ-2", Summary: "Big unfinished item", Estimate: models.SprintReportEstimate{Value: 5}, CurrentEstimate: models.SprintReportEstimate{Value: 8}},
					{Key: "PROJ-6", Summary: "Added late", Estimate: models.SprintReportEstimate{Value: 2}, CurrentEstimate: models.SprintReportEstimate{Value: 2}, AddedDuringSprint: true},
				},
				RemovedIssues:            []models.SprintReportIssue{{Key: "PROJ-3", Summary: "Removed", Estimate: models.SprintReportEstimate{Value: 2}, CurrentEstimate: models.SprintReportEstimate{Value: 2}}},
				CompletedInAnotherSprint: []models.SprintReportIssue{{Key: "PROJ-4", Summary: "Finished later", Estimate: models.SprintReportEstimate{Value: 1}, CurrentEstimate: models.SprintReportEstimate{Value: 1}}},
				AddedIssueKeys:           []string{"PROJ-5", "PROJ-6"},
			},
		},
		attachmentStore: &stubAttachmentStore{},
	}

	report, err := service.GetSprintHealthReport(context.Background(), models.GetSprintHealthReportRequest{SprintID: "7"})
	require.NoError(t, err)
	require.NotNil(t, report)

	require.Equal(t, 4, report.Summary.CommittedIssues)
	require.Equal(t, 1, report.Summary.CompletedCommittedIssues)
	require.Equal(t, 2, report.Summary.CompletedIssues)
	require.Equal(t, 2, report.Summary.IncompleteIssues)
	require.Equal(t, 1, report.Summary.CompletedElsewhereIssues)
	require.Equal(t, 2, report.Summary.CarryOverIssues)
	require.Equal(t, 1, report.Summary.RemovedIssues)
	require.Equal(t, 2, report.Summary.AddedDuringSprint)
	require.Equal(t, 1, report.Summary.UnplannedCompletedIssues)
	require.InDelta(t, 0.25, report.Summary.CompletionRatio, 0.0001)
	require.InDelta(t, 11.0, report.Summary.CommittedEstimate, 0.0001)
	require.InDelta(t, 3.0, report.Summary.CompletedCommittedEstimate, 0.0001)
	require.InDelta(t, 3.0/11.0, report.Summary.EstimateCompletionRatio, 0.0001)
	require.Contains(t, report.Risks, "low_commitment_completion")
	require.Contains(t, report.Risks, "high_scope_creep")
	require.Contains(t, report.Risks, "high_scope_removal")
	require.Contains(t, report.Risks, "carry_over_risk")
	require.Contains(t, report.Risks, "high_unplanned_delivery")
	require.Contains(t, report.Risks, "low_estimate_completion")
	require.Len(t, report.TopIncompleteIssues, 2)
	require.Equal(t, "PROJ-2", report.TopIncompleteIssues[0].Key)
}
