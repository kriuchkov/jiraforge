package main

import (
	"context"
	"testing"
	"time"

	"google.golang.org/genai"

	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
)

func TestParseStreamingMode(t *testing.T) {
	t.Parallel()

	mode, err := parseStreamingMode("")
	if err != nil {
		t.Fatalf("parseStreamingMode returned error: %v", err)
	}
	if mode != "" {
		t.Fatalf("unexpected empty mode result: %q", mode)
	}

	mode, err = parseStreamingMode("sse")
	if err != nil {
		t.Fatalf("parseStreamingMode returned error: %v", err)
	}
	if mode != "sse" {
		t.Fatalf("unexpected streaming mode: %q", mode)
	}

	if _, err = parseStreamingMode("bad"); err == nil {
		t.Fatal("expected invalid streaming mode to fail")
	}
}

func TestResolveConsoleSessionCreatesNamedSession(t *testing.T) {
	ctx := context.Background()
	svc := session.InMemoryService()

	resolved, created, err := resolveConsoleSession(ctx, svc, "alice", "planning", false)
	if err != nil {
		t.Fatalf("resolveConsoleSession returned error: %v", err)
	}
	if !created {
		t.Fatal("expected session to be created")
	}
	if resolved.ID() != "planning" {
		t.Fatalf("unexpected session id: got %q want %q", resolved.ID(), "planning")
	}
}

func TestResolveConsoleSessionResumeLastUsesMostRecent(t *testing.T) {
	ctx := context.Background()
	svc := session.InMemoryService()

	older := mustCreateConsoleSession(t, ctx, svc, "alice", "older")
	appendConsoleEvent(t, ctx, svc, older, "older event", time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC))

	newer := mustCreateConsoleSession(t, ctx, svc, "alice", "newer")
	appendConsoleEvent(t, ctx, svc, newer, "newer event", time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC))

	resolved, created, err := resolveConsoleSession(ctx, svc, "alice", "", true)
	if err != nil {
		t.Fatalf("resolveConsoleSession returned error: %v", err)
	}
	if created {
		t.Fatal("expected an existing session to be resumed")
	}
	if resolved.ID() != "newer" {
		t.Fatalf("unexpected session id: got %q want %q", resolved.ID(), "newer")
	}
}

func TestResolveConsoleSessionRejectsConflictingFlags(t *testing.T) {
	ctx := context.Background()
	svc := session.InMemoryService()

	_, _, err := resolveConsoleSession(ctx, svc, "alice", "named", true)
	if err == nil {
		t.Fatal("expected conflicting session options to fail")
	}
}

func mustCreateConsoleSession(t *testing.T, ctx context.Context, svc session.Service, userID, sessionID string) session.Session {
	t.Helper()
	resp, err := svc.Create(ctx, &session.CreateRequest{AppName: agentAppName, UserID: userID, SessionID: sessionID})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	return resp.Session
}

func appendConsoleEvent(t *testing.T, ctx context.Context, svc session.Service, current session.Session, text string, timestamp time.Time) {
	t.Helper()
	event := session.NewEvent("console-test")
	event.Author = "user"
	event.Timestamp = timestamp
	event.LLMResponse = model.LLMResponse{Content: genai.NewContentFromText(text, genai.RoleUser)}
	if err := svc.AppendEvent(ctx, current, event); err != nil {
		t.Fatalf("AppendEvent returned error: %v", err)
	}
}
