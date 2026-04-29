package ports

import (
	"context"

	"github.com/kriuchkov/jiraforge/internal/core/models"
)

// IssueQueryService defines read-only issue lookup operations exposed by the application service.
type IssueQueryService interface {
	// GetIssue returns a single issue with the requested fields and expansions.
	GetIssue(context.Context, models.GetIssueRequest) (*models.Issue, error)
	// ListIssueTypes returns issue types available for the requested project.
	ListIssueTypes(context.Context, models.ListIssueTypesRequest) ([]models.IssueType, error)
	// SearchIssues executes a JQL search and returns the matching issues.
	SearchIssues(context.Context, models.SearchIssuesRequest) (*models.SearchIssuesResult, error)
}

// IssueMutationService defines issue write operations exposed by the application service.
type IssueMutationService interface {
	// CreateIssue creates a new issue in the target project.
	CreateIssue(context.Context, models.CreateIssueRequest) (*models.IssueMutationResult, error)
	// CreateChildIssue creates a child issue under an existing parent issue.
	CreateChildIssue(context.Context, models.CreateChildIssueRequest) (*models.IssueMutationResult, error)
	// UpdateIssue updates mutable fields on an existing issue.
	UpdateIssue(context.Context, models.UpdateIssueRequest) (*models.IssueMutationResult, error)
	// DeleteIssue removes an existing issue from Jira.
	DeleteIssue(context.Context, models.DeleteIssueRequest) (*models.IssueMutationResult, error)
}

// SprintService defines sprint-related read operations exposed by the application service.
type SprintService interface {
	// ListSprints returns sprints for the requested board or project.
	ListSprints(context.Context, models.ListSprintsRequest) (*models.SprintCollection, error)
	// GetSprint returns sprint details for a sprint ID.
	GetSprint(context.Context, models.GetSprintRequest) (*models.Sprint, error)
	// GetSprintReport returns the sprint report with completed, incomplete, and removed issues.
	GetSprintReport(context.Context, models.GetSprintReportRequest) (*models.SprintReport, error)
	// GetSprintHealthReport returns a management-focused sprint health summary derived from the sprint report.
	GetSprintHealthReport(context.Context, models.GetSprintHealthReportRequest) (*models.SprintHealthReport, error)
	// GetActiveSprint returns the current active sprint for a board or project.
	GetActiveSprint(context.Context, models.GetActiveSprintRequest) (*models.Sprint, error)
	// SearchSprints finds sprints by name within a board or project scope.
	SearchSprints(context.Context, models.SearchSprintsRequest) (*models.SprintCollection, error)
}

// ReportService defines computed management/reporting operations exposed by the application service.
type ReportService interface {
	// GetAgingReport returns issues ordered by how long they have remained in their current status.
	GetAgingReport(context.Context, models.GetAgingReportRequest) (*models.AgingReport, error)
	// GetBlockedIssuesReport returns issues that appear blocked by status or dependency links.
	GetBlockedIssuesReport(context.Context, models.GetBlockedIssuesReportRequest) (*models.BlockedIssuesReport, error)
	// GetFlowEfficiencyReport returns active-versus-blocked time metrics computed from issue history.
	GetFlowEfficiencyReport(context.Context, models.GetFlowEfficiencyReportRequest) (*models.FlowEfficiencyReport, error)
	// GetSprintWorkloadReport returns a per-assignee breakdown of issues, story points, and statuses for a sprint.
	GetSprintWorkloadReport(context.Context, models.GetSprintWorkloadReportRequest) (*models.SprintWorkloadReport, error)
	// GetUtilizationReport aggregates worklog activity per assignee within a time window.
	GetUtilizationReport(context.Context, models.GetUtilizationReportRequest) (*models.UtilizationReport, error)
	// GetCycleTimeReport computes cycle time, lead time, and throughput metrics from issue history.
	GetCycleTimeReport(context.Context, models.GetCycleTimeReportRequest) (*models.CycleTimeReport, error)
	// GetTeamWipSnapshot returns a per-assignee snapshot of in-flight work grouped by status.
	GetTeamWipSnapshot(context.Context, models.GetTeamWipSnapshotRequest) (*models.TeamWipSnapshot, error)
}

// CommentService defines comment operations exposed by the application service.
type CommentService interface {
	// AddComment posts a new comment to an issue.
	AddComment(context.Context, models.AddCommentRequest) (*models.Comment, error)
	// GetComments returns comments associated with an issue.
	GetComments(context.Context, models.GetCommentsRequest) (*models.CommentList, error)
}

// WorklogService defines worklog operations exposed by the application service.
type WorklogService interface {
	// AddWorklog records time spent against an issue.
	AddWorklog(context.Context, models.AddWorklogRequest) (*models.Worklog, error)
	// GetWorklogs returns worklogs recorded against an issue, optionally filtered by start time.
	GetWorklogs(context.Context, models.GetWorklogsRequest) (*models.WorklogList, error)
}

// WorkflowService defines workflow and status operations exposed by the application service.
type WorkflowService interface {
	// GetTransitions returns workflow transitions currently available for an issue.
	GetTransitions(context.Context, models.GetTransitionsRequest) ([]models.Transition, error)
	// TransitionIssue moves an issue through the requested workflow transition.
	TransitionIssue(context.Context, models.TransitionIssueRequest) (*models.IssueMutationResult, error)
	// ListStatuses returns statuses available in the requested project context.
	ListStatuses(context.Context, models.ListStatusesRequest) (*models.StatusCatalog, error)
	// GetIssueHistory returns changelog and history information for an issue.
	GetIssueHistory(context.Context, models.GetIssueHistoryRequest) (*models.IssueHistory, error)
}

// RelationshipService defines issue relationship operations exposed by the application service.
type RelationshipService interface {
	// GetRelatedIssues returns issues related or linked to the requested issue.
	GetRelatedIssues(context.Context, models.GetRelatedIssuesRequest) ([]models.IssueRelation, error)
	// LinkIssues creates a relationship between two issues.
	LinkIssues(context.Context, models.LinkIssuesRequest) (*models.IssueMutationResult, error)
}

// VersionService defines project version operations exposed by the application service.
type VersionService interface {
	// GetVersion returns version details for a version ID.
	GetVersion(context.Context, models.GetVersionRequest) (*models.Version, error)
	// ListProjectVersions returns versions defined for a project.
	ListProjectVersions(context.Context, models.ListProjectVersionsRequest) (*models.VersionCollection, error)
}

// DevelopmentService defines development metadata operations exposed by the application service.
type DevelopmentService interface {
	// GetDevelopmentInfo returns branches, pull requests, and commits linked to an issue.
	GetDevelopmentInfo(context.Context, models.GetDevelopmentInfoRequest) (*models.DevelopmentInfo, error)
}

// AttachmentService defines attachment download operations exposed by the application service.
type AttachmentService interface {
	// DownloadAttachment downloads an attachment and returns the stored file metadata.
	DownloadAttachment(context.Context, models.DownloadAttachmentRequest) (*models.StoredAttachment, error)
}

// JiraService composes all application-facing Jira capabilities.
type JiraService interface {
	IssueQueryService
	IssueMutationService
	SprintService
	ReportService
	CommentService
	WorklogService
	WorkflowService
	RelationshipService
	VersionService
	DevelopmentService
	AttachmentService
}

// IssueQueryGateway defines read-only issue lookup operations required from the Jira adapter.
type IssueQueryGateway interface {
	// GetIssue returns a single issue with the requested fields and expansions.
	GetIssue(context.Context, models.GetIssueRequest) (*models.Issue, error)
	// ListIssueTypes returns issue types available for the requested project.
	ListIssueTypes(context.Context, models.ListIssueTypesRequest) ([]models.IssueType, error)
	// SearchIssues executes a JQL search and returns the matching issues.
	SearchIssues(context.Context, models.SearchIssuesRequest) (*models.SearchIssuesResult, error)
}

// IssueMutationGateway defines issue write operations required from the Jira adapter.
type IssueMutationGateway interface {
	// CreateIssue creates a new issue in the target project.
	CreateIssue(context.Context, models.CreateIssueRequest) (*models.IssueMutationResult, error)
	// CreateChildIssue creates a child issue under an existing parent issue.
	CreateChildIssue(context.Context, models.CreateChildIssueRequest) (*models.IssueMutationResult, error)
	// UpdateIssue updates mutable fields on an existing issue.
	UpdateIssue(context.Context, models.UpdateIssueRequest) (*models.IssueMutationResult, error)
	// DeleteIssue removes an existing issue from Jira.
	DeleteIssue(context.Context, models.DeleteIssueRequest) (*models.IssueMutationResult, error)
}

// SprintGateway defines sprint-related read operations required from the Jira adapter.
type SprintGateway interface {
	// ListSprints returns sprints for the requested board or project.
	ListSprints(context.Context, models.ListSprintsRequest) (*models.SprintCollection, error)
	// GetSprint returns sprint details for a sprint ID.
	GetSprint(context.Context, models.GetSprintRequest) (*models.Sprint, error)
	// GetSprintReport returns the sprint report with completed, incomplete, and removed issues.
	GetSprintReport(context.Context, models.GetSprintReportRequest) (*models.SprintReport, error)
	// GetActiveSprint returns the current active sprint for a board or project.
	GetActiveSprint(context.Context, models.GetActiveSprintRequest) (*models.Sprint, error)
	// SearchSprints finds sprints by name within a board or project scope.
	SearchSprints(context.Context, models.SearchSprintsRequest) (*models.SprintCollection, error)
}

// CommentGateway defines comment operations required from the Jira adapter.
type CommentGateway interface {
	// AddComment posts a new comment to an issue.
	AddComment(context.Context, models.AddCommentRequest) (*models.Comment, error)
	// GetComments returns comments associated with an issue.
	GetComments(context.Context, models.GetCommentsRequest) (*models.CommentList, error)
}

// WorklogGateway defines worklog operations required from the Jira adapter.
type WorklogGateway interface {
	// AddWorklog records time spent against an issue.
	AddWorklog(context.Context, models.AddWorklogRequest) (*models.Worklog, error)
	// GetWorklogs returns worklogs recorded against an issue.
	GetWorklogs(context.Context, models.GetWorklogsRequest) (*models.WorklogList, error)
}

// WorkflowGateway defines workflow and status operations required from the Jira adapter.
type WorkflowGateway interface {
	// GetTransitions returns workflow transitions currently available for an issue.
	GetTransitions(context.Context, models.GetTransitionsRequest) ([]models.Transition, error)
	// TransitionIssue moves an issue through the requested workflow transition.
	TransitionIssue(context.Context, models.TransitionIssueRequest) (*models.IssueMutationResult, error)
	// ListStatuses returns statuses available in the requested project context.
	ListStatuses(context.Context, models.ListStatusesRequest) (*models.StatusCatalog, error)
	// GetIssueHistory returns changelog and history information for an issue.
	GetIssueHistory(context.Context, models.GetIssueHistoryRequest) (*models.IssueHistory, error)
}

// RelationshipGateway defines issue relationship operations required from the Jira adapter.
type RelationshipGateway interface {
	// GetRelatedIssues returns issues related or linked to the requested issue.
	GetRelatedIssues(context.Context, models.GetRelatedIssuesRequest) ([]models.IssueRelation, error)
	// LinkIssues creates a relationship between two issues.
	LinkIssues(context.Context, models.LinkIssuesRequest) (*models.IssueMutationResult, error)
}

// VersionGateway defines project version operations required from the Jira adapter.
type VersionGateway interface {
	// GetVersion returns version details for a version ID.
	GetVersion(context.Context, models.GetVersionRequest) (*models.Version, error)
	// ListProjectVersions returns versions defined for a project.
	ListProjectVersions(context.Context, models.ListProjectVersionsRequest) (*models.VersionCollection, error)
}

// DevelopmentGateway defines development metadata operations required from the Jira adapter.
type DevelopmentGateway interface {
	// GetDevelopmentInfo returns branches, pull requests, and commits linked to an issue.
	GetDevelopmentInfo(context.Context, models.GetDevelopmentInfoRequest) (*models.DevelopmentInfo, error)
}

// AttachmentGateway defines attachment download operations required from the Jira adapter.
type AttachmentGateway interface {
	// DownloadAttachment retrieves raw attachment content from Jira.
	DownloadAttachment(context.Context, models.DownloadAttachmentRequest) (*models.AttachmentContent, error)
}

// JiraGateway composes all outbound Jira adapter capabilities.
type JiraGateway interface {
	IssueQueryGateway
	IssueMutationGateway
	SprintGateway
	CommentGateway
	WorklogGateway
	WorkflowGateway
	RelationshipGateway
	VersionGateway
	DevelopmentGateway
	AttachmentGateway
}

// AttachmentStore persists downloaded attachment data and returns stored file metadata.
type AttachmentStore interface {
	// Save writes attachment content to storage and returns the stored attachment descriptor.
	Save(context.Context, models.AttachmentContent) (*models.StoredAttachment, error)
}
