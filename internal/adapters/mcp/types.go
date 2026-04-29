package mcp

type getIssueInput struct {
	IssueKey string `json:"issue_key" validate:"required"`
	Fields   string `json:"fields,omitempty"`
	Expand   string `json:"expand,omitempty"`
}

type createIssueInput struct {
	ProjectKey  string `json:"project_key" validate:"required"`
	Summary     string `json:"summary" validate:"required"`
	Description string `json:"description" validate:"required"`
	IssueType   string `json:"issue_type" validate:"required"`
}

type createChildIssueInput struct {
	ParentIssueKey string `json:"parent_issue_key" validate:"required"`
	Summary        string `json:"summary" validate:"required"`
	Description    string `json:"description" validate:"required"`
	IssueType      string `json:"issue_type,omitempty"`
}

type updateIssueInput struct {
	IssueKey    string `json:"issue_key" validate:"required"`
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
}

type deleteIssueInput struct {
	IssueKey string `json:"issue_key" validate:"required"`
}

type listIssueTypesInput struct {
	ProjectKey string `json:"project_key" validate:"required"`
}

type searchIssuesInput struct {
	JQL    string `json:"jql" validate:"required"`
	Fields string `json:"fields,omitempty"`
	Expand string `json:"expand,omitempty"`
}

type getAgingReportInput struct {
	ProjectKey      string `json:"project_key,omitempty"`
	JQL             string `json:"jql,omitempty"`
	Statuses        string `json:"statuses,omitempty"`
	Assignee        string `json:"assignee,omitempty"`
	MinDaysInStatus int    `json:"min_days_in_status,omitempty"`
	MaxResults      int    `json:"max_results,omitempty"`
}

type getBlockedIssuesReportInput struct {
	ProjectKey      string `json:"project_key,omitempty"`
	JQL             string `json:"jql,omitempty"`
	Statuses        string `json:"statuses,omitempty"`
	BlockedStatuses string `json:"blocked_statuses,omitempty"`
	LinkTypes       string `json:"link_types,omitempty"`
	Assignee        string `json:"assignee,omitempty"`
	MinDaysInStatus int    `json:"min_days_in_status,omitempty"`
	MaxResults      int    `json:"max_results,omitempty"`
}

type getFlowEfficiencyReportInput struct {
	ProjectKey      string `json:"project_key,omitempty"`
	JQL             string `json:"jql,omitempty"`
	Statuses        string `json:"statuses,omitempty"`
	ActiveStatuses  string `json:"active_statuses,omitempty"`
	BlockedStatuses string `json:"blocked_statuses,omitempty"`
	Assignee        string `json:"assignee,omitempty"`
	MinBlockedDays  int    `json:"min_blocked_days,omitempty"`
	StartDate       string `json:"start_date,omitempty"`
	EndDate         string `json:"end_date,omitempty"`
	WindowDays      int    `json:"window_days,omitempty"`
	MaxResults      int    `json:"max_results,omitempty"`
}

type listSprintsInput struct {
	BoardID    string `json:"board_id,omitempty"`
	ProjectKey string `json:"project_key,omitempty"`
}

type getSprintInput struct {
	SprintID string `json:"sprint_id" validate:"required"`
}

type getSprintReportInput struct {
	SprintID string `json:"sprint_id" validate:"required"`
}

type getSprintHealthReportInput struct {
	SprintID string `json:"sprint_id" validate:"required"`
}

type getActiveSprintInput struct {
	BoardID    string `json:"board_id,omitempty"`
	ProjectKey string `json:"project_key,omitempty"`
}

type searchSprintsInput struct {
	Name       string `json:"name" validate:"required"`
	BoardID    string `json:"board_id,omitempty"`
	ProjectKey string `json:"project_key,omitempty"`
	ExactMatch bool   `json:"exact_match,omitempty"`
}

type addCommentInput struct {
	IssueKey string `json:"issue_key" validate:"required"`
	Comment  string `json:"comment" validate:"required"`
}

type getCommentsInput struct {
	IssueKey string `json:"issue_key" validate:"required"`
}

type addWorklogInput struct {
	IssueKey  string `json:"issue_key" validate:"required"`
	TimeSpent string `json:"time_spent" validate:"required"`
	Comment   string `json:"comment,omitempty"`
	Started   string `json:"started,omitempty"`
}

type getTransitionsInput struct {
	IssueKey string `json:"issue_key" validate:"required"`
}

type transitionIssueInput struct {
	IssueKey     string `json:"issue_key" validate:"required"`
	TransitionID string `json:"transition_id" validate:"required"`
	Comment      string `json:"comment,omitempty"`
}

type listStatusesInput struct {
	ProjectKey string `json:"project_key" validate:"required"`
}

type getIssueHistoryInput struct {
	IssueKey string `json:"issue_key" validate:"required"`
}

type getRelatedIssuesInput struct {
	IssueKey string `json:"issue_key" validate:"required"`
}

type linkIssuesInput struct {
	InwardIssue  string `json:"inward_issue" validate:"required"`
	OutwardIssue string `json:"outward_issue" validate:"required"`
	LinkType     string `json:"link_type" validate:"required"`
	Comment      string `json:"comment,omitempty"`
}

type getVersionInput struct {
	VersionID string `json:"version_id" validate:"required"`
}

type listProjectVersionsInput struct {
	ProjectKey string `json:"project_key" validate:"required"`
}

type getDevelopmentInfoInput struct {
	IssueKey            string `json:"issue_key" validate:"required"`
	IncludeBranches     bool   `json:"include_branches,omitempty"`
	IncludePullRequests bool   `json:"include_pull_requests,omitempty"`
	IncludeCommits      bool   `json:"include_commits,omitempty"`
	IncludeBuilds       bool   `json:"include_builds,omitempty"`
}

type downloadAttachmentInput struct {
	AttachmentID string `json:"attachment_id" validate:"required"`
}

type getSprintWorkloadReportInput struct {
	SprintID string `json:"sprint_id" validate:"required"`
}

type getUtilizationReportInput struct {
	ProjectKey       string `json:"project_key,omitempty"`
	JQL              string `json:"jql,omitempty"`
	Assignee         string `json:"assignee,omitempty"`
	StartDate        string `json:"start_date,omitempty"`
	EndDate          string `json:"end_date,omitempty"`
	WindowDays       int    `json:"window_days,omitempty"`
	MaxResults       int    `json:"max_results,omitempty"`
	TopIssuesPerUser int    `json:"top_issues_per_user,omitempty"`
}

type getCycleTimeReportInput struct {
	ProjectKey    string `json:"project_key,omitempty"`
	JQL           string `json:"jql,omitempty"`
	StartStatuses string `json:"start_statuses,omitempty"`
	DoneStatuses  string `json:"done_statuses,omitempty"`
	Assignee      string `json:"assignee,omitempty"`
	IssueType     string `json:"issue_type,omitempty"`
	StartDate     string `json:"start_date,omitempty"`
	EndDate       string `json:"end_date,omitempty"`
	WindowDays    int    `json:"window_days,omitempty"`
	MaxResults    int    `json:"max_results,omitempty"`
}

type getTeamWipSnapshotInput struct {
	ProjectKey string `json:"project_key,omitempty"`
	JQL        string `json:"jql,omitempty"`
	Statuses   string `json:"statuses,omitempty"`
	Assignees  string `json:"assignees,omitempty"`
	WipLimit   int    `json:"wip_limit,omitempty"`
	MaxResults int    `json:"max_results,omitempty"`
}

type getWorklogsInput struct {
	IssueKey     string `json:"issue_key" validate:"required"`
	StartedAfter string `json:"started_after,omitempty"`
	MaxResults   int    `json:"max_results,omitempty"`
}
