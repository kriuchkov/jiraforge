package jira

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-faster/errors"

	coreerrors "github.com/kriuchkov/jiraforge/internal/core/errors"
	"github.com/kriuchkov/jiraforge/internal/core/models"
	"github.com/kriuchkov/jiraforge/internal/core/ports"
)

type Service struct {
	gateway         ports.JiraGateway
	attachmentStore ports.AttachmentStore
}

func NewService(gateway ports.JiraGateway, attachmentStore ports.AttachmentStore) (ports.JiraService, error) {
	if gateway == nil {
		return nil, coreerrors.New(coreerrors.CodeInvalidArgument, "service.jira.NewService", "jira gateway is required")
	}
	if attachmentStore == nil {
		return nil, coreerrors.New(coreerrors.CodeInvalidArgument, "service.jira.NewService", "attachment store is required")
	}
	return &Service{gateway: gateway, attachmentStore: attachmentStore}, nil
}

func (s *Service) GetIssue(ctx context.Context, req models.GetIssueRequest) (*models.Issue, error) {
	req = req.Normalized()
	if !req.Validate() {
		return nil, invalid("service.jira.GetIssue", fmt.Sprintf("invalid issue key %q", req.IssueKey))
	}
	issue, err := s.gateway.GetIssue(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "get issue")
	}
	return issue, nil
}

func (s *Service) CreateIssue(ctx context.Context, req models.CreateIssueRequest) (*models.IssueMutationResult, error) {
	if !models.HasText(req.ProjectKey, req.Summary, req.Description, req.IssueType) {
		return nil, invalid("service.jira.CreateIssue", "project_key, summary, description, and issue_type are required")
	}
	if strings.TrimSpace(req.ProjectKey) == "" || strings.TrimSpace(req.Summary) == "" || strings.TrimSpace(req.Description) == "" || strings.TrimSpace(req.IssueType) == "" {
		return nil, invalid("service.jira.CreateIssue", "project_key, summary, description, and issue_type are required")
	}
	result, err := s.gateway.CreateIssue(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "create issue")
	}
	return result, nil
}

func (s *Service) CreateChildIssue(ctx context.Context, req models.CreateChildIssueRequest) (*models.IssueMutationResult, error) {
	if !models.IsIssueKey(req.ParentIssueKey) {
		return nil, invalid("service.jira.CreateChildIssue", fmt.Sprintf("invalid parent issue key %q", req.ParentIssueKey))
	}
	if strings.TrimSpace(req.Summary) == "" || strings.TrimSpace(req.Description) == "" {
		return nil, invalid("service.jira.CreateChildIssue", "summary and description are required")
	}
	result, err := s.gateway.CreateChildIssue(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "create child issue")
	}
	return result, nil
}

func (s *Service) UpdateIssue(ctx context.Context, req models.UpdateIssueRequest) (*models.IssueMutationResult, error) {
	if !models.IsIssueKey(req.IssueKey) {
		return nil, invalid("service.jira.UpdateIssue", fmt.Sprintf("invalid issue key %q", req.IssueKey))
	}
	if !models.HasText(req.Summary, req.Description) {
		return nil, invalid("service.jira.UpdateIssue", "at least one mutable field is required")
	}
	result, err := s.gateway.UpdateIssue(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "update issue")
	}
	return result, nil
}

func (s *Service) DeleteIssue(ctx context.Context, req models.DeleteIssueRequest) (*models.IssueMutationResult, error) {
	if !models.IsIssueKey(req.IssueKey) {
		return nil, invalid("service.jira.DeleteIssue", fmt.Sprintf("invalid issue key %q", req.IssueKey))
	}
	result, err := s.gateway.DeleteIssue(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "delete issue")
	}
	return result, nil
}

func (s *Service) ListIssueTypes(ctx context.Context, req models.ListIssueTypesRequest) ([]models.IssueType, error) {
	if strings.TrimSpace(req.ProjectKey) == "" {
		return nil, invalid("service.jira.ListIssueTypes", "project_key is required")
	}
	issueTypes, err := s.gateway.ListIssueTypes(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "list issue types")
	}
	return issueTypes, nil
}

func (s *Service) SearchIssues(ctx context.Context, req models.SearchIssuesRequest) (*models.SearchIssuesResult, error) {
	req = req.Normalized()
	if strings.TrimSpace(req.JQL) == "" {
		return nil, invalid("service.jira.SearchIssues", "jql is required")
	}
	result, err := s.gateway.SearchIssues(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "search issues")
	}
	return result, nil
}

func (s *Service) ListSprints(ctx context.Context, req models.ListSprintsRequest) (*models.SprintCollection, error) {
	if strings.TrimSpace(req.BoardID) == "" && strings.TrimSpace(req.ProjectKey) == "" {
		return nil, invalid("service.jira.ListSprints", "either board_id or project_key is required")
	}
	result, err := s.gateway.ListSprints(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "list sprints")
	}
	return result, nil
}

func (s *Service) GetSprint(ctx context.Context, req models.GetSprintRequest) (*models.Sprint, error) {
	if strings.TrimSpace(req.SprintID) == "" {
		return nil, invalid("service.jira.GetSprint", "sprint_id is required")
	}
	sprint, err := s.gateway.GetSprint(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "get sprint")
	}
	return sprint, nil
}

func (s *Service) GetSprintReport(ctx context.Context, req models.GetSprintReportRequest) (*models.SprintReport, error) {
	if strings.TrimSpace(req.SprintID) == "" {
		return nil, invalid("service.jira.GetSprintReport", "sprint_id is required")
	}
	report, err := s.gateway.GetSprintReport(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "get sprint report")
	}
	return report, nil
}

func (s *Service) GetActiveSprint(ctx context.Context, req models.GetActiveSprintRequest) (*models.Sprint, error) {
	if strings.TrimSpace(req.BoardID) == "" && strings.TrimSpace(req.ProjectKey) == "" {
		return nil, invalid("service.jira.GetActiveSprint", "either board_id or project_key is required")
	}
	sprint, err := s.gateway.GetActiveSprint(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "get active sprint")
	}
	return sprint, nil
}

func (s *Service) SearchSprints(ctx context.Context, req models.SearchSprintsRequest) (*models.SprintCollection, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, invalid("service.jira.SearchSprints", "name is required")
	}
	if strings.TrimSpace(req.BoardID) == "" && strings.TrimSpace(req.ProjectKey) == "" {
		return nil, invalid("service.jira.SearchSprints", "either board_id or project_key is required")
	}
	result, err := s.gateway.SearchSprints(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "search sprints")
	}
	return result, nil
}

func (s *Service) AddComment(ctx context.Context, req models.AddCommentRequest) (*models.Comment, error) {
	if !models.IsIssueKey(req.IssueKey) {
		return nil, invalid("service.jira.AddComment", fmt.Sprintf("invalid issue key %q", req.IssueKey))
	}
	if strings.TrimSpace(req.Comment) == "" {
		return nil, invalid("service.jira.AddComment", "comment is required")
	}
	comment, err := s.gateway.AddComment(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "add comment")
	}
	return comment, nil
}

func (s *Service) GetComments(ctx context.Context, req models.GetCommentsRequest) (*models.CommentList, error) {
	if !models.IsIssueKey(req.IssueKey) {
		return nil, invalid("service.jira.GetComments", fmt.Sprintf("invalid issue key %q", req.IssueKey))
	}
	comments, err := s.gateway.GetComments(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "get comments")
	}
	return comments, nil
}

func (s *Service) AddWorklog(ctx context.Context, req models.AddWorklogRequest) (*models.Worklog, error) {
	if !models.IsIssueKey(req.IssueKey) {
		return nil, invalid("service.jira.AddWorklog", fmt.Sprintf("invalid issue key %q", req.IssueKey))
	}
	if strings.TrimSpace(req.TimeSpent) == "" {
		return nil, invalid("service.jira.AddWorklog", "time_spent is required")
	}
	worklog, err := s.gateway.AddWorklog(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "add worklog")
	}
	return worklog, nil
}

func (s *Service) GetWorklogs(ctx context.Context, req models.GetWorklogsRequest) (*models.WorklogList, error) {
	if !models.IsIssueKey(req.IssueKey) {
		return nil, invalid("service.jira.GetWorklogs", fmt.Sprintf("invalid issue key %q", req.IssueKey))
	}
	worklogs, err := s.gateway.GetWorklogs(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "get worklogs")
	}
	return worklogs, nil
}

func (s *Service) GetTransitions(ctx context.Context, req models.GetTransitionsRequest) ([]models.Transition, error) {
	if !models.IsIssueKey(req.IssueKey) {
		return nil, invalid("service.jira.GetTransitions", fmt.Sprintf("invalid issue key %q", req.IssueKey))
	}
	transitions, err := s.gateway.GetTransitions(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "get transitions")
	}
	return transitions, nil
}

func (s *Service) TransitionIssue(ctx context.Context, req models.TransitionIssueRequest) (*models.IssueMutationResult, error) {
	if !models.IsIssueKey(req.IssueKey) {
		return nil, invalid("service.jira.TransitionIssue", fmt.Sprintf("invalid issue key %q", req.IssueKey))
	}
	if strings.TrimSpace(req.TransitionID) == "" {
		return nil, invalid("service.jira.TransitionIssue", "transition_id is required")
	}
	result, err := s.gateway.TransitionIssue(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "transition issue")
	}
	return result, nil
}

func (s *Service) ListStatuses(ctx context.Context, req models.ListStatusesRequest) (*models.StatusCatalog, error) {
	if strings.TrimSpace(req.ProjectKey) == "" {
		return nil, invalid("service.jira.ListStatuses", "project_key is required")
	}
	statuses, err := s.gateway.ListStatuses(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "list statuses")
	}
	return statuses, nil
}

func (s *Service) GetIssueHistory(ctx context.Context, req models.GetIssueHistoryRequest) (*models.IssueHistory, error) {
	if !models.IsIssueKey(req.IssueKey) {
		return nil, invalid("service.jira.GetIssueHistory", fmt.Sprintf("invalid issue key %q", req.IssueKey))
	}
	history, err := s.gateway.GetIssueHistory(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "get issue history")
	}
	return history, nil
}

func (s *Service) GetRelatedIssues(ctx context.Context, req models.GetRelatedIssuesRequest) ([]models.IssueRelation, error) {
	if !models.IsIssueKey(req.IssueKey) {
		return nil, invalid("service.jira.GetRelatedIssues", fmt.Sprintf("invalid issue key %q", req.IssueKey))
	}
	relations, err := s.gateway.GetRelatedIssues(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "get related issues")
	}
	return relations, nil
}

func (s *Service) LinkIssues(ctx context.Context, req models.LinkIssuesRequest) (*models.IssueMutationResult, error) {
	if !models.IsIssueKey(req.InwardIssue) || !models.IsIssueKey(req.OutwardIssue) {
		return nil, invalid("service.jira.LinkIssues", "inward_issue and outward_issue must be valid issue keys")
	}
	if strings.TrimSpace(req.LinkType) == "" {
		return nil, invalid("service.jira.LinkIssues", "link_type is required")
	}
	result, err := s.gateway.LinkIssues(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "link issues")
	}
	return result, nil
}

func (s *Service) GetVersion(ctx context.Context, req models.GetVersionRequest) (*models.Version, error) {
	if strings.TrimSpace(req.VersionID) == "" {
		return nil, invalid("service.jira.GetVersion", "version_id is required")
	}
	version, err := s.gateway.GetVersion(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "get version")
	}
	return version, nil
}

func (s *Service) ListProjectVersions(ctx context.Context, req models.ListProjectVersionsRequest) (*models.VersionCollection, error) {
	if strings.TrimSpace(req.ProjectKey) == "" {
		return nil, invalid("service.jira.ListProjectVersions", "project_key is required")
	}
	versions, err := s.gateway.ListProjectVersions(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "list project versions")
	}
	return versions, nil
}

func (s *Service) GetDevelopmentInfo(ctx context.Context, req models.GetDevelopmentInfoRequest) (*models.DevelopmentInfo, error) {
	req = req.Normalized()
	if !models.IsIssueKey(req.IssueKey) {
		return nil, invalid("service.jira.GetDevelopmentInfo", fmt.Sprintf("invalid issue key %q", req.IssueKey))
	}
	info, err := s.gateway.GetDevelopmentInfo(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "get development info")
	}
	return info, nil
}

func (s *Service) DownloadAttachment(ctx context.Context, req models.DownloadAttachmentRequest) (*models.StoredAttachment, error) {
	if strings.TrimSpace(req.AttachmentID) == "" {
		return nil, invalid("service.jira.DownloadAttachment", "attachment_id is required")
	}
	attachment, err := s.gateway.DownloadAttachment(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "download attachment")
	}
	storedAttachment, err := s.attachmentStore.Save(ctx, *attachment)
	if err != nil {
		return nil, errors.Wrap(err, "store attachment")
	}
	return storedAttachment, nil
}

func invalid(op, message string) error {
	return coreerrors.New(coreerrors.CodeInvalidArgument, op, message)
}
