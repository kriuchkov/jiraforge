package atlassian

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	jiramodels "github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/go-faster/errors"

	coreerrors "github.com/kriuchkov/jiraforge/internal/core/errors"
	coremodels "github.com/kriuchkov/jiraforge/internal/core/models"
)

func (g *Gateway) ListSprints(ctx context.Context, req coremodels.ListSprintsRequest) (*coremodels.SprintCollection, error) {
	boardIDs, err := g.resolveBoardIDs(ctx, req.BoardID, req.ProjectKey)
	if err != nil {
		return nil, errors.Wrap(err, "resolve board IDs")
	}

	collection := &coremodels.SprintCollection{Scope: sprintScope(req.BoardID, req.ProjectKey)}
	for _, boardID := range boardIDs {
		sprints, response, err := g.agileClient.Board.Sprints(ctx, boardID, 0, 50, []string{"active", "future"})
		if err != nil {
			return nil, wrapAgileError("adapters.atlassian.ListSprints", "list sprints", response, err)
		}
		for _, sprint := range sprints.Values {
			collection.Sprints = append(collection.Sprints, mapBoardSprint(sprint, boardID))
		}
	}
	return collection, nil
}

func (g *Gateway) GetSprint(ctx context.Context, req coremodels.GetSprintRequest) (*coremodels.Sprint, error) {
	sprintID, err := strconv.Atoi(req.SprintID)
	if err != nil {
		return nil, coreerrors.Wrap(coreerrors.CodeInvalidArgument, "adapters.atlassian.GetSprint", "parse sprint id", err)
	}

	sprint, response, err := g.agileClient.Sprint.Get(ctx, sprintID)
	if err != nil {
		return nil, wrapAgileError("adapters.atlassian.GetSprint", "get sprint", response, err)
	}

	mapped := mapSprint(sprint, sprint.OriginBoardID)
	return &mapped, nil
}

func (g *Gateway) GetSprintReport(ctx context.Context, req coremodels.GetSprintReportRequest) (*coremodels.SprintReport, error) {
	sprintID, err := strconv.Atoi(req.SprintID)
	if err != nil {
		return nil, coreerrors.Wrap(coreerrors.CodeInvalidArgument, "adapters.atlassian.GetSprintReport", "parse sprint id", err)
	}

	sprint, response, err := g.agileClient.Sprint.Get(ctx, sprintID)
	if err != nil {
		return nil, wrapAgileError("adapters.atlassian.GetSprintReport", "get sprint", response, err)
	}
	if sprint.OriginBoardID == 0 {
		return nil, coreerrors.New(coreerrors.CodeInternal, "adapters.atlassian.GetSprintReport", fmt.Sprintf("sprint %d does not expose origin board id", sprintID))
	}

	path := fmt.Sprintf("rest/greenhopper/1.0/rapid/charts/sprintreport?rapidViewId=%d&sprintId=%d", sprint.OriginBoardID, sprintID)
	httpRequest, err := g.agileClient.NewRequest(ctx, http.MethodGet, path, "", nil)
	if err != nil {
		return nil, coreerrors.Wrap(coreerrors.CodeInternal, "adapters.atlassian.GetSprintReport", "build sprint report request", err)
	}

	payload := new(sprintReportResponse)
	response, err = g.agileClient.Call(httpRequest, payload)
	if err != nil {
		return nil, wrapAgileError("adapters.atlassian.GetSprintReport", "get sprint report", response, err)
	}

	report := mapSprintReport(sprint, payload)
	return &report, nil
}

func (g *Gateway) GetActiveSprint(ctx context.Context, req coremodels.GetActiveSprintRequest) (*coremodels.Sprint, error) {
	boardIDs, err := g.resolveBoardIDs(ctx, req.BoardID, req.ProjectKey)
	if err != nil {
		return nil, errors.Wrap(err, "resolve board IDs")
	}

	for _, boardID := range boardIDs {
		sprints, response, err := g.agileClient.Board.Sprints(ctx, boardID, 0, 50, []string{"active"})
		if err != nil {
			return nil, wrapAgileError("adapters.atlassian.GetActiveSprint", "get active sprint", response, err)
		}
		if len(sprints.Values) > 0 {
			sprint := mapBoardSprint(sprints.Values[0], boardID)
			return &sprint, nil
		}
	}

	return nil, coreerrors.New(coreerrors.CodeNotFound, "adapters.atlassian.GetActiveSprint", "no active sprint found")
}

func (g *Gateway) SearchSprints(ctx context.Context, req coremodels.SearchSprintsRequest) (*coremodels.SprintCollection, error) {
	boardIDs, err := g.resolveBoardIDs(ctx, req.BoardID, req.ProjectKey)
	if err != nil {
		return nil, errors.Wrap(err, "resolve board IDs")
	}

	needle := strings.ToLower(strings.TrimSpace(req.Name))
	collection := &coremodels.SprintCollection{Scope: sprintScope(req.BoardID, req.ProjectKey)}
	for _, boardID := range boardIDs {
		sprints, response, err := g.agileClient.Board.Sprints(ctx, boardID, 0, 100, []string{"active", "future", "closed"})
		if err != nil {
			return nil, wrapAgileError("adapters.atlassian.SearchSprints", "search sprints", response, err)
		}
		for _, sprint := range sprints.Values {
			name := strings.ToLower(sprint.Name)
			matched := name == needle
			if !req.ExactMatch {
				matched = strings.Contains(name, needle)
			}
			if matched {
				collection.Sprints = append(collection.Sprints, mapBoardSprint(sprint, boardID))
			}
		}
	}
	return collection, nil
}

type sprintReportResponse struct {
	Contents sprintReportContents `json:"contents"`
}

type sprintReportContents struct {
	CompletedIssues                   []sprintReportIssue `json:"completedIssues"`
	IssuesNotCompletedInCurrentSprint []sprintReportIssue `json:"issuesNotCompletedInCurrentSprint"`
	PuntedIssues                      []sprintReportIssue `json:"puntedIssues"`
	IssuesCompletedInAnotherSprint    []sprintReportIssue `json:"issuesCompletedInAnotherSprint"`
	AllIssuesEstimateSum              sprintReportStat    `json:"allIssuesEstimateSum"`
	CompletedIssuesEstimateSum        sprintReportStat    `json:"completedIssuesEstimateSum"`
	IssuesNotCompletedEstimateSum     sprintReportStat    `json:"issuesNotCompletedEstimateSum"`
	PuntedIssuesEstimateSum           sprintReportStat    `json:"puntedIssuesEstimateSum"`
	IssueKeysAddedDuringSprint        map[string]bool     `json:"issueKeysAddedDuringSprint"`
}

type sprintReportIssue struct {
	Key                      string           `json:"key"`
	Summary                  string           `json:"summary"`
	TypeName                 string           `json:"typeName"`
	StatusName               string           `json:"statusName"`
	CurrentStatus            string           `json:"currentStatus"`
	EstimateStatistic        sprintReportStat `json:"estimateStatistic"`
	CurrentEstimateStatistic sprintReportStat `json:"currentEstimateStatistic"`
}

type sprintReportStat struct {
	StatFieldValue sprintReportEstimate `json:"statFieldValue"`
	Text           string               `json:"text"`
	Value          float64              `json:"value"`
}

type sprintReportEstimate struct {
	Text  string  `json:"text"`
	Value float64 `json:"value"`
}

func mapSprintReport(sprint *jiramodels.SprintScheme, payload *sprintReportResponse) coremodels.SprintReport {
	addedIssueKeys := collectAddedIssueKeys(payload.Contents.IssueKeysAddedDuringSprint)
	addedSet := make(map[string]struct{}, len(addedIssueKeys))
	for _, key := range addedIssueKeys {
		addedSet[key] = struct{}{}
	}

	return coremodels.SprintReport{
		Sprint:                   mapSprint(sprint, sprint.OriginBoardID),
		CompletedIssues:          mapSprintReportIssues(payload.Contents.CompletedIssues, addedSet),
		IncompleteIssues:         mapSprintReportIssues(payload.Contents.IssuesNotCompletedInCurrentSprint, addedSet),
		RemovedIssues:            mapSprintReportIssues(payload.Contents.PuntedIssues, addedSet),
		CompletedInAnotherSprint: mapSprintReportIssues(payload.Contents.IssuesCompletedInAnotherSprint, addedSet),
		AddedIssueKeys:           addedIssueKeys,
		AllIssuesEstimate:        mapSprintReportStat(payload.Contents.AllIssuesEstimateSum),
		CompletedIssuesEstimate:  mapSprintReportStat(payload.Contents.CompletedIssuesEstimateSum),
		IncompleteIssuesEstimate: mapSprintReportStat(payload.Contents.IssuesNotCompletedEstimateSum),
		RemovedIssuesEstimate:    mapSprintReportStat(payload.Contents.PuntedIssuesEstimateSum),
	}
}

func mapSprintReportIssues(items []sprintReportIssue, addedSet map[string]struct{}) []coremodels.SprintReportIssue {
	if len(items) == 0 {
		return nil
	}

	result := make([]coremodels.SprintReportIssue, 0, len(items))
	for _, item := range items {
		status := item.StatusName
		if strings.TrimSpace(status) == "" {
			status = item.CurrentStatus
		}
		_, addedDuringSprint := addedSet[item.Key]
		result = append(result, coremodels.SprintReportIssue{
			Key:               item.Key,
			Summary:           item.Summary,
			Type:              item.TypeName,
			Status:            status,
			Estimate:          mapSprintReportEstimate(item.EstimateStatistic.StatFieldValue),
			CurrentEstimate:   mapSprintReportEstimate(item.CurrentEstimateStatistic.StatFieldValue),
			AddedDuringSprint: addedDuringSprint,
		})
	}
	return result
}

func mapSprintReportEstimate(value sprintReportEstimate) coremodels.SprintReportEstimate {
	return coremodels.SprintReportEstimate{Text: value.Text, Value: value.Value}
}

func mapSprintReportStat(value sprintReportStat) coremodels.SprintReportEstimate {
	if value.StatFieldValue.Text != "" || value.StatFieldValue.Value != 0 {
		return mapSprintReportEstimate(value.StatFieldValue)
	}
	return coremodels.SprintReportEstimate{Text: value.Text, Value: value.Value}
}

func collectAddedIssueKeys(values map[string]bool) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	for key, added := range values {
		if added {
			result = append(result, key)
		}
	}
	sort.Strings(result)
	return result
}
