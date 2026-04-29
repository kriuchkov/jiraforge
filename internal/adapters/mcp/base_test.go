package mcp

import (
	"testing"

	"github.com/kriuchkov/jiraforge/internal/adapters/mcp/mocks"
	"github.com/stretchr/testify/require"
)

func TestSplitCSV(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "empty", input: "", want: nil},
		{name: "whitespace", input: "   ", want: nil},
		{name: "single value", input: "summary", want: []string{"summary"}},
		{name: "trimmed values", input: " summary , status , changelog ", want: []string{"summary", "status", "changelog"}},
		{name: "skip empty values", input: "summary,, status, ", want: []string{"summary", "status"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, splitCSV(tt.input))
		})
	}
}

func TestNewServerInitializeCapabilities(t *testing.T) {
	t.Parallel()

	srv := NewServer("JiraForge Test", "9.9.9", mocks.NewJiraService(t))
	result := initializeServer(t, srv)

	require.Equal(t, "2025-03-26", result.ProtocolVersion)
	require.Equal(t, "JiraForge Test", result.ServerInfo.Name)
	require.Equal(t, "9.9.9", result.ServerInfo.Version)
	require.NotNil(t, result.Capabilities.Logging)
	require.NotNil(t, result.Capabilities.Tools)
	require.True(t, result.Capabilities.Tools.ListChanged)
	require.NotNil(t, result.Capabilities.Prompts)
	require.True(t, result.Capabilities.Prompts.ListChanged)
	require.NotNil(t, result.Capabilities.Resources)
	require.True(t, result.Capabilities.Resources.Subscribe)
	require.True(t, result.Capabilities.Resources.ListChanged)
}

func TestNewServerRegistersAllTools(t *testing.T) {
	t.Parallel()

	srv := NewServer("JiraForge Test", "9.9.9", mocks.NewJiraService(t))
	tools := srv.ListTools()

	require.Len(t, tools, 34)
	for _, name := range []string{
		"jira_get_issue",
		"jira_create_issue",
		"jira_create_child_issue",
		"jira_update_issue",
		"jira_delete_issue",
		"jira_list_issue_types",
		"jira_search_issue",
		"jira_get_aging_report",
		"jira_get_blocked_issues_report",
		"jira_get_flow_efficiency_report",
		"jira_get_sprint_workload_report",
		"jira_get_utilization_report",
		"jira_get_cycle_time_report",
		"jira_get_team_wip_snapshot",
		"jira_list_sprints",
		"jira_get_sprint",
		"jira_get_sprint_report",
		"jira_get_sprint_health_report",
		"jira_get_active_sprint",
		"jira_search_sprint_by_name",
		"jira_add_comment",
		"jira_get_comments",
		"jira_add_worklog",
		"jira_get_worklogs",
		"jira_get_transitions",
		"jira_transition_issue",
		"jira_list_statuses",
		"jira_get_issue_history",
		"jira_get_related_issues",
		"jira_link_issues",
		"jira_get_version",
		"jira_list_project_versions",
		"jira_get_development_information",
		"jira_download_attachment",
	} {
		require.Contains(t, tools, name)
	}
}
