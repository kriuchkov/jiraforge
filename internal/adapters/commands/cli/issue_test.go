package cli

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/kriuchkov/jiraforge/internal/core/models"
	"github.com/kriuchkov/jiraforge/internal/core/ports"
	"github.com/stretchr/testify/require"
)

func TestIssueCommandValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		wantError string
	}{
		{name: "get issue", args: []string{"get-issue"}, wantError: "--issue-key is required"},
		{name: "search issues", args: []string{"search-issues"}, wantError: "--jql is required"},
		{name: "create issue", args: []string{"create-issue"}, wantError: "--project-key, --summary, --description, and --issue-type are required"},
		{name: "create child issue", args: []string{"create-child-issue"}, wantError: "--parent-issue-key, --summary, and --description are required"},
		{name: "update issue missing key", args: []string{"update-issue"}, wantError: "--issue-key is required"},
		{name: "update issue missing fields", args: []string{"update-issue", "--issue-key", "PROJ-1"}, wantError: "at least one of --summary or --description is required"},
		{name: "delete issue", args: []string{"delete-issue"}, wantError: "--issue-key is required"},
		{name: "list issue types", args: []string{"list-issue-types"}, wantError: "--project-key is required"},
		{name: "add comment", args: []string{"add-comment"}, wantError: "--issue-key and --comment are required"},
		{name: "get comments", args: []string{"get-comments"}, wantError: "--issue-key is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertValidationError(t, tt.args, tt.wantError)
		})
	}
}

func TestIssueCommandDispatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		service *stubCLIService
		assert  func(*testing.T, *stubCLIService, string, string)
	}{
		{
			name:    "get issue json",
			args:    []string{"get-issue", "--env", ".env.local", "--issue-key", "PROJ-42", "--fields", "summary,status", "--expand", " transitions , changelog ", "--output", "json"},
			service: &stubCLIService{getIssueResult: &models.Issue{Key: "PROJ-42", Summary: "Investigate"}},
			assert: func(t *testing.T, service *stubCLIService, stdout, stderr string) {
				require.Empty(t, stderr)
				require.Equal(t, ".env.local", service.envArg)
				require.Equal(t, "GetIssue", service.lastCall)
				require.Equal(t, "PROJ-42", service.getIssueReq.IssueKey)
				require.Equal(t, []string{"summary", "status"}, service.getIssueReq.Fields)
				require.Equal(t, []string{"transitions", "changelog"}, service.getIssueReq.Expand)
				var payload models.Issue
				require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
				require.Equal(t, "PROJ-42", payload.Key)
				require.Equal(t, "Investigate", payload.Summary)
			},
		},
		{
			name:    "search issues",
			args:    []string{"search-issues", "--jql", "project = PROJ", "--max-results", "15", "--output", "json"},
			service: &stubCLIService{searchIssuesResult: &models.SearchIssuesResult{Query: "project = PROJ", Total: 2}},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "SearchIssues", service.lastCall)
				require.Equal(t, "project = PROJ", service.searchIssuesReq.JQL)
				require.Equal(t, 15, service.searchIssuesReq.MaxResults)
			},
		},
		{
			name:    "create issue",
			args:    []string{"create-issue", "--project-key", "PROJ", "--summary", "Investigate login", "--description", "Broken redirect", "--issue-type", "Bug"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "CreateIssue", service.lastCall)
				require.Equal(t, "PROJ", service.createIssueReq.ProjectKey)
				require.Equal(t, "Investigate login", service.createIssueReq.Summary)
				require.Equal(t, "Broken redirect", service.createIssueReq.Description)
				require.Equal(t, "Bug", service.createIssueReq.IssueType)
			},
		},
		{
			name:    "create child issue",
			args:    []string{"create-child-issue", "--parent-issue-key", "PROJ-1", "--summary", "Investigate child", "--description", "Details", "--issue-type", "Sub-task"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "CreateChildIssue", service.lastCall)
				require.Equal(t, "PROJ-1", service.createChildIssueReq.ParentIssueKey)
				require.Equal(t, "Investigate child", service.createChildIssueReq.Summary)
				require.Equal(t, "Details", service.createChildIssueReq.Description)
				require.Equal(t, "Sub-task", service.createChildIssueReq.IssueType)
			},
		},
		{
			name:    "update issue",
			args:    []string{"update-issue", "--issue-key", "PROJ-2", "--summary", "Updated summary", "--description", "Updated description"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "UpdateIssue", service.lastCall)
				require.Equal(t, "PROJ-2", service.updateIssueReq.IssueKey)
				require.Equal(t, "Updated summary", service.updateIssueReq.Summary)
				require.Equal(t, "Updated description", service.updateIssueReq.Description)
			},
		},
		{
			name:    "delete issue",
			args:    []string{"delete-issue", "--issue-key", "PROJ-3"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "DeleteIssue", service.lastCall)
				require.Equal(t, "PROJ-3", service.deleteIssueReq.IssueKey)
			},
		},
		{
			name:    "list issue types",
			args:    []string{"list-issue-types", "--project-key", "PROJ"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "ListIssueTypes", service.lastCall)
				require.Equal(t, "PROJ", service.listIssueTypesReq.ProjectKey)
			},
		},
		{
			name:    "add comment",
			args:    []string{"add-comment", "--issue-key", "PROJ-7", "--comment", "hello"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "AddComment", service.lastCall)
				require.Equal(t, "PROJ-7", service.addCommentReq.IssueKey)
				require.Equal(t, "hello", service.addCommentReq.Comment)
			},
		},
		{
			name:    "get comments",
			args:    []string{"get-comments", "--issue-key", "PROJ-7"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "GetComments", service.lastCall)
				require.Equal(t, "PROJ-7", service.getCommentsReq.IssueKey)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertDispatch(t, tt.args, tt.service, tt.assert)
		})
	}
}

func TestGetIssueCommandUnsupportedOutputFormat(t *testing.T) {
	t.Parallel()

	service := &stubCLIService{getIssueResult: &models.Issue{Key: "PROJ-42"}}
	_, stderr, exitCode := executeCLI(t, []string{"get-issue", "--issue-key", "PROJ-42", "--output", "yaml"}, service, newStubFactory(service))

	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr, "unsupported output format \"yaml\"")
}

func TestIssueCommandPropagatesFactoryError(t *testing.T) {
	t.Parallel()

	factoryErr := errors.New("config failed")
	_, stderr, exitCode := executeCLI(t, []string{"get-issue", "--issue-key", "PROJ-42"}, nil, func(string) (ports.JiraService, error) {
		return nil, factoryErr
	})

	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr, factoryErr.Error())
}
