//go:build integration

package atlassian

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-faster/errors"

	appconfig "github.com/kriuchkov/jiraforge/internal/config"
	coreerrors "github.com/kriuchkov/jiraforge/internal/core/errors"
	coremodels "github.com/kriuchkov/jiraforge/internal/core/models"
	"github.com/kriuchkov/jiraforge/internal/core/ports"
)

const (
	integrationTestTimeout = 45 * time.Second

	envTestIssueKey            = "JIRAFORGE_TEST_ISSUE_KEY"
	envTestProjectKey          = "JIRAFORGE_TEST_PROJECT_KEY"
	envTestBoardID             = "JIRAFORGE_TEST_BOARD_ID"
	envTestAttachmentID        = "JIRAFORGE_TEST_ATTACHMENT_ID"
	envTestMutationProjectKey  = "JIRAFORGE_TEST_MUTATION_PROJECT_KEY"
	envTestCreateIssueType     = "JIRAFORGE_TEST_CREATE_ISSUE_TYPE"
	envTestEnableMutations     = "JIRAFORGE_TEST_ENABLE_MUTATIONS"
	envTestLinkTargetIssueKey  = "JIRAFORGE_TEST_LINK_TARGET_ISSUE_KEY"
	envTestLinkType            = "JIRAFORGE_TEST_LINK_TYPE"
	defaultIntegrationLinkType = "Relates"
)

type integrationConfig struct {
	atlassian           appconfig.AtlassianConfig
	issueKey            string
	projectKey          string
	boardID             string
	attachmentID        string
	mutationProjectKey  string
	createIssueType     string
	enableMutations     bool
	linkTargetIssueKey  string
	linkType            string
	explicitLinkTypeEnv bool
}

type integrationHarness struct {
	gateway ports.JiraGateway
	cfg     integrationConfig
}

func TestGatewayIntegrationIssueReadOperations(t *testing.T) {
	h := newIntegrationHarness(t)
	issueKey := h.requireIssueKey(t)
	issue := h.mustGetIssue(t, issueKey)

	if issue.Key != issueKey {
		t.Fatal(errors.Errorf("unexpected issue key: got %q want %q", issue.Key, issueKey))
	}
	if strings.TrimSpace(issue.ID) == "" {
		t.Fatal(errors.New("expected issue ID to be populated"))
	}
	if strings.TrimSpace(issue.Summary) == "" {
		t.Fatal(errors.New("expected issue summary to be populated"))
	}

	t.Run("SearchIssues", func(t *testing.T) {
		result, err := h.gateway.SearchIssues(h.context(t), coremodels.SearchIssuesRequest{
			JQL:        fmt.Sprintf("key = %s", issueKey),
			Fields:     []string{"summary", "status", "issuetype"},
			MaxResults: 1,
		})
		if err != nil {
			t.Fatal(errors.Wrap(err, "search issues"))
		}
		if result == nil {
			t.Fatal(errors.New("expected search result"))
		}
		if result.Total == 0 || len(result.Issues) == 0 {
			t.Fatal(errors.Errorf("expected search results for %s", issueKey))
		}
		if result.Issues[0].Key != issueKey {
			t.Fatal(errors.Errorf("unexpected search hit: got %q want %q", result.Issues[0].Key, issueKey))
		}
	})

	t.Run("GetComments", func(t *testing.T) {
		comments, err := h.gateway.GetComments(h.context(t), coremodels.GetCommentsRequest{IssueKey: issueKey})
		if err != nil {
			t.Fatal(errors.Wrap(err, "get comments"))
		}
		if comments == nil {
			t.Fatal(errors.New("expected comment list"))
		}
		if comments.IssueKey != issueKey {
			t.Fatal(errors.Errorf("unexpected comment list issue key: got %q want %q", comments.IssueKey, issueKey))
		}
	})

	t.Run("GetTransitions", func(t *testing.T) {
		transitions, err := h.gateway.GetTransitions(h.context(t), coremodels.GetTransitionsRequest{IssueKey: issueKey})
		if err != nil {
			t.Fatal(errors.Wrap(err, "get transitions"))
		}
		for _, transition := range transitions {
			if strings.TrimSpace(transition.ID) == "" || strings.TrimSpace(transition.Name) == "" {
				t.Fatal(errors.Errorf("unexpected transition payload: %+v", transition))
			}
		}
	})

	t.Run("GetIssueHistory", func(t *testing.T) {
		history, err := h.gateway.GetIssueHistory(h.context(t), coremodels.GetIssueHistoryRequest{IssueKey: issueKey})
		if err != nil {
			t.Fatal(errors.Wrap(err, "get issue history"))
		}
		if history == nil {
			t.Fatal(errors.New("expected history payload"))
		}
		if history.IssueKey != issueKey {
			t.Fatal(errors.Errorf("unexpected history issue key: got %q want %q", history.IssueKey, issueKey))
		}
	})

	t.Run("GetRelatedIssues", func(t *testing.T) {
		relations, err := h.gateway.GetRelatedIssues(h.context(t), coremodels.GetRelatedIssuesRequest{IssueKey: issueKey})
		if err != nil {
			t.Fatal(errors.Wrap(err, "get related issues"))
		}
		for _, relation := range relations {
			if strings.TrimSpace(relation.Issue.Key) == "" {
				t.Fatal(errors.Errorf("unexpected related issue payload: %+v", relation))
			}
		}
	})

	t.Run("DownloadAttachment", func(t *testing.T) {
		attachmentID := h.cfg.attachmentID
		if attachmentID == "" && len(issue.Attachments) > 0 {
			attachmentID = issue.Attachments[0].ID
		}
		if attachmentID == "" {
			t.Skipf("set %s or use an issue with attachments", envTestAttachmentID)
		}

		attachment, err := h.gateway.DownloadAttachment(h.context(t), coremodels.DownloadAttachmentRequest{AttachmentID: attachmentID})
		if err != nil {
			t.Fatal(errors.Wrap(err, "download attachment"))
		}
		if attachment == nil {
			t.Fatal(errors.New("expected attachment payload"))
		}
		if attachment.ID != attachmentID {
			t.Fatal(errors.Errorf("unexpected attachment ID: got %q want %q", attachment.ID, attachmentID))
		}
		if strings.TrimSpace(attachment.Filename) == "" {
			t.Fatal(errors.New("expected attachment filename"))
		}
		if attachment.Size > 0 && len(attachment.Data) == 0 {
			t.Fatal(errors.New("expected attachment data for non-empty attachment"))
		}
	})

	t.Run("GetDevelopmentInfo", func(t *testing.T) {
		info, err := h.gateway.GetDevelopmentInfo(h.context(t), coremodels.GetDevelopmentInfoRequest{
			IssueKey:            issueKey,
			IncludeBranches:     true,
			IncludePullRequests: true,
			IncludeCommits:      true,
			IncludeBuilds:       true,
		})
		if err != nil {
			if coreerrors.IsCode(err, coreerrors.CodeUnavailable) {
				t.Skip("development endpoint is not available or Jira dev integrations are disabled")
			}
			t.Fatal(errors.Wrap(err, "get development info"))
		}
		if info == nil {
			t.Fatal(errors.New("expected development info payload"))
		}
		if info.IssueKey != issueKey {
			t.Fatal(errors.Errorf("unexpected development issue key: got %q want %q", info.IssueKey, issueKey))
		}
		if strings.TrimSpace(info.IssueID) == "" {
			t.Fatal(errors.New("expected development info issue ID"))
		}
	})
}

func TestGatewayIntegrationProjectReadOperations(t *testing.T) {
	h := newIntegrationHarness(t)
	var referenceIssue *coremodels.Issue
	if strings.TrimSpace(h.cfg.projectKey) == "" {
		referenceIssue = h.referenceIssue(t)
	}
	projectKey := h.requireProjectKey(t, referenceIssue)

	t.Run("ListIssueTypes", func(t *testing.T) {
		issueTypes := h.mustListIssueTypes(t, projectKey)
		if len(issueTypes) == 0 {
			t.Fatal(errors.New("expected at least one issue type"))
		}
		foundNamedType := false
		for _, issueType := range issueTypes {
			if strings.TrimSpace(issueType.Name) != "" {
				foundNamedType = true
				break
			}
		}
		if !foundNamedType {
			t.Fatal(errors.New("expected at least one named issue type"))
		}
	})

	t.Run("ListStatuses", func(t *testing.T) {
		catalog, err := h.gateway.ListStatuses(h.context(t), coremodels.ListStatusesRequest{ProjectKey: projectKey})
		if err != nil {
			t.Fatal(errors.Wrap(err, "list statuses"))
		}
		if catalog == nil {
			t.Fatal(errors.New("expected status catalog"))
		}
		if catalog.ProjectKey != projectKey {
			t.Fatal(errors.Errorf("unexpected project key: got %q want %q", catalog.ProjectKey, projectKey))
		}
		if len(catalog.Groups) == 0 {
			t.Fatal(errors.New("expected at least one status group"))
		}
	})

	t.Run("Versions", func(t *testing.T) {
		versions, err := h.gateway.ListProjectVersions(h.context(t), coremodels.ListProjectVersionsRequest{ProjectKey: projectKey})
		if err != nil {
			t.Fatal(errors.Wrap(err, "list project versions"))
		}
		if versions == nil {
			t.Fatal(errors.New("expected version collection"))
		}
		if versions.ProjectKey != projectKey {
			t.Fatal(errors.Errorf("unexpected versions project key: got %q want %q", versions.ProjectKey, projectKey))
		}
		if len(versions.Versions) == 0 {
			t.Skip("project has no versions")
		}

		version, err := h.gateway.GetVersion(h.context(t), coremodels.GetVersionRequest{VersionID: versions.Versions[0].ID})
		if err != nil {
			t.Fatal(errors.Wrap(err, "get version"))
		}
		if version == nil {
			t.Fatal(errors.New("expected version payload"))
		}
		if version.ID != versions.Versions[0].ID {
			t.Fatal(errors.Errorf("unexpected version ID: got %q want %q", version.ID, versions.Versions[0].ID))
		}
	})

	t.Run("Sprints", func(t *testing.T) {
		request := coremodels.ListSprintsRequest{BoardID: h.cfg.boardID, ProjectKey: projectKey}
		collection, err := h.gateway.ListSprints(h.context(t), request)
		if err != nil {
			if coreerrors.IsCode(err, coreerrors.CodeNotFound) {
				t.Skip("no Jira boards found for the configured sprint scope")
			}
			t.Fatal(errors.Wrap(err, "list sprints"))
		}
		if collection == nil {
			t.Fatal(errors.New("expected sprint collection"))
		}
		if len(collection.Sprints) == 0 {
			t.Skip("no sprints found for the configured sprint scope")
		}

		firstSprint := collection.Sprints[0]
		if firstSprint.ID == 0 {
			t.Fatal(errors.Errorf("unexpected sprint payload: %+v", firstSprint))
		}

		sprint, err := h.gateway.GetSprint(h.context(t), coremodels.GetSprintRequest{SprintID: fmt.Sprintf("%d", firstSprint.ID)})
		if err != nil {
			t.Fatal(errors.Wrap(err, "get sprint"))
		}
		if sprint == nil {
			t.Fatal(errors.New("expected sprint payload"))
		}
		if sprint.ID != firstSprint.ID {
			t.Fatal(errors.Errorf("unexpected sprint ID: got %d want %d", sprint.ID, firstSprint.ID))
		}

		searchResult, err := h.gateway.SearchSprints(h.context(t), coremodels.SearchSprintsRequest{
			BoardID:    h.cfg.boardID,
			ProjectKey: projectKey,
			Name:       firstSprint.Name,
			ExactMatch: true,
		})
		if err != nil {
			t.Fatal(errors.Wrap(err, "search sprints"))
		}
		if searchResult == nil {
			t.Fatal(errors.New("expected sprint search result"))
		}

		found := false
		for _, sprint := range searchResult.Sprints {
			if sprint.ID == firstSprint.ID {
				found = true
				break
			}
		}
		if !found {
			t.Fatal(errors.Errorf("expected sprint %d to be returned by exact-name search", firstSprint.ID))
		}

		activeSprint, err := h.gateway.GetActiveSprint(h.context(t), coremodels.GetActiveSprintRequest{BoardID: h.cfg.boardID, ProjectKey: projectKey})
		if err != nil {
			if coreerrors.IsCode(err, coreerrors.CodeNotFound) {
				t.Skip("no active sprint for the configured sprint scope")
			}
			t.Fatal(errors.Wrap(err, "get active sprint"))
		}
		if activeSprint == nil {
			t.Fatal(errors.New("expected active sprint payload"))
		}
	})
}

func TestGatewayIntegrationMutations(t *testing.T) {
	h := newIntegrationHarness(t)
	if !h.cfg.enableMutations {
		t.Skipf("set %s=1 to run create/update/comment/worklog/transition/delete scenarios", envTestEnableMutations)
	}

	var referenceIssue *coremodels.Issue
	if strings.TrimSpace(h.cfg.mutationProjectKey) == "" && strings.TrimSpace(h.cfg.projectKey) == "" {
		referenceIssue = h.referenceIssue(t)
	}
	projectKey := h.requireMutationProjectKey(t, referenceIssue)
	issueTypes := h.mustListIssueTypes(t, projectKey)
	issueType := h.createIssueType(issueTypes)
	if issueType == "" {
		t.Fatal(errors.New("expected a non-subtask issue type for creation"))
	}

	uniqueSuffix := time.Now().UTC().Format("20060102-150405")
	createdSummary := fmt.Sprintf("JiraForge integration %s", uniqueSuffix)
	createdDescription := "created by internal/adapters/atlassian integration test"

	created, err := h.gateway.CreateIssue(h.context(t), coremodels.CreateIssueRequest{
		ProjectKey:  projectKey,
		Summary:     createdSummary,
		Description: createdDescription,
		IssueType:   issueType,
	})
	if err != nil {
		t.Fatal(errors.Wrap(err, "create issue"))
	}
	if created == nil {
		t.Fatal(errors.New("expected create issue result"))
	}
	if strings.TrimSpace(created.Key) == "" {
		t.Fatal(errors.New("expected created issue key"))
	}

	createdIssueKey := created.Key
	deleted := false
	childIssueKeys := make([]string, 0, 1)
	t.Cleanup(func() {
		for index := len(childIssueKeys) - 1; index >= 0; index-- {
			h.cleanupDeleteIssue(t, childIssueKeys[index])
		}
		if !deleted {
			h.cleanupDeleteIssue(t, createdIssueKey)
		}
	})

	t.Run("GetAndUpdateIssue", func(t *testing.T) {
		issue := h.mustGetIssue(t, createdIssueKey)
		if issue.Summary != createdSummary {
			t.Fatal(errors.Errorf("unexpected created summary: got %q want %q", issue.Summary, createdSummary))
		}

		updatedSummary := createdSummary + " updated"
		updatedDescription := createdDescription + " updated"
		result, err := h.gateway.UpdateIssue(h.context(t), coremodels.UpdateIssueRequest{
			IssueKey:    createdIssueKey,
			Summary:     updatedSummary,
			Description: updatedDescription,
		})
		if err != nil {
			t.Fatal(errors.Wrap(err, "update issue"))
		}
		if result == nil {
			t.Fatal(errors.New("expected update issue result"))
		}

		issue = h.mustGetIssue(t, createdIssueKey)
		if issue.Summary != updatedSummary {
			t.Fatal(errors.Errorf("unexpected updated summary: got %q want %q", issue.Summary, updatedSummary))
		}
		if !strings.Contains(issue.Description, "updated") {
			t.Fatal(errors.Errorf("expected updated description, got %q", issue.Description))
		}
	})

	t.Run("AddCommentAndReadBack", func(t *testing.T) {
		commentBody := fmt.Sprintf("integration comment %s", uniqueSuffix)
		comment, err := h.gateway.AddComment(h.context(t), coremodels.AddCommentRequest{IssueKey: createdIssueKey, Comment: commentBody})
		if err != nil {
			t.Fatal(errors.Wrap(err, "add comment"))
		}
		if comment == nil {
			t.Fatal(errors.New("expected comment payload"))
		}
		if strings.TrimSpace(comment.ID) == "" {
			t.Fatal(errors.New("expected comment ID"))
		}

		comments, err := h.gateway.GetComments(h.context(t), coremodels.GetCommentsRequest{IssueKey: createdIssueKey})
		if err != nil {
			t.Fatal(errors.Wrap(err, "get comments after add"))
		}
		if comments == nil {
			t.Fatal(errors.New("expected comment list"))
		}

		found := false
		for _, item := range comments.Comments {
			if item.ID == comment.ID || strings.Contains(item.Body, commentBody) {
				found = true
				break
			}
		}
		if !found {
			t.Fatal(errors.Errorf("expected comment %q to be present in issue comments", commentBody))
		}
	})

	t.Run("AddWorklog", func(t *testing.T) {
		worklog, err := h.gateway.AddWorklog(h.context(t), coremodels.AddWorklogRequest{
			IssueKey:  createdIssueKey,
			TimeSpent: "60",
			Comment:   fmt.Sprintf("integration worklog %s", uniqueSuffix),
		})
		if err != nil {
			t.Fatal(errors.Wrap(err, "add worklog"))
		}
		if worklog == nil {
			t.Fatal(errors.New("expected worklog payload"))
		}
		if worklog.TimeSpentSeconds != 60 {
			t.Fatal(errors.Errorf("unexpected worklog seconds: got %d want 60", worklog.TimeSpentSeconds))
		}
	})

	t.Run("CreateChildIssue", func(t *testing.T) {
		subtaskType := h.subtaskIssueType(issueTypes)
		if subtaskType == "" {
			t.Skip("project does not expose a subtask issue type")
		}

		child, err := h.gateway.CreateChildIssue(h.context(t), coremodels.CreateChildIssueRequest{
			ParentIssueKey: createdIssueKey,
			Summary:        fmt.Sprintf("Child integration %s", uniqueSuffix),
			Description:    "child issue created by integration test",
			IssueType:      subtaskType,
		})
		if err != nil {
			t.Fatal(errors.Wrap(err, "create child issue"))
		}
		if child == nil {
			t.Fatal(errors.New("expected child issue result"))
		}
		if strings.TrimSpace(child.Key) == "" {
			t.Fatal(errors.New("expected child issue key"))
		}
		childIssueKeys = append(childIssueKeys, child.Key)

		childIssue := h.mustGetIssue(t, child.Key)
		if childIssue.Parent == nil || childIssue.Parent.Key != createdIssueKey {
			t.Fatal(errors.Errorf("expected child issue parent %q, got %+v", createdIssueKey, childIssue.Parent))
		}
	})

	t.Run("LinkIssues", func(t *testing.T) {
		targetIssueKey := h.linkTargetIssueKey(createdIssueKey)
		if targetIssueKey == "" {
			t.Skipf("set %s or %s to run issue link scenario", envTestLinkTargetIssueKey, envTestIssueKey)
		}
		if targetIssueKey == createdIssueKey {
			t.Skip("link target must differ from the created issue")
		}

		_, err := h.gateway.LinkIssues(h.context(t), coremodels.LinkIssuesRequest{
			InwardIssue:  createdIssueKey,
			OutwardIssue: targetIssueKey,
			LinkType:     h.linkType(),
			Comment:      fmt.Sprintf("integration link %s", uniqueSuffix),
		})
		if err != nil {
			if !h.cfg.explicitLinkTypeEnv && coreerrors.IsCode(err, coreerrors.CodeInvalidArgument) {
				t.Skipf("default link type %q is unavailable in this Jira; set %s", defaultIntegrationLinkType, envTestLinkType)
			}
			t.Fatal(errors.Wrap(err, "link issues"))
		}

		relations, err := h.gateway.GetRelatedIssues(h.context(t), coremodels.GetRelatedIssuesRequest{IssueKey: createdIssueKey})
		if err != nil {
			t.Fatal(errors.Wrap(err, "get related issues after link"))
		}

		found := false
		for _, relation := range relations {
			if relation.Issue.Key == targetIssueKey {
				found = true
				break
			}
		}
		if !found {
			t.Fatal(errors.Errorf("expected linked issue %q to appear in related issues", targetIssueKey))
		}
	})

	t.Run("TransitionIssue", func(t *testing.T) {
		transitions, err := h.gateway.GetTransitions(h.context(t), coremodels.GetTransitionsRequest{IssueKey: createdIssueKey})
		if err != nil {
			t.Fatal(errors.Wrap(err, "get transitions before move"))
		}
		if len(transitions) == 0 {
			t.Skip("created issue has no available transitions")
		}

		result, err := h.gateway.TransitionIssue(h.context(t), coremodels.TransitionIssueRequest{
			IssueKey:     createdIssueKey,
			TransitionID: transitions[0].ID,
		})
		if err != nil {
			t.Fatal(errors.Wrap(err, "transition issue"))
		}
		if result == nil {
			t.Fatal(errors.New("expected transition result"))
		}
		if result.Key != createdIssueKey {
			t.Fatal(errors.Errorf("unexpected transition result key: got %q want %q", result.Key, createdIssueKey))
		}
	})

	t.Run("DeleteIssue", func(t *testing.T) {
		for index := len(childIssueKeys) - 1; index >= 0; index-- {
			_, err := h.gateway.DeleteIssue(h.context(t), coremodels.DeleteIssueRequest{IssueKey: childIssueKeys[index]})
			if err != nil && !coreerrors.IsCode(err, coreerrors.CodeNotFound) {
				t.Fatal(errors.Wrap(err, "delete child issue"))
			}
		}
		childIssueKeys = nil

		_, err := h.gateway.DeleteIssue(h.context(t), coremodels.DeleteIssueRequest{IssueKey: createdIssueKey})
		if err != nil {
			t.Fatal(errors.Wrap(err, "delete issue"))
		}
		deleted = true

		_, err = h.gateway.GetIssue(h.context(t), coremodels.GetIssueRequest{IssueKey: createdIssueKey})
		if !coreerrors.IsCode(err, coreerrors.CodeNotFound) {
			t.Fatal(errors.Errorf("expected deleted issue lookup to return not_found, got %v", err))
		}
	})
}

func newIntegrationHarness(t *testing.T) *integrationHarness {
	t.Helper()

	linkType, explicitLinkTypeEnv := lookupTrimmedEnv(envTestLinkType)
	config := integrationConfig{
		atlassian: appconfig.AtlassianConfig{
			Host:     envString("ATLASSIAN_HOST"),
			Email:    envString("ATLASSIAN_EMAIL"),
			Token:    envString("ATLASSIAN_TOKEN"),
			ProxyURL: envString("PROXY_URL"),
		},
		issueKey:            envString(envTestIssueKey),
		projectKey:          envString(envTestProjectKey),
		boardID:             envString(envTestBoardID),
		attachmentID:        envString(envTestAttachmentID),
		mutationProjectKey:  envString(envTestMutationProjectKey),
		createIssueType:     envString(envTestCreateIssueType),
		enableMutations:     envBool(envTestEnableMutations),
		linkTargetIssueKey:  envString(envTestLinkTargetIssueKey),
		linkType:            linkType,
		explicitLinkTypeEnv: explicitLinkTypeEnv,
	}

	if err := config.atlassian.Validate(); err != nil {
		t.Skip(errors.Wrap(err, "configure live Atlassian adapter tests"))
	}

	gateway, err := NewGateway(config.atlassian)
	if err != nil {
		t.Fatal(errors.Wrap(err, "create live Atlassian gateway"))
	}

	return &integrationHarness{gateway: gateway, cfg: config}
}

func (h *integrationHarness) context(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	t.Cleanup(cancel)
	return ctx
}

func (h *integrationHarness) requireIssueKey(t *testing.T) string {
	t.Helper()
	if strings.TrimSpace(h.cfg.issueKey) == "" {
		t.Skipf("set %s to run issue-level live integration tests", envTestIssueKey)
	}
	return h.cfg.issueKey
}

func (h *integrationHarness) referenceIssue(t *testing.T) *coremodels.Issue {
	t.Helper()
	if strings.TrimSpace(h.cfg.issueKey) == "" {
		return nil
	}
	return h.mustGetIssue(t, h.cfg.issueKey)
}

func (h *integrationHarness) requireProjectKey(t *testing.T, issue *coremodels.Issue) string {
	t.Helper()
	if strings.TrimSpace(h.cfg.projectKey) != "" {
		return h.cfg.projectKey
	}
	if issue != nil && strings.TrimSpace(issue.ProjectKey) != "" {
		return issue.ProjectKey
	}
	t.Skipf("set %s or provide %s that belongs to a project", envTestProjectKey, envTestIssueKey)
	return ""
}

func (h *integrationHarness) requireMutationProjectKey(t *testing.T, issue *coremodels.Issue) string {
	t.Helper()
	if strings.TrimSpace(h.cfg.mutationProjectKey) != "" {
		return h.cfg.mutationProjectKey
	}
	return h.requireProjectKey(t, issue)
}

func (h *integrationHarness) mustGetIssue(t *testing.T, issueKey string) *coremodels.Issue {
	t.Helper()
	issue, err := h.gateway.GetIssue(h.context(t), coremodels.GetIssueRequest{
		IssueKey: issueKey,
		Fields:   []string{"summary", "description", "project", "attachment", "issuetype", "status", "parent"},
	})
	if err != nil {
		t.Fatal(errors.Wrap(err, "get issue"))
	}
	if issue == nil {
		t.Fatal(errors.New("expected issue payload"))
	}
	return issue
}

func (h *integrationHarness) mustListIssueTypes(t *testing.T, projectKey string) []coremodels.IssueType {
	t.Helper()
	issueTypes, err := h.gateway.ListIssueTypes(h.context(t), coremodels.ListIssueTypesRequest{ProjectKey: projectKey})
	if err != nil {
		t.Fatal(errors.Wrap(err, "list issue types"))
	}
	return issueTypes
}

func (h *integrationHarness) createIssueType(issueTypes []coremodels.IssueType) string {
	if strings.TrimSpace(h.cfg.createIssueType) != "" {
		return h.cfg.createIssueType
	}
	for _, issueType := range issueTypes {
		if !issueType.Subtask && strings.EqualFold(issueType.Name, "Task") {
			return issueType.Name
		}
	}
	for _, issueType := range issueTypes {
		if !issueType.Subtask && strings.TrimSpace(issueType.Name) != "" {
			return issueType.Name
		}
	}
	return ""
}

func (h *integrationHarness) subtaskIssueType(issueTypes []coremodels.IssueType) string {
	for _, issueType := range issueTypes {
		if issueType.Subtask && strings.TrimSpace(issueType.Name) != "" {
			return issueType.Name
		}
	}
	return ""
}

func (h *integrationHarness) linkTargetIssueKey(createdIssueKey string) string {
	if strings.TrimSpace(h.cfg.linkTargetIssueKey) != "" {
		return h.cfg.linkTargetIssueKey
	}
	if strings.TrimSpace(h.cfg.issueKey) != "" && h.cfg.issueKey != createdIssueKey {
		return h.cfg.issueKey
	}
	return ""
}

func (h *integrationHarness) linkType() string {
	if strings.TrimSpace(h.cfg.linkType) != "" {
		return h.cfg.linkType
	}
	return defaultIntegrationLinkType
}

func (h *integrationHarness) cleanupDeleteIssue(t *testing.T, issueKey string) {
	t.Helper()
	if strings.TrimSpace(issueKey) == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	_, err := h.gateway.DeleteIssue(ctx, coremodels.DeleteIssueRequest{IssueKey: issueKey})
	if err != nil && !coreerrors.IsCode(err, coreerrors.CodeNotFound) {
		t.Error(errors.Wrap(err, "cleanup delete issue"))
	}
}

func envString(name string) string {
	value, _ := lookupTrimmedEnv(name)
	return value
}

func envBool(name string) bool {
	switch strings.ToLower(envString(name)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func lookupTrimmedEnv(name string) (string, bool) {
	value, ok := os.LookupEnv(name)
	trimmed := strings.TrimSpace(value)
	return trimmed, ok && trimmed != ""
}
