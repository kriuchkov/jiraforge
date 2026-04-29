package mcp

import (
	"testing"

	"github.com/kriuchkov/jiraforge/internal/adapters/mcp/mocks"
	"github.com/kriuchkov/jiraforge/internal/core/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegisterWorkflowToolsRegistersExpectedTools(t *testing.T) {
	t.Parallel()

	srv := newTestServer()
	service := mocks.NewJiraService(t)

	registerCommentTools(srv, service)
	registerWorklogTools(srv, service)
	registerTransitionTools(srv, service)
	registerStatusTools(srv, service)
	registerHistoryTools(srv, service)

	tools := srv.ListTools()
	require.Len(t, tools, 8)
	for _, name := range []string{
		"jira_add_comment",
		"jira_get_comments",
		"jira_add_worklog",
		"jira_get_worklogs",
		"jira_get_transitions",
		"jira_transition_issue",
		"jira_list_statuses",
		"jira_get_issue_history",
	} {
		require.Contains(t, tools, name)
	}
}

func TestWorkflowToolHandlers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		tool   string
		args   map[string]any
		setup  func(*mocks.JiraService)
		assert func(*testing.T, string)
	}{
		{
			name: "add comment",
			tool: "jira_add_comment",
			args: map[string]any{"issue_key": "PROJ-1", "comment": "Looks good"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().AddComment(mock.Anything, models.AddCommentRequest{IssueKey: "PROJ-1", Comment: "Looks good"}).Return(&models.Comment{ID: "1", Body: "Looks good"}, nil).Once()
			},
			assert: func(t *testing.T, text string) { require.Contains(t, text, "Looks good") },
		},
		{
			name: "get comments",
			tool: "jira_get_comments",
			args: map[string]any{"issue_key": "PROJ-1"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().GetComments(mock.Anything, models.GetCommentsRequest{IssueKey: "PROJ-1"}).Return(&models.CommentList{IssueKey: "PROJ-1", Comments: []models.Comment{{ID: "1", Body: "Looks good"}}}, nil).Once()
			},
			assert: func(t *testing.T, text string) { require.Contains(t, text, "Looks good") },
		},
		{
			name: "add worklog",
			tool: "jira_add_worklog",
			args: map[string]any{"issue_key": "PROJ-1", "time_spent": "1h", "comment": "Investigated", "started": "2026-04-09T10:00:00Z"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().AddWorklog(mock.Anything, models.AddWorklogRequest{IssueKey: "PROJ-1", TimeSpent: "1h", Comment: "Investigated", Started: "2026-04-09T10:00:00Z"}).Return(&models.Worklog{ID: "wl-1", IssueKey: "PROJ-1", TimeSpent: "1h", Comment: "Investigated"}, nil).Once()
			},
			assert: func(t *testing.T, text string) { require.Contains(t, text, "1h") },
		},
		{
			name: "get transitions",
			tool: "jira_get_transitions",
			args: map[string]any{"issue_key": "PROJ-1"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().GetTransitions(mock.Anything, models.GetTransitionsRequest{IssueKey: "PROJ-1"}).Return([]models.Transition{{ID: "31", Name: "Done"}}, nil).Once()
			},
			assert: func(t *testing.T, text string) { require.Contains(t, text, "Done") },
		},
		{
			name: "transition issue",
			tool: "jira_transition_issue",
			args: map[string]any{"issue_key": "PROJ-1", "transition_id": "31", "comment": "Ship it"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().TransitionIssue(mock.Anything, models.TransitionIssueRequest{IssueKey: "PROJ-1", TransitionID: "31", Comment: "Ship it"}).Return(&models.IssueMutationResult{Key: "PROJ-1", Message: "transitioned"}, nil).Once()
			},
			assert: func(t *testing.T, text string) { require.Contains(t, text, "transitioned") },
		},
		{
			name: "list statuses",
			tool: "jira_list_statuses",
			args: map[string]any{"project_key": "PROJ"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().ListStatuses(mock.Anything, models.ListStatusesRequest{ProjectKey: "PROJ"}).Return(&models.StatusCatalog{ProjectKey: "PROJ", Groups: []models.StatusGroup{{IssueType: models.IssueType{Name: "Bug"}, Statuses: []models.Status{{Name: "To Do"}}}}}, nil).Once()
			},
			assert: func(t *testing.T, text string) { require.Contains(t, text, "To Do") },
		},
		{
			name: "issue history",
			tool: "jira_get_issue_history",
			args: map[string]any{"issue_key": "PROJ-1"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().GetIssueHistory(mock.Anything, models.GetIssueHistoryRequest{IssueKey: "PROJ-1"}).Return(&models.IssueHistory{IssueKey: "PROJ-1", Entries: []models.HistoryEntry{{Date: "2026-04-09", Author: "Alice", Changes: []models.HistoryChange{{Field: "status", From: "To Do", To: "Done"}}}}}, nil).Once()
			},
			assert: func(t *testing.T, text string) { require.Contains(t, text, "status") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := newTestServer()
			service := mocks.NewJiraService(t)
			registerCommentTools(srv, service)
			registerWorklogTools(srv, service)
			registerTransitionTools(srv, service)
			registerStatusTools(srv, service)
			registerHistoryTools(srv, service)
			tt.setup(service)

			result := callTool(t, srv, tt.tool, tt.args)
			tt.assert(t, toolText(t, result))
		})
	}
}
