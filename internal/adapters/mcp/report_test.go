package mcp

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kriuchkov/jiraforge/internal/adapters/mcp/mocks"
	"github.com/kriuchkov/jiraforge/internal/core/models"
)

func TestRegisterReportToolsRegistersExpectedTools(t *testing.T) {
	t.Parallel()

	srv := newTestServer()
	service := mocks.NewJiraService(t)

	registerReportTools(srv, service)

	tools := srv.ListTools()
	require.Len(t, tools, 7)
	require.Contains(t, tools, "jira_get_aging_report")
	require.Contains(t, tools, "jira_get_blocked_issues_report")
	require.Contains(t, tools, "jira_get_flow_efficiency_report")
	require.Contains(t, tools, "jira_get_sprint_workload_report")
	require.Contains(t, tools, "jira_get_utilization_report")
	require.Contains(t, tools, "jira_get_cycle_time_report")
	require.Contains(t, tools, "jira_get_team_wip_snapshot")
}

func TestReportToolHandlers(t *testing.T) {
	t.Parallel()

	srv := newTestServer()
	service := mocks.NewJiraService(t)
	registerReportTools(srv, service)

	service.EXPECT().GetAgingReport(mock.Anything, models.GetAgingReportRequest{
		ProjectKey:      "PROJ",
		Statuses:        []string{"In Progress", "Code Review"},
		Assignee:        "alice",
		MinDaysInStatus: 5,
		MaxResults:      50,
	}).Return(&models.AgingReport{Summary: models.AgingReportSummary{MatchingIssues: 1}, Items: []models.AgingReportItem{{Key: "PROJ-1", Summary: "Investigate regression", DaysInStatus: 6}}}, nil).Once()

	result := callTool(t, srv, "jira_get_aging_report", map[string]any{
		"project_key":        "PROJ",
		"statuses":           "In Progress, Code Review",
		"assignee":           "alice",
		"min_days_in_status": 5,
		"max_results":        50,
	})

	text := toolText(t, result)
	require.Contains(t, text, "Aging Report")
	require.Contains(t, text, "PROJ-1")
	require.Contains(t, text, "Days In Status: 6")

	service.EXPECT().GetBlockedIssuesReport(mock.Anything, models.GetBlockedIssuesReportRequest{
		ProjectKey:      "PROJ",
		Statuses:        []string{"In Progress", "Code Review"},
		BlockedStatuses: []string{"Blocked"},
		LinkTypes:       []string{"is blocked by"},
		Assignee:        "alice",
		MinDaysInStatus: 3,
		MaxResults:      40,
	}).Return(&models.BlockedIssuesReport{Summary: models.BlockedIssuesReportSummary{BlockedIssues: 1}, Items: []models.BlockedIssuesReportItem{{Key: "PROJ-2", Summary: "Waiting for dependency", DaysInStatus: 9, Reasons: []string{"blocking_dependency"}}}}, nil).Once()

	result = callTool(t, srv, "jira_get_blocked_issues_report", map[string]any{
		"project_key":        "PROJ",
		"statuses":           "In Progress, Code Review",
		"blocked_statuses":   "Blocked",
		"link_types":         "is blocked by",
		"assignee":           "alice",
		"min_days_in_status": 3,
		"max_results":        40,
	})

	text = toolText(t, result)
	require.Contains(t, text, "Blocked Issues Report")
	require.Contains(t, text, "PROJ-2")
	require.Contains(t, text, "blocking_dependency")

	service.EXPECT().GetFlowEfficiencyReport(mock.Anything, models.GetFlowEfficiencyReportRequest{
		ProjectKey:      "PROJ",
		Statuses:        []string{"In Progress", "Code Review", "Blocked"},
		ActiveStatuses:  []string{"In Progress", "Code Review"},
		BlockedStatuses: []string{"Blocked"},
		Assignee:        "alice",
		MinBlockedDays:  2,
		WindowDays:      7,
		MaxResults:      25,
	}).Return(&models.FlowEfficiencyReport{WindowDays: 7, WindowStart: "2026-04-03T12:00:00Z", WindowEnd: "2026-04-10T12:00:00Z", Summary: models.FlowEfficiencyReportSummary{MatchingIssues: 1}, Assignees: []models.FlowEfficiencyAssigneeSummary{{Assignee: "alice", Issues: 1, PortfolioFlowEfficiency: 0.4}}, Items: []models.FlowEfficiencyReportItem{{Key: "PROJ-3", Summary: "Slow review", FlowEfficiency: 0.4, BlockedDays: 3}}}, nil).Once()

	result = callTool(t, srv, "jira_get_flow_efficiency_report", map[string]any{
		"project_key":      "PROJ",
		"statuses":         "In Progress, Code Review, Blocked",
		"active_statuses":  "In Progress, Code Review",
		"blocked_statuses": "Blocked",
		"assignee":         "alice",
		"min_blocked_days": 2,
		"window_days":      7,
		"max_results":      25,
	})

	text = toolText(t, result)
	require.Contains(t, text, "Flow Efficiency Report")
	require.Contains(t, text, "PROJ-3")
	require.Contains(t, text, "40.0%")
	require.Contains(t, text, "Window Days: 7")
	require.Contains(t, text, "Assignees (1)")

	service.EXPECT().GetFlowEfficiencyReport(mock.Anything, models.GetFlowEfficiencyReportRequest{
		ProjectKey:      "PROJ",
		Statuses:        []string{"In Progress", "Blocked"},
		ActiveStatuses:  []string{"In Progress"},
		BlockedStatuses: []string{"Blocked"},
		StartDate:       "2026-04-01",
		EndDate:         "2026-04-07",
		MaxResults:      10,
	}).Return(&models.FlowEfficiencyReport{WindowStart: "2026-04-01T00:00:00Z", WindowEnd: "2026-04-08T00:00:00Z", Summary: models.FlowEfficiencyReportSummary{MatchingIssues: 1}, Items: []models.FlowEfficiencyReportItem{{Key: "PROJ-4", Summary: "Release prep", FlowEfficiency: 0.5, BlockedDays: 2}}}, nil).Once()

	result = callTool(t, srv, "jira_get_flow_efficiency_report", map[string]any{
		"project_key":      "PROJ",
		"statuses":         "In Progress, Blocked",
		"active_statuses":  "In Progress",
		"blocked_statuses": "Blocked",
		"start_date":       "2026-04-01",
		"end_date":         "2026-04-07",
		"max_results":      10,
	})

	text = toolText(t, result)
	require.Contains(t, text, "Window Start: 2026-04-01T00:00:00Z")
	require.Contains(t, text, "Window End: 2026-04-08T00:00:00Z")
	require.Contains(t, text, "PROJ-4")
}
