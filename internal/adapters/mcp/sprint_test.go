package mcp

import (
	"testing"

	"github.com/kriuchkov/jiraforge/internal/adapters/mcp/mocks"
	"github.com/kriuchkov/jiraforge/internal/core/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegisterSprintToolsRegistersExpectedTools(t *testing.T) {
	t.Parallel()

	srv := newTestServer()
	service := mocks.NewJiraService(t)

	registerSprintTools(srv, service)

	tools := srv.ListTools()
	require.Len(t, tools, 6)
	for _, name := range []string{
		"jira_list_sprints",
		"jira_get_sprint",
		"jira_get_sprint_report",
		"jira_get_sprint_health_report",
		"jira_get_active_sprint",
		"jira_search_sprint_by_name",
	} {
		require.Contains(t, tools, name)
	}
}

func TestSprintToolHandlers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		tool   string
		args   map[string]any
		setup  func(*mocks.JiraService)
		assert func(*testing.T, string)
	}{
		{
			name: "list sprints",
			tool: "jira_list_sprints",
			args: map[string]any{"board_id": "42"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().ListSprints(mock.Anything, models.ListSprintsRequest{BoardID: "42"}).Return(&models.SprintCollection{Scope: "board 42", Sprints: []models.Sprint{{ID: 1, Name: "Sprint 1", State: "active"}}}, nil).Once()
			},
			assert: func(t *testing.T, text string) { require.Contains(t, text, "Sprint 1") },
		},
		{
			name: "get sprint",
			tool: "jira_get_sprint",
			args: map[string]any{"sprint_id": "7"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().GetSprint(mock.Anything, models.GetSprintRequest{SprintID: "7"}).Return(&models.Sprint{ID: 7, Name: "Sprint 7", State: "future"}, nil).Once()
			},
			assert: func(t *testing.T, text string) { require.Contains(t, text, "Sprint 7") },
		},
		{
			name: "get sprint report",
			tool: "jira_get_sprint_report",
			args: map[string]any{"sprint_id": "7"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().GetSprintReport(mock.Anything, models.GetSprintReportRequest{SprintID: "7"}).Return(&models.SprintReport{Sprint: models.Sprint{ID: 7, Name: "Sprint 7"}, CompletedIssues: []models.SprintReportIssue{{Key: "PROJ-1", Summary: "Done"}}}, nil).Once()
			},
			assert: func(t *testing.T, text string) {
				require.Contains(t, text, "Sprint Report")
				require.Contains(t, text, "PROJ-1")
			},
		},
		{
			name: "get sprint health report",
			tool: "jira_get_sprint_health_report",
			args: map[string]any{"sprint_id": "7"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().GetSprintHealthReport(mock.Anything, models.GetSprintHealthReportRequest{SprintID: "7"}).Return(&models.SprintHealthReport{Sprint: models.Sprint{ID: 7, Name: "Sprint 7"}, Risks: []string{"low_commitment_completion"}, TopIncompleteIssues: []models.SprintReportIssue{{Key: "PROJ-2", Summary: "Big unfinished item"}}}, nil).Once()
			},
			assert: func(t *testing.T, text string) {
				require.Contains(t, text, "Sprint Health Report")
				require.Contains(t, text, "low_commitment_completion")
				require.Contains(t, text, "PROJ-2")
			},
		},
		{
			name: "get active sprint",
			tool: "jira_get_active_sprint",
			args: map[string]any{"project_key": "PROJ"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().GetActiveSprint(mock.Anything, models.GetActiveSprintRequest{ProjectKey: "PROJ"}).Return(&models.Sprint{ID: 3, Name: "Sprint Active", State: "active"}, nil).Once()
			},
			assert: func(t *testing.T, text string) { require.Contains(t, text, "Sprint Active") },
		},
		{
			name: "search sprint",
			tool: "jira_search_sprint_by_name",
			args: map[string]any{"name": "Sprint", "project_key": "PROJ", "exact_match": true},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().SearchSprints(mock.Anything, models.SearchSprintsRequest{Name: "Sprint", ProjectKey: "PROJ", ExactMatch: true}).Return(&models.SprintCollection{Scope: "project PROJ", Sprints: []models.Sprint{{ID: 10, Name: "Sprint", State: "active"}}}, nil).Once()
			},
			assert: func(t *testing.T, text string) { require.Contains(t, text, "Sprint") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := newTestServer()
			service := mocks.NewJiraService(t)
			registerSprintTools(srv, service)
			tt.setup(service)

			result := callTool(t, srv, tt.tool, tt.args)
			tt.assert(t, toolText(t, result))
		})
	}
}
