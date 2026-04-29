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

func registerReportTools(s *server.MCPServer, service ports.ReportService) {
	tool := markmcp.NewTool("jira_get_aging_report",
		markmcp.WithDescription("Retrieve issues ordered by how long they have remained in their current status"),
		markmcp.WithString("project_key", markmcp.Description("Project key to analyze when JQL is not provided")),
		markmcp.WithString("jql", markmcp.Description("Custom JQL query to analyze instead of project/status filters")),
		markmcp.WithString("statuses", markmcp.Description("Comma-separated statuses to include")),
		markmcp.WithString("assignee", markmcp.Description("Assignee to filter by when building JQL")),
		markmcp.WithNumber("min_days_in_status", markmcp.Description("Only include issues that have stayed at least this many days in the same status")),
		markmcp.WithNumber("max_results", markmcp.Description("Maximum number of issues to analyze")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getAgingReportInput) (*markmcp.CallToolResult, error) {
		report, err := service.GetAgingReport(ctx, models.GetAgingReportRequest{
			ProjectKey:      input.ProjectKey,
			JQL:             input.JQL,
			Statuses:        splitCSV(input.Statuses),
			Assignee:        input.Assignee,
			MinDaysInStatus: input.MinDaysInStatus,
			MaxResults:      input.MaxResults,
		})
		if err != nil {
			return nil, errors.Wrap(err, "get aging report")
		}
		return markmcp.NewToolResultText(textfmt.FormatAgingReport(report)), nil
	}))

	tool = markmcp.NewTool("jira_get_blocked_issues_report",
		markmcp.WithDescription("Retrieve issues that appear blocked by status or dependency links"),
		markmcp.WithString("project_key", markmcp.Description("Project key to analyze when JQL is not provided")),
		markmcp.WithString("jql", markmcp.Description("Custom JQL query to analyze instead of project/status filters")),
		markmcp.WithString("statuses", markmcp.Description("Comma-separated statuses to include")),
		markmcp.WithString("blocked_statuses", markmcp.Description("Comma-separated statuses treated as blocked")),
		markmcp.WithString("link_types", markmcp.Description("Comma-separated link types treated as blocking")),
		markmcp.WithString("assignee", markmcp.Description("Assignee to filter by when building JQL")),
		markmcp.WithNumber("min_days_in_status", markmcp.Description("Only include issues that have stayed at least this many days in the same status")),
		markmcp.WithNumber("max_results", markmcp.Description("Maximum number of issues to analyze")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getBlockedIssuesReportInput) (*markmcp.CallToolResult, error) {
		report, err := service.GetBlockedIssuesReport(ctx, models.GetBlockedIssuesReportRequest{
			ProjectKey:      input.ProjectKey,
			JQL:             input.JQL,
			Statuses:        splitCSV(input.Statuses),
			BlockedStatuses: splitCSV(input.BlockedStatuses),
			LinkTypes:       splitCSV(input.LinkTypes),
			Assignee:        input.Assignee,
			MinDaysInStatus: input.MinDaysInStatus,
			MaxResults:      input.MaxResults,
		})
		if err != nil {
			return nil, errors.Wrap(err, "get blocked issues report")
		}
		return markmcp.NewToolResultText(textfmt.FormatBlockedIssuesReport(report)), nil
	}))

	tool = markmcp.NewTool("jira_get_flow_efficiency_report",
		markmcp.WithDescription("Retrieve active-versus-blocked flow metrics derived from issue history"),
		markmcp.WithString("project_key", markmcp.Description("Project key to analyze when JQL is not provided")),
		markmcp.WithString("jql", markmcp.Description("Custom JQL query to analyze instead of project/status filters")),
		markmcp.WithString("statuses", markmcp.Description("Comma-separated current statuses to include")),
		markmcp.WithString("active_statuses", markmcp.Description("Comma-separated statuses treated as active work")),
		markmcp.WithString("blocked_statuses", markmcp.Description("Comma-separated statuses treated as blocked work")),
		markmcp.WithString("assignee", markmcp.Description("Assignee to filter by when building JQL")),
		markmcp.WithNumber("min_blocked_days", markmcp.Description("Only include issues with at least this many blocked days")),
		markmcp.WithString("start_date", markmcp.Description("Explicit range start for active and blocked time, in RFC3339 or YYYY-MM-DD")),
		markmcp.WithString("end_date", markmcp.Description("Explicit range end for active and blocked time, in RFC3339 or YYYY-MM-DD")),
		markmcp.WithNumber("window_days", markmcp.Description("Only count active and blocked time from the last N days")),
		markmcp.WithNumber("max_results", markmcp.Description("Maximum number of issues to analyze")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getFlowEfficiencyReportInput) (*markmcp.CallToolResult, error) {
		report, err := service.GetFlowEfficiencyReport(ctx, models.GetFlowEfficiencyReportRequest{
			ProjectKey:      input.ProjectKey,
			JQL:             input.JQL,
			Statuses:        splitCSV(input.Statuses),
			ActiveStatuses:  splitCSV(input.ActiveStatuses),
			BlockedStatuses: splitCSV(input.BlockedStatuses),
			Assignee:        input.Assignee,
			MinBlockedDays:  input.MinBlockedDays,
			StartDate:       input.StartDate,
			EndDate:         input.EndDate,
			WindowDays:      input.WindowDays,
			MaxResults:      input.MaxResults,
		})
		if err != nil {
			return nil, errors.Wrap(err, "get flow efficiency report")
		}
		return markmcp.NewToolResultText(textfmt.FormatFlowEfficiencyReport(report)), nil
	}))

	tool = markmcp.NewTool("jira_get_sprint_workload_report",
		markmcp.WithDescription("Per-assignee breakdown of issues, story points, and statuses for a sprint"),
		markmcp.WithString("sprint_id", markmcp.Required(), markmcp.Description("Sprint identifier")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getSprintWorkloadReportInput) (*markmcp.CallToolResult, error) {
		report, err := service.GetSprintWorkloadReport(ctx, models.GetSprintWorkloadReportRequest{SprintID: input.SprintID})
		if err != nil {
			return nil, errors.Wrap(err, "get sprint workload report")
		}
		return markmcp.NewToolResultText(textfmt.FormatSprintWorkloadReport(report)), nil
	}))

	tool = markmcp.NewTool("jira_get_utilization_report",
		markmcp.WithDescription("Aggregate worklog activity per assignee within a time window"),
		markmcp.WithString("project_key", markmcp.Description("Project key when JQL is not provided")),
		markmcp.WithString("jql", markmcp.Description("Custom JQL query for issues whose worklogs will be aggregated")),
		markmcp.WithString("assignee", markmcp.Description("Assignee to filter worklog authors")),
		markmcp.WithString("start_date", markmcp.Description("Window start, RFC3339 or YYYY-MM-DD")),
		markmcp.WithString("end_date", markmcp.Description("Window end, RFC3339 or YYYY-MM-DD")),
		markmcp.WithNumber("window_days", markmcp.Description("Use last N days when start_date is not provided (default 7)")),
		markmcp.WithNumber("max_results", markmcp.Description("Maximum number of issues to scan for worklogs")),
		markmcp.WithNumber("top_issues_per_user", markmcp.Description("Limit how many top issues are returned per assignee")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getUtilizationReportInput) (*markmcp.CallToolResult, error) {
		report, err := service.GetUtilizationReport(ctx, models.GetUtilizationReportRequest{
			ProjectKey:       input.ProjectKey,
			JQL:              input.JQL,
			Assignee:         input.Assignee,
			StartDate:        input.StartDate,
			EndDate:          input.EndDate,
			WindowDays:       input.WindowDays,
			MaxResults:       input.MaxResults,
			TopIssuesPerUser: input.TopIssuesPerUser,
		})
		if err != nil {
			return nil, errors.Wrap(err, "get utilization report")
		}
		return markmcp.NewToolResultText(textfmt.FormatUtilizationReport(report)), nil
	}))

	tool = markmcp.NewTool("jira_get_cycle_time_report",
		markmcp.WithDescription("Cycle time, lead time, and throughput metrics derived from issue history"),
		markmcp.WithString("project_key", markmcp.Description("Project key when JQL is not provided")),
		markmcp.WithString("jql", markmcp.Description("Custom JQL query for completed issues to analyze")),
		markmcp.WithString("start_statuses", markmcp.Description("Comma-separated statuses that mark active work has started (default: In Progress)")),
		markmcp.WithString("done_statuses", markmcp.Description("Comma-separated statuses that mark completion (default: Done, Closed, Resolved)")),
		markmcp.WithString("assignee", markmcp.Description("Assignee to filter by when building JQL")),
		markmcp.WithString("issue_type", markmcp.Description("Issue type to filter by when building JQL")),
		markmcp.WithString("start_date", markmcp.Description("Window start, RFC3339 or YYYY-MM-DD")),
		markmcp.WithString("end_date", markmcp.Description("Window end, RFC3339 or YYYY-MM-DD")),
		markmcp.WithNumber("window_days", markmcp.Description("Use last N days when start_date is not provided (default 30)")),
		markmcp.WithNumber("max_results", markmcp.Description("Maximum number of issues to analyze")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getCycleTimeReportInput) (*markmcp.CallToolResult, error) {
		report, err := service.GetCycleTimeReport(ctx, models.GetCycleTimeReportRequest{
			ProjectKey:    input.ProjectKey,
			JQL:           input.JQL,
			StartStatuses: splitCSV(input.StartStatuses),
			DoneStatuses:  splitCSV(input.DoneStatuses),
			Assignee:      input.Assignee,
			IssueType:     input.IssueType,
			StartDate:     input.StartDate,
			EndDate:       input.EndDate,
			WindowDays:    input.WindowDays,
			MaxResults:    input.MaxResults,
		})
		if err != nil {
			return nil, errors.Wrap(err, "get cycle time report")
		}
		return markmcp.NewToolResultText(textfmt.FormatCycleTimeReport(report)), nil
	}))

	tool = markmcp.NewTool("jira_get_team_wip_snapshot",
		markmcp.WithDescription("Per-assignee snapshot of in-flight work grouped by status"),
		markmcp.WithString("project_key", markmcp.Description("Project key when JQL is not provided")),
		markmcp.WithString("jql", markmcp.Description("Custom JQL query overriding default in-progress filter")),
		markmcp.WithString("statuses", markmcp.Description("Comma-separated statuses to include (default: statusCategory=In Progress)")),
		markmcp.WithString("assignees", markmcp.Description("Comma-separated assignees to include")),
		markmcp.WithNumber("wip_limit", markmcp.Description("Optional WIP limit per assignee; values above are flagged")),
		markmcp.WithNumber("max_results", markmcp.Description("Maximum number of issues to scan")),
	)
	s.AddTool(tool, markmcp.NewTypedToolHandler(func(ctx context.Context, _ markmcp.CallToolRequest, input getTeamWipSnapshotInput) (*markmcp.CallToolResult, error) {
		snapshot, err := service.GetTeamWipSnapshot(ctx, models.GetTeamWipSnapshotRequest{
			ProjectKey: input.ProjectKey,
			JQL:        input.JQL,
			Statuses:   splitCSV(input.Statuses),
			Assignees:  splitCSV(input.Assignees),
			WipLimit:   input.WipLimit,
			MaxResults: input.MaxResults,
		})
		if err != nil {
			return nil, errors.Wrap(err, "get team wip snapshot")
		}
		return markmcp.NewToolResultText(textfmt.FormatTeamWipSnapshot(snapshot)), nil
	}))
}
