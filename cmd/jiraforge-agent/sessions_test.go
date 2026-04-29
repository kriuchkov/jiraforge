package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"google.golang.org/genai"

	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
)

func TestRunSessionCommandListOrdersNewestFirst(t *testing.T) {
	ctx := context.Background()
	svc := session.InMemoryService()

	older := mustCreateAgentSession(t, ctx, svc, "alice", "older")
	appendAgentTextEvent(t, ctx, svc, older, "user", genai.RoleUser, "older preview", time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC))

	newer := mustCreateAgentSession(t, ctx, svc, "alice", "newer")
	appendAgentTextEvent(t, ctx, svc, newer, "user", genai.RoleUser, "newer preview", time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC))

	bob := mustCreateAgentSession(t, ctx, svc, "bob", "other")
	appendAgentTextEvent(t, ctx, svc, bob, "user", genai.RoleUser, "other user preview", time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC))

	var output bytes.Buffer
	if err := runSessionCommand(ctx, svc, []string{"list", "--user-id", "alice"}, "default", &output); err != nil {
		t.Fatalf("runSessionCommand returned error: %v", err)
	}

	text := output.String()
	if !strings.Contains(text, "Sessions for user alice (2)") {
		t.Fatalf("unexpected list output: %s", text)
	}
	if strings.Contains(text, "other") {
		t.Fatalf("list output should not include other users: %s", text)
	}
	if strings.Index(text, "- newer") > strings.Index(text, "- older") {
		t.Fatalf("expected newer session before older: %s", text)
	}
	if !strings.Contains(text, "last: user: newer preview") {
		t.Fatalf("missing last preview in list output: %s", text)
	}
}

func TestRunSessionCommandInspectPrintsTranscript(t *testing.T) {
	ctx := context.Background()
	svc := session.InMemoryService()

	current := mustCreateAgentSession(t, ctx, svc, "alice", "release-audit")
	appendAgentEvent(t, ctx, svc, current, "user", []*genai.Part{genai.NewPartFromText("Summarize blockers")}, time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC))
	appendAgentEvent(t, ctx, svc, current, "jiraforge_agent", []*genai.Part{genai.NewPartFromFunctionCall("load_memory", map[string]any{"query": "blockers"})}, time.Date(2026, 4, 10, 10, 1, 0, 0, time.UTC))

	var output bytes.Buffer
	if err := runSessionCommand(ctx, svc, []string{"inspect", "--user-id", "alice", "--session-id", "release-audit", "--recent", "0"}, "default", &output); err != nil {
		t.Fatalf("runSessionCommand returned error: %v", err)
	}

	text := output.String()
	if !strings.Contains(text, "Session: release-audit") {
		t.Fatalf("missing session header: %s", text)
	}
	if !strings.Contains(text, "Summarize blockers") {
		t.Fatalf("missing text event: %s", text)
	}
	if !strings.Contains(text, "tool call load_memory") {
		t.Fatalf("missing tool call event: %s", text)
	}
}

func TestRunSessionCommandDeleteRemovesSession(t *testing.T) {
	ctx := context.Background()
	svc := session.InMemoryService()

	current := mustCreateAgentSession(t, ctx, svc, "alice", "cleanup")
	appendAgentTextEvent(t, ctx, svc, current, "user", genai.RoleUser, "delete me", time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC))

	var output bytes.Buffer
	if err := runSessionCommand(ctx, svc, []string{"delete", "--user-id", "alice", "--session-id", "cleanup"}, "default", &output); err != nil {
		t.Fatalf("runSessionCommand returned error: %v", err)
	}

	text := output.String()
	if !strings.Contains(text, "Deleted session cleanup for user alice.") {
		t.Fatalf("unexpected delete output: %s", text)
	}

	list, err := svc.List(ctx, &session.ListRequest{AppName: agentAppName, UserID: "alice"})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(list.Sessions) != 0 {
		t.Fatalf("expected no sessions after delete, got %d", len(list.Sessions))
	}
}

func mustCreateAgentSession(t *testing.T, ctx context.Context, svc session.Service, userID, sessionID string) session.Session {
	t.Helper()
	resp, err := svc.Create(ctx, &session.CreateRequest{AppName: agentAppName, UserID: userID, SessionID: sessionID})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	return resp.Session
}

func appendAgentTextEvent(t *testing.T, ctx context.Context, svc session.Service, current session.Session, author string, role genai.Role, text string, timestamp time.Time) {
	t.Helper()
	appendAgentEvent(t, ctx, svc, current, author, []*genai.Part{genai.NewPartFromText(text)}, timestamp)
}

func appendAgentEvent(t *testing.T, ctx context.Context, svc session.Service, current session.Session, author string, parts []*genai.Part, timestamp time.Time) {
	t.Helper()
	event := session.NewEvent("session-command-test")
	event.Author = author
	event.Timestamp = timestamp
	event.LLMResponse = model.LLMResponse{Content: &genai.Content{Role: genai.RoleModel, Parts: parts}}
	if err := svc.AppendEvent(ctx, current, event); err != nil {
		t.Fatalf("AppendEvent returned error: %v", err)
	}
}
