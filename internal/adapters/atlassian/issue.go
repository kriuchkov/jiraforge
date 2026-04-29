package atlassian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	jiramodels "github.com/ctreminiom/go-atlassian/pkg/infra/models"

	coreerrors "github.com/kriuchkov/jiraforge/internal/core/errors"
	coremodels "github.com/kriuchkov/jiraforge/internal/core/models"
)

func (g *Gateway) GetIssue(ctx context.Context, req coremodels.GetIssueRequest) (*coremodels.Issue, error) {
	issue, response, err := g.jiraClient.Issue.Get(ctx, req.IssueKey, req.Fields, req.Expand)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.GetIssue", "get issue", response, err)
	}
	return mapIssue(issue), nil
}

func (g *Gateway) CreateIssue(ctx context.Context, req coremodels.CreateIssueRequest) (*coremodels.IssueMutationResult, error) {
	payload := &jiramodels.IssueScheme{
		Fields: &jiramodels.IssueFieldsScheme{
			Summary:     req.Summary,
			Project:     &jiramodels.ProjectScheme{Key: req.ProjectKey},
			Description: MarkdownToADF(req.Description),
			IssueType:   &jiramodels.IssueTypeScheme{Name: req.IssueType},
		},
	}
	if strings.TrimSpace(req.AssigneeAccountID) != "" {
		payload.Fields.Assignee = &jiramodels.UserScheme{AccountID: req.AssigneeAccountID}
	}
	if strings.TrimSpace(req.Priority) != "" {
		payload.Fields.Priority = &jiramodels.PriorityScheme{Name: req.Priority}
	}

	issue, response, err := g.jiraClient.Issue.Create(ctx, payload, nil)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.CreateIssue", "create issue", response, err)
	}

	return &coremodels.IssueMutationResult{Key: issue.Key, ID: issue.ID, URL: issue.Self, Message: "issue created"}, nil
}

func (g *Gateway) CreateChildIssue(ctx context.Context, req coremodels.CreateChildIssueRequest) (*coremodels.IssueMutationResult, error) {
	parent, response, err := g.jiraClient.Issue.Get(ctx, req.ParentIssueKey, nil, nil)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.CreateChildIssue", "get parent issue", response, err)
	}

	issueType := strings.TrimSpace(req.IssueType)
	if issueType == "" {
		issueType = "Subtask"
	}

	payload := &jiramodels.IssueScheme{
		Fields: &jiramodels.IssueFieldsScheme{
			Summary:     req.Summary,
			Project:     &jiramodels.ProjectScheme{Key: parent.Fields.Project.Key},
			Description: MarkdownToADF(req.Description),
			IssueType:   &jiramodels.IssueTypeScheme{Name: issueType},
			Parent:      &jiramodels.ParentScheme{Key: req.ParentIssueKey},
		},
	}

	issue, response, err := g.jiraClient.Issue.Create(ctx, payload, nil)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.CreateChildIssue", "create child issue", response, err)
	}

	return &coremodels.IssueMutationResult{Key: issue.Key, ID: issue.ID, URL: issue.Self, ParentKey: req.ParentIssueKey, Message: "child issue created"}, nil
}

func (g *Gateway) UpdateIssue(ctx context.Context, req coremodels.UpdateIssueRequest) (*coremodels.IssueMutationResult, error) {
	payload := &jiramodels.IssueScheme{Fields: &jiramodels.IssueFieldsScheme{}}
	if strings.TrimSpace(req.Summary) != "" {
		payload.Fields.Summary = req.Summary
	}
	if strings.TrimSpace(req.Description) != "" {
		payload.Fields.Description = MarkdownToADF(req.Description)
	}

	response, err := g.jiraClient.Issue.Update(ctx, req.IssueKey, true, payload, nil, nil)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.UpdateIssue", "update issue", response, err)
	}

	return &coremodels.IssueMutationResult{Key: req.IssueKey, Message: "issue updated"}, nil
}

func (g *Gateway) DeleteIssue(ctx context.Context, req coremodels.DeleteIssueRequest) (*coremodels.IssueMutationResult, error) {
	response, err := g.jiraClient.Issue.Delete(ctx, req.IssueKey, false)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.DeleteIssue", "delete issue", response, err)
	}
	return &coremodels.IssueMutationResult{Key: req.IssueKey, Message: "issue deleted"}, nil
}

func (g *Gateway) ListIssueTypes(ctx context.Context, _ coremodels.ListIssueTypesRequest) ([]coremodels.IssueType, error) {
	issueTypes, response, err := g.jiraClient.Issue.Type.Gets(ctx)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.ListIssueTypes", "list issue types", response, err)
	}

	result := make([]coremodels.IssueType, 0, len(issueTypes))
	for _, issueType := range issueTypes {
		result = append(result, mapIssueType(issueType))
	}
	return result, nil
}

func (g *Gateway) SearchIssues(ctx context.Context, req coremodels.SearchIssuesRequest) (*coremodels.SearchIssuesResult, error) {
	params := url.Values{}
	params.Set("jql", req.JQL)
	if len(req.Fields) > 0 {
		params.Set("fields", strings.Join(req.Fields, ","))
	}
	if len(req.Expand) > 0 {
		params.Set("expand", strings.Join(req.Expand, ","))
	}
	if req.StartAt > 0 {
		params.Set("startAt", strconv.Itoa(req.StartAt))
	}
	if req.MaxResults > 0 {
		params.Set("maxResults", strconv.Itoa(req.MaxResults))
	}

	endpoint := fmt.Sprintf("%s/rest/api/3/search/jql?%s", g.jiraClient.Site.String(), params.Encode())
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, coreerrors.Wrap(coreerrors.CodeInternal, "adapters.atlassian.SearchIssues", "create search request", err)
	}
	if g.jiraClient.Auth != nil && g.jiraClient.Auth.HasBasicAuth() {
		username, password := g.jiraClient.Auth.GetBasicAuth()
		httpRequest.SetBasicAuth(username, password)
	}
	httpRequest.Header.Set("Accept", "application/json")

	response, err := g.jiraClient.HTTP.Do(httpRequest)
	if err != nil {
		return nil, coreerrors.Wrap(coreerrors.CodeUnavailable, "adapters.atlassian.SearchIssues", "execute search request", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, coreerrors.New(coreerrors.CodeInternal, "adapters.atlassian.SearchIssues", fmt.Sprintf("jira search failed with status %d", response.StatusCode))
	}

	var result jiramodels.IssueSearchScheme
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, coreerrors.Wrap(coreerrors.CodeInternal, "adapters.atlassian.SearchIssues", "decode search response", err)
	}

	issues := make([]coremodels.Issue, 0, len(result.Issues))
	for _, issue := range result.Issues {
		mapped := mapIssue(issue)
		if mapped != nil {
			issues = append(issues, *mapped)
		}
	}

	return &coremodels.SearchIssuesResult{
		Query:      req.JQL,
		StartAt:    req.StartAt,
		MaxResults: req.MaxResults,
		Total:      result.Total,
		Issues:     issues,
	}, nil
}

func (g *Gateway) AddComment(ctx context.Context, req coremodels.AddCommentRequest) (*coremodels.Comment, error) {
	payload := &jiramodels.CommentPayloadScheme{Body: MarkdownToADF(req.Comment)}
	comment, response, err := g.jiraClient.Issue.Comment.Add(ctx, req.IssueKey, payload, nil)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.AddComment", "add comment", response, err)
	}
	return mapComment(comment), nil
}

func (g *Gateway) GetComments(ctx context.Context, req coremodels.GetCommentsRequest) (*coremodels.CommentList, error) {
	comments, response, err := g.jiraClient.Issue.Comment.Gets(ctx, req.IssueKey, "", nil, 0, 50)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.GetComments", "get comments", response, err)
	}

	result := &coremodels.CommentList{IssueKey: req.IssueKey}
	for _, comment := range comments.Comments {
		mapped := mapComment(comment)
		if mapped != nil {
			result.Comments = append(result.Comments, *mapped)
		}
	}
	return result, nil
}

func (g *Gateway) AddWorklog(ctx context.Context, req coremodels.AddWorklogRequest) (*coremodels.Worklog, error) {
	seconds, err := parseTimeSpent(req.TimeSpent)
	if err != nil {
		return nil, coreerrors.Wrap(coreerrors.CodeInvalidArgument, "adapters.atlassian.AddWorklog", "parse time spent", err)
	}

	started := strings.TrimSpace(req.Started)
	if started == "" {
		started = time.Now().Format("2006-01-02T15:04:05.000-0700")
	}

	payload := &jiramodels.WorklogADFPayloadScheme{
		TimeSpentSeconds: seconds,
		Started:          started,
	}
	if strings.TrimSpace(req.Comment) != "" {
		payload.Comment = MarkdownToADF(req.Comment)
	}

	worklog, response, err := g.jiraClient.Issue.Worklog.Add(ctx, req.IssueKey, payload, &jiramodels.WorklogOptionsScheme{
		Notify:         true,
		AdjustEstimate: "auto",
	})
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.AddWorklog", "add worklog", response, err)
	}

	return &coremodels.Worklog{
		ID:               worklog.ID,
		IssueKey:         req.IssueKey,
		TimeSpent:        req.TimeSpent,
		TimeSpentSeconds: worklog.TimeSpentSeconds,
		Started:          worklog.Started,
		Author:           mapUserDetail(worklog.Author),
		Comment:          req.Comment,
	}, nil
}

func (g *Gateway) GetWorklogs(ctx context.Context, req coremodels.GetWorklogsRequest) (*coremodels.WorklogList, error) {
	req = req.Normalized()
	if strings.TrimSpace(req.IssueKey) == "" {
		return nil, coreerrors.New(coreerrors.CodeInvalidArgument, "adapters.atlassian.GetWorklogs", "issue_key is required")
	}

	var after int
	if req.StartedAfter != "" {
		parsed, err := time.Parse(time.RFC3339, req.StartedAfter)
		if err != nil {
			if parsed2, err2 := time.ParseInLocation("2006-01-02", req.StartedAfter, time.UTC); err2 == nil {
				parsed = parsed2
			} else {
				return nil, coreerrors.Wrap(coreerrors.CodeInvalidArgument, "adapters.atlassian.GetWorklogs", "parse started_after", err)
			}
		}
		after = int(parsed.UnixMilli())
	}

	page, response, err := g.jiraClient.Issue.Worklog.Issue(ctx, req.IssueKey, 0, req.MaxResults, after, nil)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.GetWorklogs", "get worklogs", response, err)
	}

	result := &coremodels.WorklogList{IssueKey: req.IssueKey, Total: page.Total}
	for _, worklog := range page.Worklogs {
		if worklog == nil {
			continue
		}
		result.Worklogs = append(result.Worklogs, coremodels.Worklog{
			ID:               worklog.ID,
			IssueKey:         req.IssueKey,
			TimeSpent:        worklog.TimeSpent,
			TimeSpentSeconds: worklog.TimeSpentSeconds,
			Started:          worklog.Started,
			Author:           mapUserDetail(worklog.Author),
		})
	}
	return result, nil
}

func (g *Gateway) GetTransitions(ctx context.Context, req coremodels.GetTransitionsRequest) ([]coremodels.Transition, error) {
	transitions, response, err := g.jiraClient.Issue.Transitions(ctx, req.IssueKey)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.GetTransitions", "get transitions", response, err)
	}

	result := make([]coremodels.Transition, 0, len(transitions.Transitions))
	for _, transition := range transitions.Transitions {
		result = append(result, coremodels.Transition{ID: transition.ID, Name: transition.Name})
	}
	return result, nil
}

func (g *Gateway) TransitionIssue(ctx context.Context, req coremodels.TransitionIssueRequest) (*coremodels.IssueMutationResult, error) {
	response, err := g.jiraClient.Issue.Move(ctx, req.IssueKey, req.TransitionID, nil)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.TransitionIssue", "transition issue", response, err)
	}
	return &coremodels.IssueMutationResult{Key: req.IssueKey, Message: "issue transitioned"}, nil
}

func (g *Gateway) ListStatuses(ctx context.Context, req coremodels.ListStatusesRequest) (*coremodels.StatusCatalog, error) {
	issueTypes, response, err := g.jiraClient.Project.Statuses(ctx, req.ProjectKey)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.ListStatuses", "list statuses", response, err)
	}

	result := &coremodels.StatusCatalog{ProjectKey: req.ProjectKey}
	for _, issueType := range issueTypes {
		group := coremodels.StatusGroup{IssueType: coremodels.IssueType{ID: issueType.ID, Name: issueType.Name, Subtask: issueType.Subtask}}
		for _, status := range issueType.Statuses {
			group.Statuses = append(group.Statuses, coremodels.Status{ID: status.ID, Name: status.Name, Description: status.Description})
		}
		result.Groups = append(result.Groups, group)
	}
	return result, nil
}

func (g *Gateway) GetIssueHistory(ctx context.Context, req coremodels.GetIssueHistoryRequest) (*coremodels.IssueHistory, error) {
	issue, response, err := g.jiraClient.Issue.Get(ctx, req.IssueKey, nil, []string{"changelog"})
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.GetIssueHistory", "get issue history", response, err)
	}

	history := &coremodels.IssueHistory{IssueKey: req.IssueKey}
	if issue.Changelog == nil {
		return history, nil
	}
	for _, entry := range issue.Changelog.Histories {
		historyEntry := coremodels.HistoryEntry{Date: formatHistoryDate(entry.Created)}
		if entry.Author != nil {
			historyEntry.Author = entry.Author.DisplayName
		}
		for _, item := range entry.Items {
			historyEntry.Changes = append(historyEntry.Changes, coremodels.HistoryChange{
				Field: item.Field,
				From:  defaultEmpty(item.FromString),
				To:    defaultEmpty(item.ToString),
			})
		}
		history.Entries = append(history.Entries, historyEntry)
	}
	return history, nil
}

func (g *Gateway) GetRelatedIssues(ctx context.Context, req coremodels.GetRelatedIssuesRequest) ([]coremodels.IssueRelation, error) {
	issue, response, err := g.jiraClient.Issue.Get(ctx, req.IssueKey, nil, []string{"issuelinks"})
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.GetRelatedIssues", "get related issues", response, err)
	}

	return mapIssueRelations(issue), nil
}

func (g *Gateway) LinkIssues(ctx context.Context, req coremodels.LinkIssuesRequest) (*coremodels.IssueMutationResult, error) {
	payload := &jiramodels.LinkPayloadSchemeV3{
		InwardIssue:  &jiramodels.LinkedIssueScheme{Key: req.InwardIssue},
		OutwardIssue: &jiramodels.LinkedIssueScheme{Key: req.OutwardIssue},
		Type:         &jiramodels.LinkTypeScheme{Name: req.LinkType},
	}
	if strings.TrimSpace(req.Comment) != "" {
		payload.Comment = &jiramodels.CommentPayloadScheme{Body: MarkdownToADF(req.Comment)}
	}

	response, err := g.jiraClient.Issue.Link.Create(ctx, payload)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.LinkIssues", "link issues", response, err)
	}

	return &coremodels.IssueMutationResult{Key: req.InwardIssue, ParentKey: req.OutwardIssue, Message: "issues linked"}, nil
}

func (g *Gateway) GetVersion(ctx context.Context, req coremodels.GetVersionRequest) (*coremodels.Version, error) {
	version, response, err := g.jiraClient.Project.Version.Get(ctx, req.VersionID, nil)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.GetVersion", "get version", response, err)
	}
	result := mapVersion(version)
	return &result, nil
}

func (g *Gateway) ListProjectVersions(ctx context.Context, req coremodels.ListProjectVersionsRequest) (*coremodels.VersionCollection, error) {
	versions, response, err := g.jiraClient.Project.Version.Gets(ctx, req.ProjectKey)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.ListProjectVersions", "list project versions", response, err)
	}

	result := &coremodels.VersionCollection{ProjectKey: req.ProjectKey}
	for _, version := range versions {
		mapped := mapVersion(version)
		result.Versions = append(result.Versions, mapped)
	}
	return result, nil
}

func (g *Gateway) DownloadAttachment(ctx context.Context, req coremodels.DownloadAttachmentRequest) (*coremodels.AttachmentContent, error) {
	metadata, response, err := g.jiraClient.Issue.Attachment.Metadata(ctx, req.AttachmentID)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.DownloadAttachment", "get attachment metadata", response, err)
	}

	downloadResponse, err := g.jiraClient.Issue.Attachment.Download(ctx, req.AttachmentID, true)
	if err != nil {
		return nil, wrapJiraError("adapters.atlassian.DownloadAttachment", "download attachment", downloadResponse, err)
	}

	return &coremodels.AttachmentContent{
		ID:       req.AttachmentID,
		Filename: metadata.Filename,
		MimeType: metadata.MimeType,
		Size:     int64(metadata.Size),
		Data:     downloadResponse.Bytes.Bytes(),
	}, nil
}
