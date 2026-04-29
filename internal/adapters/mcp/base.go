package mcp

import (
	"strings"

	"github.com/mark3labs/mcp-go/server"

	"github.com/kriuchkov/jiraforge/internal/core/ports"
)

func NewServer(name, version string, service ports.JiraService) *server.MCPServer {
	mcpServer := server.NewMCPServer(
		name,
		version,
		server.WithLogging(),
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
		server.WithResourceCapabilities(true, true),
		server.WithRecovery(),
	)

	registerIssueTools(mcpServer, service)
	registerSearchTools(mcpServer, service)
	registerReportTools(mcpServer, service)
	registerSprintTools(mcpServer, service)
	registerCommentTools(mcpServer, service)
	registerWorklogTools(mcpServer, service)
	registerTransitionTools(mcpServer, service)
	registerStatusTools(mcpServer, service)
	registerHistoryTools(mcpServer, service)
	registerRelationshipTools(mcpServer, service)
	registerVersionTools(mcpServer, service)
	registerDevelopmentTools(mcpServer, service)
	registerAttachmentTools(mcpServer, service)
	registerPrompts(mcpServer)

	return mcpServer
}

func splitCSV(value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
