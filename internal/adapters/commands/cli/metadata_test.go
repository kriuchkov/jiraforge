package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetadataCommandValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		wantError string
	}{
		{name: "get related issues", args: []string{"get-related-issues"}, wantError: "--issue-key is required"},
		{name: "link issues", args: []string{"link-issues"}, wantError: "--inward-issue, --outward-issue, and --link-type are required"},
		{name: "get version", args: []string{"get-version"}, wantError: "--version-id is required"},
		{name: "list project versions", args: []string{"list-project-versions"}, wantError: "--project-key is required"},
		{name: "get development info", args: []string{"get-development-info"}, wantError: "--issue-key is required"},
		{name: "download attachment", args: []string{"download-attachment"}, wantError: "--attachment-id is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertValidationError(t, tt.args, tt.wantError)
		})
	}
}

func TestMetadataCommandDispatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		service *stubCLIService
		assert  func(*testing.T, *stubCLIService, string, string)
	}{
		{
			name:    "get related issues",
			args:    []string{"get-related-issues", "--issue-key", "PROJ-7"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "GetRelatedIssues", service.lastCall)
				require.Equal(t, "PROJ-7", service.getRelatedIssuesReq.IssueKey)
			},
		},
		{
			name:    "link issues",
			args:    []string{"link-issues", "--inward-issue", "PROJ-1", "--outward-issue", "PROJ-2", "--link-type", "blocks", "--comment", "linked"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "LinkIssues", service.lastCall)
				require.Equal(t, "PROJ-1", service.linkIssuesReq.InwardIssue)
				require.Equal(t, "PROJ-2", service.linkIssuesReq.OutwardIssue)
				require.Equal(t, "blocks", service.linkIssuesReq.LinkType)
				require.Equal(t, "linked", service.linkIssuesReq.Comment)
			},
		},
		{
			name:    "get version",
			args:    []string{"get-version", "--version-id", "1001"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "GetVersion", service.lastCall)
				require.Equal(t, "1001", service.getVersionReq.VersionID)
			},
		},
		{
			name:    "list project versions",
			args:    []string{"list-project-versions", "--project-key", "PROJ"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "ListProjectVersions", service.lastCall)
				require.Equal(t, "PROJ", service.listProjectVersionsReq.ProjectKey)
			},
		},
		{
			name:    "get development info",
			args:    []string{"get-development-info", "--issue-key", "PROJ-9", "--include-branches", "--include-pull-requests", "--include-commits", "--include-builds"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "GetDevelopmentInfo", service.lastCall)
				require.Equal(t, "PROJ-9", service.getDevelopmentInfoReq.IssueKey)
				require.True(t, service.getDevelopmentInfoReq.IncludeBranches)
				require.True(t, service.getDevelopmentInfoReq.IncludePullRequests)
				require.True(t, service.getDevelopmentInfoReq.IncludeCommits)
				require.True(t, service.getDevelopmentInfoReq.IncludeBuilds)
			},
		},
		{
			name:    "download attachment",
			args:    []string{"download-attachment", "--attachment-id", "att-1"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "DownloadAttachment", service.lastCall)
				require.Equal(t, "att-1", service.downloadAttachmentReq.AttachmentID)
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
