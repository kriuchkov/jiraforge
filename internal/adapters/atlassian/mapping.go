package atlassian

import (
	"strings"

	jiramodels "github.com/ctreminiom/go-atlassian/pkg/infra/models"

	coremodels "github.com/kriuchkov/jiraforge/internal/core/models"
)

func mapIssue(issue *jiramodels.IssueScheme) *coremodels.Issue {
	if issue == nil {
		return nil
	}

	result := &coremodels.Issue{Key: issue.Key, ID: issue.ID, URL: issue.Self}
	if issue.Fields == nil {
		return result
	}

	fields := issue.Fields
	result.Summary = fields.Summary
	result.Description = RenderADF(fields.Description)
	result.Type = mapIssueTypePtr(fields.IssueType)
	result.Status = mapStatus(fields.Status)
	if fields.Priority != nil {
		result.Priority = fields.Priority.Name
	}
	if fields.Resolution != nil {
		result.Resolution = fields.Resolution.Name
		result.ResolutionDescription = fields.Resolution.Description
	}
	result.ResolutionDate = fields.Resolutiondate
	result.Reporter = mapPerson(fields.Reporter)
	result.Assignee = mapPerson(fields.Assignee)
	result.Creator = mapPerson(fields.Creator)
	result.Created = fields.Created
	result.Updated = fields.Updated
	result.LastViewed = fields.LastViewed
	result.StatusCategoryChange = fields.StatusCategoryChangeDate
	if fields.Project != nil {
		result.ProjectKey = fields.Project.Key
		result.ProjectName = fields.Project.Name
	}
	if fields.Parent != nil {
		result.Parent = mapParentRef(fields.Parent)
	}
	result.Labels = slicesOrNil(fields.Labels)
	result.Components = componentNames(fields.Components)
	result.FixVersions = versionNames(fields.FixVersions)
	result.AffectedVersions = versionNames(fields.Versions)
	if fields.Security != nil {
		result.SecurityLevel = fields.Security.Name
	}
	for _, subtask := range fields.Subtasks {
		if ref := mapIssueRef(subtask.Key, subtask.Fields); ref != nil {
			result.Subtasks = append(result.Subtasks, *ref)
		}
	}
	result.RelatedIssues = mapIssueLinks(fields.IssueLinks)
	for _, attachment := range fields.Attachment {
		result.Attachments = append(result.Attachments, coremodels.AttachmentMeta{
			ID:       attachment.ID,
			Filename: attachment.Title,
			URL:      attachment.DownloadLink,
			MimeType: attachment.MediaType,
			Size:     int64(attachment.FileSize),
		})
	}
	if fields.Watcher != nil {
		result.Watchers = fields.Watcher.WatchCount
	}
	if fields.Votes != nil {
		result.Votes = fields.Votes.Votes
	}
	if fields.Comment != nil {
		result.CommentCount = fields.Comment.Total
	}
	if fields.Worklog != nil {
		result.WorklogCount = fields.Worklog.Total
	}
	for _, transition := range issue.Transitions {
		result.Transitions = append(result.Transitions, coremodels.Transition{ID: transition.ID, Name: transition.Name})
	}
	if issue.Changelog != nil {
		for _, history := range issue.Changelog.Histories {
			for _, item := range history.Items {
				if item.Field == "Story point estimate" && strings.TrimSpace(item.ToString) != "" {
					result.StoryPointEstimate = item.ToString
				}
			}
		}
	}

	return result
}

func mapIssueType(value *jiramodels.IssueTypeScheme) coremodels.IssueType {
	result := coremodels.IssueType{}
	if value == nil {
		return result
	}
	result.ID = value.ID
	result.Name = value.Name
	result.Description = value.Description
	result.IconURL = value.IconURL
	result.Subtask = value.Subtask
	if value.Scope != nil {
		result.Scope = value.Scope.Type
	}
	return result
}

func mapIssueTypePtr(value *jiramodels.IssueTypeScheme) *coremodels.IssueType {
	if value == nil {
		return nil
	}
	mapped := mapIssueType(value)
	return &mapped
}

func mapStatus(value *jiramodels.StatusScheme) *coremodels.Status {
	if value == nil {
		return nil
	}
	return &coremodels.Status{ID: value.ID, Name: value.Name, Description: value.Description}
}

func mapPerson(value *jiramodels.UserScheme) *coremodels.Person {
	if value == nil {
		return nil
	}
	return &coremodels.Person{DisplayName: value.DisplayName, Email: value.EmailAddress}
}

func mapUserDetail(value *jiramodels.UserDetailScheme) *coremodels.Person {
	if value == nil {
		return nil
	}
	return &coremodels.Person{DisplayName: value.DisplayName, Email: value.EmailAddress}
}

func mapIssueRef(key string, fields *jiramodels.IssueFieldsScheme) *coremodels.IssueRef {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	ref := &coremodels.IssueRef{Key: key}
	if fields != nil {
		ref.Summary = fields.Summary
		ref.Status = mapStatus(fields.Status)
	}
	return ref
}

func mapParentRef(parent *jiramodels.ParentScheme) *coremodels.IssueRef {
	if parent == nil {
		return nil
	}
	ref := &coremodels.IssueRef{Key: parent.Key}
	if parent.Fields != nil {
		ref.Summary = parent.Fields.Summary
		ref.Status = mapStatus(parent.Fields.Status)
	}
	return ref
}

func mapLinkedIssueRef(issue *jiramodels.LinkedIssueScheme) *coremodels.IssueRef {
	if issue == nil {
		return nil
	}
	ref := &coremodels.IssueRef{Key: issue.Key}
	if issue.Fields != nil {
		ref.Summary = issue.Fields.Summary
		ref.Status = mapStatus(issue.Fields.Status)
	}
	return ref
}

func mapIssueLinks(links []*jiramodels.IssueLinkScheme) []coremodels.IssueRelation {
	result := make([]coremodels.IssueRelation, 0, len(links))
	for _, link := range links {
		if link == nil || link.Type == nil {
			continue
		}
		if link.InwardIssue != nil {
			ref := mapLinkedIssueRef(link.InwardIssue)
			if ref != nil {
				result = append(result, coremodels.IssueRelation{Type: link.Type.Inward, Direction: "inward", Issue: *ref})
			}
		}
		if link.OutwardIssue != nil {
			ref := mapLinkedIssueRef(link.OutwardIssue)
			if ref != nil {
				result = append(result, coremodels.IssueRelation{Type: link.Type.Outward, Direction: "outward", Issue: *ref})
			}
		}
	}
	return result
}

func mapIssueRelations(issue *jiramodels.IssueScheme) []coremodels.IssueRelation {
	if issue == nil || issue.Fields == nil {
		return nil
	}
	return mapIssueLinks(issue.Fields.IssueLinks)
}

func componentNames(values []*jiramodels.ComponentScheme) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != nil && strings.TrimSpace(value.Name) != "" {
			result = append(result, value.Name)
		}
	}
	return result
}

func versionNames(values []*jiramodels.VersionScheme) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != nil && strings.TrimSpace(value.Name) != "" {
			result = append(result, value.Name)
		}
	}
	return result
}

func slicesOrNil(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}

func mapComment(comment *jiramodels.IssueCommentScheme) *coremodels.Comment {
	if comment == nil {
		return nil
	}
	return &coremodels.Comment{
		ID:      comment.ID,
		Author:  mapPerson(comment.Author),
		Created: comment.Created,
		Updated: comment.Updated,
		Body:    RenderADF(comment.Body),
	}
}

func mapSprint(value *jiramodels.SprintScheme, boardID int) coremodels.Sprint {
	return coremodels.Sprint{
		ID:           value.ID,
		Name:         value.Name,
		State:        value.State,
		StartDate:    formatTime(value.StartDate),
		EndDate:      formatTime(value.EndDate),
		CompleteDate: formatTime(value.CompleteDate),
		BoardID:      boardID,
		Goal:         value.Goal,
	}
}

func mapBoardSprint(value *jiramodels.BoardSprintScheme, boardID int) coremodels.Sprint {
	return coremodels.Sprint{
		ID:           value.ID,
		Name:         value.Name,
		State:        value.State,
		StartDate:    formatTime(value.StartDate),
		EndDate:      formatTime(value.EndDate),
		CompleteDate: formatTime(value.CompleteDate),
		BoardID:      boardID,
		Goal:         value.Goal,
	}
}

func mapVersion(value *jiramodels.VersionScheme) coremodels.Version {
	status := "In Development"
	if value.Released {
		status = "Released"
	}
	if value.Archived {
		status = "Archived"
	}
	return coremodels.Version{
		ID:          value.ID,
		Name:        value.Name,
		Description: value.Description,
		ProjectID:   value.ProjectID,
		Released:    value.Released,
		Archived:    value.Archived,
		ReleaseDate: value.ReleaseDate,
		URL:         value.Self,
		Status:      status,
	}
}
