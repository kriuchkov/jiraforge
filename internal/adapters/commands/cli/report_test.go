package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReportCommandValidationErrors(t *testing.T) {
	t.Parallel()

	assertValidationError(t, []string{"get-aging-report"}, "either --project-key or --jql is required")
	assertValidationError(t, []string{"get-blocked-issues-report"}, "either --project-key or --jql is required")
	assertValidationError(t, []string{"get-flow-efficiency-report"}, "either --project-key or --jql is required")
}

func TestReportCommandDispatch(t *testing.T) {
	t.Parallel()

	assertDispatch(t, []string{"get-aging-report", "--project-key", "PROJ", "--status", "In Progress", "--status", "Code Review", "--assignee", "alice", "--min-days-in-status", "5", "--max-results", "50"}, &stubCLIService{}, func(t *testing.T, service *stubCLIService, _, _ string) {
		require.Equal(t, "GetAgingReport", service.lastCall)
		require.Equal(t, "PROJ", service.getAgingReportReq.ProjectKey)
		require.Equal(t, []string{"In Progress", "Code Review"}, service.getAgingReportReq.Statuses)
		require.Equal(t, "alice", service.getAgingReportReq.Assignee)
		require.Equal(t, 5, service.getAgingReportReq.MinDaysInStatus)
		require.Equal(t, 50, service.getAgingReportReq.MaxResults)
	})

	assertDispatch(t, []string{"get-blocked-issues-report", "--project-key", "PROJ", "--status", "In Progress", "--blocked-status", "Blocked", "--link-type", "is blocked by", "--assignee", "alice", "--min-days-in-status", "3", "--max-results", "40"}, &stubCLIService{}, func(t *testing.T, service *stubCLIService, _, _ string) {
		require.Equal(t, "GetBlockedIssuesReport", service.lastCall)
		require.Equal(t, "PROJ", service.getBlockedIssuesReportReq.ProjectKey)
		require.Equal(t, []string{"In Progress"}, service.getBlockedIssuesReportReq.Statuses)
		require.Equal(t, []string{"Blocked"}, service.getBlockedIssuesReportReq.BlockedStatuses)
		require.Equal(t, []string{"is blocked by"}, service.getBlockedIssuesReportReq.LinkTypes)
		require.Equal(t, "alice", service.getBlockedIssuesReportReq.Assignee)
		require.Equal(t, 3, service.getBlockedIssuesReportReq.MinDaysInStatus)
		require.Equal(t, 40, service.getBlockedIssuesReportReq.MaxResults)
	})

	assertDispatch(t, []string{"get-flow-efficiency-report", "--project-key", "PROJ", "--status", "In Progress", "--status", "Code Review", "--status", "Blocked", "--active-status", "In Progress", "--active-status", "Code Review", "--blocked-status", "Blocked", "--assignee", "alice", "--min-blocked-days", "2", "--window-days", "7", "--max-results", "25"}, &stubCLIService{}, func(t *testing.T, service *stubCLIService, _, _ string) {
		require.Equal(t, "GetFlowEfficiencyReport", service.lastCall)
		require.Equal(t, "PROJ", service.getFlowEfficiencyReportReq.ProjectKey)
		require.Equal(t, []string{"In Progress", "Code Review", "Blocked"}, service.getFlowEfficiencyReportReq.Statuses)
		require.Equal(t, []string{"In Progress", "Code Review"}, service.getFlowEfficiencyReportReq.ActiveStatuses)
		require.Equal(t, []string{"Blocked"}, service.getFlowEfficiencyReportReq.BlockedStatuses)
		require.Equal(t, "alice", service.getFlowEfficiencyReportReq.Assignee)
		require.Equal(t, 2, service.getFlowEfficiencyReportReq.MinBlockedDays)
		require.Equal(t, 7, service.getFlowEfficiencyReportReq.WindowDays)
		require.Equal(t, 25, service.getFlowEfficiencyReportReq.MaxResults)
	})

	assertDispatch(t, []string{"get-flow-efficiency-report", "--project-key", "PROJ", "--status", "In Progress", "--status", "Blocked", "--active-status", "In Progress", "--blocked-status", "Blocked", "--start-date", "2026-04-01", "--end-date", "2026-04-07", "--max-results", "15"}, &stubCLIService{}, func(t *testing.T, service *stubCLIService, _, _ string) {
		require.Equal(t, "GetFlowEfficiencyReport", service.lastCall)
		require.Equal(t, "2026-04-01", service.getFlowEfficiencyReportReq.StartDate)
		require.Equal(t, "2026-04-07", service.getFlowEfficiencyReportReq.EndDate)
		require.Zero(t, service.getFlowEfficiencyReportReq.WindowDays)
		require.Equal(t, 15, service.getFlowEfficiencyReportReq.MaxResults)
	})
}
