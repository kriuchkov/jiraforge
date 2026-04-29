package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-faster/errors"
	markmcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	textfmt "github.com/kriuchkov/jiraforge/internal/adapters/presentation/text"
	"github.com/kriuchkov/jiraforge/internal/core/models"
	"github.com/kriuchkov/jiraforge/internal/core/ports"
)

func registerRelationshipTools(s *server.MCPServer, service ports.RelationshipService) {
	tool := markmcp.NewTool("jira_get_related_issues",
		markmcp.WithDescription("Retrieve issues linked to a Jira issue"),
		markmcp.WithString("issue_key", markmcp.Required(), markmcp.Description("Issue key")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getRelatedIssuesInput) (*markmcp.CallToolResult, error) {
		relations, err := service.GetRelatedIssues(ctx, models.GetRelatedIssuesRequest{IssueKey: input.IssueKey})
		if err != nil {
			return nil, errors.Wrap(err, "get related issues")
		}
		return markmcp.NewToolResultText(textfmt.FormatRelations(input.IssueKey, relations)), nil
	}))

	tool = markmcp.NewTool("jira_link_issues",
		markmcp.WithDescription("Create a relationship between two Jira issues"),
		markmcp.WithString("inward_issue", markmcp.Required(), markmcp.Description("Inward issue key")),
		markmcp.WithString("outward_issue", markmcp.Required(), markmcp.Description("Outward issue key")),
		markmcp.WithString("link_type", markmcp.Required(), markmcp.Description("Link type, for example Blocks or Relates")),
		markmcp.WithString("comment", markmcp.Description("Optional link comment")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input linkIssuesInput) (*markmcp.CallToolResult, error) {
		result, err := service.LinkIssues(ctx, models.LinkIssuesRequest{InwardIssue: input.InwardIssue, OutwardIssue: input.OutwardIssue, LinkType: input.LinkType, Comment: input.Comment})
		if err != nil {
			return nil, errors.Wrap(err, "link issues")
		}
		return markmcp.NewToolResultText(textfmt.FormatMutation(result)), nil
	}))
}

func registerVersionTools(s *server.MCPServer, service ports.VersionService) {
	tool := markmcp.NewTool("jira_get_version",
		markmcp.WithDescription("Retrieve details about a project version"),
		markmcp.WithString("version_id", markmcp.Required(), markmcp.Description("Version ID")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getVersionInput) (*markmcp.CallToolResult, error) {
		version, err := service.GetVersion(ctx, models.GetVersionRequest{VersionID: input.VersionID})
		if err != nil {
			return nil, errors.Wrap(err, "get version")
		}
		return markmcp.NewToolResultText(textfmt.FormatVersion(version)), nil
	}))

	tool = markmcp.NewTool("jira_list_project_versions",
		markmcp.WithDescription("List versions defined in a Jira project"),
		markmcp.WithString("project_key", markmcp.Required(), markmcp.Description("Project key")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input listProjectVersionsInput) (*markmcp.CallToolResult, error) {
		versions, err := service.ListProjectVersions(ctx, models.ListProjectVersionsRequest{ProjectKey: input.ProjectKey})
		if err != nil {
			return nil, errors.Wrap(err, "list project versions")
		}
		return markmcp.NewToolResultText(textfmt.FormatVersions(versions)), nil
	}))
}

func registerDevelopmentTools(s *server.MCPServer, service ports.DevelopmentService) {
	tool := markmcp.NewTool("jira_get_development_information",
		markmcp.WithDescription("Retrieve branches, pull requests, commits, and builds linked to a Jira issue via Jira development integrations"),
		markmcp.WithString("issue_key", markmcp.Required(), markmcp.Description("Issue key")),
		markmcp.WithBoolean("include_branches", markmcp.Description("Include branches in the response. Defaults to true")),
		markmcp.WithBoolean("include_pull_requests", markmcp.Description("Include pull requests in the response. Defaults to true")),
		markmcp.WithBoolean("include_commits", markmcp.Description("Include commits in the response. Defaults to true")),
		markmcp.WithBoolean("include_builds", markmcp.Description("Include builds in the response. Defaults to true")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getDevelopmentInfoInput) (*markmcp.CallToolResult, error) {
		info, err := service.GetDevelopmentInfo(ctx, models.GetDevelopmentInfoRequest{
			IssueKey:            input.IssueKey,
			IncludeBranches:     input.IncludeBranches,
			IncludePullRequests: input.IncludePullRequests,
			IncludeCommits:      input.IncludeCommits,
			IncludeBuilds:       input.IncludeBuilds,
		})
		if err != nil {
			return nil, errors.Wrap(err, "get development info")
		}
		return markmcp.NewToolResultText(textfmt.FormatDevelopmentInfo(info)), nil
	}))
}

func registerAttachmentTools(s *server.MCPServer, service ports.AttachmentService) {
	tool := markmcp.NewTool("jira_download_attachment",
		markmcp.WithDescription("Download a Jira attachment into a local temporary file and return the absolute file path"),
		markmcp.WithString("attachment_id", markmcp.Required(), markmcp.Description("Attachment ID")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input downloadAttachmentInput) (*markmcp.CallToolResult, error) {
		attachment, err := service.DownloadAttachment(ctx, models.DownloadAttachmentRequest{AttachmentID: input.AttachmentID})
		if err != nil {
			return nil, errors.Wrap(err, "download attachment")
		}
		return markmcp.NewToolResultText(textfmt.FormatStoredAttachment(attachment)), nil
	}))
}

func registerPrompts(s *server.MCPServer) {
	s.AddPrompt(markmcp.NewPrompt("issue_development_tree",
		markmcp.WithPromptDescription("List all development work for an issue and its subtasks"),
		markmcp.WithArgument("issue_key", markmcp.ArgumentDescription("The Jira issue key, for example PROJ-123"), markmcp.RequiredArgument()),
	), func(_ context.Context, request markmcp.GetPromptRequest) (*markmcp.GetPromptResult, error) {
		issueKey := request.Params.Arguments["issue_key"]
		if strings.TrimSpace(issueKey) == "" {
			return nil, fmt.Errorf("issue_key is required")
		}
		return markmcp.NewGetPromptResult(
			"Development work tree for issue and subtasks",
			[]markmcp.PromptMessage{
				markmcp.NewPromptMessage(markmcp.RoleUser, markmcp.NewTextContent(fmt.Sprintf(`Please analyze all development work for issue %s and its child issues:

1. Use jira_get_issue with issue_key=%s and expand=subtasks to retrieve the parent issue and its subtasks.
2. Use jira_get_development_information for the parent issue %s.
3. For each discovered subtask, call jira_get_development_information as well.
4. Present the result as a hierarchy with the parent issue first and each subtask below it.

Provide a concise summary of the full development tree.`, issueKey, issueKey, issueKey))),
			},
		), nil
	})

	s.AddPrompt(markmcp.NewPrompt("release_development_overview",
		markmcp.WithPromptDescription("List issues and related development work for a specific release"),
		markmcp.WithArgument("version", markmcp.ArgumentDescription("Version or release name"), markmcp.RequiredArgument()),
		markmcp.WithArgument("project_key", markmcp.ArgumentDescription("Jira project key"), markmcp.RequiredArgument()),
	), func(_ context.Context, request markmcp.GetPromptRequest) (*markmcp.GetPromptResult, error) {
		version := request.Params.Arguments["version"]
		projectKey := request.Params.Arguments["project_key"]
		if strings.TrimSpace(version) == "" || strings.TrimSpace(projectKey) == "" {
			return nil, fmt.Errorf("version and project_key are required")
		}
		return markmcp.NewGetPromptResult(
			"Development overview for release",
			[]markmcp.PromptMessage{
				markmcp.NewPromptMessage(markmcp.RoleUser, markmcp.NewTextContent(fmt.Sprintf(`Please prepare a development overview for release %q in project %s:

1. Use jira_search_issue with the JQL: fixVersion = %q AND project = %s.
2. For every issue in the result set, call jira_get_development_information.
3. Summarize issue status, pull request state, commit coverage, and build state.

Format the result so it is easy to review release readiness.`, version, projectKey, version, projectKey))),
			},
		), nil
	})
}
