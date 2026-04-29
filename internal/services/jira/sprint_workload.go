package jira

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/go-faster/errors"

	"github.com/kriuchkov/jiraforge/internal/core/models"
)

type sprintWorkloadAccumulator struct {
	assignee                  string
	committedKeys             map[string]struct{}
	completedKeys             map[string]struct{}
	incompleteKeys            map[string]struct{}
	addedKeys                 map[string]struct{}
	removedKeys               map[string]struct{}
	allKeys                   map[string]struct{}
	statusBreakdown           map[string]int
	committedEstimate         float64
	completedEstimate         float64
}

func newSprintWorkloadAccumulator(assignee string) *sprintWorkloadAccumulator {
	return &sprintWorkloadAccumulator{
		assignee:        assignee,
		committedKeys:   map[string]struct{}{},
		completedKeys:   map[string]struct{}{},
		incompleteKeys:  map[string]struct{}{},
		addedKeys:       map[string]struct{}{},
		removedKeys:     map[string]struct{}{},
		allKeys:         map[string]struct{}{},
		statusBreakdown: map[string]int{},
	}
}

func (s *Service) GetSprintWorkloadReport(ctx context.Context, req models.GetSprintWorkloadReportRequest) (*models.SprintWorkloadReport, error) {
	if strings.TrimSpace(req.SprintID) == "" {
		return nil, invalid("service.jira.GetSprintWorkloadReport", "sprint_id is required")
	}

	report, err := s.gateway.GetSprintReport(ctx, models.GetSprintReportRequest{SprintID: req.SprintID})
	if err != nil {
		return nil, errors.Wrap(err, "get sprint report")
	}
	if report == nil {
		return &models.SprintWorkloadReport{}, nil
	}

	enrichment, err := s.gateway.SearchIssues(ctx, models.SearchIssuesRequest{
		JQL:        fmt.Sprintf("sprint = %s", strings.TrimSpace(req.SprintID)),
		Fields:     []string{"summary", "status", "assignee", "issuetype"},
		MaxResults: 200,
	})
	if err != nil {
		return nil, errors.Wrap(err, "enrich sprint issues")
	}

	assigneeByKey := map[string]string{}
	if enrichment != nil {
		for _, issue := range enrichment.Issues {
			assigneeByKey[issue.Key] = flowEfficiencyAssigneeLabel(personLabel(issue.Assignee))
		}
	}

	addedSet := makeStringSet(report.AddedIssueKeys)
	accumulators := map[string]*sprintWorkloadAccumulator{}
	getAccumulator := func(key string) *sprintWorkloadAccumulator {
		assignee := assigneeByKey[key]
		if assignee == "" {
			assignee = "Unassigned"
		}
		acc, ok := accumulators[assignee]
		if !ok {
			acc = newSprintWorkloadAccumulator(assignee)
			accumulators[assignee] = acc
		}
		return acc
	}

	classify := func(items []models.SprintReportIssue, bucket func(*sprintWorkloadAccumulator, models.SprintReportIssue)) {
		for _, item := range items {
			acc := getAccumulator(item.Key)
			if _, ok := acc.allKeys[item.Key]; ok {
				continue
			}
			acc.allKeys[item.Key] = struct{}{}
			bucket(acc, item)
			_, isAdded := addedSet[item.Key]
			if !isAdded {
				acc.committedKeys[item.Key] = struct{}{}
				acc.committedEstimate += issueEstimate(item.Estimate)
			} else {
				acc.addedKeys[item.Key] = struct{}{}
			}
			if status := strings.TrimSpace(item.Status); status != "" {
				acc.statusBreakdown[status]++
			}
		}
	}

	classify(report.CompletedIssues, func(acc *sprintWorkloadAccumulator, item models.SprintReportIssue) {
		acc.completedKeys[item.Key] = struct{}{}
		acc.completedEstimate += issueEstimate(item.Estimate)
	})
	classify(report.IncompleteIssues, func(acc *sprintWorkloadAccumulator, item models.SprintReportIssue) {
		acc.incompleteKeys[item.Key] = struct{}{}
	})
	classify(report.CompletedInAnotherSprint, func(acc *sprintWorkloadAccumulator, item models.SprintReportIssue) {
		acc.incompleteKeys[item.Key] = struct{}{}
	})
	classify(report.RemovedIssues, func(acc *sprintWorkloadAccumulator, item models.SprintReportIssue) {
		acc.removedKeys[item.Key] = struct{}{}
	})

	result := &models.SprintWorkloadReport{Sprint: report.Sprint}
	for _, acc := range accumulators {
		summary := models.SprintWorkloadAssignee{
			Assignee:                acc.assignee,
			TotalIssues:             len(acc.allKeys),
			CommittedIssues:         len(acc.committedKeys),
			CompletedIssues:         len(acc.completedKeys),
			IncompleteIssues:        len(acc.incompleteKeys),
			AddedDuringSprintIssues: len(acc.addedKeys),
			RemovedIssues:           len(acc.removedKeys),
			CommittedEstimate:       acc.committedEstimate,
			CompletedEstimate:       acc.completedEstimate,
		}
		if summary.CommittedIssues > 0 {
			completedCommitted := 0
			for key := range acc.completedKeys {
				if _, ok := acc.committedKeys[key]; ok {
					completedCommitted++
				}
			}
			summary.CommitmentCompletionRatio = ratioInt(completedCommitted, summary.CommittedIssues)
		}
		summary.StatusBreakdown = sortedStatusBreakdown(acc.statusBreakdown)
		summary.CompletedIssueKeys = sortedKeys(acc.completedKeys)
		summary.IncompleteIssueKeys = sortedKeys(acc.incompleteKeys)
		summary.AddedDuringSprintIssueKeys = sortedKeys(acc.addedKeys)
		result.Assignees = append(result.Assignees, summary)
	}

	sort.SliceStable(result.Assignees, func(i, j int) bool {
		if result.Assignees[i].TotalIssues != result.Assignees[j].TotalIssues {
			return result.Assignees[i].TotalIssues > result.Assignees[j].TotalIssues
		}
		return result.Assignees[i].Assignee < result.Assignees[j].Assignee
	})

	return result, nil
}

func sortedKeys(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedStatusBreakdown(counts map[string]int) []models.SprintWorkloadStatusBreakdown {
	if len(counts) == 0 {
		return nil
	}
	result := make([]models.SprintWorkloadStatusBreakdown, 0, len(counts))
	for status, count := range counts {
		result = append(result, models.SprintWorkloadStatusBreakdown{Status: status, Issues: count})
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Issues != result[j].Issues {
			return result[i].Issues > result[j].Issues
		}
		return result[i].Status < result[j].Status
	})
	return result
}
