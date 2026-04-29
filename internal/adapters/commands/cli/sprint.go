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

func (r runner) newListSprintsCommand(state *commandState) *cobra.Command {
	var boardID, projectKey string
	cmd := r.newCommand(state, "list-sprints", "List sprints for a board or project", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(boardID) == "" && strings.TrimSpace(projectKey) == "" {
			return nil, "", fmt.Errorf("either --board-id or --project-key is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.ListSprints(ctx, models.ListSprintsRequest{BoardID: boardID, ProjectKey: projectKey})
		if err != nil {
			return nil, "", errors.Wrap(err, "list sprints")
		}
		return result, textfmt.FormatSprints(result), nil
	})
	cmd.Flags().StringVar(&boardID, "board-id", "", "Board ID")
	cmd.Flags().StringVar(&projectKey, "project-key", "", "Project key")
	return cmd
}

func (r runner) newGetSprintCommand(state *commandState) *cobra.Command {
	var sprintID string
	cmd := r.newCommand(state, "get-sprint", "Get a sprint by ID", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(sprintID) == "" {
			return nil, "", fmt.Errorf("--sprint-id is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		sprint, err := service.GetSprint(ctx, models.GetSprintRequest{SprintID: sprintID})
		if err != nil {
			return nil, "", errors.Wrap(err, "get sprint")
		}
		return sprint, textfmt.FormatSprint(sprint), nil
	})
	cmd.Flags().StringVar(&sprintID, "sprint-id", "", "Sprint ID")
	return cmd
}

func (r runner) newGetSprintReportCommand(state *commandState) *cobra.Command {
	var sprintID string
	cmd := r.newCommand(state, "get-sprint-report", "Get the sprint report for a sprint", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(sprintID) == "" {
			return nil, "", fmt.Errorf("--sprint-id is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		report, err := service.GetSprintReport(ctx, models.GetSprintReportRequest{SprintID: sprintID})
		if err != nil {
			return nil, "", errors.Wrap(err, "get sprint report")
		}
		return report, textfmt.FormatSprintReport(report), nil
	})
	cmd.Flags().StringVar(&sprintID, "sprint-id", "", "Sprint ID")
	return cmd
}

func (r runner) newGetSprintHealthReportCommand(state *commandState) *cobra.Command {
	var sprintID string
	cmd := r.newCommand(state, "get-sprint-health-report", "Get a health summary for a sprint", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(sprintID) == "" {
			return nil, "", fmt.Errorf("--sprint-id is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		report, err := service.GetSprintHealthReport(ctx, models.GetSprintHealthReportRequest{SprintID: sprintID})
		if err != nil {
			return nil, "", errors.Wrap(err, "get sprint health report")
		}
		return report, textfmt.FormatSprintHealthReport(report), nil
	})
	cmd.Flags().StringVar(&sprintID, "sprint-id", "", "Sprint ID")
	return cmd
}

func (r runner) newGetActiveSprintCommand(state *commandState) *cobra.Command {
	var boardID, projectKey string
	cmd := r.newCommand(state, "get-active-sprint", "Get the active sprint for a board or project", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(boardID) == "" && strings.TrimSpace(projectKey) == "" {
			return nil, "", fmt.Errorf("either --board-id or --project-key is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		sprint, err := service.GetActiveSprint(ctx, models.GetActiveSprintRequest{BoardID: boardID, ProjectKey: projectKey})
		if err != nil {
			return nil, "", errors.Wrap(err, "get active sprint")
		}
		return sprint, textfmt.FormatSprint(sprint), nil
	})
	cmd.Flags().StringVar(&boardID, "board-id", "", "Board ID")
	cmd.Flags().StringVar(&projectKey, "project-key", "", "Project key")
	return cmd
}

func (r runner) newSearchSprintCommand(state *commandState) *cobra.Command {
	var name, boardID, projectKey string
	var exactMatch bool
	cmd := r.newCommand(state, "search-sprint", "Search sprints by name", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(name) == "" {
			return nil, "", fmt.Errorf("--name is required")
		}
		if strings.TrimSpace(boardID) == "" && strings.TrimSpace(projectKey) == "" {
			return nil, "", fmt.Errorf("either --board-id or --project-key is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.SearchSprints(ctx, models.SearchSprintsRequest{Name: name, BoardID: boardID, ProjectKey: projectKey, ExactMatch: exactMatch})
		if err != nil {
			return nil, "", errors.Wrap(err, "search sprints")
		}
		return result, textfmt.FormatSprints(result), nil
	})
	cmd.Flags().StringVar(&name, "name", "", "Sprint name")
	cmd.Flags().StringVar(&boardID, "board-id", "", "Board ID")
	cmd.Flags().StringVar(&projectKey, "project-key", "", "Project key")
	cmd.Flags().BoolVar(&exactMatch, "exact-match", false, "Require an exact name match")
	return cmd
}
