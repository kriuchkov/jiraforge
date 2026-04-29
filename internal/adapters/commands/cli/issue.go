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

func (r runner) newGetIssueCommand(state *commandState) *cobra.Command {
	var issueKey, fields, expand string

	cmd := r.newCommand(state, "get-issue", "Get a Jira issue by key", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(issueKey) == "" {
			return nil, "", fmt.Errorf("--issue-key is required")
		}

		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}

		issue, err := service.GetIssue(ctx, models.GetIssueRequest{IssueKey: issueKey, Fields: splitCSV(fields), Expand: splitCSV(expand)})
		if err != nil {
			return nil, "", errors.Wrap(err, "get issue")
		}

		return issue, textfmt.FormatIssue(issue), nil
	})

	cmd.Flags().StringVar(&issueKey, "issue-key", "", "Issue key, for example PROJ-123")
	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated fields to retrieve")
	cmd.Flags().StringVar(&expand, "expand", "", "Comma-separated expansions")
	return cmd
}

func (r runner) newSearchIssuesCommand(state *commandState) *cobra.Command {
	var jql, fields, expand string
	var maxResults int
	cmd := r.newCommand(state, "search-issues", "Search issues using JQL", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(jql) == "" {
			return nil, "", fmt.Errorf("--jql is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}

		result, err := service.SearchIssues(ctx, models.SearchIssuesRequest{JQL: jql, Fields: splitCSV(fields), Expand: splitCSV(expand), MaxResults: maxResults})
		if err != nil {
			return nil, "", errors.Wrap(err, "search issues")
		}

		return result, textfmt.FormatSearchIssues(result), nil
	})

	cmd.Flags().StringVar(&jql, "jql", "", "JQL query")
	cmd.Flags().IntVar(&maxResults, "max-results", 30, "Maximum number of results")
	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated fields to retrieve")
	cmd.Flags().StringVar(&expand, "expand", "", "Comma-separated expansions")
	return cmd
}

func (r runner) newCreateIssueCommand(state *commandState) *cobra.Command {
	var projectKey, summary, description, issueType string
	cmd := r.newCommand(state, "create-issue", "Create a new issue", func(ctx context.Context) (any, string, error) {
		if !models.HasText(projectKey, summary, description, issueType) || strings.TrimSpace(projectKey) == "" || strings.TrimSpace(summary) == "" || strings.TrimSpace(description) == "" || strings.TrimSpace(issueType) == "" {
			return nil, "", fmt.Errorf("--project-key, --summary, --description, and --issue-type are required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.CreateIssue(ctx, models.CreateIssueRequest{ProjectKey: projectKey, Summary: summary, Description: description, IssueType: issueType})
		if err != nil {
			return nil, "", errors.Wrap(err, "create issue")
		}
		return result, textfmt.FormatMutation(result), nil
	})

	cmd.Flags().StringVar(&projectKey, "project-key", "", "Project key")
	cmd.Flags().StringVar(&summary, "summary", "", "Issue summary")
	cmd.Flags().StringVar(&description, "description", "", "Issue description in Markdown")
	cmd.Flags().StringVar(&issueType, "issue-type", "", "Issue type")
	return cmd
}

func (r runner) newCreateChildIssueCommand(state *commandState) *cobra.Command {
	var parentIssueKey, summary, description, issueType string
	cmd := r.newCommand(state, "create-child-issue", "Create a child or subtask issue", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(parentIssueKey) == "" || strings.TrimSpace(summary) == "" || strings.TrimSpace(description) == "" {
			return nil, "", fmt.Errorf("--parent-issue-key, --summary, and --description are required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.CreateChildIssue(ctx, models.CreateChildIssueRequest{ParentIssueKey: parentIssueKey, Summary: summary, Description: description, IssueType: issueType})
		if err != nil {
			return nil, "", errors.Wrap(err, "create child issue")
		}
		return result, textfmt.FormatMutation(result), nil
	})
	cmd.Flags().StringVar(&parentIssueKey, "parent-issue-key", "", "Parent issue key")
	cmd.Flags().StringVar(&summary, "summary", "", "Issue summary")
	cmd.Flags().StringVar(&description, "description", "", "Issue description in Markdown")
	cmd.Flags().StringVar(&issueType, "issue-type", "", "Issue type")
	return cmd
}

func (r runner) newUpdateIssueCommand(state *commandState) *cobra.Command {
	var issueKey, summary, description string
	cmd := r.newCommand(state, "update-issue", "Update an existing issue", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(issueKey) == "" {
			return nil, "", fmt.Errorf("--issue-key is required")
		}
		if !models.HasText(summary, description) {
			return nil, "", fmt.Errorf("at least one of --summary or --description is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.UpdateIssue(ctx, models.UpdateIssueRequest{IssueKey: issueKey, Summary: summary, Description: description})
		if err != nil {
			return nil, "", errors.Wrap(err, "update issue")
		}
		return result, textfmt.FormatMutation(result), nil
	})
	cmd.Flags().StringVar(&issueKey, "issue-key", "", "Issue key")
	cmd.Flags().StringVar(&summary, "summary", "", "Updated summary")
	cmd.Flags().StringVar(&description, "description", "", "Updated description in Markdown")
	return cmd
}

func (r runner) newDeleteIssueCommand(state *commandState) *cobra.Command {
	var issueKey string
	cmd := r.newCommand(state, "delete-issue", "Delete an issue", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(issueKey) == "" {
			return nil, "", fmt.Errorf("--issue-key is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.DeleteIssue(ctx, models.DeleteIssueRequest{IssueKey: issueKey})
		if err != nil {
			return nil, "", errors.Wrap(err, "delete issue")
		}
		return result, textfmt.FormatMutation(result), nil
	})
	cmd.Flags().StringVar(&issueKey, "issue-key", "", "Issue key")
	return cmd
}

func (r runner) newListIssueTypesCommand(state *commandState) *cobra.Command {
	var projectKey string
	cmd := r.newCommand(state, "list-issue-types", "List issue types for a project", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(projectKey) == "" {
			return nil, "", fmt.Errorf("--project-key is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		issueTypes, err := service.ListIssueTypes(ctx, models.ListIssueTypesRequest{ProjectKey: projectKey})
		if err != nil {
			return nil, "", errors.Wrap(err, "list issue types")
		}
		return issueTypes, textfmt.FormatIssueTypes(issueTypes), nil
	})
	cmd.Flags().StringVar(&projectKey, "project-key", "", "Project key")
	return cmd
}

func (r runner) newAddCommentCommand(state *commandState) *cobra.Command {
	var issueKey, comment string
	cmd := r.newCommand(state, "add-comment", "Add a comment to an issue", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(issueKey) == "" || strings.TrimSpace(comment) == "" {
			return nil, "", fmt.Errorf("--issue-key and --comment are required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.AddComment(ctx, models.AddCommentRequest{IssueKey: issueKey, Comment: comment})
		if err != nil {
			return nil, "", errors.Wrap(err, "add comment")
		}
		return result, textfmt.FormatComment(result), nil
	})
	cmd.Flags().StringVar(&issueKey, "issue-key", "", "Issue key")
	cmd.Flags().StringVar(&comment, "comment", "", "Comment body in Markdown")
	return cmd
}

func (r runner) newGetCommentsCommand(state *commandState) *cobra.Command {
	var issueKey string
	cmd := r.newCommand(state, "get-comments", "Get comments for an issue", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(issueKey) == "" {
			return nil, "", fmt.Errorf("--issue-key is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.GetComments(ctx, models.GetCommentsRequest{IssueKey: issueKey})
		if err != nil {
			return nil, "", errors.Wrap(err, "get comments")
		}
		return result, textfmt.FormatComments(result), nil
	})
	cmd.Flags().StringVar(&issueKey, "issue-key", "", "Issue key")
	return cmd
}
