package atlassian

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	adagile "github.com/ctreminiom/go-atlassian/jira/agile"
	"github.com/stretchr/testify/require"

	coremodels "github.com/kriuchkov/jiraforge/internal/core/models"
)

func TestGatewayGetSprintReport(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/agile/1.0/sprint/7", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":7,"name":"Sprint 7","state":"closed","originBoardId":42,"goal":"Ship sprint report","startDate":"2025-01-01T00:00:00.000Z","endDate":"2025-01-14T00:00:00.000Z","completeDate":"2025-01-15T00:00:00.000Z"}`))
	})
	mux.HandleFunc("/rest/greenhopper/1.0/rapid/charts/sprintreport", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("rapidViewId") != "42" || r.URL.Query().Get("sprintId") != "7" {
			http.Error(w, "unexpected sprint report query", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"contents":{"completedIssues":[{"key":"PROJ-2","summary":"Done work","typeName":"Story","statusName":"Done","estimateStatistic":{"statFieldValue":{"text":"2","value":2}},"currentEstimateStatistic":{"statFieldValue":{"text":"3","value":3}}}],"issuesNotCompletedInCurrentSprint":[{"key":"PROJ-1","summary":"Carry over","typeName":"Bug","currentStatus":"In Progress","estimateStatistic":{"statFieldValue":{"text":"1","value":1}},"currentEstimateStatistic":{"statFieldValue":{"text":"2","value":2}}}],"puntedIssues":[{"key":"PROJ-3","summary":"Removed work","typeName":"Task","statusName":"To Do"}],"issuesCompletedInAnotherSprint":[{"key":"PROJ-4","summary":"Finished later","typeName":"Task","statusName":"Done"}],"allIssuesEstimateSum":{"text":"5","value":5},"completedIssuesEstimateSum":{"text":"3","value":3},"issuesNotCompletedEstimateSum":{"text":"2","value":2},"puntedIssuesEstimateSum":{"text":"1","value":1},"issueKeysAddedDuringSprint":{"PROJ-2":true,"PROJ-3":true}}}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	agileClient, err := adagile.New(server.Client(), server.URL)
	require.NoError(t, err)

	gateway := &Gateway{agileClient: agileClient}
	report, err := gateway.GetSprintReport(context.Background(), coremodels.GetSprintReportRequest{SprintID: "7"})
	require.NoError(t, err)

	require.Equal(t, 7, report.Sprint.ID)
	require.Equal(t, 42, report.Sprint.BoardID)
	require.Equal(t, "Sprint 7", report.Sprint.Name)
	require.Equal(t, []string{"PROJ-2", "PROJ-3"}, report.AddedIssueKeys)
	require.Len(t, report.CompletedIssues, 1)
	require.True(t, report.CompletedIssues[0].AddedDuringSprint)
	require.Equal(t, "Done", report.CompletedIssues[0].Status)
	require.Len(t, report.IncompleteIssues, 1)
	require.Equal(t, "In Progress", report.IncompleteIssues[0].Status)
	require.Len(t, report.RemovedIssues, 1)
	require.True(t, report.RemovedIssues[0].AddedDuringSprint)
	require.Len(t, report.CompletedInAnotherSprint, 1)
	require.Equal(t, "3", report.CompletedIssuesEstimate.Text)
	require.Equal(t, 2.0, report.IncompleteIssuesEstimate.Value)
}
