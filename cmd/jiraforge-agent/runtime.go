package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/mcptoolset"
	"google.golang.org/genai"

	adkadapter "github.com/kriuchkov/jiraforge/internal/adapters/adk"
	appconfig "github.com/kriuchkov/jiraforge/internal/config"
)

type agentRuntimeOptions struct {
	envFile    string
	modelName  string
	mcpCommand string
	mcpArgs    string
	sessionDB  string
}

type agentRuntime struct {
	sessionService    session.Service
	resolvedSessionDB string
	jiraAgent         agent.Agent
}

func newAgentRuntime(ctx context.Context, options agentRuntimeOptions) (*agentRuntime, error) {
	sessionService, resolvedSessionDB, err := newPersistentSessionService(options.sessionDB)
	if err != nil {
		return nil, fmt.Errorf("create persistent session service: %w", err)
	}

	settings, err := appconfig.LoadSettingsFromEnvFile(options.envFile)
	if err != nil {
		return nil, fmt.Errorf("load env file: %w", err)
	}
	if err := settings.Gemini.Validate(); err != nil {
		return nil, fmt.Errorf("configure Gemini: %w", err)
	}
	if err := settings.Atlassian.Validate(); err != nil {
		return nil, fmt.Errorf("configure Atlassian: %w", err)
	}

	model, err := gemini.NewModel(ctx, firstNonEmpty(options.modelName, settings.Gemini.Model), &genai.ClientConfig{
		APIKey: settings.Gemini.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create Gemini model: %w", err)
	}

	commandName, commandArgs := resolveMCPCommand(options.mcpCommand, options.mcpArgs)
	mcpToolSet, err := mcptoolset.New(mcptoolset.Config{
		Transport:                   &mcp.CommandTransport{Command: exec.Command(commandName, commandArgs...)},
		RequireConfirmationProvider: requireConfirmation,
	})
	if err != nil {
		return nil, fmt.Errorf("create MCP tool set: %w", err)
	}

	memorySearch := adkadapter.NewSessionMemorySearch(sessionService)
	preloadMemoryTool := adkadapter.NewPreloadMemoryTool(memorySearch)
	loadMemoryTool, err := adkadapter.NewLoadMemoryTool(memorySearch)
	if err != nil {
		return nil, fmt.Errorf("create memory tool: %w", err)
	}

	jiraAgent, err := llmagent.New(llmagent.Config{
		Name:        "jiraforge_agent",
		Model:       model,
		Description: "Gemini agent for Jira workflows through the local JiraForge MCP server.",
		Instruction: `You are a JiraForge operations agent. Use the available Jira MCP tools to inspect issues, comments, sprints, versions, statuses, linked work, and development information. Relevant snippets from previous sessions may be preloaded automatically. If earlier conversations may matter and the preloaded context is insufficient, call load_memory. Prefer read-only tools first. For write operations, explain what you are changing and wait for confirmation if the runtime requests it. Keep outputs concise, accurate, and action-oriented.`,
		Tools: []tool.Tool{
			preloadMemoryTool,
			loadMemoryTool,
		},
		Toolsets: []tool.Toolset{
			mcpToolSet,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create Jira agent: %w", err)
	}

	return &agentRuntime{
		sessionService:    sessionService,
		resolvedSessionDB: resolvedSessionDB,
		jiraAgent:         jiraAgent,
	}, nil
}

func requireConfirmation(name string, _ any) bool {
	return isMutatingTool(name)
}

func isMutatingTool(name string) bool {
	switch strings.TrimSpace(name) {
	case "jira_create_issue",
		"jira_create_child_issue",
		"jira_update_issue",
		"jira_delete_issue",
		"jira_add_comment",
		"jira_add_worklog",
		"jira_transition_issue",
		"jira_link_issues":
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return appconfig.DefaultGeminiModel
}

func splitArgs(value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return strings.Fields(trimmed)
}

func resolveMCPCommand(commandFlag, argsFlag string) (string, []string) {
	if strings.TrimSpace(commandFlag) != "" {
		return commandFlag, splitArgs(argsFlag)
	}

	if executable, err := os.Executable(); err == nil {
		directory := filepath.Dir(executable)
		for _, candidate := range []string{"jiraforge", "jiraforge-mcp"} {
			path := filepath.Join(directory, candidate)
			if info, statErr := os.Stat(path); statErr == nil && !info.IsDir() {
				return path, nil
			}
		}
	}
	return "go", []string{"run", "."}
}
