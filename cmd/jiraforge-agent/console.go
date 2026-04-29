package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/glebarez/sqlite"
	"google.golang.org/genai"

	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	sessiondb "google.golang.org/adk/session/database"
)

const agentAppName = "jiraforge_agent"

type consoleOptions struct {
	userID        string
	sessionID     string
	resumeLast    bool
	sessionDB     string
	streamingMode adkagent.StreamingMode
}

func runConsoleAgent(ctx context.Context, rootAgent adkagent.Agent, sessionService session.Service, options consoleOptions) error {
	if rootAgent == nil {
		return fmt.Errorf("root agent is required")
	}
	if sessionService == nil {
		return fmt.Errorf("session service is required")
	}
	if strings.TrimSpace(options.userID) == "" {
		return fmt.Errorf("user id is required")
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	currentSession, created, err := resolveConsoleSession(ctx, sessionService, options.userID, options.sessionID, options.resumeLast)
	if err != nil {
		return err
	}

	r, err := runner.New(runner.Config{
		AppName:        agentAppName,
		Agent:          rootAgent,
		SessionService: sessionService,
	})
	if err != nil {
		return fmt.Errorf("create runner: %w", err)
	}

	if created {
		fmt.Printf("Started new session %s for user %s\n", currentSession.ID(), options.userID)
	} else {
		fmt.Printf("Resumed session %s for user %s\n", currentSession.ID(), options.userID)
	}
	if strings.TrimSpace(options.sessionDB) != "" {
		fmt.Printf("Session DB: %s\n", options.sessionDB)
	}

	inputChan := make(chan string)
	readErrChan := make(chan error, 1)

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			userInput, err := reader.ReadString('\n')
			if err != nil {
				readErrChan <- err
				return
			}
			inputChan <- userInput
		}
	}()

	defaultStreamingMode := options.streamingMode
	if defaultStreamingMode == "" {
		if info, err := os.Stdout.Stat(); err == nil && (info.Mode()&os.ModeCharDevice) != 0 {
			defaultStreamingMode = adkagent.StreamingModeSSE
		} else {
			defaultStreamingMode = adkagent.StreamingModeNone
		}
	}

	fmt.Println()
	fmt.Print("User -> ")

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-readErrChan:
			if errors.Is(err, io.EOF) {
				fmt.Println("\nEOF detected, exiting...")
				return nil
			}
			return fmt.Errorf("read stdin: %w", err)
		case userInput := <-inputChan:
			if strings.TrimSpace(userInput) == "" {
				fmt.Print("User -> ")
				continue
			}

			userMsg := genai.NewContentFromText(userInput, genai.RoleUser)
			fmt.Print("\nAgent -> ")

			captured := ""
			for event, err := range r.Run(ctx, options.userID, currentSession.ID(), userMsg, adkagent.RunConfig{
				StreamingMode: defaultStreamingMode,
			}) {
				if err != nil {
					fmt.Printf("\nAGENT_ERROR: %v\n", err)
					continue
				}
				if event == nil || event.LLMResponse.Content == nil {
					continue
				}

				text := contentText(event.LLMResponse.Content)
				if defaultStreamingMode != adkagent.StreamingModeSSE {
					fmt.Print(text)
					continue
				}

				if !event.IsFinalResponse() {
					fmt.Print(text)
					captured += text
					continue
				}

				if text != captured {
					fmt.Print(text)
				}
				captured = ""
			}
			fmt.Print("\nUser -> ")
		}
	}
}

func parseStreamingMode(value string) (adkagent.StreamingMode, error) {
	mode := adkagent.StreamingMode(strings.TrimSpace(value))
	if mode == "" || mode == adkagent.StreamingModeNone || mode == adkagent.StreamingModeSSE {
		return mode, nil
	}
	return "", fmt.Errorf("invalid streaming mode %q; supported values are %s or %s", value, adkagent.StreamingModeNone, adkagent.StreamingModeSSE)
}

func newPersistentSessionService(path string) (session.Service, string, error) {
	resolvedPath := resolveSessionDBPath(path)
	if err := os.MkdirAll(filepath.Dir(resolvedPath), 0o755); err != nil {
		return nil, "", fmt.Errorf("create session db directory: %w", err)
	}

	service, err := sessiondb.NewSessionService(sqlite.Open(resolvedPath))
	if err != nil {
		return nil, "", fmt.Errorf("create session service: %w", err)
	}
	if err := sessiondb.AutoMigrate(service); err != nil {
		return nil, "", fmt.Errorf("migrate session service: %w", err)
	}
	return service, resolvedPath, nil
}

func resolveConsoleSession(ctx context.Context, sessionService session.Service, userID, sessionID string, resumeLast bool) (session.Session, bool, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, false, fmt.Errorf("user id is required")
	}
	if strings.TrimSpace(sessionID) != "" && resumeLast {
		return nil, false, fmt.Errorf("--session-id and --resume-last cannot be used together")
	}

	list, err := sessionService.List(ctx, &session.ListRequest{AppName: agentAppName, UserID: userID})
	if err != nil {
		return nil, false, fmt.Errorf("list sessions: %w", err)
	}

	trimmedSessionID := strings.TrimSpace(sessionID)
	if trimmedSessionID != "" {
		for _, existing := range list.Sessions {
			if existing.ID() == trimmedSessionID {
				return getSession(ctx, sessionService, userID, trimmedSessionID)
			}
		}
		return createSession(ctx, sessionService, userID, trimmedSessionID)
	}

	if resumeLast {
		latest := latestSession(list.Sessions)
		if latest != nil {
			return getSession(ctx, sessionService, userID, latest.ID())
		}
	}

	return createSession(ctx, sessionService, userID, "")
}

func createSession(ctx context.Context, sessionService session.Service, userID, sessionID string) (session.Session, bool, error) {
	resp, err := sessionService.Create(ctx, &session.CreateRequest{AppName: agentAppName, UserID: userID, SessionID: sessionID})
	if err != nil {
		return nil, false, fmt.Errorf("create session: %w", err)
	}
	return resp.Session, true, nil
}

func getSession(ctx context.Context, sessionService session.Service, userID, sessionID string) (session.Session, bool, error) {
	resp, err := sessionService.Get(ctx, &session.GetRequest{AppName: agentAppName, UserID: userID, SessionID: sessionID})
	if err != nil {
		return nil, false, fmt.Errorf("get session %s: %w", sessionID, err)
	}
	return resp.Session, false, nil
}

func latestSession(sessions []session.Session) session.Session {
	var latest session.Session
	for _, current := range sessions {
		if latest == nil {
			latest = current
			continue
		}
		if current.LastUpdateTime().After(latest.LastUpdateTime()) {
			latest = current
			continue
		}
		if current.LastUpdateTime().Equal(latest.LastUpdateTime()) && current.ID() > latest.ID() {
			latest = current
		}
	}
	return latest
}

func resolveSessionDBPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed != "" {
		return filepath.Clean(trimmed)
	}

	configDir, err := os.UserConfigDir()
	if err != nil || strings.TrimSpace(configDir) == "" {
		return filepath.Join(os.TempDir(), "jiraforge", "agent-sessions.db")
	}
	return filepath.Join(configDir, "jiraforge", "agent-sessions.db")
}

func defaultConsoleUserID() string {
	for _, key := range []string{"JIRAFORGE_AGENT_USER_ID", "USER", "USERNAME"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return "default"
}

func contentText(content *genai.Content) string {
	if content == nil {
		return ""
	}
	parts := make([]string, 0, len(content.Parts))
	for _, part := range content.Parts {
		if part == nil || part.Text == "" {
			continue
		}
		parts = append(parts, part.Text)
	}
	return strings.Join(parts, "")
}
