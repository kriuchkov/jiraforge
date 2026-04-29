package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkflowCommandValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		wantError string
	}{
		{name: "add worklog", args: []string{"add-worklog"}, wantError: "--issue-key and --time-spent are required"},
		{name: "get transitions", args: []string{"get-transitions"}, wantError: "--issue-key is required"},
		{name: "transition issue", args: []string{"transition-issue"}, wantError: "--issue-key and --transition-id are required"},
		{name: "list statuses", args: []string{"list-statuses"}, wantError: "--project-key is required"},
		{name: "get issue history", args: []string{"get-issue-history"}, wantError: "--issue-key is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertValidationError(t, tt.args, tt.wantError)
		})
	}
}

func TestWorkflowCommandDispatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		service *stubCLIService
		assert  func(*testing.T, *stubCLIService, string, string)
	}{
		{
			name:    "add worklog",
			args:    []string{"add-worklog", "--issue-key", "PROJ-7", "--time-spent", "1h30m", "--comment", "done", "--started", "2026-04-09T10:00:00Z"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "AddWorklog", service.lastCall)
				require.Equal(t, "PROJ-7", service.addWorklogReq.IssueKey)
				require.Equal(t, "1h30m", service.addWorklogReq.TimeSpent)
				require.Equal(t, "done", service.addWorklogReq.Comment)
				require.Equal(t, "2026-04-09T10:00:00Z", service.addWorklogReq.Started)
			},
		},
		{
			name:    "get transitions",
			args:    []string{"get-transitions", "--issue-key", "PROJ-7"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "GetTransitions", service.lastCall)
				require.Equal(t, "PROJ-7", service.getTransitionsReq.IssueKey)
			},
		},
		{
			name:    "transition issue",
			args:    []string{"transition-issue", "--issue-key", "PROJ-7", "--transition-id", "31", "--comment", "ship it"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "TransitionIssue", service.lastCall)
				require.Equal(t, "PROJ-7", service.transitionIssueReq.IssueKey)
				require.Equal(t, "31", service.transitionIssueReq.TransitionID)
				require.Equal(t, "ship it", service.transitionIssueReq.Comment)
			},
		},
		{
			name:    "list statuses",
			args:    []string{"list-statuses", "--project-key", "PROJ"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "ListStatuses", service.lastCall)
				require.Equal(t, "PROJ", service.listStatusesReq.ProjectKey)
			},
		},
		{
			name:    "get issue history",
			args:    []string{"get-issue-history", "--issue-key", "PROJ-7"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "GetIssueHistory", service.lastCall)
				require.Equal(t, "PROJ-7", service.getIssueHistoryReq.IssueKey)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertDispatch(t, tt.args, tt.service, tt.assert)
		})
	}
}
