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

func (r runner) newGetRelatedIssuesCommand(state *commandState) *cobra.Command {
	var issueKey string
	cmd := r.newCommand(state, "get-related-issues", "Get issues linked to an issue", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(issueKey) == "" {
			return nil, "", fmt.Errorf("--issue-key is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.GetRelatedIssues(ctx, models.GetRelatedIssuesRequest{IssueKey: issueKey})
		if err != nil {
			return nil, "", errors.Wrap(err, "get related issues")
		}
		return result, textfmt.FormatRelations(issueKey, result), nil
	})
	cmd.Flags().StringVar(&issueKey, "issue-key", "", "Issue key")
	return cmd
}

func (r runner) newLinkIssuesCommand(state *commandState) *cobra.Command {
	var inwardIssue, outwardIssue, linkType, comment string
	cmd := r.newCommand(state, "link-issues", "Create a link between two issues", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(inwardIssue) == "" || strings.TrimSpace(outwardIssue) == "" || strings.TrimSpace(linkType) == "" {
			return nil, "", fmt.Errorf("--inward-issue, --outward-issue, and --link-type are required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.LinkIssues(ctx, models.LinkIssuesRequest{InwardIssue: inwardIssue, OutwardIssue: outwardIssue, LinkType: linkType, Comment: comment})
		if err != nil {
			return nil, "", errors.Wrap(err, "link issues")
		}
		return result, textfmt.FormatMutation(result), nil
	})
	cmd.Flags().StringVar(&inwardIssue, "inward-issue", "", "Inward issue key")
	cmd.Flags().StringVar(&outwardIssue, "outward-issue", "", "Outward issue key")
	cmd.Flags().StringVar(&linkType, "link-type", "", "Link type")
	cmd.Flags().StringVar(&comment, "comment", "", "Optional comment")
	return cmd
}

func (r runner) newGetVersionCommand(state *commandState) *cobra.Command {
	var versionID string
	cmd := r.newCommand(state, "get-version", "Get a project version by ID", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(versionID) == "" {
			return nil, "", fmt.Errorf("--version-id is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.GetVersion(ctx, models.GetVersionRequest{VersionID: versionID})
		if err != nil {
			return nil, "", errors.Wrap(err, "get version")
		}
		return result, textfmt.FormatVersion(result), nil
	})
	cmd.Flags().StringVar(&versionID, "version-id", "", "Version ID")
	return cmd
}

func (r runner) newListProjectVersionsCommand(state *commandState) *cobra.Command {
	var projectKey string
	cmd := r.newCommand(state, "list-project-versions", "List all versions for a project", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(projectKey) == "" {
			return nil, "", fmt.Errorf("--project-key is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.ListProjectVersions(ctx, models.ListProjectVersionsRequest{ProjectKey: projectKey})
		if err != nil {
			return nil, "", errors.Wrap(err, "list project versions")
		}
		return result, textfmt.FormatVersions(result), nil
	})
	cmd.Flags().StringVar(&projectKey, "project-key", "", "Project key")
	return cmd
}

func (r runner) newGetDevelopmentInfoCommand(state *commandState) *cobra.Command {
	var issueKey string
	var includeBranches, includePullRequests, includeCommits, includeBuilds bool
	cmd := r.newCommand(state, "get-development-info", "Get branches, pull requests, commits, and builds for an issue", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(issueKey) == "" {
			return nil, "", fmt.Errorf("--issue-key is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.GetDevelopmentInfo(ctx, models.GetDevelopmentInfoRequest{
			IssueKey:            issueKey,
			IncludeBranches:     includeBranches,
			IncludePullRequests: includePullRequests,
			IncludeCommits:      includeCommits,
			IncludeBuilds:       includeBuilds,
		})
		if err != nil {
			return nil, "", errors.Wrap(err, "get development info")
		}
		return result, textfmt.FormatDevelopmentInfo(result), nil
	})
	cmd.Flags().StringVar(&issueKey, "issue-key", "", "Issue key")
	cmd.Flags().BoolVar(&includeBranches, "include-branches", false, "Include branches")
	cmd.Flags().BoolVar(&includePullRequests, "include-pull-requests", false, "Include pull requests")
	cmd.Flags().BoolVar(&includeCommits, "include-commits", false, "Include commits")
	cmd.Flags().BoolVar(&includeBuilds, "include-builds", false, "Include builds")
	return cmd
}

func (r runner) newDownloadAttachmentCommand(state *commandState) *cobra.Command {
	var attachmentID string
	cmd := r.newCommand(state, "download-attachment", "Download an issue attachment", func(ctx context.Context) (any, string, error) {
		if strings.TrimSpace(attachmentID) == "" {
			return nil, "", fmt.Errorf("--attachment-id is required")
		}
		service, err := r.service(state.envFile)
		if err != nil {
			return nil, "", errors.Wrap(err, "create jira service")
		}
		result, err := service.DownloadAttachment(ctx, models.DownloadAttachmentRequest{AttachmentID: attachmentID})
		if err != nil {
			return nil, "", errors.Wrap(err, "download attachment")
		}
		return result, textfmt.FormatStoredAttachment(result), nil
	})
	cmd.Flags().StringVar(&attachmentID, "attachment-id", "", "Attachment ID")
	return cmd
}
