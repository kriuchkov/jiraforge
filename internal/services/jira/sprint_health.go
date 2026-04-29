package jira

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/go-faster/errors"

	"github.com/kriuchkov/jiraforge/internal/core/models"
)

const defaultSprintHealthTopIssues = 5

func (s *Service) GetSprintHealthReport(ctx context.Context, req models.GetSprintHealthReportRequest) (*models.SprintHealthReport, error) {
	if strings.TrimSpace(req.SprintID) == "" {
		return nil, invalid("service.jira.GetSprintHealthReport", "sprint_id is required")
	}

	report, err := s.gateway.GetSprintReport(ctx, models.GetSprintReportRequest{SprintID: req.SprintID})
	if err != nil {
		return nil, errors.Wrap(err, "get sprint report")
	}

	return buildSprintHealthReport(report), nil
}

func buildSprintHealthReport(report *models.SprintReport) *models.SprintHealthReport {
	if report == nil {
		return nil
	}

	addedSet := makeStringSet(report.AddedIssueKeys)
	committedIssues := uniqueIssueCountAcross(addedSet, false, report.CompletedIssues, report.IncompleteIssues, report.RemovedIssues, report.CompletedInAnotherSprint)
	completedCommittedIssues := uniqueIssueCount(report.CompletedIssues, addedSet, false)
	completedIssues := uniqueIssueCount(report.CompletedIssues, nil, false)
	incompleteIssues := uniqueIssueCount(report.IncompleteIssues, nil, false)
	completedElsewhereIssues := uniqueIssueCount(report.CompletedInAnotherSprint, nil, false)
	carryOverIssues := uniqueIssueCount(report.IncompleteIssues, addedSet, false) + uniqueIssueCount(report.CompletedInAnotherSprint, addedSet, false)
	removedIssues := uniqueIssueCount(report.RemovedIssues, nil, false)
	addedDuringSprint := len(addedSet)
	unplannedCompletedIssues := uniqueIssueCount(report.CompletedIssues, addedSet, true)
	committedEstimate := uniqueEstimateSumAcross(addedSet, false, report.CompletedIssues, report.IncompleteIssues, report.RemovedIssues, report.CompletedInAnotherSprint)
	completedCommittedEstimate := uniqueEstimateSum(report.CompletedIssues, addedSet, false)

	summary := models.SprintHealthSummary{
		CommittedIssues:            committedIssues,
		CompletedCommittedIssues:   completedCommittedIssues,
		CompletedIssues:            completedIssues,
		IncompleteIssues:           incompleteIssues,
		CompletedElsewhereIssues:   completedElsewhereIssues,
		CarryOverIssues:            carryOverIssues,
		RemovedIssues:              removedIssues,
		AddedDuringSprint:          addedDuringSprint,
		UnplannedCompletedIssues:   unplannedCompletedIssues,
		CompletionRatio:            ratioInt(completedCommittedIssues, committedIssues),
		CommittedEstimate:          committedEstimate,
		CompletedCommittedEstimate: completedCommittedEstimate,
		EstimateCompletionRatio:    ratioFloat(completedCommittedEstimate, committedEstimate),
	}

	return &models.SprintHealthReport{
		Sprint:              report.Sprint,
		Summary:             summary,
		Risks:               detectSprintHealthRisks(summary),
		TopIncompleteIssues: topSprintHealthIssues(report.IncompleteIssues, addedSet, defaultSprintHealthTopIssues),
	}
}

func detectSprintHealthRisks(summary models.SprintHealthSummary) []string {
	var risks []string

	if summary.CommittedIssues > 0 {
		if summary.CompletionRatio < 0.7 {
			risks = append(risks, "low_commitment_completion")
		}
		if float64(summary.AddedDuringSprint)/float64(summary.CommittedIssues) >= 0.3 {
			risks = append(risks, "high_scope_creep")
		}
		if float64(summary.RemovedIssues)/float64(summary.CommittedIssues) >= 0.2 {
			risks = append(risks, "high_scope_removal")
		}
		if float64(summary.CarryOverIssues)/float64(summary.CommittedIssues) >= 0.2 {
			risks = append(risks, "carry_over_risk")
		}
	}

	if summary.CompletedIssues > 0 && float64(summary.UnplannedCompletedIssues)/float64(summary.CompletedIssues) >= 0.3 {
		risks = append(risks, "high_unplanned_delivery")
	}

	if summary.CommittedEstimate > 0 && summary.EstimateCompletionRatio < 0.7 {
		risks = append(risks, "low_estimate_completion")
	}

	return risks
}

func topSprintHealthIssues(items []models.SprintReportIssue, addedSet map[string]struct{}, limit int) []models.SprintReportIssue {
	if len(items) == 0 || limit <= 0 {
		return nil
	}

	sorted := make([]models.SprintReportIssue, len(items))
	copy(sorted, items)
	sort.SliceStable(sorted, func(i, j int) bool {
		_, leftAdded := addedSet[sorted[i].Key]
		_, rightAdded := addedSet[sorted[j].Key]
		if leftAdded != rightAdded {
			return !leftAdded
		}

		leftCurrent := issueEstimate(sorted[i].CurrentEstimate)
		rightCurrent := issueEstimate(sorted[j].CurrentEstimate)
		if leftCurrent != rightCurrent {
			return leftCurrent > rightCurrent
		}

		leftBase := issueEstimate(sorted[i].Estimate)
		rightBase := issueEstimate(sorted[j].Estimate)
		if leftBase != rightBase {
			return leftBase > rightBase
		}

		return sorted[i].Key < sorted[j].Key
	})

	if len(sorted) > limit {
		sorted = sorted[:limit]
	}
	return sorted
}

func uniqueIssueCountAcross(filter map[string]struct{}, wantAdded bool, groups ...[]models.SprintReportIssue) int {
	seen := make(map[string]struct{})
	for _, group := range groups {
		for _, issue := range group {
			if !includeIssue(issue.Key, filter, wantAdded) {
				continue
			}
			seen[issue.Key] = struct{}{}
		}
	}
	return len(seen)
}

func uniqueIssueCount(items []models.SprintReportIssue, filter map[string]struct{}, wantAdded bool) int {
	return uniqueIssueCountAcross(filter, wantAdded, items)
}

func uniqueEstimateSumAcross(filter map[string]struct{}, wantAdded bool, groups ...[]models.SprintReportIssue) float64 {
	seen := make(map[string]struct{})
	var total float64
	for _, group := range groups {
		for _, issue := range group {
			if !includeIssue(issue.Key, filter, wantAdded) {
				continue
			}
			if _, ok := seen[issue.Key]; ok {
				continue
			}
			seen[issue.Key] = struct{}{}
			total += issueEstimate(issue.Estimate)
			if issueEstimate(issue.Estimate) == 0 {
				total += issueEstimate(issue.CurrentEstimate)
			}
		}
	}
	return total
}

func uniqueEstimateSum(items []models.SprintReportIssue, filter map[string]struct{}, wantAdded bool) float64 {
	return uniqueEstimateSumAcross(filter, wantAdded, items)
}

func issueEstimate(estimate models.SprintReportEstimate) float64 {
	if estimate.Value != 0 {
		return estimate.Value
	}
	if parsed, err := strconv.ParseFloat(strings.TrimSpace(estimate.Text), 64); err == nil {
		return parsed
	}
	return 0
}

func ratioInt(part, whole int) float64 {
	if whole <= 0 {
		return 0
	}
	return float64(part) / float64(whole)
}

func ratioFloat(part, whole float64) float64 {
	if whole <= 0 {
		return 0
	}
	return part / whole
}

func makeStringSet(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			set[trimmed] = struct{}{}
		}
	}
	return set
}

func includeIssue(key string, filter map[string]struct{}, wantAdded bool) bool {
	if len(filter) == 0 {
		return !wantAdded
	}
	_, isAdded := filter[key]
	return isAdded == wantAdded
}
