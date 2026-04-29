package mcp

import (
	"testing"

	"github.com/kriuchkov/jiraforge/internal/adapters/mcp/mocks"
	"github.com/kriuchkov/jiraforge/internal/core/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegisterMetadataToolsAndPrompts(t *testing.T) {
	t.Parallel()

	srv := newTestServer()
	service := mocks.NewJiraService(t)

	registerRelationshipTools(srv, service)
	registerVersionTools(srv, service)
	registerDevelopmentTools(srv, service)
	registerAttachmentTools(srv, service)
	registerPrompts(srv)

	tools := srv.ListTools()
	require.Len(t, tools, 6)
	for _, name := range []string{
		"jira_get_related_issues",
		"jira_link_issues",
		"jira_get_version",
		"jira_list_project_versions",
		"jira_get_development_information",
		"jira_download_attachment",
	} {
		require.Contains(t, tools, name)
	}

	prompts := listPrompts(t, srv)
	require.Len(t, prompts.Prompts, 2)
	require.Equal(t, "issue_development_tree", prompts.Prompts[0].Name)
	require.Equal(t, "release_development_overview", prompts.Prompts[1].Name)
}

func TestMetadataToolHandlers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		tool   string
		args   map[string]any
		setup  func(*mocks.JiraService)
		assert func(*testing.T, string)
	}{
		{
			name: "get related issues",
			tool: "jira_get_related_issues",
			args: map[string]any{"issue_key": "PROJ-1"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().GetRelatedIssues(mock.Anything, models.GetRelatedIssuesRequest{IssueKey: "PROJ-1"}).Return([]models.IssueRelation{{Type: "blocks", Direction: "outward", Issue: models.IssueRef{Key: "PROJ-2", Summary: "Blocked task"}}}, nil).Once()
			},
			assert: func(t *testing.T, text string) { require.Contains(t, text, "PROJ-2") },
		},
		{
			name: "link issues",
			tool: "jira_link_issues",
			args: map[string]any{"inward_issue": "PROJ-1", "outward_issue": "PROJ-2", "link_type": "blocks", "comment": "dependency"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().LinkIssues(mock.Anything, models.LinkIssuesRequest{InwardIssue: "PROJ-1", OutwardIssue: "PROJ-2", LinkType: "blocks", Comment: "dependency"}).Return(&models.IssueMutationResult{Key: "PROJ-1", Message: "linked"}, nil).Once()
			},
			assert: func(t *testing.T, text string) { require.Contains(t, text, "linked") },
		},
		{
			name: "get version",
			tool: "jira_get_version",
			args: map[string]any{"version_id": "1001"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().GetVersion(mock.Anything, models.GetVersionRequest{VersionID: "1001"}).Return(&models.Version{ID: "1001", Name: "1.0.0", Status: "released"}, nil).Once()
			},
			assert: func(t *testing.T, text string) { require.Contains(t, text, "1.0.0") },
		},
		{
			name: "list project versions",
			tool: "jira_list_project_versions",
			args: map[string]any{"project_key": "PROJ"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().ListProjectVersions(mock.Anything, models.ListProjectVersionsRequest{ProjectKey: "PROJ"}).Return(&models.VersionCollection{ProjectKey: "PROJ", Versions: []models.Version{{ID: "1001", Name: "1.0.0"}}}, nil).Once()
			},
			assert: func(t *testing.T, text string) { require.Contains(t, text, "1.0.0") },
		},
		{
			name: "development info",
			tool: "jira_get_development_information",
			args: map[string]any{"issue_key": "PROJ-1", "include_branches": true, "include_pull_requests": true, "include_commits": true, "include_builds": true},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().GetDevelopmentInfo(mock.Anything, models.GetDevelopmentInfoRequest{IssueKey: "PROJ-1", IncludeBranches: true, IncludePullRequests: true, IncludeCommits: true, IncludeBuilds: true}).Return(&models.DevelopmentInfo{IssueKey: "PROJ-1", Branches: []models.DevelopmentBranch{{Name: "feature/PROJ-1"}}}, nil).Once()
			},
			assert: func(t *testing.T, text string) { require.Contains(t, text, "feature/PROJ-1") },
		},
		{
			name: "download attachment",
			tool: "jira_download_attachment",
			args: map[string]any{"attachment_id": "att-1"},
			setup: func(service *mocks.JiraService) {
				service.EXPECT().DownloadAttachment(mock.Anything, models.DownloadAttachmentRequest{AttachmentID: "att-1"}).Return(&models.StoredAttachment{Path: "/tmp/evidence.txt", Filename: "evidence.txt", Size: 128}, nil).Once()
			},
			assert: func(t *testing.T, text string) {
				require.Contains(t, text, "/tmp/evidence.txt")
				require.Contains(t, text, "evidence.txt")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := newTestServer()
			service := mocks.NewJiraService(t)
			registerRelationshipTools(srv, service)
			registerVersionTools(srv, service)
			registerDevelopmentTools(srv, service)
			registerAttachmentTools(srv, service)
			tt.setup(service)

			result := callTool(t, srv, tt.tool, tt.args)
			tt.assert(t, toolText(t, result))
		})
	}
}

func TestRegisterPromptsHandlers(t *testing.T) {
	t.Parallel()

	srv := newTestServer()
	registerPrompts(srv)

	issueTree := getPrompt(t, srv, "issue_development_tree", map[string]string{"issue_key": "PROJ-123"})
	require.Contains(t, promptText(t, issueTree), "PROJ-123")
	require.Contains(t, promptText(t, issueTree), "jira_get_issue")

	releaseOverview := getPrompt(t, srv, "release_development_overview", map[string]string{"version": "1.0.0", "project_key": "PROJ"})
	require.Contains(t, promptText(t, releaseOverview), "fixVersion = \"1.0.0\"")
	require.Contains(t, promptText(t, releaseOverview), "jira_get_development_information")
}
