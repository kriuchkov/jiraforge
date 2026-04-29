package adk

import (
	"context"
	"testing"
	"time"

	"google.golang.org/genai"

	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
)

func TestSessionMemorySearchScopesUserAndSkipsCurrentSession(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	svc := session.InMemoryService()
	search := NewSessionMemorySearch(svc)

	past := mustCreateSession(t, ctx, svc, "jiraforge_agent", "alice", "past")
	appendTextEvent(t, ctx, svc, past, "user", genai.RoleUser, "Need help with the retro agenda", time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC))

	current := mustCreateSession(t, ctx, svc, "jiraforge_agent", "alice", "current")
	appendTextEvent(t, ctx, svc, current, "user", genai.RoleUser, "This current session should be skipped", time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC))

	otherUser := mustCreateSession(t, ctx, svc, "jiraforge_agent", "bob", "other-user")
	appendTextEvent(t, ctx, svc, otherUser, "user", genai.RoleUser, "retro agenda for another user", time.Date(2026, 4, 7, 10, 0, 0, 0, time.UTC))

	hits, err := search.Search(ctx, SessionMemorySearchRequest{
		AppName:          "jiraforge_agent",
		UserID:           "alice",
		CurrentSessionID: "current",
		Query:            "retro agenda",
		Limit:            10,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("unexpected hit count: got %d want 1", len(hits))
	}
	if hits[0].SessionID != "past" {
		t.Fatalf("unexpected session id: got %q want %q", hits[0].SessionID, "past")
	}
	if hits[0].Text != "Need help with the retro agenda" {
		t.Fatalf("unexpected hit text: got %q", hits[0].Text)
	}
}

func TestSessionMemorySearchOrdersMostRelevantThenNewest(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	svc := session.InMemoryService()
	search := NewSessionMemorySearch(svc)

	older := mustCreateSession(t, ctx, svc, "jiraforge_agent", "alice", "older")
	appendTextEvent(t, ctx, svc, older, "user", genai.RoleUser, "Sprint planning for release board", time.Date(2026, 4, 7, 10, 0, 0, 0, time.UTC))

	newer := mustCreateSession(t, ctx, svc, "jiraforge_agent", "alice", "newer")
	appendTextEvent(t, ctx, svc, newer, "user", genai.RoleUser, "Sprint planning with release board and blockers", time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC))

	hits, err := search.Search(ctx, SessionMemorySearchRequest{
		AppName: "jiraforge_agent",
		UserID:  "alice",
		Query:   "release board blockers",
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(hits) != 2 {
		t.Fatalf("unexpected hit count: got %d want 2", len(hits))
	}
	if hits[0].SessionID != "newer" {
		t.Fatalf("unexpected first session: got %q want %q", hits[0].SessionID, "newer")
	}
	if hits[1].SessionID != "older" {
		t.Fatalf("unexpected second session: got %q want %q", hits[1].SessionID, "older")
	}
}

func appendTextEvent(t *testing.T, ctx context.Context, svc session.Service, current session.Session, author string, role genai.Role, text string, timestamp time.Time) {
	t.Helper()
	event := session.NewEvent("test-invocation")
	event.Author = author
	event.Timestamp = timestamp
	event.LLMResponse = model.LLMResponse{Content: genai.NewContentFromText(text, role)}
	if err := svc.AppendEvent(ctx, current, event); err != nil {
		t.Fatalf("AppendEvent returned error: %v", err)
	}
}

func mustCreateSession(t *testing.T, ctx context.Context, svc session.Service, appName, userID, sessionID string) session.Session {
	t.Helper()
	resp, err := svc.Create(ctx, &session.CreateRequest{AppName: appName, UserID: userID, SessionID: sessionID})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	return resp.Session
}
