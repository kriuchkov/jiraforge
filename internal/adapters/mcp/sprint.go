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

func registerSprintTools(s *server.MCPServer, service ports.SprintService) {
	tool := markmcp.NewTool("jira_list_sprints",
		markmcp.WithDescription("List active and future sprints for a board or project"),
		markmcp.WithString("board_id", markmcp.Description("Board ID")),
		markmcp.WithString("project_key", markmcp.Description("Project key")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input listSprintsInput) (*markmcp.CallToolResult, error) {
		result, err := service.ListSprints(ctx, models.ListSprintsRequest{BoardID: input.BoardID, ProjectKey: input.ProjectKey})
		if err != nil {
			return nil, errors.Wrap(err, "list sprints")
		}
		return markmcp.NewToolResultText(textfmt.FormatSprints(result)), nil
	}))

	tool = markmcp.NewTool("jira_get_sprint",
		markmcp.WithDescription("Retrieve detailed information about a specific sprint"),
		markmcp.WithString("sprint_id", markmcp.Required(), markmcp.Description("Sprint ID")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getSprintInput) (*markmcp.CallToolResult, error) {
		sprint, err := service.GetSprint(ctx, models.GetSprintRequest{SprintID: input.SprintID})
		if err != nil {
			return nil, errors.Wrap(err, "get sprint")
		}
		return markmcp.NewToolResultText(textfmt.FormatSprint(sprint)), nil
	}))

	tool = markmcp.NewTool("jira_get_sprint_report",
		markmcp.WithDescription("Retrieve the sprint report with completed, incomplete, and removed issues for a sprint"),
		markmcp.WithString("sprint_id", markmcp.Required(), markmcp.Description("Sprint ID")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getSprintReportInput) (*markmcp.CallToolResult, error) {
		report, err := service.GetSprintReport(ctx, models.GetSprintReportRequest{SprintID: input.SprintID})
		if err != nil {
			return nil, errors.Wrap(err, "get sprint report")
		}
		return markmcp.NewToolResultText(textfmt.FormatSprintReport(report)), nil
	}))

	tool = markmcp.NewTool("jira_get_sprint_health_report",
		markmcp.WithDescription("Retrieve a management-focused sprint health summary for a sprint"),
		markmcp.WithString("sprint_id", markmcp.Required(), markmcp.Description("Sprint ID")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getSprintHealthReportInput) (*markmcp.CallToolResult, error) {
		report, err := service.GetSprintHealthReport(ctx, models.GetSprintHealthReportRequest{SprintID: input.SprintID})
		if err != nil {
			return nil, errors.Wrap(err, "get sprint health report")
		}
		return markmcp.NewToolResultText(textfmt.FormatSprintHealthReport(report)), nil
	}))

	tool = markmcp.NewTool("jira_get_active_sprint",
		markmcp.WithDescription("Get the active sprint for a board or project"),
		markmcp.WithString("board_id", markmcp.Description("Board ID")),
		markmcp.WithString("project_key", markmcp.Description("Project key")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getActiveSprintInput) (*markmcp.CallToolResult, error) {
		sprint, err := service.GetActiveSprint(ctx, models.GetActiveSprintRequest{BoardID: input.BoardID, ProjectKey: input.ProjectKey})
		if err != nil {
			return nil, errors.Wrap(err, "get active sprint")
		}
		return markmcp.NewToolResultText(textfmt.FormatSprint(sprint)), nil
	}))

	tool = markmcp.NewTool("jira_search_sprint_by_name",
		markmcp.WithDescription("Search for sprints by name across one board or project"),
		markmcp.WithString("name", markmcp.Required(), markmcp.Description("Sprint name to search for")),
		markmcp.WithString("board_id", markmcp.Description("Board ID")),
		markmcp.WithString("project_key", markmcp.Description("Project key")),
		markmcp.WithBoolean("exact_match", markmcp.Description("Whether the sprint name must match exactly")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input searchSprintsInput) (*markmcp.CallToolResult, error) {
		result, err := service.SearchSprints(ctx, models.SearchSprintsRequest{Name: input.Name, BoardID: input.BoardID, ProjectKey: input.ProjectKey, ExactMatch: input.ExactMatch})
		if err != nil {
			return nil, errors.Wrap(err, "search sprints")
		}
		return markmcp.NewToolResultText(textfmt.FormatSprints(result)), nil
	}))
}
