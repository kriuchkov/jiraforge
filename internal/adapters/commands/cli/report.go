package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-faster/errors"
	"github.com/spf13/cobra"

	textfmt "github.com/kriuchkov/jiraforge/internal/adapters/presentation/text"
	"github.com/kriuchkov/jiraforge/internal/core/models"
)

func (r runner) newGetAgingReportCommand(state *commandState) *cobra.Command {
	var projectKey, jql, assignee string
	var statuses []string
	var minDaysInStatus, maxResults int

	cmd := r.newCommand(state, "get-aging-report", "Get issues ordered by how long they have stayed in their current status", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(projectKey) == "" && strings.TrimSpace(jql) == "" {
			return nil, "", fmt.Errorf("either --project-key or --jql is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		report, err := service.GetAgingReport(ctx, models.GetAgingReportRequest{
			ProjectKey:      projectKey,
			JQL:             jql,
			Statuses:        append([]string(nil), statuses...),
			Assignee:        assignee,
			MinDaysInStatus: minDaysInStatus,
			MaxResults:      maxResults,
		})
		if err != nil {
			return nil, "", errors.Wrap(err, "get aging report")
		}
		return report, textfmt.FormatAgingReport(report), nil
	})

	cmd.Flags().StringVar(&projectKey, "project-key", "", "Project key")
	cmd.Flags().StringVar(&jql, "jql", "", "Custom JQL query")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "Status to include; repeat the flag to pass multiple values")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Assignee to filter by")
	cmd.Flags().IntVar(&minDaysInStatus, "min-days-in-status", 0, "Only include issues that have stayed at least this many days in their current status")
	cmd.Flags().IntVar(&maxResults, "max-results", 30, "Maximum number of issues to analyze")
	return cmd
}

func (r runner) newGetBlockedIssuesReportCommand(state *commandState) *cobra.Command {
	var projectKey, jql, assignee string
	var statuses []string
	var blockedStatuses []string
	var linkTypes []string
	var minDaysInStatus, maxResults int

	cmd := r.newCommand(state, "get-blocked-issues-report", "Get issues that appear blocked by status or dependency links", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(projectKey) == "" && strings.TrimSpace(jql) == "" {
			return nil, "", fmt.Errorf("either --project-key or --jql is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		report, err := service.GetBlockedIssuesReport(ctx, models.GetBlockedIssuesReportRequest{
			ProjectKey:      projectKey,
			JQL:             jql,
			Statuses:        append([]string(nil), statuses...),
			BlockedStatuses: append([]string(nil), blockedStatuses...),
			LinkTypes:       append([]string(nil), linkTypes...),
			Assignee:        assignee,
			MinDaysInStatus: minDaysInStatus,
			MaxResults:      maxResults,
		})
		if err != nil {
			return nil, "", errors.Wrap(err, "get blocked issues report")
		}
		return report, textfmt.FormatBlockedIssuesReport(report), nil
	})

	cmd.Flags().StringVar(&projectKey, "project-key", "", "Project key")
	cmd.Flags().StringVar(&jql, "jql", "", "Custom JQL query")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "Status to include; repeat the flag to pass multiple values")
	cmd.Flags().StringArrayVar(&blockedStatuses, "blocked-status", nil, "Status treated as blocked; repeat the flag to pass multiple values")
	cmd.Flags().StringArrayVar(&linkTypes, "link-type", nil, "Link type treated as blocking; repeat the flag to pass multiple values")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Assignee to filter by")
	cmd.Flags().IntVar(&minDaysInStatus, "min-days-in-status", 0, "Only include issues that have stayed at least this many days in the same status")
	cmd.Flags().IntVar(&maxResults, "max-results", 30, "Maximum number of issues to analyze")
	return cmd
}

func (r runner) newGetFlowEfficiencyReportCommand(state *commandState) *cobra.Command {
	var projectKey, jql, assignee string
	var statuses []string
	var activeStatuses []string
	var blockedStatuses []string
	var startDate, endDate string
	var minBlockedDays, windowDays, maxResults int

	cmd := r.newCommand(state, "get-flow-efficiency-report", "Get active-versus-blocked flow metrics from issue history", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(projectKey) == "" && strings.TrimSpace(jql) == "" {
			return nil, "", fmt.Errorf("either --project-key or --jql is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		report, err := service.GetFlowEfficiencyReport(ctx, models.GetFlowEfficiencyReportRequest{
			ProjectKey:      projectKey,
			JQL:             jql,
			Statuses:        append([]string(nil), statuses...),
			ActiveStatuses:  append([]string(nil), activeStatuses...),
			BlockedStatuses: append([]string(nil), blockedStatuses...),
			Assignee:        assignee,
			MinBlockedDays:  minBlockedDays,
			StartDate:       startDate,
			EndDate:         endDate,
			WindowDays:      windowDays,
			MaxResults:      maxResults,
		})
		if err != nil {
			return nil, "", errors.Wrap(err, "get flow efficiency report")
		}
		return report, textfmt.FormatFlowEfficiencyReport(report), nil
	})

	cmd.Flags().StringVar(&projectKey, "project-key", "", "Project key")
	cmd.Flags().StringVar(&jql, "jql", "", "Custom JQL query")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "Current status to include; repeat the flag to pass multiple values")
	cmd.Flags().StringArrayVar(&activeStatuses, "active-status", nil, "Status treated as active work; repeat the flag to pass multiple values")
	cmd.Flags().StringArrayVar(&blockedStatuses, "blocked-status", nil, "Status treated as blocked work; repeat the flag to pass multiple values")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Assignee to filter by")
	cmd.Flags().IntVar(&minBlockedDays, "min-blocked-days", 0, "Only include issues with at least this many blocked days")
	cmd.Flags().StringVar(&startDate, "start-date", "", "Explicit range start for active and blocked time, in RFC3339 or YYYY-MM-DD")
	cmd.Flags().StringVar(&endDate, "end-date", "", "Explicit range end for active and blocked time, in RFC3339 or YYYY-MM-DD")
	cmd.Flags().IntVar(&windowDays, "window-days", 0, "Only count active and blocked time from the last N days")
	cmd.Flags().IntVar(&maxResults, "max-results", 30, "Maximum number of issues to analyze")
	return cmd
}

func (r runner) newGetSprintWorkloadReportCommand(state *commandState) *cobra.Command {
	var sprintID string
	cmd := r.newCommand(state, "get-sprint-workload-report", "Get per-assignee workload breakdown for a sprint", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(sprintID) == "" {
			return nil, "", fmt.Errorf("--sprint-id is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		report, err := service.GetSprintWorkloadReport(ctx, models.GetSprintWorkloadReportRequest{SprintID: sprintID})
		if err != nil {
			return nil, "", errors.Wrap(err, "get sprint workload report")
		}
		return report, textfmt.FormatSprintWorkloadReport(report), nil
	})
	cmd.Flags().StringVar(&sprintID, "sprint-id", "", "Sprint ID")
	return cmd
}

func (r runner) newGetUtilizationReportCommand(state *commandState) *cobra.Command {
	var projectKey, jql, assignee, startDate, endDate string
	var windowDays, maxResults, topIssuesPerUser int
	cmd := r.newCommand(state, "get-utilization-report", "Aggregate worklog activity per assignee within a time window", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(projectKey) == "" && strings.TrimSpace(jql) == "" {
			return nil, "", fmt.Errorf("either --project-key or --jql is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		report, err := service.GetUtilizationReport(ctx, models.GetUtilizationReportRequest{
			ProjectKey:       projectKey,
			JQL:              jql,
			Assignee:         assignee,
			StartDate:        startDate,
			EndDate:          endDate,
			WindowDays:       windowDays,
			MaxResults:       maxResults,
			TopIssuesPerUser: topIssuesPerUser,
		})
		if err != nil {
			return nil, "", errors.Wrap(err, "get utilization report")
		}
		return report, textfmt.FormatUtilizationReport(report), nil
	})
	cmd.Flags().StringVar(&projectKey, "project-key", "", "Project key")
	cmd.Flags().StringVar(&jql, "jql", "", "Custom JQL query")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Worklog author to filter by")
	cmd.Flags().StringVar(&startDate, "start-date", "", "Window start (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "Window end (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().IntVar(&windowDays, "window-days", 0, "Use last N days when start-date is not provided (default 7)")
	cmd.Flags().IntVar(&maxResults, "max-results", 100, "Maximum number of issues to scan")
	cmd.Flags().IntVar(&topIssuesPerUser, "top-issues-per-user", 5, "Top issues per assignee in output")
	return cmd
}

func (r runner) newGetCycleTimeReportCommand(state *commandState) *cobra.Command {
	var projectKey, jql, assignee, issueType, startDate, endDate string
	var startStatuses, doneStatuses []string
	var windowDays, maxResults int
	cmd := r.newCommand(state, "get-cycle-time-report", "Cycle time, lead time, and throughput from issue history", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(projectKey) == "" && strings.TrimSpace(jql) == "" {
			return nil, "", fmt.Errorf("either --project-key or --jql is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		report, err := service.GetCycleTimeReport(ctx, models.GetCycleTimeReportRequest{
			ProjectKey:    projectKey,
			JQL:           jql,
			StartStatuses: append([]string(nil), startStatuses...),
			DoneStatuses:  append([]string(nil), doneStatuses...),
			Assignee:      assignee,
			IssueType:     issueType,
			StartDate:     startDate,
			EndDate:       endDate,
			WindowDays:    windowDays,
			MaxResults:    maxResults,
		})
		if err != nil {
			return nil, "", errors.Wrap(err, "get cycle time report")
		}
		return report, textfmt.FormatCycleTimeReport(report), nil
	})
	cmd.Flags().StringVar(&projectKey, "project-key", "", "Project key")
	cmd.Flags().StringVar(&jql, "jql", "", "Custom JQL query")
	cmd.Flags().StringArrayVar(&startStatuses, "start-status", nil, "Status that marks active work has started; repeat for multiple")
	cmd.Flags().StringArrayVar(&doneStatuses, "done-status", nil, "Status that marks completion; repeat for multiple")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Assignee to filter by")
	cmd.Flags().StringVar(&issueType, "issue-type", "", "Issue type to filter by")
	cmd.Flags().StringVar(&startDate, "start-date", "", "Window start (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "Window end (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().IntVar(&windowDays, "window-days", 0, "Use last N days when start-date is not provided (default 30)")
	cmd.Flags().IntVar(&maxResults, "max-results", 100, "Maximum number of issues to analyze")
	return cmd
}

func (r runner) newGetTeamWipSnapshotCommand(state *commandState) *cobra.Command {
	var projectKey, jql string
	var statuses, assignees []string
	var wipLimit, maxResults int
	cmd := r.newCommand(state, "get-team-wip-snapshot", "Per-assignee snapshot of in-flight work grouped by status", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(projectKey) == "" && strings.TrimSpace(jql) == "" {
			return nil, "", fmt.Errorf("either --project-key or --jql is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		snapshot, err := service.GetTeamWipSnapshot(ctx, models.GetTeamWipSnapshotRequest{
			ProjectKey: projectKey,
			JQL:        jql,
			Statuses:   append([]string(nil), statuses...),
			Assignees:  append([]string(nil), assignees...),
			WipLimit:   wipLimit,
			MaxResults: maxResults,
		})
		if err != nil {
			return nil, "", errors.Wrap(err, "get team wip snapshot")
		}
		return snapshot, textfmt.FormatTeamWipSnapshot(snapshot), nil
	})
	cmd.Flags().StringVar(&projectKey, "project-key", "", "Project key")
	cmd.Flags().StringVar(&jql, "jql", "", "Custom JQL query")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "Status to include; repeat for multiple values (default: statusCategory=In Progress)")
	cmd.Flags().StringArrayVar(&assignees, "assignee", nil, "Assignee to include; repeat for multiple values")
	cmd.Flags().IntVar(&wipLimit, "wip-limit", 0, "Optional WIP limit per assignee; values above are flagged")
	cmd.Flags().IntVar(&maxResults, "max-results", 200, "Maximum number of issues to scan")
	return cmd
}
