package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSprintCommandValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		wantError string
	}{
		{name: "list sprints", args: []string{"list-sprints"}, wantError: "either --board-id or --project-key is required"},
		{name: "get sprint", args: []string{"get-sprint"}, wantError: "--sprint-id is required"},
		{name: "get sprint report", args: []string{"get-sprint-report"}, wantError: "--sprint-id is required"},
		{name: "get sprint health report", args: []string{"get-sprint-health-report"}, wantError: "--sprint-id is required"},
		{name: "get active sprint", args: []string{"get-active-sprint"}, wantError: "either --board-id or --project-key is required"},
		{name: "search sprint missing name", args: []string{"search-sprint"}, wantError: "--name is required"},
		{name: "search sprint missing scope", args: []string{"search-sprint", "--name", "Sprint 1"}, wantError: "either --board-id or --project-key is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertValidationError(t, tt.args, tt.wantError)
		})
	}
}

func TestSprintCommandDispatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		service *stubCLIService
		assert  func(*testing.T, *stubCLIService, string, string)
	}{
		{
			name:    "list sprints",
			args:    []string{"list-sprints", "--board-id", "42"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "ListSprints", service.lastCall)
				require.Equal(t, "42", service.listSprintsReq.BoardID)
			},
		},
		{
			name:    "get sprint",
			args:    []string{"get-sprint", "--sprint-id", "42"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "GetSprint", service.lastCall)
				require.Equal(t, "42", service.getSprintReq.SprintID)
			},
		},
		{
			name:    "get sprint report",
			args:    []string{"get-sprint-report", "--sprint-id", "42"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "GetSprintReport", service.lastCall)
				require.Equal(t, "42", service.getSprintReportReq.SprintID)
			},
		},
		{
			name:    "get sprint health report",
			args:    []string{"get-sprint-health-report", "--sprint-id", "42"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "GetSprintHealthReport", service.lastCall)
				require.Equal(t, "42", service.getSprintHealthReportReq.SprintID)
			},
		},
		{
			name:    "get active sprint",
			args:    []string{"get-active-sprint", "--project-key", "PROJ"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "GetActiveSprint", service.lastCall)
				require.Equal(t, "PROJ", service.getActiveSprintReq.ProjectKey)
			},
		},
		{
			name:    "search sprint",
			args:    []string{"search-sprint", "--name", "Sprint 10", "--project-key", "PROJ", "--exact-match"},
			service: &stubCLIService{},
			assert: func(t *testing.T, service *stubCLIService, _, _ string) {
				require.Equal(t, "SearchSprints", service.lastCall)
				require.Equal(t, "Sprint 10", service.searchSprintsReq.Name)
				require.Equal(t, "PROJ", service.searchSprintsReq.ProjectKey)
				require.True(t, service.searchSprintsReq.ExactMatch)
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
