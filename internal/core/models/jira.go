package models

import (
	"regexp"
	"slices"
	"strings"
)

var issueKeyPattern = regexp.MustCompile(`(?i)^[A-Z][A-Z0-9_]*-\d+$`)

func IsIssueKey(value string) bool {
	return issueKeyPattern.MatchString(strings.TrimSpace(value))
}

type Person struct {
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email,omitempty"`
}

type Status struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type IssueType struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	IconURL     string `json:"icon_url,omitempty"`
	Subtask     bool   `json:"subtask,omitempty"`
	Scope       string `json:"scope,omitempty"`
}

type Transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AttachmentMeta struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	URL      string `json:"url,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	Size     int64  `json:"size,omitempty"`
}

type IssueRef struct {
	Key     string  `json:"key"`
	Summary string  `json:"summary,omitempty"`
	Status  *Status `json:"status,omitempty"`
}

type IssueRelation struct {
	Type      string   `json:"type"`
	Direction string   `json:"direction"`
	Issue     IssueRef `json:"issue"`
}

type Issue struct {
	Key                   string           `json:"key"`
	ID                    string           `json:"id,omitempty"`
	URL                   string           `json:"url,omitempty"`
	ProjectKey            string           `json:"project_key,omitempty"`
	ProjectName           string           `json:"project_name,omitempty"`
	Summary               string           `json:"summary,omitempty"`
	Description           string           `json:"description,omitempty"`
	Type                  *IssueType       `json:"type,omitempty"`
	Status                *Status          `json:"status,omitempty"`
	Priority              string           `json:"priority,omitempty"`
	Resolution            string           `json:"resolution,omitempty"`
	ResolutionDescription string           `json:"resolution_description,omitempty"`
	ResolutionDate        string           `json:"resolution_date,omitempty"`
	Reporter              *Person          `json:"reporter,omitempty"`
	Assignee              *Person          `json:"assignee,omitempty"`
	Creator               *Person          `json:"creator,omitempty"`
	Created               string           `json:"created,omitempty"`
	Updated               string           `json:"updated,omitempty"`
	LastViewed            string           `json:"last_viewed,omitempty"`
	StatusCategoryChange  string           `json:"status_category_change,omitempty"`
	Parent                *IssueRef        `json:"parent,omitempty"`
	Labels                []string         `json:"labels,omitempty"`
	Components            []string         `json:"components,omitempty"`
	FixVersions           []string         `json:"fix_versions,omitempty"`
	AffectedVersions      []string         `json:"affected_versions,omitempty"`
	SecurityLevel         string           `json:"security_level,omitempty"`
	Subtasks              []IssueRef       `json:"subtasks,omitempty"`
	RelatedIssues         []IssueRelation  `json:"related_issues,omitempty"`
	Attachments           []AttachmentMeta `json:"attachments,omitempty"`
	Watchers              int              `json:"watchers,omitempty"`
	Votes                 int              `json:"votes,omitempty"`
	CommentCount          int              `json:"comment_count,omitempty"`
	WorklogCount          int              `json:"worklog_count,omitempty"`
	Transitions           []Transition     `json:"transitions,omitempty"`
	StoryPointEstimate    string           `json:"story_point_estimate,omitempty"`
}

type IssueMutationResult struct {
	Key       string `json:"key,omitempty"`
	ID        string `json:"id,omitempty"`
	URL       string `json:"url,omitempty"`
	ParentKey string `json:"parent_key,omitempty"`
	Message   string `json:"message,omitempty"`
}

type SearchIssuesResult struct {
	Query      string  `json:"query"`
	StartAt    int     `json:"start_at,omitempty"`
	MaxResults int     `json:"max_results,omitempty"`
	Total      int     `json:"total,omitempty"`
	Issues     []Issue `json:"issues,omitempty"`
}

type Comment struct {
	ID      string  `json:"id"`
	Author  *Person `json:"author,omitempty"`
	Created string  `json:"created,omitempty"`
	Updated string  `json:"updated,omitempty"`
	Body    string  `json:"body,omitempty"`
}

type CommentList struct {
	IssueKey string    `json:"issue_key"`
	Comments []Comment `json:"comments,omitempty"`
}

type Worklog struct {
	ID               string  `json:"id"`
	IssueKey         string  `json:"issue_key"`
	TimeSpent        string  `json:"time_spent,omitempty"`
	TimeSpentSeconds int     `json:"time_spent_seconds,omitempty"`
	Started          string  `json:"started,omitempty"`
	Author           *Person `json:"author,omitempty"`
	Comment          string  `json:"comment,omitempty"`
}

type StatusGroup struct {
	IssueType IssueType `json:"issue_type"`
	Statuses  []Status  `json:"statuses,omitempty"`
}

type StatusCatalog struct {
	ProjectKey string        `json:"project_key"`
	Groups     []StatusGroup `json:"groups,omitempty"`
}

type Sprint struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	State        string `json:"state,omitempty"`
	StartDate    string `json:"start_date,omitempty"`
	EndDate      string `json:"end_date,omitempty"`
	CompleteDate string `json:"complete_date,omitempty"`
	BoardID      int    `json:"board_id,omitempty"`
	Goal         string `json:"goal,omitempty"`
}

type SprintReportEstimate struct {
	Text  string  `json:"text,omitempty"`
	Value float64 `json:"value,omitempty"`
}

type SprintReportIssue struct {
	Key               string               `json:"key"`
	Summary           string               `json:"summary,omitempty"`
	Type              string               `json:"type,omitempty"`
	Status            string               `json:"status,omitempty"`
	Estimate          SprintReportEstimate `json:"estimate,omitempty"`
	CurrentEstimate   SprintReportEstimate `json:"current_estimate,omitempty"`
	AddedDuringSprint bool                 `json:"added_during_sprint,omitempty"`
}

type SprintReport struct {
	Sprint                   Sprint               `json:"sprint"`
	CompletedIssues          []SprintReportIssue  `json:"completed_issues,omitempty"`
	IncompleteIssues         []SprintReportIssue  `json:"incomplete_issues,omitempty"`
	RemovedIssues            []SprintReportIssue  `json:"removed_issues,omitempty"`
	CompletedInAnotherSprint []SprintReportIssue  `json:"completed_in_another_sprint,omitempty"`
	AddedIssueKeys           []string             `json:"added_issue_keys,omitempty"`
	AllIssuesEstimate        SprintReportEstimate `json:"all_issues_estimate,omitempty"`
	CompletedIssuesEstimate  SprintReportEstimate `json:"completed_issues_estimate,omitempty"`
	IncompleteIssuesEstimate SprintReportEstimate `json:"incomplete_issues_estimate,omitempty"`
	RemovedIssuesEstimate    SprintReportEstimate `json:"removed_issues_estimate,omitempty"`
}

type SprintHealthSummary struct {
	CommittedIssues            int     `json:"committed_issues,omitempty"`
	CompletedCommittedIssues   int     `json:"completed_committed_issues,omitempty"`
	CompletedIssues            int     `json:"completed_issues,omitempty"`
	IncompleteIssues           int     `json:"incomplete_issues,omitempty"`
	CompletedElsewhereIssues   int     `json:"completed_elsewhere_issues,omitempty"`
	CarryOverIssues            int     `json:"carry_over_issues,omitempty"`
	RemovedIssues              int     `json:"removed_issues,omitempty"`
	AddedDuringSprint          int     `json:"added_during_sprint,omitempty"`
	UnplannedCompletedIssues   int     `json:"unplanned_completed_issues,omitempty"`
	CompletionRatio            float64 `json:"completion_ratio,omitempty"`
	CommittedEstimate          float64 `json:"committed_estimate,omitempty"`
	CompletedCommittedEstimate float64 `json:"completed_committed_estimate,omitempty"`
	EstimateCompletionRatio    float64 `json:"estimate_completion_ratio,omitempty"`
}

type SprintHealthReport struct {
	Sprint              Sprint              `json:"sprint"`
	Summary             SprintHealthSummary `json:"summary"`
	Risks               []string            `json:"risks,omitempty"`
	TopIncompleteIssues []SprintReportIssue `json:"top_incomplete_issues,omitempty"`
}

type AgingReportItem struct {
	Key          string `json:"key"`
	Summary      string `json:"summary,omitempty"`
	Type         string `json:"type,omitempty"`
	Status       string `json:"status,omitempty"`
	Assignee     string `json:"assignee,omitempty"`
	Updated      string `json:"updated,omitempty"`
	StatusSince  string `json:"status_since,omitempty"`
	DaysInStatus int    `json:"days_in_status,omitempty"`
}

type AgingReportSummary struct {
	AnalyzedIssues      int     `json:"analyzed_issues,omitempty"`
	MatchingIssues      int     `json:"matching_issues,omitempty"`
	OldestDaysInStatus  int     `json:"oldest_days_in_status,omitempty"`
	AverageDaysInStatus float64 `json:"average_days_in_status,omitempty"`
}

type AgingReport struct {
	Query           string             `json:"query,omitempty"`
	ProjectKey      string             `json:"project_key,omitempty"`
	Statuses        []string           `json:"statuses,omitempty"`
	Assignee        string             `json:"assignee,omitempty"`
	MinDaysInStatus int                `json:"min_days_in_status,omitempty"`
	Summary         AgingReportSummary `json:"summary"`
	Items           []AgingReportItem  `json:"items,omitempty"`
}

type BlockedIssueDependency struct {
	Type      string `json:"type,omitempty"`
	Direction string `json:"direction,omitempty"`
	Key       string `json:"key,omitempty"`
	Summary   string `json:"summary,omitempty"`
	Status    string `json:"status,omitempty"`
	External  bool   `json:"external,omitempty"`
}

type BlockedIssuesReportItem struct {
	Key             string                   `json:"key"`
	Summary         string                   `json:"summary,omitempty"`
	Type            string                   `json:"type,omitempty"`
	Status          string                   `json:"status,omitempty"`
	Assignee        string                   `json:"assignee,omitempty"`
	Updated         string                   `json:"updated,omitempty"`
	StatusSince     string                   `json:"status_since,omitempty"`
	DaysInStatus    int                      `json:"days_in_status,omitempty"`
	BlockedByStatus bool                     `json:"blocked_by_status,omitempty"`
	BlockingLinks   []BlockedIssueDependency `json:"blocking_links,omitempty"`
	Reasons         []string                 `json:"reasons,omitempty"`
}

type BlockedIssuesReportSummary struct {
	AnalyzedIssues           int     `json:"analyzed_issues,omitempty"`
	BlockedIssues            int     `json:"blocked_issues,omitempty"`
	BlockedByStatusIssues    int     `json:"blocked_by_status_issues,omitempty"`
	BlockedByLinkIssues      int     `json:"blocked_by_link_issues,omitempty"`
	ExternalDependencyIssues int     `json:"external_dependency_issues,omitempty"`
	OldestDaysInStatus       int     `json:"oldest_days_in_status,omitempty"`
	AverageDaysInStatus      float64 `json:"average_days_in_status,omitempty"`
}

type BlockedIssuesReport struct {
	Query           string                     `json:"query,omitempty"`
	ProjectKey      string                     `json:"project_key,omitempty"`
	Statuses        []string                   `json:"statuses,omitempty"`
	BlockedStatuses []string                   `json:"blocked_statuses,omitempty"`
	LinkTypes       []string                   `json:"link_types,omitempty"`
	Assignee        string                     `json:"assignee,omitempty"`
	MinDaysInStatus int                        `json:"min_days_in_status,omitempty"`
	Summary         BlockedIssuesReportSummary `json:"summary"`
	Items           []BlockedIssuesReportItem  `json:"items,omitempty"`
}

type FlowEfficiencyReportItem struct {
	Key                string  `json:"key"`
	Summary            string  `json:"summary,omitempty"`
	Type               string  `json:"type,omitempty"`
	Status             string  `json:"status,omitempty"`
	Assignee           string  `json:"assignee,omitempty"`
	Updated            string  `json:"updated,omitempty"`
	CurrentStatusSince string  `json:"current_status_since,omitempty"`
	CurrentStatusDays  int     `json:"current_status_days,omitempty"`
	ObservedDays       int     `json:"observed_days,omitempty"`
	ActiveDays         int     `json:"active_days,omitempty"`
	BlockedDays        int     `json:"blocked_days,omitempty"`
	FlowEfficiency     float64 `json:"flow_efficiency,omitempty"`
	BlockedTransitions int     `json:"blocked_transitions,omitempty"`
	CurrentlyBlocked   bool    `json:"currently_blocked,omitempty"`
}

type FlowEfficiencyReportSummary struct {
	AnalyzedIssues          int     `json:"analyzed_issues,omitempty"`
	MatchingIssues          int     `json:"matching_issues,omitempty"`
	IssuesWithBlockedTime   int     `json:"issues_with_blocked_time,omitempty"`
	CurrentlyBlockedIssues  int     `json:"currently_blocked_issues,omitempty"`
	TotalObservedDays       int     `json:"total_observed_days,omitempty"`
	TotalActiveDays         int     `json:"total_active_days,omitempty"`
	TotalBlockedDays        int     `json:"total_blocked_days,omitempty"`
	AverageFlowEfficiency   float64 `json:"average_flow_efficiency,omitempty"`
	PortfolioFlowEfficiency float64 `json:"portfolio_flow_efficiency,omitempty"`
}

type FlowEfficiencyAssigneeSummary struct {
	Assignee                string  `json:"assignee,omitempty"`
	Issues                  int     `json:"issues,omitempty"`
	IssuesWithBlockedTime   int     `json:"issues_with_blocked_time,omitempty"`
	CurrentlyBlockedIssues  int     `json:"currently_blocked_issues,omitempty"`
	TotalObservedDays       int     `json:"total_observed_days,omitempty"`
	TotalActiveDays         int     `json:"total_active_days,omitempty"`
	TotalBlockedDays        int     `json:"total_blocked_days,omitempty"`
	AverageFlowEfficiency   float64 `json:"average_flow_efficiency,omitempty"`
	PortfolioFlowEfficiency float64 `json:"portfolio_flow_efficiency,omitempty"`
}

type FlowEfficiencyReport struct {
	Query           string                          `json:"query,omitempty"`
	ProjectKey      string                          `json:"project_key,omitempty"`
	Statuses        []string                        `json:"statuses,omitempty"`
	ActiveStatuses  []string                        `json:"active_statuses,omitempty"`
	BlockedStatuses []string                        `json:"blocked_statuses,omitempty"`
	Assignee        string                          `json:"assignee,omitempty"`
	MinBlockedDays  int                             `json:"min_blocked_days,omitempty"`
	WindowDays      int                             `json:"window_days,omitempty"`
	WindowStart     string                          `json:"window_start,omitempty"`
	WindowEnd       string                          `json:"window_end,omitempty"`
	Summary         FlowEfficiencyReportSummary     `json:"summary"`
	Assignees       []FlowEfficiencyAssigneeSummary `json:"assignees,omitempty"`
	Items           []FlowEfficiencyReportItem      `json:"items,omitempty"`
}

type SprintCollection struct {
	Scope   string   `json:"scope,omitempty"`
	Sprints []Sprint `json:"sprints,omitempty"`
}

type SprintWorkloadStatusBreakdown struct {
	Status string `json:"status,omitempty"`
	Issues int    `json:"issues,omitempty"`
}

type SprintWorkloadAssignee struct {
	Assignee                   string                          `json:"assignee,omitempty"`
	TotalIssues                int                             `json:"total_issues,omitempty"`
	CommittedIssues            int                             `json:"committed_issues,omitempty"`
	CompletedIssues            int                             `json:"completed_issues,omitempty"`
	IncompleteIssues           int                             `json:"incomplete_issues,omitempty"`
	AddedDuringSprintIssues    int                             `json:"added_during_sprint_issues,omitempty"`
	RemovedIssues              int                             `json:"removed_issues,omitempty"`
	CommittedEstimate          float64                         `json:"committed_estimate,omitempty"`
	CompletedEstimate          float64                         `json:"completed_estimate,omitempty"`
	CommitmentCompletionRatio  float64                         `json:"commitment_completion_ratio,omitempty"`
	StatusBreakdown            []SprintWorkloadStatusBreakdown `json:"status_breakdown,omitempty"`
	IncompleteIssueKeys        []string                        `json:"incomplete_issue_keys,omitempty"`
	CompletedIssueKeys         []string                        `json:"completed_issue_keys,omitempty"`
	AddedDuringSprintIssueKeys []string                        `json:"added_during_sprint_issue_keys,omitempty"`
}

type SprintWorkloadReport struct {
	Sprint    Sprint                   `json:"sprint"`
	Assignees []SprintWorkloadAssignee `json:"assignees,omitempty"`
}

type GetSprintWorkloadReportRequest struct {
	SprintID string `json:"sprint_id"`
}

type WorklogList struct {
	IssueKey string    `json:"issue_key"`
	Total    int       `json:"total,omitempty"`
	Worklogs []Worklog `json:"worklogs,omitempty"`
}

type GetWorklogsRequest struct {
	IssueKey   string `json:"issue_key"`
	StartedAfter string `json:"started_after,omitempty"`
	MaxResults int    `json:"max_results,omitempty"`
}

func (r GetWorklogsRequest) Normalized() GetWorklogsRequest {
	r.StartedAfter = strings.TrimSpace(r.StartedAfter)
	if r.MaxResults <= 0 {
		r.MaxResults = 1000
	}
	return r
}

type UtilizationIssueDetail struct {
	Key             string  `json:"key"`
	Summary         string  `json:"summary,omitempty"`
	Status          string  `json:"status,omitempty"`
	Type            string  `json:"type,omitempty"`
	SecondsLogged   int     `json:"seconds_logged,omitempty"`
	HoursLogged     float64 `json:"hours_logged,omitempty"`
	WorklogEntries  int     `json:"worklog_entries,omitempty"`
}

type UtilizationAssigneeSummary struct {
	Assignee        string                   `json:"assignee,omitempty"`
	WorklogEntries  int                      `json:"worklog_entries,omitempty"`
	SecondsLogged   int                      `json:"seconds_logged,omitempty"`
	HoursLogged     float64                  `json:"hours_logged,omitempty"`
	DailyAverageHrs float64                  `json:"daily_average_hours,omitempty"`
	IssuesTouched   int                      `json:"issues_touched,omitempty"`
	TopIssues       []UtilizationIssueDetail `json:"top_issues,omitempty"`
}

type UtilizationReportSummary struct {
	AnalyzedIssues  int     `json:"analyzed_issues,omitempty"`
	WorklogEntries  int     `json:"worklog_entries,omitempty"`
	SecondsLogged   int     `json:"seconds_logged,omitempty"`
	HoursLogged     float64 `json:"hours_logged,omitempty"`
	DistinctAuthors int     `json:"distinct_authors,omitempty"`
	WindowDays      int     `json:"window_days,omitempty"`
}

type UtilizationReport struct {
	Query       string                       `json:"query,omitempty"`
	ProjectKey  string                       `json:"project_key,omitempty"`
	WindowDays  int                          `json:"window_days,omitempty"`
	WindowStart string                       `json:"window_start,omitempty"`
	WindowEnd   string                       `json:"window_end,omitempty"`
	Summary     UtilizationReportSummary     `json:"summary"`
	Assignees   []UtilizationAssigneeSummary `json:"assignees,omitempty"`
}

type GetUtilizationReportRequest struct {
	ProjectKey       string `json:"project_key,omitempty"`
	JQL              string `json:"jql,omitempty"`
	Assignee         string `json:"assignee,omitempty"`
	StartDate        string `json:"start_date,omitempty"`
	EndDate          string `json:"end_date,omitempty"`
	WindowDays       int    `json:"window_days,omitempty"`
	MaxResults       int    `json:"max_results,omitempty"`
	TopIssuesPerUser int    `json:"top_issues_per_user,omitempty"`
}

func (r GetUtilizationReportRequest) Normalized() GetUtilizationReportRequest {
	r.StartDate = strings.TrimSpace(r.StartDate)
	r.EndDate = strings.TrimSpace(r.EndDate)
	if r.WindowDays < 0 {
		r.WindowDays = 0
	}
	if r.WindowDays == 0 && r.StartDate == "" && r.EndDate == "" {
		r.WindowDays = 7
	}
	if r.MaxResults <= 0 {
		r.MaxResults = 100
	}
	if r.TopIssuesPerUser <= 0 {
		r.TopIssuesPerUser = 5
	}
	return r
}

type CycleTimeReportItem struct {
	Key          string  `json:"key"`
	Summary      string  `json:"summary,omitempty"`
	Type         string  `json:"type,omitempty"`
	Status       string  `json:"status,omitempty"`
	Assignee     string  `json:"assignee,omitempty"`
	Created      string  `json:"created,omitempty"`
	StartedAt    string  `json:"started_at,omitempty"`
	CompletedAt  string  `json:"completed_at,omitempty"`
	CycleHours   float64 `json:"cycle_hours,omitempty"`
	LeadHours    float64 `json:"lead_hours,omitempty"`
}

type CycleTimeAssigneeSummary struct {
	Assignee         string  `json:"assignee,omitempty"`
	CompletedIssues  int     `json:"completed_issues,omitempty"`
	AverageCycleHrs  float64 `json:"average_cycle_hours,omitempty"`
	MedianCycleHrs   float64 `json:"median_cycle_hours,omitempty"`
	AverageLeadHrs   float64 `json:"average_lead_hours,omitempty"`
}

type CycleTimeTypeBreakdown struct {
	Type            string  `json:"type,omitempty"`
	CompletedIssues int     `json:"completed_issues,omitempty"`
	AverageCycleHrs float64 `json:"average_cycle_hours,omitempty"`
	MedianCycleHrs  float64 `json:"median_cycle_hours,omitempty"`
}

type CycleTimeReportSummary struct {
	AnalyzedIssues       int     `json:"analyzed_issues,omitempty"`
	CompletedIssues      int     `json:"completed_issues,omitempty"`
	AverageCycleHours    float64 `json:"average_cycle_hours,omitempty"`
	MedianCycleHours     float64 `json:"median_cycle_hours,omitempty"`
	P85CycleHours        float64 `json:"p85_cycle_hours,omitempty"`
	AverageLeadHours     float64 `json:"average_lead_hours,omitempty"`
	ThroughputTotal      int     `json:"throughput_total,omitempty"`
	ThroughputPerWeek    float64 `json:"throughput_per_week,omitempty"`
}

type CycleTimeReport struct {
	Query         string                     `json:"query,omitempty"`
	ProjectKey    string                     `json:"project_key,omitempty"`
	StartStatuses []string                   `json:"start_statuses,omitempty"`
	DoneStatuses  []string                   `json:"done_statuses,omitempty"`
	WindowDays    int                        `json:"window_days,omitempty"`
	WindowStart   string                     `json:"window_start,omitempty"`
	WindowEnd     string                     `json:"window_end,omitempty"`
	Summary       CycleTimeReportSummary     `json:"summary"`
	Assignees     []CycleTimeAssigneeSummary `json:"assignees,omitempty"`
	Types         []CycleTimeTypeBreakdown   `json:"types,omitempty"`
	Items         []CycleTimeReportItem      `json:"items,omitempty"`
}

type GetCycleTimeReportRequest struct {
	ProjectKey    string   `json:"project_key,omitempty"`
	JQL           string   `json:"jql,omitempty"`
	StartStatuses []string `json:"start_statuses,omitempty"`
	DoneStatuses  []string `json:"done_statuses,omitempty"`
	Assignee      string   `json:"assignee,omitempty"`
	IssueType     string   `json:"issue_type,omitempty"`
	StartDate     string   `json:"start_date,omitempty"`
	EndDate       string   `json:"end_date,omitempty"`
	WindowDays    int      `json:"window_days,omitempty"`
	MaxResults    int      `json:"max_results,omitempty"`
}

func (r GetCycleTimeReportRequest) Normalized() GetCycleTimeReportRequest {
	r.StartStatuses = normalizeReportStringSlice(r.StartStatuses)
	r.DoneStatuses = normalizeReportStringSlice(r.DoneStatuses)
	if len(r.StartStatuses) == 0 {
		r.StartStatuses = []string{"In Progress"}
	}
	if len(r.DoneStatuses) == 0 {
		r.DoneStatuses = []string{"Done", "Closed", "Resolved"}
	}
	r.StartDate = strings.TrimSpace(r.StartDate)
	r.EndDate = strings.TrimSpace(r.EndDate)
	if r.WindowDays < 0 {
		r.WindowDays = 0
	}
	if r.WindowDays == 0 && r.StartDate == "" && r.EndDate == "" {
		r.WindowDays = 30
	}
	if r.MaxResults <= 0 {
		r.MaxResults = 100
	}
	return r
}

type TeamWipIssue struct {
	Key      string `json:"key"`
	Summary  string `json:"summary,omitempty"`
	Status   string `json:"status,omitempty"`
	Type     string `json:"type,omitempty"`
	Priority string `json:"priority,omitempty"`
	Updated  string `json:"updated,omitempty"`
}

type TeamWipStatusBreakdown struct {
	Status string         `json:"status,omitempty"`
	Issues int            `json:"issues,omitempty"`
	Items  []TeamWipIssue `json:"items,omitempty"`
}

type TeamWipAssignee struct {
	Assignee   string                   `json:"assignee,omitempty"`
	TotalWip   int                      `json:"total_wip,omitempty"`
	OverLimit  bool                     `json:"over_limit,omitempty"`
	Statuses   []TeamWipStatusBreakdown `json:"statuses,omitempty"`
}

type TeamWipSnapshotSummary struct {
	AnalyzedIssues   int            `json:"analyzed_issues,omitempty"`
	DistinctAssignees int           `json:"distinct_assignees,omitempty"`
	UnassignedIssues int            `json:"unassigned_issues,omitempty"`
	StatusTotals     map[string]int `json:"status_totals,omitempty"`
}

type TeamWipSnapshot struct {
	Query      string                 `json:"query,omitempty"`
	ProjectKey string                 `json:"project_key,omitempty"`
	Statuses   []string               `json:"statuses,omitempty"`
	WipLimit   int                    `json:"wip_limit,omitempty"`
	Summary    TeamWipSnapshotSummary `json:"summary"`
	Assignees  []TeamWipAssignee      `json:"assignees,omitempty"`
}

type GetTeamWipSnapshotRequest struct {
	ProjectKey string   `json:"project_key,omitempty"`
	JQL        string   `json:"jql,omitempty"`
	Statuses   []string `json:"statuses,omitempty"`
	Assignees  []string `json:"assignees,omitempty"`
	WipLimit   int      `json:"wip_limit,omitempty"`
	MaxResults int      `json:"max_results,omitempty"`
}

func (r GetTeamWipSnapshotRequest) Normalized() GetTeamWipSnapshotRequest {
	r.Statuses = normalizeReportStringSlice(r.Statuses)
	r.Assignees = normalizeReportStringSlice(r.Assignees)
	if r.WipLimit < 0 {
		r.WipLimit = 0
	}
	if r.MaxResults <= 0 {
		r.MaxResults = 200
	}
	return r
}

type HistoryChange struct {
	Field string `json:"field"`
	From  string `json:"from,omitempty"`
	To    string `json:"to,omitempty"`
}

type HistoryEntry struct {
	Date    string          `json:"date,omitempty"`
	Author  string          `json:"author,omitempty"`
	Changes []HistoryChange `json:"changes,omitempty"`
}

type IssueHistory struct {
	IssueKey string         `json:"issue_key"`
	Entries  []HistoryEntry `json:"entries,omitempty"`
}

type Version struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ProjectID   int    `json:"project_id,omitempty"`
	Released    bool   `json:"released,omitempty"`
	Archived    bool   `json:"archived,omitempty"`
	ReleaseDate string `json:"release_date,omitempty"`
	URL         string `json:"url,omitempty"`
	Status      string `json:"status,omitempty"`
}

type VersionCollection struct {
	ProjectKey string    `json:"project_key"`
	Versions   []Version `json:"versions,omitempty"`
}

type DevelopmentRepositoryRef struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

type DevelopmentCommitFile struct {
	Path         string `json:"path,omitempty"`
	URL          string `json:"url,omitempty"`
	ChangeType   string `json:"change_type,omitempty"`
	LinesAdded   int    `json:"lines_added,omitempty"`
	LinesRemoved int    `json:"lines_removed,omitempty"`
}

type DevelopmentCommit struct {
	ID              string                  `json:"id,omitempty"`
	DisplayID       string                  `json:"display_id,omitempty"`
	Message         string                  `json:"message,omitempty"`
	Author          *Person                 `json:"author,omitempty"`
	AuthorTimestamp string                  `json:"author_timestamp,omitempty"`
	URL             string                  `json:"url,omitempty"`
	FileCount       int                     `json:"file_count,omitempty"`
	Merge           bool                    `json:"merge,omitempty"`
	Files           []DevelopmentCommitFile `json:"files,omitempty"`
}

type DevelopmentBranch struct {
	Name                 string                   `json:"name,omitempty"`
	URL                  string                   `json:"url,omitempty"`
	CreatePullRequestURL string                   `json:"create_pull_request_url,omitempty"`
	Repository           DevelopmentRepositoryRef `json:"repository,omitempty"`
	LastCommit           DevelopmentCommit        `json:"last_commit,omitempty"`
}

type DevelopmentReviewer struct {
	Name     string `json:"name,omitempty"`
	Approved bool   `json:"approved,omitempty"`
}

type DevelopmentBranchRef struct {
	Branch string `json:"branch,omitempty"`
	URL    string `json:"url,omitempty"`
}

type DevelopmentPullRequest struct {
	ID             string                `json:"id,omitempty"`
	Title          string                `json:"title,omitempty"`
	URL            string                `json:"url,omitempty"`
	Status         string                `json:"status,omitempty"`
	Author         *Person               `json:"author,omitempty"`
	LastUpdate     string                `json:"last_update,omitempty"`
	Source         DevelopmentBranchRef  `json:"source,omitempty"`
	Destination    DevelopmentBranchRef  `json:"destination,omitempty"`
	CommentCount   int                   `json:"comment_count,omitempty"`
	Reviewers      []DevelopmentReviewer `json:"reviewers,omitempty"`
	RepositoryID   string                `json:"repository_id,omitempty"`
	RepositoryName string                `json:"repository_name,omitempty"`
	RepositoryURL  string                `json:"repository_url,omitempty"`
}

type DevelopmentRepository struct {
	ID      string              `json:"id,omitempty"`
	Name    string              `json:"name,omitempty"`
	URL     string              `json:"url,omitempty"`
	Avatar  string              `json:"avatar,omitempty"`
	Commits []DevelopmentCommit `json:"commits,omitempty"`
}

type BuildTestSummary struct {
	TotalNumber   int `json:"total_number,omitempty"`
	NumberPassed  int `json:"number_passed,omitempty"`
	SuccessNumber int `json:"success_number,omitempty"`
	NumberFailed  int `json:"number_failed,omitempty"`
	FailedNumber  int `json:"failed_number,omitempty"`
	SkippedNumber int `json:"skipped_number,omitempty"`
}

type BuildCommitRef struct {
	ID            string `json:"id,omitempty"`
	DisplayID     string `json:"display_id,omitempty"`
	RepositoryURI string `json:"repository_uri,omitempty"`
}

type BuildRefInfo struct {
	Name string `json:"name,omitempty"`
	URI  string `json:"uri,omitempty"`
}

type BuildReference struct {
	Commit BuildCommitRef `json:"commit,omitempty"`
	Ref    BuildRefInfo   `json:"ref,omitempty"`
}

type DevelopmentBuild struct {
	ID             string            `json:"id,omitempty"`
	Name           string            `json:"name,omitempty"`
	DisplayName    string            `json:"display_name,omitempty"`
	Description    string            `json:"description,omitempty"`
	URL            string            `json:"url,omitempty"`
	State          string            `json:"state,omitempty"`
	CreatedAt      string            `json:"created_at,omitempty"`
	LastUpdated    string            `json:"last_updated,omitempty"`
	BuildNumber    any               `json:"build_number,omitempty"`
	Tests          *BuildTestSummary `json:"tests,omitempty"`
	References     []BuildReference  `json:"references,omitempty"`
	PipelineID     string            `json:"pipeline_id,omitempty"`
	PipelineName   string            `json:"pipeline_name,omitempty"`
	ProviderID     string            `json:"provider_id,omitempty"`
	ProviderType   string            `json:"provider_type,omitempty"`
	RepositoryID   string            `json:"repository_id,omitempty"`
	RepositoryName string            `json:"repository_name,omitempty"`
	RepositoryURL  string            `json:"repository_url,omitempty"`
}

type DevelopmentInfo struct {
	IssueKey     string                   `json:"issue_key"`
	IssueID      string                   `json:"issue_id,omitempty"`
	Branches     []DevelopmentBranch      `json:"branches,omitempty"`
	PullRequests []DevelopmentPullRequest `json:"pull_requests,omitempty"`
	Repositories []DevelopmentRepository  `json:"repositories,omitempty"`
	Builds       []DevelopmentBuild       `json:"builds,omitempty"`
}

type AttachmentContent struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Data     []byte `json:"-"`
}

type StoredAttachment struct {
	Path     string `json:"path"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type,omitempty"`
	Size     int64  `json:"size,omitempty"`
}

type GetIssueRequest struct {
	IssueKey string   `json:"issue_key"`
	Fields   []string `json:"fields,omitempty"`
	Expand   []string `json:"expand,omitempty"`
}

func (r GetIssueRequest) Normalized() GetIssueRequest {
	r.Fields = slices.Clone(r.Fields)
	r.Expand = slices.Clone(r.Expand)
	if len(r.Expand) == 0 {
		r.Expand = []string{"transitions", "changelog", "subtasks", "description"}
	}
	return r
}

func (r GetIssueRequest) Validate() bool {
	return IsIssueKey(r.IssueKey)
}

type CreateIssueRequest struct {
	ProjectKey        string `json:"project_key"`
	Summary           string `json:"summary"`
	Description       string `json:"description"`
	IssueType         string `json:"issue_type"`
	AssigneeAccountID string `json:"assignee_account_id,omitempty"`
	Priority          string `json:"priority,omitempty"`
}

type CreateChildIssueRequest struct {
	ParentIssueKey string `json:"parent_issue_key"`
	Summary        string `json:"summary"`
	Description    string `json:"description"`
	IssueType      string `json:"issue_type,omitempty"`
}

type UpdateIssueRequest struct {
	IssueKey    string `json:"issue_key"`
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
}

type DeleteIssueRequest struct {
	IssueKey string `json:"issue_key"`
}

type ListIssueTypesRequest struct {
	ProjectKey string `json:"project_key"`
}

type SearchIssuesRequest struct {
	JQL        string   `json:"jql"`
	Fields     []string `json:"fields,omitempty"`
	Expand     []string `json:"expand,omitempty"`
	StartAt    int      `json:"start_at,omitempty"`
	MaxResults int      `json:"max_results,omitempty"`
}

func (r SearchIssuesRequest) Normalized() SearchIssuesRequest {
	r.Fields = slices.Clone(r.Fields)
	r.Expand = slices.Clone(r.Expand)
	if len(r.Expand) == 0 {
		r.Expand = []string{"transitions", "changelog", "subtasks", "description"}
	}
	if r.MaxResults <= 0 {
		r.MaxResults = 30
	}
	return r
}

type ListSprintsRequest struct {
	BoardID    string `json:"board_id,omitempty"`
	ProjectKey string `json:"project_key,omitempty"`
}

type GetSprintRequest struct {
	SprintID string `json:"sprint_id"`
}

type GetSprintReportRequest struct {
	SprintID string `json:"sprint_id"`
}

type GetSprintHealthReportRequest struct {
	SprintID string `json:"sprint_id"`
}

type GetAgingReportRequest struct {
	ProjectKey      string   `json:"project_key,omitempty"`
	JQL             string   `json:"jql,omitempty"`
	Statuses        []string `json:"statuses,omitempty"`
	Assignee        string   `json:"assignee,omitempty"`
	MinDaysInStatus int      `json:"min_days_in_status,omitempty"`
	MaxResults      int      `json:"max_results,omitempty"`
}

func (r GetAgingReportRequest) Normalized() GetAgingReportRequest {
	r.Statuses = slices.Clone(r.Statuses)
	for index := 0; index < len(r.Statuses); index++ {
		r.Statuses[index] = strings.TrimSpace(r.Statuses[index])
	}
	r.Statuses = slices.DeleteFunc(r.Statuses, func(value string) bool {
		return value == ""
	})
	if r.MinDaysInStatus < 0 {
		r.MinDaysInStatus = 0
	}
	if r.MaxResults <= 0 {
		r.MaxResults = 30
	}
	return r
}

type GetBlockedIssuesReportRequest struct {
	ProjectKey      string   `json:"project_key,omitempty"`
	JQL             string   `json:"jql,omitempty"`
	Statuses        []string `json:"statuses,omitempty"`
	BlockedStatuses []string `json:"blocked_statuses,omitempty"`
	LinkTypes       []string `json:"link_types,omitempty"`
	Assignee        string   `json:"assignee,omitempty"`
	MinDaysInStatus int      `json:"min_days_in_status,omitempty"`
	MaxResults      int      `json:"max_results,omitempty"`
}

func (r GetBlockedIssuesReportRequest) Normalized() GetBlockedIssuesReportRequest {
	r.Statuses = normalizeReportStringSlice(r.Statuses)
	r.BlockedStatuses = normalizeReportStringSlice(r.BlockedStatuses)
	r.LinkTypes = normalizeReportStringSlice(r.LinkTypes)
	if len(r.BlockedStatuses) == 0 {
		r.BlockedStatuses = []string{"Blocked"}
	}
	if r.MinDaysInStatus < 0 {
		r.MinDaysInStatus = 0
	}
	if r.MaxResults <= 0 {
		r.MaxResults = 30
	}
	return r
}

type GetFlowEfficiencyReportRequest struct {
	ProjectKey      string   `json:"project_key,omitempty"`
	JQL             string   `json:"jql,omitempty"`
	Statuses        []string `json:"statuses,omitempty"`
	ActiveStatuses  []string `json:"active_statuses,omitempty"`
	BlockedStatuses []string `json:"blocked_statuses,omitempty"`
	Assignee        string   `json:"assignee,omitempty"`
	MinBlockedDays  int      `json:"min_blocked_days,omitempty"`
	StartDate       string   `json:"start_date,omitempty"`
	EndDate         string   `json:"end_date,omitempty"`
	WindowDays      int      `json:"window_days,omitempty"`
	MaxResults      int      `json:"max_results,omitempty"`
}

func (r GetFlowEfficiencyReportRequest) Normalized() GetFlowEfficiencyReportRequest {
	r.Statuses = normalizeReportStringSlice(r.Statuses)
	r.ActiveStatuses = normalizeReportStringSlice(r.ActiveStatuses)
	r.BlockedStatuses = normalizeReportStringSlice(r.BlockedStatuses)
	if len(r.BlockedStatuses) == 0 {
		r.BlockedStatuses = []string{"Blocked"}
	}
	if len(r.ActiveStatuses) == 0 && len(r.Statuses) > 0 {
		blocked := make(map[string]struct{}, len(r.BlockedStatuses))
		for _, status := range r.BlockedStatuses {
			blocked[strings.ToLower(status)] = struct{}{}
		}
		for _, status := range r.Statuses {
			if _, ok := blocked[strings.ToLower(status)]; !ok {
				r.ActiveStatuses = append(r.ActiveStatuses, status)
			}
		}
	}
	if r.MinBlockedDays < 0 {
		r.MinBlockedDays = 0
	}
	r.StartDate = strings.TrimSpace(r.StartDate)
	r.EndDate = strings.TrimSpace(r.EndDate)
	if r.WindowDays < 0 {
		r.WindowDays = 0
	}
	if r.MaxResults <= 0 {
		r.MaxResults = 30
	}
	return r
}

func normalizeReportStringSlice(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

type GetActiveSprintRequest struct {
	BoardID    string `json:"board_id,omitempty"`
	ProjectKey string `json:"project_key,omitempty"`
}

type SearchSprintsRequest struct {
	Name       string `json:"name"`
	BoardID    string `json:"board_id,omitempty"`
	ProjectKey string `json:"project_key,omitempty"`
	ExactMatch bool   `json:"exact_match,omitempty"`
}

type AddCommentRequest struct {
	IssueKey string `json:"issue_key"`
	Comment  string `json:"comment"`
}

type GetCommentsRequest struct {
	IssueKey string `json:"issue_key"`
}

type AddWorklogRequest struct {
	IssueKey  string `json:"issue_key"`
	TimeSpent string `json:"time_spent"`
	Comment   string `json:"comment,omitempty"`
	Started   string `json:"started,omitempty"`
}

type GetTransitionsRequest struct {
	IssueKey string `json:"issue_key"`
}

type TransitionIssueRequest struct {
	IssueKey     string `json:"issue_key"`
	TransitionID string `json:"transition_id"`
	Comment      string `json:"comment,omitempty"`
}

type ListStatusesRequest struct {
	ProjectKey string `json:"project_key"`
}

type GetIssueHistoryRequest struct {
	IssueKey string `json:"issue_key"`
}

type GetRelatedIssuesRequest struct {
	IssueKey string `json:"issue_key"`
}

type LinkIssuesRequest struct {
	InwardIssue  string `json:"inward_issue"`
	OutwardIssue string `json:"outward_issue"`
	LinkType     string `json:"link_type"`
	Comment      string `json:"comment,omitempty"`
}

type GetVersionRequest struct {
	VersionID string `json:"version_id"`
}

type ListProjectVersionsRequest struct {
	ProjectKey string `json:"project_key"`
}

type GetDevelopmentInfoRequest struct {
	IssueKey            string `json:"issue_key"`
	IncludeBranches     bool   `json:"include_branches,omitempty"`
	IncludePullRequests bool   `json:"include_pull_requests,omitempty"`
	IncludeCommits      bool   `json:"include_commits,omitempty"`
	IncludeBuilds       bool   `json:"include_builds,omitempty"`
}

func (r GetDevelopmentInfoRequest) Normalized() GetDevelopmentInfoRequest {
	if !r.IncludeBranches && !r.IncludePullRequests && !r.IncludeCommits && !r.IncludeBuilds {
		r.IncludeBranches = true
		r.IncludePullRequests = true
		r.IncludeCommits = true
		r.IncludeBuilds = true
	}
	return r
}

type DownloadAttachmentRequest struct {
	AttachmentID string `json:"attachment_id"`
}

func HasText(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}
