package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"

	"github.com/kriuchkov/jiraforge/internal/app"
	appconfig "github.com/kriuchkov/jiraforge/internal/config"
)

type rootOptions struct {
	envFile    string
	modelName  string
	mcpCommand string
	mcpArgs    string
	sessionDB  string
}

func newRootCommand(stdout, stderr io.Writer) *cobra.Command {
	options := &rootOptions{}

	root := &cobra.Command{
		Use:           "jiraforge-agent",
		Short:         "Gemini Jira agent powered by JiraForge MCP",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.SetOut(stdout)
	root.SetErr(stderr)

	flags := root.PersistentFlags()
	flags.StringVar(&options.envFile, "env", "", "Path to environment file")
	flags.StringVar(&options.modelName, "model", appconfig.DefaultGeminiModel, "Gemini model to use")
	flags.StringVar(&options.mcpCommand, "mcp-command", "", "Command used to start the local MCP server")
	flags.StringVar(&options.mcpArgs, "mcp-args", "", "Arguments passed to the MCP server command")
	flags.StringVar(&options.sessionDB, "session-db", "", "Path to the persistent session database")

	root.AddCommand(
		newConsoleCommand(options),
		newWebCommand(options),
		newSessionsCommand(options),
	)

	return root
}

func newConsoleCommand(rootOptions *rootOptions) *cobra.Command {
	var userID string
	var sessionID string
	var resumeLast bool
	var streamingMode string

	command := &cobra.Command{
		Use:   "console",
		Short: "Run the interactive console agent",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			mode, err := parseStreamingMode(streamingMode)
			if err != nil {
				return err
			}

			runtime, err := newAgentRuntime(cmd.Context(), agentRuntimeOptions{
				envFile:    rootOptions.envFile,
				modelName:  rootOptions.modelName,
				mcpCommand: rootOptions.mcpCommand,
				mcpArgs:    rootOptions.mcpArgs,
				sessionDB:  rootOptions.sessionDB,
			})
			if err != nil {
				return err
			}

			return runConsoleAgent(cmd.Context(), runtime.jiraAgent, runtime.sessionService, consoleOptions{
				userID:        userID,
				sessionID:     sessionID,
				resumeLast:    resumeLast,
				sessionDB:     runtime.resolvedSessionDB,
				streamingMode: mode,
			})
		},
	}

	flags := command.Flags()
	flags.StringVar(&userID, "user-id", defaultConsoleUserID(), "User identity for persistent agent sessions")
	flags.StringVar(&sessionID, "session-id", "", "Explicit session ID to create or resume")
	flags.BoolVar(&resumeLast, "resume-last", false, "Resume the most recently updated session for the user")
	flags.StringVar(&streamingMode, "streaming-mode", "", fmt.Sprintf("Streaming mode (%s|%s)", agent.StreamingModeNone, agent.StreamingModeSSE))
	command.MarkFlagsMutuallyExclusive("session-id", "resume-last")

	return command
}

func newWebCommand(rootOptions *rootOptions) *cobra.Command {
	command := &cobra.Command{
		Use:   "web [launcher args...]",
		Short: "Run the ADK web launcher",
		Long:  "Run the ADK web launcher. Pass launcher-specific arguments after --.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := newAgentRuntime(cmd.Context(), agentRuntimeOptions{
				envFile:    rootOptions.envFile,
				modelName:  rootOptions.modelName,
				mcpCommand: rootOptions.mcpCommand,
				mcpArgs:    rootOptions.mcpArgs,
				sessionDB:  rootOptions.sessionDB,
			})
			if err != nil {
				return err
			}

			config := &launcher.Config{
				SessionService: runtime.sessionService,
				AgentLoader:    agent.NewSingleLoader(runtime.jiraAgent),
			}

			webLauncher := full.NewLauncher()
			if err := webLauncher.Execute(cmd.Context(), config, append([]string{"web"}, args...)); err != nil && !app.IsContextCanceled(err) {
				return fmt.Errorf("%w\n\n%s", err, webLauncher.CommandLineSyntax())
			}
			return nil
		},
	}

	return command
}

func newSessionsCommand(rootOptions *rootOptions) *cobra.Command {
	var userID string

	command := &cobra.Command{
		Use:           "sessions",
		Short:         "Manage persistent agent sessions",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	command.PersistentFlags().StringVar(&userID, "user-id", defaultConsoleUserID(), "User identity that owns the sessions")

	command.AddCommand(
		newSessionsListCommand(rootOptions, &userID),
		newSessionsInspectCommand(rootOptions, &userID),
		newSessionsDeleteCommand(rootOptions, &userID),
	)

	return command
}

func newSessionsListCommand(rootOptions *rootOptions, userID *string) *cobra.Command {
	var limit int

	command := &cobra.Command{
		Use:   "list",
		Short: "List saved sessions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			sessionService, _, err := newPersistentSessionService(rootOptions.sessionDB)
			if err != nil {
				return err
			}
			return runSessionList(cmd.Context(), sessionService, sessionListOptions{UserID: *userID, Limit: limit}, cmd.OutOrStdout())
		},
	}

	command.Flags().IntVar(&limit, "limit", 20, "Maximum number of sessions to print; use 0 to print all")
	return command
}

func newSessionsInspectCommand(rootOptions *rootOptions, userID *string) *cobra.Command {
	var sessionID string
	var recent int

	command := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect a saved session transcript",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			sessionService, _, err := newPersistentSessionService(rootOptions.sessionDB)
			if err != nil {
				return err
			}
			return runSessionInspect(cmd.Context(), sessionService, sessionInspectOptions{UserID: *userID, SessionID: sessionID, Recent: recent}, cmd.OutOrStdout())
		},
	}

	command.Flags().StringVar(&sessionID, "session-id", "", "Session ID to inspect")
	command.Flags().IntVar(&recent, "recent", 20, "Number of most recent events to print; use 0 to print all")
	_ = command.MarkFlagRequired("session-id")

	return command
}

func newSessionsDeleteCommand(rootOptions *rootOptions, userID *string) *cobra.Command {
	var sessionID string

	command := &cobra.Command{
		Use:   "delete",
		Short: "Delete a saved session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			sessionService, _, err := newPersistentSessionService(rootOptions.sessionDB)
			if err != nil {
				return err
			}
			return runSessionDelete(cmd.Context(), sessionService, sessionDeleteOptions{UserID: *userID, SessionID: sessionID}, cmd.OutOrStdout())
		},
	}

	command.Flags().StringVar(&sessionID, "session-id", "", "Session ID to delete")
	_ = command.MarkFlagRequired("session-id")

	return command
}
