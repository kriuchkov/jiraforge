package mcp

import (
	"reflect"
	"testing"

	"github.com/kriuchkov/jiraforge/internal/adapters/mcp/mocks"
	"github.com/kriuchkov/jiraforge/internal/core/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegisterIssueToolsRegistersExpectedTools(t *testing.T) {
	t.Parallel()

	srv := newTestServer()
	service := mocks.NewJiraService(t)

	registerIssueTools(srv, service)
	registerSearchTools(srv, service)

	tools := srv.ListTools()
	require.Len(t, tools, 7)
	for _, name := range []string{
		"jira_get_issue",
		"jira_create_issue",
		"jira_create_child_issue",
		"jira_update_issue",
		"jira_delete_issue",
		"jira_list_issue_types",
		"jira_search_issue",
	} {
		require.Contains(t, tools, name)
	}
}

func TestIssueToolHandlers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		tool   string
		args   map[string]any
		setup  func(*mocks.JiraService)
		assert func(*testing.T, string)
	}{
		{
			name: "get issue",
			tool: "jira_get_issue",
			args: map[string]any{"issue_key": "PROJ-1", "fields": "summary,status", "expand": "transitions, changelog"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().GetIssue(mock.Anything, mock.MatchedBy(func(req models.GetIssueRequest) bool {
					return req.IssueKey == "PROJ-1" &&
						reflect.DeepEqual(req.Fields, []string{"summary", "status"}) &&
						reflect.DeepEqual(req.Expand, []string{"transitions", "changelog"})
				})).Return(&models.Issue{Key: "PROJ-1", Summary: "Fix login", Status: &models.Status{Name: "In Progress"}, Type: &models.IssueType{Name: "Bug"}}, nil).Once()
			},
			assert: func(t *testing.T, text string) {
				require.Contains(t, text, "PROJ-1")
				require.Contains(t, text, "Fix login")
			},
		},
		{
			name: "create issue",
			tool: "jira_create_issue",
			args: map[string]any{"project_key": "PROJ", "summary": "Fix login", "description": "Broken OAuth redirect", "issue_type": "Bug"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().CreateIssue(mock.Anything, models.CreateIssueRequest{ProjectKey: "PROJ", Summary: "Fix login", Description: "Broken OAuth redirect", IssueType: "Bug"}).Return(&models.IssueMutationResult{Key: "PROJ-1", Message: "created"}, nil).Once()
			},
			assert: func(t *testing.T, text string) {
				require.Contains(t, text, "PROJ-1")
			},
		},
		{
			name: "create child issue",
			tool: "jira_create_child_issue",
			args: map[string]any{"parent_issue_key": "PROJ-1", "summary": "Write regression test", "description": "Need coverage", "issue_type": "Sub-task"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().CreateChildIssue(mock.Anything, models.CreateChildIssueRequest{ParentIssueKey: "PROJ-1", Summary: "Write regression test", Description: "Need coverage", IssueType: "Sub-task"}).Return(&models.IssueMutationResult{Key: "PROJ-2", ParentKey: "PROJ-1", Message: "created"}, nil).Once()
			},
			assert: func(t *testing.T, text string) {
				require.Contains(t, text, "PROJ-2")
				require.Contains(t, text, "PROJ-1")
			},
		},
		{
			name: "update issue",
			tool: "jira_update_issue",
			args: map[string]any{"issue_key": "PROJ-1", "summary": "Fix login callback", "description": "Updated details"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().UpdateIssue(mock.Anything, models.UpdateIssueRequest{IssueKey: "PROJ-1", Summary: "Fix login callback", Description: "Updated details"}).Return(&models.IssueMutationResult{Key: "PROJ-1", Message: "updated"}, nil).Once()
			},
			assert: func(t *testing.T, text string) {
				require.Contains(t, text, "updated")
			},
		},
		{
			name: "delete issue",
			tool: "jira_delete_issue",
			args: map[string]any{"issue_key": "PROJ-9"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().DeleteIssue(mock.Anything, models.DeleteIssueRequest{IssueKey: "PROJ-9"}).Return(&models.IssueMutationResult{Key: "PROJ-9", Message: "deleted"}, nil).Once()
			},
			assert: func(t *testing.T, text string) {
				require.Contains(t, text, "deleted")
			},
		},
		{
			name: "list issue types",
			tool: "jira_list_issue_types",
			args: map[string]any{"project_key": "PROJ"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().ListIssueTypes(mock.Anything, models.ListIssueTypesRequest{ProjectKey: "PROJ"}).Return([]models.IssueType{{ID: "1", Name: "Bug"}, {ID: "2", Name: "Story"}}, nil).Once()
			},
			assert: func(t *testing.T, text string) {
				require.Contains(t, text, "Bug")
				require.Contains(t, text, "Story")
			},
		},
		{
			name: "search issues",
			tool: "jira_search_issue",
			args: map[string]any{"jql": "project = PROJ", "fields": "summary", "expand": "changelog"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().SearchIssues(mock.Anything, mock.MatchedBy(func(req models.SearchIssuesRequest) bool {
					return req.JQL == "project = PROJ" &&
						reflect.DeepEqual(req.Fields, []string{"summary"}) &&
						reflect.DeepEqual(req.Expand, []string{"changelog"})
				})).Return(&models.SearchIssuesResult{Query: "project = PROJ", Total: 1, Issues: []models.Issue{{Key: "PROJ-1", Summary: "Fix login"}}}, nil).Once()
			},
			assert: func(t *testing.T, text string) {
				require.Contains(t, text, "PROJ-1")
				require.Contains(t, text, "Fix login")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := newTestServer()
			service := mocks.NewJiraService(t)
			registerIssueTools(srv, service)
			registerSearchTools(srv, service)
			tt.setup(service)

			result := callTool(t, srv, tt.tool, tt.args)
			tt.assert(t, toolText(t, result))
		})
	}
}
