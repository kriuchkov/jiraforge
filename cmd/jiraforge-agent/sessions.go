package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"

	"google.golang.org/adk/session"
)

type sessionSummary struct {
	ID          string
	UpdatedAt   string
	EventCount  int
	LastPreview string
}

type sessionListOptions struct {
	UserID string
	Limit  int
}

type sessionInspectOptions struct {
	UserID    string
	SessionID string
	Recent    int
}

type sessionDeleteOptions struct {
	UserID    string
	SessionID string
}

func runSessionCommand(ctx context.Context, sessionService session.Service, args []string, defaultUserID string, stdout io.Writer) error {
	if sessionService == nil {
		return fmt.Errorf("session service is required")
	}
	if stdout == nil {
		stdout = io.Discard
	}
	if len(args) == 0 {
		return fmt.Errorf("session subcommand is required: list, inspect, delete")
	}

	switch strings.TrimSpace(args[0]) {
	case "list":
		return runSessionListCommand(ctx, sessionService, args[1:], defaultUserID, stdout)
	case "inspect":
		return runSessionInspectCommand(ctx, sessionService, args[1:], defaultUserID, stdout)
	case "delete":
		return runSessionDeleteCommand(ctx, sessionService, args[1:], defaultUserID, stdout)
	default:
		return fmt.Errorf("unsupported session subcommand %q; use list, inspect, or delete", args[0])
	}
}

func runSessionListCommand(ctx context.Context, sessionService session.Service, args []string, defaultUserID string, stdout io.Writer) error {
	flags := flag.NewFlagSet("sessions list", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	userID := defaultUserID
	limit := 20
	flags.StringVar(&userID, "user-id", userID, "User identity that owns the sessions")
	flags.IntVar(&limit, "limit", limit, "Maximum number of sessions to print; use 0 to print all")

	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse sessions list flags: %w", err)
	}
	if len(flags.Args()) > 0 {
		return fmt.Errorf("cannot parse following arguments: %v", flags.Args())
	}
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("--user-id is required")
	}
	if limit < 0 {
		return fmt.Errorf("--limit must be >= 0")
	}

	return runSessionList(ctx, sessionService, sessionListOptions{UserID: userID, Limit: limit}, stdout)
}

func runSessionList(ctx context.Context, sessionService session.Service, options sessionListOptions, stdout io.Writer) error {
	userID := strings.TrimSpace(options.UserID)
	if userID == "" {
		return fmt.Errorf("--user-id is required")
	}
	if options.Limit < 0 {
		return fmt.Errorf("--limit must be >= 0")
	}

	summaries, err := listSessionSummaries(ctx, sessionService, userID)
	if err != nil {
		return err
	}
	if options.Limit > 0 && len(summaries) > options.Limit {
		summaries = summaries[:options.Limit]
	}

	if len(summaries) == 0 {
		_, err = fmt.Fprintf(stdout, "No sessions found for user %s.\n", userID)
		return err
	}

	if _, err := fmt.Fprintf(stdout, "Sessions for user %s (%d)\n", userID, len(summaries)); err != nil {
		return err
	}
	for _, summary := range summaries {
		if _, err := fmt.Fprintf(stdout, "\n- %s\n  updated: %s\n  events: %d\n", summary.ID, summary.UpdatedAt, summary.EventCount); err != nil {
			return err
		}
		if strings.TrimSpace(summary.LastPreview) != "" {
			if _, err := fmt.Fprintf(stdout, "  last: %s\n", summary.LastPreview); err != nil {
				return err
			}
		}
	}
	return nil
}

func runSessionInspectCommand(ctx context.Context, sessionService session.Service, args []string, defaultUserID string, stdout io.Writer) error {
	flags := flag.NewFlagSet("sessions inspect", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	userID := defaultUserID
	sessionID := ""
	recent := 20
	flags.StringVar(&userID, "user-id", userID, "User identity that owns the session")
	flags.StringVar(&sessionID, "session-id", sessionID, "Session ID to inspect")
	flags.IntVar(&recent, "recent", recent, "Number of most recent events to print; use 0 to print all")

	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse sessions inspect flags: %w", err)
	}
	if len(flags.Args()) > 0 {
		return fmt.Errorf("cannot parse following arguments: %v", flags.Args())
	}
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("--user-id is required")
	}
	if strings.TrimSpace(sessionID) == "" {
		return fmt.Errorf("--session-id is required")
	}
	if recent < 0 {
		return fmt.Errorf("--recent must be >= 0")
	}

	return runSessionInspect(ctx, sessionService, sessionInspectOptions{UserID: userID, SessionID: sessionID, Recent: recent}, stdout)
}

func runSessionInspect(ctx context.Context, sessionService session.Service, options sessionInspectOptions, stdout io.Writer) error {
	userID := strings.TrimSpace(options.UserID)
	sessionID := strings.TrimSpace(options.SessionID)
	if userID == "" {
		return fmt.Errorf("--user-id is required")
	}
	if sessionID == "" {
		return fmt.Errorf("--session-id is required")
	}
	if options.Recent < 0 {
		return fmt.Errorf("--recent must be >= 0")
	}

	fullSession, _, err := getSession(ctx, sessionService, userID, sessionID)
	if err != nil {
		return err
	}

	events := collectEvents(fullSession.Events())
	visibleEvents := events
	if options.Recent > 0 && len(visibleEvents) > options.Recent {
		visibleEvents = visibleEvents[len(visibleEvents)-options.Recent:]
	}

	if _, err := fmt.Fprintf(stdout, "Session: %s\nUser: %s\nUpdated: %s\nEvents: %d\n", fullSession.ID(), fullSession.UserID(), formatSessionTime(fullSession.LastUpdateTime()), len(events)); err != nil {
		return err
	}

	if len(events) == 0 {
		_, err = fmt.Fprintln(stdout, "\nNo events recorded.")
		return err
	}

	if len(visibleEvents) != len(events) {
		if _, err := fmt.Fprintf(stdout, "Showing last %d events.\n", len(visibleEvents)); err != nil {
			return err
		}
	}

	for _, event := range visibleEvents {
		if _, err := fmt.Fprintf(stdout, "\n[%s] %s\n%s\n", formatSessionTime(event.Timestamp), eventAuthor(event), eventDisplayText(event)); err != nil {
			return err
		}
	}

	return nil
}

func runSessionDeleteCommand(ctx context.Context, sessionService session.Service, args []string, defaultUserID string, stdout io.Writer) error {
	flags := flag.NewFlagSet("sessions delete", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	userID := defaultUserID
	sessionID := ""
	flags.StringVar(&userID, "user-id", userID, "User identity that owns the session")
	flags.StringVar(&sessionID, "session-id", sessionID, "Session ID to delete")

	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse sessions delete flags: %w", err)
	}
	if len(flags.Args()) > 0 {
		return fmt.Errorf("cannot parse following arguments: %v", flags.Args())
	}
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("--user-id is required")
	}
	if strings.TrimSpace(sessionID) == "" {
		return fmt.Errorf("--session-id is required")
	}

	return runSessionDelete(ctx, sessionService, sessionDeleteOptions{UserID: userID, SessionID: sessionID}, stdout)
}

func runSessionDelete(ctx context.Context, sessionService session.Service, options sessionDeleteOptions, stdout io.Writer) error {
	userID := strings.TrimSpace(options.UserID)
	sessionID := strings.TrimSpace(options.SessionID)
	if userID == "" {
		return fmt.Errorf("--user-id is required")
	}
	if sessionID == "" {
		return fmt.Errorf("--session-id is required")
	}

	if _, _, err := getSession(ctx, sessionService, userID, sessionID); err != nil {
		return err
	}

	if err := sessionService.Delete(ctx, &session.DeleteRequest{AppName: agentAppName, UserID: userID, SessionID: sessionID}); err != nil {
		return fmt.Errorf("delete session %s: %w", sessionID, err)
	}
	_, err := fmt.Fprintf(stdout, "Deleted session %s for user %s.\n", sessionID, userID)
	return err
}

func listSessionSummaries(ctx context.Context, sessionService session.Service, userID string) ([]sessionSummary, error) {
	list, err := sessionService.List(ctx, &session.ListRequest{AppName: agentAppName, UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	summaries := make([]sessionSummary, 0, len(list.Sessions))
	for _, listedSession := range list.Sessions {
		fullSession, _, err := getSession(ctx, sessionService, userID, listedSession.ID())
		if err != nil {
			return nil, err
		}

		events := collectEvents(fullSession.Events())
		summaries = append(summaries, sessionSummary{
			ID:          fullSession.ID(),
			UpdatedAt:   formatSessionTime(fullSession.LastUpdateTime()),
			EventCount:  len(events),
			LastPreview: lastPreview(events),
		})
	}

	sort.SliceStable(summaries, func(i, j int) bool {
		if summaries[i].UpdatedAt != summaries[j].UpdatedAt {
			return summaries[i].UpdatedAt > summaries[j].UpdatedAt
		}
		return summaries[i].ID < summaries[j].ID
	})

	return summaries, nil
}

func collectEvents(events session.Events) []*session.Event {
	collected := make([]*session.Event, 0, events.Len())
	for event := range events.All() {
		collected = append(collected, event)
	}
	return collected
}

func lastPreview(events []*session.Event) string {
	for index := len(events) - 1; index >= 0; index-- {
		text := eventDisplayText(events[index])
		if strings.TrimSpace(text) == "" {
			continue
		}
		return truncateSessionText(fmt.Sprintf("%s: %s", eventAuthor(events[index]), text), 120)
	}
	return ""
}

func eventDisplayText(event *session.Event) string {
	if event == nil || event.LLMResponse.Content == nil {
		return "[no content]"
	}

	parts := make([]string, 0, len(event.LLMResponse.Content.Parts))
	for _, part := range event.LLMResponse.Content.Parts {
		if part == nil {
			continue
		}
		switch {
		case strings.TrimSpace(part.Text) != "":
			parts = append(parts, normalizeSessionText(part.Text))
		case part.FunctionCall != nil:
			name := strings.TrimSpace(part.FunctionCall.Name)
			if name == "" {
				name = "unknown"
			}
			parts = append(parts, fmt.Sprintf("tool call %s", name))
		case part.FunctionResponse != nil:
			name := strings.TrimSpace(part.FunctionResponse.Name)
			if name == "" {
				name = "unknown"
			}
			parts = append(parts, fmt.Sprintf("tool response %s", name))
		case part.FileData != nil:
			parts = append(parts, fmt.Sprintf("file %s", strings.TrimSpace(part.FileData.FileURI)))
		case part.InlineData != nil:
			parts = append(parts, fmt.Sprintf("binary data %s", strings.TrimSpace(part.InlineData.MIMEType)))
		case part.ExecutableCode != nil:
			parts = append(parts, "generated code")
		case part.CodeExecutionResult != nil:
			parts = append(parts, "code execution result")
		}
	}

	if len(parts) == 0 {
		return "[no displayable content]"
	}
	return strings.Join(parts, " | ")
}

func eventAuthor(event *session.Event) string {
	if event == nil || strings.TrimSpace(event.Author) == "" {
		return "unknown"
	}
	return strings.TrimSpace(event.Author)
}

func formatSessionTime(value interface{ Format(string) string }) string {
	return value.Format("2006-01-02T15:04:05Z07:00")
}

func normalizeSessionText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func truncateSessionText(value string, limit int) string {
	if limit <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return strings.TrimSpace(string(runes[:limit])) + "..."
}
