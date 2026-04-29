package jira

import (
	"context"
	"errors"
	"testing"

	coreerrors "github.com/kriuchkov/jiraforge/internal/core/errors"
	"github.com/kriuchkov/jiraforge/internal/core/models"
)

func TestNewServiceRequiresDependencies(t *testing.T) {
	store := &stubAttachmentStore{}
	gateway := &stubGateway{}

	if _, err := NewService(nil, store); err == nil {
		t.Fatal("expected nil gateway to fail")
	}
	if _, err := NewService(gateway, nil); err == nil {
		t.Fatal("expected nil attachment store to fail")
	}
	if _, err := NewService(gateway, store); err != nil {
		t.Fatalf("expected valid dependencies to succeed, got %v", err)
	}
}

func TestServiceValidationRejectsInvalidRequests(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		run  func(*Service) error
	}{
		{name: "GetIssue", run: func(s *Service) error { _, err := s.GetIssue(ctx, models.GetIssueRequest{IssueKey: "bad"}); return err }},
		{name: "CreateIssue", run: func(s *Service) error { _, err := s.CreateIssue(ctx, models.CreateIssueRequest{}); return err }},
		{name: "CreateChildIssue", run: func(s *Service) error {
			_, err := s.CreateChildIssue(ctx, models.CreateChildIssueRequest{ParentIssueKey: "bad", Summary: "x", Description: "y"})
			return err
		}},
		{name: "UpdateIssue", run: func(s *Service) error {
			_, err := s.UpdateIssue(ctx, models.UpdateIssueRequest{IssueKey: "PROJ-1"})
			return err
		}},
		{name: "DeleteIssue", run: func(s *Service) error {
			_, err := s.DeleteIssue(ctx, models.DeleteIssueRequest{IssueKey: "bad"})
			return err
		}},
		{name: "ListIssueTypes", run: func(s *Service) error { _, err := s.ListIssueTypes(ctx, models.ListIssueTypesRequest{}); return err }},
		{name: "GetAgingReport", run: func(s *Service) error { _, err := s.GetAgingReport(ctx, models.GetAgingReportRequest{}); return err }},
		{name: "GetBlockedIssuesReport", run: func(s *Service) error {
			_, err := s.GetBlockedIssuesReport(ctx, models.GetBlockedIssuesReportRequest{})
			return err
		}},
		{name: "GetFlowEfficiencyReport", run: func(s *Service) error {
			_, err := s.GetFlowEfficiencyReport(ctx, models.GetFlowEfficiencyReportRequest{})
			return err
		}},
		{name: "SearchIssues", run: func(s *Service) error { _, err := s.SearchIssues(ctx, models.SearchIssuesRequest{}); return err }},
		{name: "ListSprints", run: func(s *Service) error { _, err := s.ListSprints(ctx, models.ListSprintsRequest{}); return err }},
		{name: "GetSprint", run: func(s *Service) error { _, err := s.GetSprint(ctx, models.GetSprintRequest{}); return err }},
		{name: "GetSprintReport", run: func(s *Service) error { _, err := s.GetSprintReport(ctx, models.GetSprintReportRequest{}); return err }},
		{name: "GetSprintHealthReport", run: func(s *Service) error {
			_, err := s.GetSprintHealthReport(ctx, models.GetSprintHealthReportRequest{})
			return err
		}},
		{name: "GetActiveSprint", run: func(s *Service) error { _, err := s.GetActiveSprint(ctx, models.GetActiveSprintRequest{}); return err }},
		{name: "SearchSprints", run: func(s *Service) error {
			_, err := s.SearchSprints(ctx, models.SearchSprintsRequest{Name: "Sprint 1"})
			return err
		}},
		{name: "AddComment", run: func(s *Service) error {
			_, err := s.AddComment(ctx, models.AddCommentRequest{IssueKey: "bad", Comment: "hello"})
			return err
		}},
		{name: "GetComments", run: func(s *Service) error {
			_, err := s.GetComments(ctx, models.GetCommentsRequest{IssueKey: "bad"})
			return err
		}},
		{name: "AddWorklog", run: func(s *Service) error {
			_, err := s.AddWorklog(ctx, models.AddWorklogRequest{IssueKey: "PROJ-1"})
			return err
		}},
		{name: "GetTransitions", run: func(s *Service) error {
			_, err := s.GetTransitions(ctx, models.GetTransitionsRequest{IssueKey: "bad"})
			return err
		}},
		{name: "TransitionIssue", run: func(s *Service) error {
			_, err := s.TransitionIssue(ctx, models.TransitionIssueRequest{IssueKey: "PROJ-1"})
			return err
		}},
		{name: "ListStatuses", run: func(s *Service) error { _, err := s.ListStatuses(ctx, models.ListStatusesRequest{}); return err }},
		{name: "GetIssueHistory", run: func(s *Service) error {
			_, err := s.GetIssueHistory(ctx, models.GetIssueHistoryRequest{IssueKey: "bad"})
			return err
		}},
		{name: "GetRelatedIssues", run: func(s *Service) error {
			_, err := s.GetRelatedIssues(ctx, models.GetRelatedIssuesRequest{IssueKey: "bad"})
			return err
		}},
		{name: "LinkIssues", run: func(s *Service) error {
			_, err := s.LinkIssues(ctx, models.LinkIssuesRequest{InwardIssue: "bad", OutwardIssue: "PROJ-2", LinkType: "Blocks"})
			return err
		}},
		{name: "GetVersion", run: func(s *Service) error { _, err := s.GetVersion(ctx, models.GetVersionRequest{}); return err }},
		{name: "ListProjectVersions", run: func(s *Service) error {
			_, err := s.ListProjectVersions(ctx, models.ListProjectVersionsRequest{})
			return err
		}},
		{name: "GetDevelopmentInfo", run: func(s *Service) error {
			_, err := s.GetDevelopmentInfo(ctx, models.GetDevelopmentInfoRequest{IssueKey: "bad"})
			return err
		}},
		{name: "DownloadAttachment", run: func(s *Service) error {
			_, err := s.DownloadAttachment(ctx, models.DownloadAttachmentRequest{})
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gateway := &stubGateway{}
			store := &stubAttachmentStore{}
			service := &Service{gateway: gateway, attachmentStore: store}

			err := tt.run(service)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if got := coreerrors.CodeOf(err); got != coreerrors.CodeInvalidArgument {
				t.Fatalf("unexpected error code: got %s want %s", got, coreerrors.CodeInvalidArgument)
			}
			if gateway.calls != 0 {
				t.Fatalf("expected gateway not to be called, got %d calls", gateway.calls)
			}
			if store.calls != 0 {
				t.Fatalf("expected store not to be called, got %d calls", store.calls)
			}
		})
	}
}

func TestGetIssueNormalizesBeforeDelegating(t *testing.T) {
	ctx := context.Background()
	gateway := &stubGateway{
		getIssueResult: &models.Issue{Key: "PROJ-42"},
	}
	service := &Service{gateway: gateway, attachmentStore: &stubAttachmentStore{}}

	issue, err := service.GetIssue(ctx, models.GetIssueRequest{IssueKey: "PROJ-42", Fields: []string{"summary"}})
	if err != nil {
		t.Fatalf("GetIssue returned error: %v", err)
	}
	if issue == nil || issue.Key != "PROJ-42" {
		t.Fatalf("unexpected issue result: %+v", issue)
	}
	if gateway.calls != 1 {
		t.Fatalf("expected one gateway call, got %d", gateway.calls)
	}
	if gateway.getIssueReq.IssueKey != "PROJ-42" {
		t.Fatalf("unexpected issue key: %q", gateway.getIssueReq.IssueKey)
	}
	if len(gateway.getIssueReq.Expand) != 4 {
		t.Fatalf("expected default expand fields, got %+v", gateway.getIssueReq.Expand)
	}
	if len(gateway.getIssueReq.Fields) != 1 || gateway.getIssueReq.Fields[0] != "summary" {
		t.Fatalf("unexpected fields: %+v", gateway.getIssueReq.Fields)
	}
}

func TestSearchIssuesNormalizesBeforeDelegating(t *testing.T) {
	ctx := context.Background()
	gateway := &stubGateway{
		searchIssuesResult: &models.SearchIssuesResult{Total: 1},
	}
	service := &Service{gateway: gateway, attachmentStore: &stubAttachmentStore{}}

	result, err := service.SearchIssues(ctx, models.SearchIssuesRequest{JQL: "project = PROJ"})
	if err != nil {
		t.Fatalf("SearchIssues returned error: %v", err)
	}
	if result == nil || result.Total != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if gateway.searchIssuesReq.MaxResults != 30 {
		t.Fatalf("expected default max results, got %d", gateway.searchIssuesReq.MaxResults)
	}
	if len(gateway.searchIssuesReq.Expand) != 4 {
		t.Fatalf("expected default expand fields, got %+v", gateway.searchIssuesReq.Expand)
	}
}

func TestGetSprintReportDelegatesToGateway(t *testing.T) {
	ctx := context.Background()
	gateway := &stubGateway{
		sprintReportResult: &models.SprintReport{Sprint: models.Sprint{ID: 7, Name: "Sprint 7"}},
	}
	service := &Service{gateway: gateway, attachmentStore: &stubAttachmentStore{}}

	report, err := service.GetSprintReport(ctx, models.GetSprintReportRequest{SprintID: "7"})
	if err != nil {
		t.Fatalf("GetSprintReport returned error: %v", err)
	}
	if report == nil || report.Sprint.ID != 7 {
		t.Fatalf("unexpected report: %+v", report)
	}
	if gateway.getSprintReportReq.SprintID != "7" {
		t.Fatalf("unexpected sprint report request: %+v", gateway.getSprintReportReq)
	}
}

func TestGetDevelopmentInfoAppliesDefaultFlags(t *testing.T) {
	ctx := context.Background()
	gateway := &stubGateway{
		developmentInfoResult: &models.DevelopmentInfo{IssueKey: "PROJ-42"},
	}
	service := &Service{gateway: gateway, attachmentStore: &stubAttachmentStore{}}

	result, err := service.GetDevelopmentInfo(ctx, models.GetDevelopmentInfoRequest{IssueKey: "PROJ-42"})
	if err != nil {
		t.Fatalf("GetDevelopmentInfo returned error: %v", err)
	}
	if result == nil || result.IssueKey != "PROJ-42" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if !gateway.developmentInfoReq.IncludeBranches || !gateway.developmentInfoReq.IncludePullRequests || !gateway.developmentInfoReq.IncludeCommits || !gateway.developmentInfoReq.IncludeBuilds {
		t.Fatalf("expected default include flags to be enabled, got %+v", gateway.developmentInfoReq)
	}
}

func TestDownloadAttachmentSavesGatewayPayload(t *testing.T) {
	ctx := context.Background()
	content := &models.AttachmentContent{ID: "att-1", Filename: "dump.txt", MimeType: "text/plain", Size: 7, Data: []byte("payload")}
	stored := &models.StoredAttachment{Path: "/tmp/dump.txt", Filename: "dump.txt", MimeType: "text/plain", Size: 7}
	gateway := &stubGateway{downloadAttachmentResult: content}
	store := &stubAttachmentStore{saveResult: stored}
	service := &Service{gateway: gateway, attachmentStore: store}

	result, err := service.DownloadAttachment(ctx, models.DownloadAttachmentRequest{AttachmentID: "att-1"})
	if err != nil {
		t.Fatalf("DownloadAttachment returned error: %v", err)
	}
	if result == nil || result.Path != "/tmp/dump.txt" {
		t.Fatalf("unexpected stored attachment: %+v", result)
	}
	if gateway.downloadAttachmentReq.AttachmentID != "att-1" {
		t.Fatalf("unexpected attachment request: %+v", gateway.downloadAttachmentReq)
	}
	if store.calls != 1 {
		t.Fatalf("expected one store call, got %d", store.calls)
	}
	if store.savedAttachment.ID != "att-1" || string(store.savedAttachment.Data) != "payload" {
		t.Fatalf("unexpected saved payload: %+v", store.savedAttachment)
	}
}

func TestDownloadAttachmentPropagatesStoreError(t *testing.T) {
	ctx := context.Background()
	storeErr := errors.New("disk full")
	gateway := &stubGateway{downloadAttachmentResult: &models.AttachmentContent{ID: "att-1", Data: []byte("payload")}}
	store := &stubAttachmentStore{saveErr: storeErr}
	service := &Service{gateway: gateway, attachmentStore: store}

	_, err := service.DownloadAttachment(ctx, models.DownloadAttachmentRequest{AttachmentID: "att-1"})
	if !errors.Is(err, storeErr) {
		t.Fatalf("expected store error to propagate, got %v", err)
	}
}

type stubGateway struct {
	calls                 int
	getIssueReq           models.GetIssueRequest
	searchIssuesReq       models.SearchIssuesRequest
	getSprintReportReq    models.GetSprintReportRequest
	getIssueHistoryReqs   []models.GetIssueHistoryRequest
	developmentInfoReq    models.GetDevelopmentInfoRequest
	downloadAttachmentReq models.DownloadAttachmentRequest

	getIssueResult           *models.Issue
	searchIssuesResult       *models.SearchIssuesResult
	sprintReportResult       *models.SprintReport
	issueHistoryResults      map[string]*models.IssueHistory
	developmentInfoResult    *models.DevelopmentInfo
	downloadAttachmentResult *models.AttachmentContent
}

func (s *stubGateway) GetIssue(_ context.Context, req models.GetIssueRequest) (*models.Issue, error) {
	s.calls++
	s.getIssueReq = req
	if s.getIssueResult == nil {
		return &models.Issue{}, nil
	}
	return s.getIssueResult, nil
}

func (s *stubGateway) CreateIssue(_ context.Context, _ models.CreateIssueRequest) (*models.IssueMutationResult, error) {
	s.calls++
	return &models.IssueMutationResult{}, nil
}

func (s *stubGateway) CreateChildIssue(_ context.Context, _ models.CreateChildIssueRequest) (*models.IssueMutationResult, error) {
	s.calls++
	return &models.IssueMutationResult{}, nil
}

func (s *stubGateway) UpdateIssue(_ context.Context, _ models.UpdateIssueRequest) (*models.IssueMutationResult, error) {
	s.calls++
	return &models.IssueMutationResult{}, nil
}

func (s *stubGateway) DeleteIssue(_ context.Context, _ models.DeleteIssueRequest) (*models.IssueMutationResult, error) {
	s.calls++
	return &models.IssueMutationResult{}, nil
}

func (s *stubGateway) ListIssueTypes(_ context.Context, _ models.ListIssueTypesRequest) ([]models.IssueType, error) {
	s.calls++
	return nil, nil
}

func (s *stubGateway) SearchIssues(_ context.Context, req models.SearchIssuesRequest) (*models.SearchIssuesResult, error) {
	s.calls++
	s.searchIssuesReq = req
	if s.searchIssuesResult == nil {
		return &models.SearchIssuesResult{}, nil
	}
	return s.searchIssuesResult, nil
}

func (s *stubGateway) ListSprints(_ context.Context, _ models.ListSprintsRequest) (*models.SprintCollection, error) {
	s.calls++
	return &models.SprintCollection{}, nil
}

func (s *stubGateway) GetSprint(_ context.Context, _ models.GetSprintRequest) (*models.Sprint, error) {
	s.calls++
	return &models.Sprint{}, nil
}

func (s *stubGateway) GetSprintReport(_ context.Context, req models.GetSprintReportRequest) (*models.SprintReport, error) {
	s.calls++
	s.getSprintReportReq = req
	if s.sprintReportResult == nil {
		return &models.SprintReport{}, nil
	}
	return s.sprintReportResult, nil
}

func (s *stubGateway) GetActiveSprint(_ context.Context, _ models.GetActiveSprintRequest) (*models.Sprint, error) {
	s.calls++
	return &models.Sprint{}, nil
}

func (s *stubGateway) SearchSprints(_ context.Context, _ models.SearchSprintsRequest) (*models.SprintCollection, error) {
	s.calls++
	return &models.SprintCollection{}, nil
}

func (s *stubGateway) AddComment(_ context.Context, _ models.AddCommentRequest) (*models.Comment, error) {
	s.calls++
	return &models.Comment{}, nil
}

func (s *stubGateway) GetComments(_ context.Context, _ models.GetCommentsRequest) (*models.CommentList, error) {
	s.calls++
	return &models.CommentList{}, nil
}

func (s *stubGateway) AddWorklog(_ context.Context, _ models.AddWorklogRequest) (*models.Worklog, error) {
	s.calls++
	return &models.Worklog{}, nil
}

func (s *stubGateway) GetWorklogs(_ context.Context, _ models.GetWorklogsRequest) (*models.WorklogList, error) {
	s.calls++
	return &models.WorklogList{}, nil
}

func (s *stubGateway) GetTransitions(_ context.Context, _ models.GetTransitionsRequest) ([]models.Transition, error) {
	s.calls++
	return nil, nil
}

func (s *stubGateway) TransitionIssue(_ context.Context, _ models.TransitionIssueRequest) (*models.IssueMutationResult, error) {
	s.calls++
	return &models.IssueMutationResult{}, nil
}

func (s *stubGateway) ListStatuses(_ context.Context, _ models.ListStatusesRequest) (*models.StatusCatalog, error) {
	s.calls++
	return &models.StatusCatalog{}, nil
}

func (s *stubGateway) GetIssueHistory(_ context.Context, req models.GetIssueHistoryRequest) (*models.IssueHistory, error) {
	s.calls++
	s.getIssueHistoryReqs = append(s.getIssueHistoryReqs, req)
	if s.issueHistoryResults != nil {
		if history, ok := s.issueHistoryResults[req.IssueKey]; ok {
			return history, nil
		}
	}
	return &models.IssueHistory{}, nil
}

func (s *stubGateway) GetRelatedIssues(_ context.Context, _ models.GetRelatedIssuesRequest) ([]models.IssueRelation, error) {
	s.calls++
	return nil, nil
}

func (s *stubGateway) LinkIssues(_ context.Context, _ models.LinkIssuesRequest) (*models.IssueMutationResult, error) {
	s.calls++
	return &models.IssueMutationResult{}, nil
}

func (s *stubGateway) GetVersion(_ context.Context, _ models.GetVersionRequest) (*models.Version, error) {
	s.calls++
	return &models.Version{}, nil
}

func (s *stubGateway) ListProjectVersions(_ context.Context, _ models.ListProjectVersionsRequest) (*models.VersionCollection, error) {
	s.calls++
	return &models.VersionCollection{}, nil
}

func (s *stubGateway) GetDevelopmentInfo(_ context.Context, req models.GetDevelopmentInfoRequest) (*models.DevelopmentInfo, error) {
	s.calls++
	s.developmentInfoReq = req
	if s.developmentInfoResult == nil {
		return &models.DevelopmentInfo{}, nil
	}
	return s.developmentInfoResult, nil
}

func (s *stubGateway) DownloadAttachment(_ context.Context, req models.DownloadAttachmentRequest) (*models.AttachmentContent, error) {
	s.calls++
	s.downloadAttachmentReq = req
	if s.downloadAttachmentResult == nil {
		return &models.AttachmentContent{}, nil
	}
	return s.downloadAttachmentResult, nil
}

type stubAttachmentStore struct {
	calls           int
	savedAttachment models.AttachmentContent
	saveResult      *models.StoredAttachment
	saveErr         error
}

func (s *stubAttachmentStore) Save(_ context.Context, attachment models.AttachmentContent) (*models.StoredAttachment, error) {
	s.calls++
	s.savedAttachment = attachment
	if s.saveErr != nil {
		return nil, s.saveErr
	}
	if s.saveResult == nil {
		return &models.StoredAttachment{}, nil
	}
	return s.saveResult, nil
}
