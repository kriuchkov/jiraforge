package atlassian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	coreerrors "github.com/kriuchkov/jiraforge/internal/core/errors"
	coremodels "github.com/kriuchkov/jiraforge/internal/core/models"
)

func (g *Gateway) GetDevelopmentInfo(ctx context.Context, req coremodels.GetDevelopmentInfoRequest) (*coremodels.DevelopmentInfo, error) {
	issue, response, err := g.jiraClient.Issue.Get(ctx, req.IssueKey, nil, []string{"id"})
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.GetDevelopmentInfo", "get issue id", response, err)
	}

	summaryRequest, err := g.jiraClient.NewRequest(ctx, http.MethodGet, fmt.Sprintf("/rest/dev-status/latest/issue/summary?issueId=%s", issue.ID), "", nil)
	if err != nil {
		return nil, coreerrors.Wrap(coreerrors.CodeInternal, "adapters.atlassian.GetDevelopmentInfo", "create summary request", err)
	}

	var summary devSummaryResponse
	summaryResponse, err := g.jiraClient.Call(summaryRequest, &summary)
	if err != nil {
		if summaryResponse != nil && summaryResponse.Code == http.StatusNotFound {
			return nil, coreerrors.New(coreerrors.CodeUnavailable, "adapters.atlassian.GetDevelopmentInfo", "jira dev-status endpoint is not available")
		}
		return nil, wrapJiraError("adapters.atlassian.GetDevelopmentInfo", "load development summary", summaryResponse, err)
	}

	endpoints := collectDevEndpoints(summary)
	result := &coremodels.DevelopmentInfo{IssueKey: req.IssueKey, IssueID: issue.ID}
	if len(endpoints) == 0 {
		return result, nil
	}

	for _, endpoint := range endpoints {
		detailRequest, err := g.jiraClient.NewRequest(ctx, http.MethodGet,
			fmt.Sprintf("/rest/dev-status/latest/issue/detail?issueId=%s&applicationType=%s&dataType=%s", issue.ID, endpoint.AppType, endpoint.DataType),
			"", nil,
		)
		if err != nil {
			return nil, coreerrors.Wrap(coreerrors.CodeInternal, "adapters.atlassian.GetDevelopmentInfo", "create detail request", err)
		}

		var detailResponse devStatusResponse
		callResponse, err := g.jiraClient.Call(detailRequest, &detailResponse)
		if err != nil {
			if callResponse != nil && callResponse.Code == http.StatusNotFound {
				continue
			}
			return nil, wrapJiraError("adapters.atlassian.GetDevelopmentInfo", "load development detail", callResponse, err)
		}
		for _, detail := range detailResponse.Detail {
			for _, branch := range detail.Branches {
				result.Branches = append(result.Branches, mapDevelopmentBranch(branch))
			}
			for _, pullRequest := range detail.PullRequests {
				result.PullRequests = append(result.PullRequests, mapDevelopmentPullRequest(pullRequest))
			}
			for _, repository := range detail.Repositories {
				result.Repositories = append(result.Repositories, mapDevelopmentRepository(repository))
			}
			for _, build := range detail.Builds {
				result.Builds = append(result.Builds, mapDevelopmentBuild(build))
			}
			for _, buildsData := range detail.JSWDDBuildsData {
				for _, build := range buildsData.Builds {
					result.Builds = append(result.Builds, mapDevelopmentBuild(build))
				}
			}
		}
	}

	if !req.IncludeBranches {
		result.Branches = nil
	}
	if !req.IncludePullRequests {
		result.PullRequests = nil
	}
	if !req.IncludeCommits {
		for index := range result.Repositories {
			result.Repositories[index].Commits = nil
		}
	}
	if !req.IncludeBuilds {
		result.Builds = nil
	}

	return result, nil
}

type devEndpoint struct {
	AppType  string
	DataType string
}

type devSummaryResponse struct {
	Summary struct {
		Repository  devSummarySection `json:"repository"`
		Branch      devSummarySection `json:"branch"`
		PullRequest devSummarySection `json:"pullrequest"`
		Build       devSummarySection `json:"build"`
	} `json:"summary"`
}

type devSummarySection struct {
	ByInstanceType map[string]json.RawMessage `json:"byInstanceType"`
}

func collectDevEndpoints(summary devSummaryResponse) []devEndpoint {
	sections := []struct {
		name    string
		section devSummarySection
	}{
		{name: "repository", section: summary.Summary.Repository},
		{name: "branch", section: summary.Summary.Branch},
		{name: "pullrequest", section: summary.Summary.PullRequest},
		{name: "build", section: summary.Summary.Build},
	}

	seen := map[string]struct{}{}
	var result []devEndpoint
	for _, section := range sections {
		for appType := range section.section.ByInstanceType {
			key := appType + ":" + section.name
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, devEndpoint{AppType: appType, DataType: section.name})
		}
	}
	return result
}

type devStatusResponse struct {
	Errors []string          `json:"errors"`
	Detail []devStatusDetail `json:"detail"`
}

type devStatusDetail struct {
	Branches        []devBranch          `json:"branches,omitempty"`
	PullRequests    []devPullRequest     `json:"pullRequests,omitempty"`
	Repositories    []devRepository      `json:"repositories,omitempty"`
	Builds          []devBuild           `json:"builds,omitempty"`
	JSWDDBuildsData []devJSWDDBuildsData `json:"jswddBuildsData,omitempty"`
}

type devJSWDDBuildsData struct {
	Builds []devBuild `json:"builds,omitempty"`
}

type devBranch struct {
	Name                 string           `json:"name"`
	URL                  string           `json:"url"`
	CreatePullRequestURL string           `json:"createPullRequestUrl,omitempty"`
	Repository           devRepositoryRef `json:"repository"`
	LastCommit           devCommit        `json:"lastCommit"`
}

type devRepositoryRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type devPullRequest struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	URL            string        `json:"url"`
	Status         string        `json:"status"`
	Author         devAuthor     `json:"author"`
	LastUpdate     string        `json:"lastUpdate"`
	Source         devBranchRef  `json:"source"`
	Destination    devBranchRef  `json:"destination"`
	CommentCount   int           `json:"commentCount,omitempty"`
	Reviewers      []devReviewer `json:"reviewers,omitempty"`
	RepositoryID   string        `json:"repositoryId,omitempty"`
	RepositoryName string        `json:"repositoryName,omitempty"`
	RepositoryURL  string        `json:"repositoryUrl,omitempty"`
}

type devReviewer struct {
	Name     string `json:"name"`
	Approved bool   `json:"approved"`
}

type devBranchRef struct {
	Branch string `json:"branch"`
	URL    string `json:"url,omitempty"`
}

type devRepository struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	URL     string      `json:"url"`
	Avatar  string      `json:"avatar,omitempty"`
	Commits []devCommit `json:"commits,omitempty"`
}

type devCommit struct {
	ID              string          `json:"id"`
	DisplayID       string          `json:"displayId"`
	Message         string          `json:"message"`
	Author          devAuthor       `json:"author"`
	AuthorTimestamp string          `json:"authorTimestamp"`
	URL             string          `json:"url,omitempty"`
	FileCount       int             `json:"fileCount,omitempty"`
	Merge           bool            `json:"merge,omitempty"`
	Files           []devCommitFile `json:"files,omitempty"`
}

type devCommitFile struct {
	Path         string `json:"path"`
	URL          string `json:"url"`
	ChangeType   string `json:"changeType"`
	LinesAdded   int    `json:"linesAdded"`
	LinesRemoved int    `json:"linesRemoved"`
}

type devAuthor struct {
	Name   string `json:"name"`
	Email  string `json:"email,omitempty"`
	Avatar string `json:"avatar,omitempty"`
}

type devBuild struct {
	ID             string            `json:"id"`
	Name           string            `json:"name,omitempty"`
	DisplayName    string            `json:"displayName,omitempty"`
	Description    string            `json:"description,omitempty"`
	URL            string            `json:"url"`
	State          string            `json:"state"`
	CreatedAt      string            `json:"createdAt,omitempty"`
	LastUpdated    string            `json:"lastUpdated"`
	BuildNumber    any               `json:"buildNumber,omitempty"`
	TestInfo       *devBuildTestInfo `json:"testInfo,omitempty"`
	TestSummary    *devBuildTestInfo `json:"testSummary,omitempty"`
	References     []devBuildRef     `json:"references,omitempty"`
	PipelineID     string            `json:"pipelineId,omitempty"`
	PipelineName   string            `json:"pipelineName,omitempty"`
	ProviderID     string            `json:"providerId,omitempty"`
	ProviderType   string            `json:"providerType,omitempty"`
	RepositoryID   string            `json:"repositoryId,omitempty"`
	RepositoryName string            `json:"repositoryName,omitempty"`
	RepositoryURL  string            `json:"repositoryUrl,omitempty"`
}

type devBuildTestInfo struct {
	TotalNumber   int `json:"totalNumber"`
	NumberPassed  int `json:"numberPassed,omitempty"`
	SuccessNumber int `json:"successNumber,omitempty"`
	NumberFailed  int `json:"numberFailed,omitempty"`
	FailedNumber  int `json:"failedNumber,omitempty"`
	SkippedNumber int `json:"skippedNumber,omitempty"`
}

type devBuildRef struct {
	Commit devBuildCommitRef `json:"commit,omitempty"`
	Ref    devBuildRefInfo   `json:"ref,omitempty"`
}

type devBuildCommitRef struct {
	ID            string `json:"id"`
	DisplayID     string `json:"displayId"`
	RepositoryURI string `json:"repositoryUri,omitempty"`
}

type devBuildRefInfo struct {
	Name string `json:"name"`
	URI  string `json:"uri,omitempty"`
}

func mapDevelopmentBranch(value devBranch) coremodels.DevelopmentBranch {
	return coremodels.DevelopmentBranch{
		Name:                 value.Name,
		URL:                  value.URL,
		CreatePullRequestURL: value.CreatePullRequestURL,
		Repository:           coremodels.DevelopmentRepositoryRef{ID: value.Repository.ID, Name: value.Repository.Name, URL: value.Repository.URL},
		LastCommit:           mapDevelopmentCommit(value.LastCommit),
	}
}

func mapDevelopmentPullRequest(value devPullRequest) coremodels.DevelopmentPullRequest {
	result := coremodels.DevelopmentPullRequest{
		ID:             value.ID,
		Title:          value.Name,
		URL:            value.URL,
		Status:         value.Status,
		Author:         mapDevelopmentAuthor(value.Author),
		LastUpdate:     value.LastUpdate,
		Source:         coremodels.DevelopmentBranchRef{Branch: value.Source.Branch, URL: value.Source.URL},
		Destination:    coremodels.DevelopmentBranchRef{Branch: value.Destination.Branch, URL: value.Destination.URL},
		CommentCount:   value.CommentCount,
		RepositoryID:   value.RepositoryID,
		RepositoryName: value.RepositoryName,
		RepositoryURL:  value.RepositoryURL,
	}
	for _, reviewer := range value.Reviewers {
		result.Reviewers = append(result.Reviewers, coremodels.DevelopmentReviewer{Name: reviewer.Name, Approved: reviewer.Approved})
	}
	return result
}

func mapDevelopmentRepository(value devRepository) coremodels.DevelopmentRepository {
	result := coremodels.DevelopmentRepository{ID: value.ID, Name: value.Name, URL: value.URL, Avatar: value.Avatar}
	for _, commit := range value.Commits {
		result.Commits = append(result.Commits, mapDevelopmentCommit(commit))
	}
	return result
}

func mapDevelopmentCommit(value devCommit) coremodels.DevelopmentCommit {
	result := coremodels.DevelopmentCommit{
		ID:              value.ID,
		DisplayID:       value.DisplayID,
		Message:         value.Message,
		Author:          mapDevelopmentAuthor(value.Author),
		AuthorTimestamp: value.AuthorTimestamp,
		URL:             value.URL,
		FileCount:       value.FileCount,
		Merge:           value.Merge,
	}
	for _, file := range value.Files {
		result.Files = append(result.Files, coremodels.DevelopmentCommitFile{
			Path:         file.Path,
			URL:          file.URL,
			ChangeType:   file.ChangeType,
			LinesAdded:   file.LinesAdded,
			LinesRemoved: file.LinesRemoved,
		})
	}
	return result
}

func mapDevelopmentAuthor(value devAuthor) *coremodels.Person {
	if strings.TrimSpace(value.Name) == "" && strings.TrimSpace(value.Email) == "" {
		return nil
	}
	return &coremodels.Person{DisplayName: value.Name, Email: value.Email}
}

func mapDevelopmentBuild(value devBuild) coremodels.DevelopmentBuild {
	build := coremodels.DevelopmentBuild{
		ID:             value.ID,
		Name:           value.Name,
		DisplayName:    value.DisplayName,
		Description:    value.Description,
		URL:            value.URL,
		State:          value.State,
		CreatedAt:      value.CreatedAt,
		LastUpdated:    value.LastUpdated,
		BuildNumber:    value.BuildNumber,
		PipelineID:     value.PipelineID,
		PipelineName:   value.PipelineName,
		ProviderID:     value.ProviderID,
		ProviderType:   value.ProviderType,
		RepositoryID:   value.RepositoryID,
		RepositoryName: value.RepositoryName,
		RepositoryURL:  value.RepositoryURL,
	}
	if value.TestInfo != nil {
		build.Tests = &coremodels.BuildTestSummary{
			TotalNumber:   value.TestInfo.TotalNumber,
			NumberPassed:  value.TestInfo.NumberPassed,
			SuccessNumber: value.TestInfo.SuccessNumber,
			NumberFailed:  value.TestInfo.NumberFailed,
			FailedNumber:  value.TestInfo.FailedNumber,
			SkippedNumber: value.TestInfo.SkippedNumber,
		}
	} else if value.TestSummary != nil {
		build.Tests = &coremodels.BuildTestSummary{
			TotalNumber:   value.TestSummary.TotalNumber,
			NumberPassed:  value.TestSummary.NumberPassed,
			SuccessNumber: value.TestSummary.SuccessNumber,
			NumberFailed:  value.TestSummary.NumberFailed,
			FailedNumber:  value.TestSummary.FailedNumber,
			SkippedNumber: value.TestSummary.SkippedNumber,
		}
	}
	for _, ref := range value.References {
		build.References = append(build.References, coremodels.BuildReference{
			Commit: coremodels.BuildCommitRef{ID: ref.Commit.ID, DisplayID: ref.Commit.DisplayID, RepositoryURI: ref.Commit.RepositoryURI},
			Ref:    coremodels.BuildRefInfo{Name: ref.Ref.Name, URI: ref.Ref.URI},
		})
	}
	return build
}
