package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/kriuchkov/jiraforge/internal/core/models"
	"github.com/kriuchkov/jiraforge/internal/core/ports"
	"github.com/stretchr/testify/require"
)

func newStubFactory(service *stubCLIService) ServiceFactory {
	return func(envFile string) (ports.JiraService, error) {
		if service != nil {
			service.envArg = envFile
			return service, nil
		}
		return &stubCLIService{}, nil
	}
}

func executeCLI(t *testing.T, args []string, service *stubCLIService, factory ServiceFactory) (string, string, int) {
	t.Helper()

	var stdout, stderr strings.Builder
	exitCode := Execute(context.Background(), args, &stdout, &stderr, factory)
	if service != nil {
		service.stdout = stdout.String()
		service.stderr = stderr.String()
	}
	return stdout.String(), stderr.String(), exitCode
}

func assertValidationError(t *testing.T, args []string, wantError string) {
	t.Helper()

	factoryCalls := 0
	stdout, stderr, exitCode := executeCLI(t, args, nil, func(string) (ports.JiraService, error) {
		factoryCalls++
		return &stubCLIService{}, nil
	})

	require.Equal(t, 1, exitCode)
	require.Zero(t, factoryCalls)
	require.Empty(t, stdout)
	require.Contains(t, stderr, wantError)
}

func assertDispatch(t *testing.T, args []string, service *stubCLIService, assert func(*testing.T, *stubCLIService, string, string)) {
	t.Helper()

	require.NotNil(t, service)
	stdout, stderr, exitCode := executeCLI(t, args, service, newStubFactory(service))
	require.Equal(t, 0, exitCode, "stderr=%q", stderr)
	assert(t, service, stdout, stderr)
}

type stubCLIService struct {
	lastCall                   string
	envArg                     string
	stdout                     string
	stderr                     string
	getIssueReq                models.GetIssueRequest
	createIssueReq             models.CreateIssueRequest
	createChildIssueReq        models.CreateChildIssueRequest
	updateIssueReq             models.UpdateIssueRequest
	deleteIssueReq             models.DeleteIssueRequest
	listIssueTypesReq          models.ListIssueTypesRequest
	getAgingReportReq          models.GetAgingReportRequest
	getBlockedIssuesReportReq  models.GetBlockedIssuesReportRequest
	getFlowEfficiencyReportReq models.GetFlowEfficiencyReportRequest
	getSprintWorkloadReportReq models.GetSprintWorkloadReportRequest
	getUtilizationReportReq    models.GetUtilizationReportRequest
	getCycleTimeReportReq      models.GetCycleTimeReportRequest
	getTeamWipSnapshotReq      models.GetTeamWipSnapshotRequest
	getWorklogsReq             models.GetWorklogsRequest
	searchIssuesReq            models.SearchIssuesRequest
	listSprintsReq             models.ListSprintsRequest
	getSprintReq               models.GetSprintRequest
	getSprintReportReq         models.GetSprintReportRequest
	getSprintHealthReportReq   models.GetSprintHealthReportRequest
	getActiveSprintReq         models.GetActiveSprintRequest
	searchSprintsReq           models.SearchSprintsRequest
	addCommentReq              models.AddCommentRequest
	getCommentsReq             models.GetCommentsRequest
	addWorklogReq              models.AddWorklogRequest
	getTransitionsReq          models.GetTransitionsRequest
	transitionIssueReq         models.TransitionIssueRequest
	listStatusesReq            models.ListStatusesRequest
	getIssueHistoryReq         models.GetIssueHistoryRequest
	getRelatedIssuesReq        models.GetRelatedIssuesRequest
	linkIssuesReq              models.LinkIssuesRequest
	getVersionReq              models.GetVersionRequest
	listProjectVersionsReq     models.ListProjectVersionsRequest
	getDevelopmentInfoReq      models.GetDevelopmentInfoRequest
	downloadAttachmentReq      models.DownloadAttachmentRequest
	getIssueResult             *models.Issue
	searchIssuesResult         *models.SearchIssuesResult
}

func (s *stubCLIService) GetIssue(_ context.Context, req models.GetIssueRequest) (*models.Issue, error) {
	s.lastCall = "GetIssue"
	s.getIssueReq = req
	if s.getIssueResult == nil {
		return &models.Issue{}, nil
	}
	return s.getIssueResult, nil
}

func (s *stubCLIService) CreateIssue(_ context.Context, req models.CreateIssueRequest) (*models.IssueMutationResult, error) {
	s.lastCall = "CreateIssue"
	s.createIssueReq = req
	return &models.IssueMutationResult{}, nil
}

func (s *stubCLIService) CreateChildIssue(_ context.Context, req models.CreateChildIssueRequest) (*models.IssueMutationResult, error) {
	s.lastCall = "CreateChildIssue"
	s.createChildIssueReq = req
	return &models.IssueMutationResult{}, nil
}

func (s *stubCLIService) UpdateIssue(_ context.Context, req models.UpdateIssueRequest) (*models.IssueMutationResult, error) {
	s.lastCall = "UpdateIssue"
	s.updateIssueReq = req
	return &models.IssueMutationResult{}, nil
}

func (s *stubCLIService) DeleteIssue(_ context.Context, req models.DeleteIssueRequest) (*models.IssueMutationResult, error) {
	s.lastCall = "DeleteIssue"
	s.deleteIssueReq = req
	return &models.IssueMutationResult{}, nil
}

func (s *stubCLIService) ListIssueTypes(_ context.Context, req models.ListIssueTypesRequest) ([]models.IssueType, error) {
	s.lastCall = "ListIssueTypes"
	s.listIssueTypesReq = req
	return []models.IssueType{}, nil
}

func (s *stubCLIService) GetAgingReport(_ context.Context, req models.GetAgingReportRequest) (*models.AgingReport, error) {
	s.lastCall = "GetAgingReport"
	s.getAgingReportReq = req
	return &models.AgingReport{}, nil
}

func (s *stubCLIService) GetBlockedIssuesReport(_ context.Context, req models.GetBlockedIssuesReportRequest) (*models.BlockedIssuesReport, error) {
	s.lastCall = "GetBlockedIssuesReport"
	s.getBlockedIssuesReportReq = req
	return &models.BlockedIssuesReport{}, nil
}

func (s *stubCLIService) GetFlowEfficiencyReport(_ context.Context, req models.GetFlowEfficiencyReportRequest) (*models.FlowEfficiencyReport, error) {
	s.lastCall = "GetFlowEfficiencyReport"
	s.getFlowEfficiencyReportReq = req
	return &models.FlowEfficiencyReport{}, nil
}

func (s *stubCLIService) GetSprintWorkloadReport(_ context.Context, req models.GetSprintWorkloadReportRequest) (*models.SprintWorkloadReport, error) {
	s.lastCall = "GetSprintWorkloadReport"
	s.getSprintWorkloadReportReq = req
	return &models.SprintWorkloadReport{}, nil
}

func (s *stubCLIService) GetUtilizationReport(_ context.Context, req models.GetUtilizationReportRequest) (*models.UtilizationReport, error) {
	s.lastCall = "GetUtilizationReport"
	s.getUtilizationReportReq = req
	return &models.UtilizationReport{}, nil
}

func (s *stubCLIService) GetCycleTimeReport(_ context.Context, req models.GetCycleTimeReportRequest) (*models.CycleTimeReport, error) {
	s.lastCall = "GetCycleTimeReport"
	s.getCycleTimeReportReq = req
	return &models.CycleTimeReport{}, nil
}

func (s *stubCLIService) GetTeamWipSnapshot(_ context.Context, req models.GetTeamWipSnapshotRequest) (*models.TeamWipSnapshot, error) {
	s.lastCall = "GetTeamWipSnapshot"
	s.getTeamWipSnapshotReq = req
	return &models.TeamWipSnapshot{}, nil
}

func (s *stubCLIService) GetWorklogs(_ context.Context, req models.GetWorklogsRequest) (*models.WorklogList, error) {
	s.lastCall = "GetWorklogs"
	s.getWorklogsReq = req
	return &models.WorklogList{IssueKey: req.IssueKey}, nil
}

func (s *stubCLIService) SearchIssues(_ context.Context, req models.SearchIssuesRequest) (*models.SearchIssuesResult, error) {
	s.lastCall = "SearchIssues"
	s.searchIssuesReq = req
	if s.searchIssuesResult == nil {
		return &models.SearchIssuesResult{}, nil
	}
	return s.searchIssuesResult, nil
}

func (s *stubCLIService) ListSprints(_ context.Context, req models.ListSprintsRequest) (*models.SprintCollection, error) {
	s.lastCall = "ListSprints"
	s.listSprintsReq = req
	return &models.SprintCollection{}, nil
}

func (s *stubCLIService) GetSprint(_ context.Context, req models.GetSprintRequest) (*models.Sprint, error) {
	s.lastCall = "GetSprint"
	s.getSprintReq = req
	return &models.Sprint{}, nil
}

func (s *stubCLIService) GetSprintReport(_ context.Context, req models.GetSprintReportRequest) (*models.SprintReport, error) {
	s.lastCall = "GetSprintReport"
	s.getSprintReportReq = req
	return &models.SprintReport{}, nil
}

func (s *stubCLIService) GetSprintHealthReport(_ context.Context, req models.GetSprintHealthReportRequest) (*models.SprintHealthReport, error) {
	s.lastCall = "GetSprintHealthReport"
	s.getSprintHealthReportReq = req
	return &models.SprintHealthReport{}, nil
}

func (s *stubCLIService) GetActiveSprint(_ context.Context, req models.GetActiveSprintRequest) (*models.Sprint, error) {
	s.lastCall = "GetActiveSprint"
	s.getActiveSprintReq = req
	return &models.Sprint{}, nil
}

func (s *stubCLIService) SearchSprints(_ context.Context, req models.SearchSprintsRequest) (*models.SprintCollection, error) {
	s.lastCall = "SearchSprints"
	s.searchSprintsReq = req
	return &models.SprintCollection{}, nil
}

func (s *stubCLIService) AddComment(_ context.Context, req models.AddCommentRequest) (*models.Comment, error) {
	s.lastCall = "AddComment"
	s.addCommentReq = req
	return &models.Comment{}, nil
}

func (s *stubCLIService) GetComments(_ context.Context, req models.GetCommentsRequest) (*models.CommentList, error) {
	s.lastCall = "GetComments"
	s.getCommentsReq = req
	return &models.CommentList{}, nil
}

func (s *stubCLIService) AddWorklog(_ context.Context, req models.AddWorklogRequest) (*models.Worklog, error) {
	s.lastCall = "AddWorklog"
	s.addWorklogReq = req
	return &models.Worklog{}, nil
}

func (s *stubCLIService) GetTransitions(_ context.Context, req models.GetTransitionsRequest) ([]models.Transition, error) {
	s.lastCall = "GetTransitions"
	s.getTransitionsReq = req
	return []models.Transition{}, nil
}

func (s *stubCLIService) TransitionIssue(_ context.Context, req models.TransitionIssueRequest) (*models.IssueMutationResult, error) {
	s.lastCall = "TransitionIssue"
	s.transitionIssueReq = req
	return &models.IssueMutationResult{}, nil
}

func (s *stubCLIService) ListStatuses(_ context.Context, req models.ListStatusesRequest) (*models.StatusCatalog, error) {
	s.lastCall = "ListStatuses"
	s.listStatusesReq = req
	return &models.StatusCatalog{}, nil
}

func (s *stubCLIService) GetIssueHistory(_ context.Context, req models.GetIssueHistoryRequest) (*models.IssueHistory, error) {
	s.lastCall = "GetIssueHistory"
	s.getIssueHistoryReq = req
	return &models.IssueHistory{}, nil
}

func (s *stubCLIService) GetRelatedIssues(_ context.Context, req models.GetRelatedIssuesRequest) ([]models.IssueRelation, error) {
	s.lastCall = "GetRelatedIssues"
	s.getRelatedIssuesReq = req
	return []models.IssueRelation{}, nil
}

func (s *stubCLIService) LinkIssues(_ context.Context, req models.LinkIssuesRequest) (*models.IssueMutationResult, error) {
	s.lastCall = "LinkIssues"
	s.linkIssuesReq = req
	return &models.IssueMutationResult{}, nil
}

func (s *stubCLIService) GetVersion(_ context.Context, req models.GetVersionRequest) (*models.Version, error) {
	s.lastCall = "GetVersion"
	s.getVersionReq = req
	return &models.Version{}, nil
}

func (s *stubCLIService) ListProjectVersions(_ context.Context, req models.ListProjectVersionsRequest) (*models.VersionCollection, error) {
	s.lastCall = "ListProjectVersions"
	s.listProjectVersionsReq = req
	return &models.VersionCollection{}, nil
}

func (s *stubCLIService) GetDevelopmentInfo(_ context.Context, req models.GetDevelopmentInfoRequest) (*models.DevelopmentInfo, error) {
	s.lastCall = "GetDevelopmentInfo"
	s.getDevelopmentInfoReq = req
	return &models.DevelopmentInfo{}, nil
}

func (s *stubCLIService) DownloadAttachment(_ context.Context, req models.DownloadAttachmentRequest) (*models.StoredAttachment, error) {
	s.lastCall = "DownloadAttachment"
	s.downloadAttachmentReq = req
	return &models.StoredAttachment{}, nil
}
