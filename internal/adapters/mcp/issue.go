package mcp

import (
	"context"

	"github.com/go-faster/errors"
	markmcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	textfmt "github.com/kriuchkov/jiraforge/internal/adapters/presentation/text"
	"github.com/kriuchkov/jiraforge/internal/core/models"
	"github.com/kriuchkov/jiraforge/internal/core/ports"
)

type issueService interface {
	ports.IssueQueryService
	ports.IssueMutationService
}

func registerIssueTools(s *server.MCPServer, service issueService) {
	tool := markmcp.NewTool("jira_get_issue",
		markmcp.WithDescription("Retrieve detailed information about a specific Jira issue including status, assignee, description, subtasks, attachments, and available transitions"),
		markmcp.WithString("issue_key", markmcp.Required(), markmcp.Description("The Jira issue key, for example PROJ-123")),
		markmcp.WithString("fields", markmcp.Description("Optional comma-separated list of fields to retrieve")),
		markmcp.WithString("expand", markmcp.Description("Optional comma-separated expansions. Defaults to transitions,changelog,subtasks,description")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getIssueInput) (*markmcp.CallToolResult, error) {
		issue, err := service.GetIssue(ctx, models.GetIssueRequest{IssueKey: input.IssueKey, Fields: splitCSV(input.Fields), Expand: splitCSV(input.Expand)})
		if err != nil {
			return nil, errors.Wrap(err, "get issue")
		}
		return markmcp.NewToolResultText(textfmt.FormatIssue(issue)), nil
	}))

	tool = markmcp.NewTool("jira_create_issue",
		markmcp.WithDescription("Create a new Jira issue with summary, description, and issue type"),
		markmcp.WithString("project_key", markmcp.Required(), markmcp.Description("Project key where the issue will be created")),
		markmcp.WithString("summary", markmcp.Required(), markmcp.Description("Issue summary")),
		markmcp.WithString("description", markmcp.Required(), markmcp.Description("Issue description in Markdown")),
		markmcp.WithString("issue_type", markmcp.Required(), markmcp.Description("Issue type, for example Bug, Task, Story")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input createIssueInput) (*markmcp.CallToolResult, error) {
		result, err := service.CreateIssue(ctx, models.CreateIssueRequest{ProjectKey: input.ProjectKey, Summary: input.Summary, Description: input.Description, IssueType: input.IssueType})
		if err != nil {
			return nil, errors.Wrap(err, "create issue")
		}
		return markmcp.NewToolResultText(textfmt.FormatMutation(result)), nil
	}))

	tool = markmcp.NewTool("jira_create_child_issue",
		markmcp.WithDescription("Create a child issue linked to a parent issue"),
		markmcp.WithString("parent_issue_key", markmcp.Required(), markmcp.Description("Parent issue key")),
		markmcp.WithString("summary", markmcp.Required(), markmcp.Description("Child issue summary")),
		markmcp.WithString("description", markmcp.Required(), markmcp.Description("Child issue description in Markdown")),
		markmcp.WithString("issue_type", markmcp.Description("Optional child issue type. Defaults to Subtask")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input createChildIssueInput) (*markmcp.CallToolResult, error) {
		result, err := service.CreateChildIssue(ctx, models.CreateChildIssueRequest{ParentIssueKey: input.ParentIssueKey, Summary: input.Summary, Description: input.Description, IssueType: input.IssueType})
		if err != nil {
			return nil, errors.Wrap(err, "create child issue")
		}
		return markmcp.NewToolResultText(textfmt.FormatMutation(result)), nil
	}))

	tool = markmcp.NewTool("jira_update_issue",
		markmcp.WithDescription("Update a Jira issue. Only specified fields are changed"),
		markmcp.WithString("issue_key", markmcp.Required(), markmcp.Description("Issue key")),
		markmcp.WithString("summary", markmcp.Description("Updated summary")),
		markmcp.WithString("description", markmcp.Description("Updated description in Markdown")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input updateIssueInput) (*markmcp.CallToolResult, error) {
		result, err := service.UpdateIssue(ctx, models.UpdateIssueRequest{IssueKey: input.IssueKey, Summary: input.Summary, Description: input.Description})
		if err != nil {
			return nil, errors.Wrap(err, "update issue")
		}
		return markmcp.NewToolResultText(textfmt.FormatMutation(result)), nil
	}))

	tool = markmcp.NewTool("jira_delete_issue",
		markmcp.WithDescription("Delete a Jira issue permanently"),
		markmcp.WithString("issue_key", markmcp.Required(), markmcp.Description("Issue key to delete")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input deleteIssueInput) (*markmcp.CallToolResult, error) {
		result, err := service.DeleteIssue(ctx, models.DeleteIssueRequest{IssueKey: input.IssueKey})
		if err != nil {
			return nil, errors.Wrap(err, "delete issue")
		}
		return markmcp.NewToolResultText(textfmt.FormatMutation(result)), nil
	}))

	tool = markmcp.NewTool("jira_list_issue_types",
		markmcp.WithDescription("List the available issue types for a Jira project"),
		markmcp.WithString("project_key", markmcp.Required(), markmcp.Description("Project key")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input listIssueTypesInput) (*markmcp.CallToolResult, error) {
		issueTypes, err := service.ListIssueTypes(ctx, models.ListIssueTypesRequest{ProjectKey: input.ProjectKey})
		if err != nil {
			return nil, errors.Wrap(err, "list issue types")
		}
		return markmcp.NewToolResultText(textfmt.FormatIssueTypes(issueTypes)), nil
	}))
}

func registerSearchTools(s *server.MCPServer, service ports.IssueQueryService) {
	tool := markmcp.NewTool("jira_search_issue",
		markmcp.WithDescription("Search for Jira issues using JQL"),
		markmcp.WithString("jql", markmcp.Required(), markmcp.Description("JQL query string")),
		markmcp.WithString("fields", markmcp.Description("Optional comma-separated list of fields to retrieve")),
		markmcp.WithString("expand", markmcp.Description("Optional comma-separated expansions")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input searchIssuesInput) (*markmcp.CallToolResult, error) {
		result, err := service.SearchIssues(ctx, models.SearchIssuesRequest{JQL: input.JQL, Fields: splitCSV(input.Fields), Expand: splitCSV(input.Expand)})
		if err != nil {
			return nil, errors.Wrap(err, "search issues")
		}
		return markmcp.NewToolResultText(textfmt.FormatSearchIssues(result)), nil
	}))
}
