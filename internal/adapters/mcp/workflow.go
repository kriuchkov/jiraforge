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

func registerCommentTools(s *server.MCPServer, service ports.CommentService) {
	tool := markmcp.NewTool("jira_add_comment",
		markmcp.WithDescription("Add a comment to a Jira issue"),
		markmcp.WithString("issue_key", markmcp.Required(), markmcp.Description("Issue key")),
		markmcp.WithString("comment", markmcp.Required(), markmcp.Description("Comment body in Markdown")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input addCommentInput) (*markmcp.CallToolResult, error) {
		comment, err := service.AddComment(ctx, models.AddCommentRequest{IssueKey: input.IssueKey, Comment: input.Comment})
		if err != nil {
			return nil, errors.Wrap(err, "add comment")
		}
		return markmcp.NewToolResultText(textfmt.FormatComment(comment)), nil
	}))

	tool = markmcp.NewTool("jira_get_comments",
		markmcp.WithDescription("Retrieve comments from a Jira issue"),
		markmcp.WithString("issue_key", markmcp.Required(), markmcp.Description("Issue key")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getCommentsInput) (*markmcp.CallToolResult, error) {
		comments, err := service.GetComments(ctx, models.GetCommentsRequest{IssueKey: input.IssueKey})
		if err != nil {
			return nil, errors.Wrap(err, "get comments")
		}
		return markmcp.NewToolResultText(textfmt.FormatComments(comments)), nil
	}))
}

func registerWorklogTools(s *server.MCPServer, service ports.WorklogService) {
	tool := markmcp.NewTool("jira_add_worklog",
		markmcp.WithDescription("Add a worklog to a Jira issue"),
		markmcp.WithString("issue_key", markmcp.Required(), markmcp.Description("Issue key")),
		markmcp.WithString("time_spent", markmcp.Required(), markmcp.Description("Time spent, for example 1h30m or 30m")),
		markmcp.WithString("comment", markmcp.Description("Optional worklog comment in Markdown")),
		markmcp.WithString("started", markmcp.Description("Optional started timestamp in Jira format")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input addWorklogInput) (*markmcp.CallToolResult, error) {
		worklog, err := service.AddWorklog(ctx, models.AddWorklogRequest{IssueKey: input.IssueKey, TimeSpent: input.TimeSpent, Comment: input.Comment, Started: input.Started})
		if err != nil {
			return nil, errors.Wrap(err, "add worklog")
		}
		return markmcp.NewToolResultText(textfmt.FormatWorklog(worklog)), nil
	}))

	tool = markmcp.NewTool("jira_get_worklogs",
		markmcp.WithDescription("List worklogs recorded against a Jira issue"),
		markmcp.WithString("issue_key", markmcp.Required(), markmcp.Description("Issue key")),
		markmcp.WithString("started_after", markmcp.Description("Only include worklogs started at or after this timestamp (RFC3339 or YYYY-MM-DD)")),
		markmcp.WithNumber("max_results", markmcp.Description("Maximum number of worklogs to return")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getWorklogsInput) (*markmcp.CallToolResult, error) {
		worklogs, err := service.GetWorklogs(ctx, models.GetWorklogsRequest{IssueKey: input.IssueKey, StartedAfter: input.StartedAfter, MaxResults: input.MaxResults})
		if err != nil {
			return nil, errors.Wrap(err, "get worklogs")
		}
		return markmcp.NewToolResultText(textfmt.FormatWorklogList(worklogs)), nil
	}))
}

func registerTransitionTools(s *server.MCPServer, service ports.WorkflowService) {
	tool := markmcp.NewTool("jira_get_transitions",
		markmcp.WithDescription("List available workflow transitions for a Jira issue"),
		markmcp.WithString("issue_key", markmcp.Required(), markmcp.Description("Issue key")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getTransitionsInput) (*markmcp.CallToolResult, error) {
		transitions, err := service.GetTransitions(ctx, models.GetTransitionsRequest{IssueKey: input.IssueKey})
		if err != nil {
			return nil, errors.Wrap(err, "get transitions")
		}
		return markmcp.NewToolResultText(textfmt.FormatTransitions(transitions)), nil
	}))

	tool = markmcp.NewTool("jira_transition_issue",
		markmcp.WithDescription("Transition an issue through its workflow using a valid transition ID"),
		markmcp.WithString("issue_key", markmcp.Required(), markmcp.Description("Issue key")),
		markmcp.WithString("transition_id", markmcp.Required(), markmcp.Description("Transition ID")),
		markmcp.WithString("comment", markmcp.Description("Optional transition comment")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input transitionIssueInput) (*markmcp.CallToolResult, error) {
		result, err := service.TransitionIssue(ctx, models.TransitionIssueRequest{IssueKey: input.IssueKey, TransitionID: input.TransitionID, Comment: input.Comment})
		if err != nil {
			return nil, errors.Wrap(err, "transition issue")
		}
		return markmcp.NewToolResultText(textfmt.FormatMutation(result)), nil
	}))
}

func registerStatusTools(s *server.MCPServer, service ports.WorkflowService) {
	tool := markmcp.NewTool("jira_list_statuses",
		markmcp.WithDescription("Retrieve statuses available in a Jira project, grouped by issue type"),
		markmcp.WithString("project_key", markmcp.Required(), markmcp.Description("Project key")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input listStatusesInput) (*markmcp.CallToolResult, error) {
		statuses, err := service.ListStatuses(ctx, models.ListStatusesRequest{ProjectKey: input.ProjectKey})
		if err != nil {
			return nil, errors.Wrap(err, "list statuses")
		}
		return markmcp.NewToolResultText(textfmt.FormatStatuses(statuses)), nil
	}))
}

func registerHistoryTools(s *server.MCPServer, service ports.WorkflowService) {
	tool := markmcp.NewTool("jira_get_issue_history",
		markmcp.WithDescription("Retrieve the change history of a Jira issue"),
		markmcp.WithString("issue_key", markmcp.Required(), markmcp.Description("Issue key")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getIssueHistoryInput) (*markmcp.CallToolResult, error) {
		history, err := service.GetIssueHistory(ctx, models.GetIssueHistoryRequest{IssueKey: input.IssueKey})
		if err != nil {
			return nil, errors.Wrap(err, "get issue history")
		}
		return markmcp.NewToolResultText(textfmt.FormatHistory(history)), nil
	}))
}
