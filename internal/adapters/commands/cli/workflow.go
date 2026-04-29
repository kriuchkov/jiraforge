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

func (r runner) newAddWorklogCommand(state *commandState) *cobra.Command {
	var issueKey, timeSpent, comment, started string
	cmd := r.newCommand(state, "add-worklog", "Add a worklog entry to an issue", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(issueKey) == "" || strings.TrimSpace(timeSpent) == "" {
			return nil, "", fmt.Errorf("--issue-key and --time-spent are required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.AddWorklog(ctx, models.AddWorklogRequest{IssueKey: issueKey, TimeSpent: timeSpent, Comment: comment, Started: started})
		if err != nil {
			return nil, "", errors.Wrap(err, "add worklog")
		}
		return result, textfmt.FormatWorklog(result), nil
	})
	cmd.Flags().StringVar(&issueKey, "issue-key", "", "Issue key")
	cmd.Flags().StringVar(&timeSpent, "time-spent", "", "Time spent, for example 1h30m")
	cmd.Flags().StringVar(&comment, "comment", "", "Optional comment in Markdown")
	cmd.Flags().StringVar(&started, "started", "", "Optional started timestamp")
	return cmd
}

func (r runner) newGetWorklogsCommand(state *commandState) *cobra.Command {
	var issueKey, startedAfter string
	var maxResults int
	cmd := r.newCommand(state, "get-worklogs", "List worklogs recorded against an issue", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(issueKey) == "" {
			return nil, "", fmt.Errorf("--issue-key is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.GetWorklogs(ctx, models.GetWorklogsRequest{IssueKey: issueKey, StartedAfter: startedAfter, MaxResults: maxResults})
		if err != nil {
			return nil, "", errors.Wrap(err, "get worklogs")
		}
		return result, textfmt.FormatWorklogList(result), nil
	})
	cmd.Flags().StringVar(&issueKey, "issue-key", "", "Issue key")
	cmd.Flags().StringVar(&startedAfter, "started-after", "", "Only include worklogs started at or after this timestamp")
	cmd.Flags().IntVar(&maxResults, "max-results", 1000, "Maximum number of worklogs to return")
	return cmd
}

func (r runner) newGetTransitionsCommand(state *commandState) *cobra.Command {
	var issueKey string
	cmd := r.newCommand(state, "get-transitions", "Get available transitions for an issue", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(issueKey) == "" {
			return nil, "", fmt.Errorf("--issue-key is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.GetTransitions(ctx, models.GetTransitionsRequest{IssueKey: issueKey})
		if err != nil {
			return nil, "", errors.Wrap(err, "get transitions")
		}
		return result, textfmt.FormatTransitions(result), nil
	})
	cmd.Flags().StringVar(&issueKey, "issue-key", "", "Issue key")
	return cmd
}

func (r runner) newTransitionIssueCommand(state *commandState) *cobra.Command {
	var issueKey, transitionID, comment string
	cmd := r.newCommand(state, "transition-issue", "Transition an issue to a new status", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(issueKey) == "" || strings.TrimSpace(transitionID) == "" {
			return nil, "", fmt.Errorf("--issue-key and --transition-id are required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.TransitionIssue(ctx, models.TransitionIssueRequest{IssueKey: issueKey, TransitionID: transitionID, Comment: comment})
		if err != nil {
			return nil, "", errors.Wrap(err, "transition issue")
		}
		return result, textfmt.FormatMutation(result), nil
	})
	cmd.Flags().StringVar(&issueKey, "issue-key", "", "Issue key")
	cmd.Flags().StringVar(&transitionID, "transition-id", "", "Transition ID")
	cmd.Flags().StringVar(&comment, "comment", "", "Optional comment")
	return cmd
}

func (r runner) newListStatusesCommand(state *commandState) *cobra.Command {
	var projectKey string
	cmd := r.newCommand(state, "list-statuses", "List statuses for a project", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(projectKey) == "" {
			return nil, "", fmt.Errorf("--project-key is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.ListStatuses(ctx, models.ListStatusesRequest{ProjectKey: projectKey})
		if err != nil {
			return nil, "", errors.Wrap(err, "list statuses")
		}
		return result, textfmt.FormatStatuses(result), nil
	})
	cmd.Flags().StringVar(&projectKey, "project-key", "", "Project key")
	return cmd
}

func (r runner) newGetIssueHistoryCommand(state *commandState) *cobra.Command {
	var issueKey string
	cmd := r.newCommand(state, "get-issue-history", "Get the change history of an issue", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(issueKey) == "" {
			return nil, "", fmt.Errorf("--issue-key is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.GetIssueHistory(ctx, models.GetIssueHistoryRequest{IssueKey: issueKey})
		if err != nil {
			return nil, "", errors.Wrap(err, "get issue history")
		}
		return result, textfmt.FormatHistory(result), nil
	})
	cmd.Flags().StringVar(&issueKey, "issue-key", "", "Issue key")
	return cmd
}
