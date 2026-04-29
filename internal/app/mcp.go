package app

import (
	"context"
	"io"
	"strings"

	"github.com/go-faster/errors"

	mcpadapter "github.com/kriuchkov/jiraforge/internal/adapters/mcp"

	"github.com/mark3labs/mcp-go/server"
)

func ServeMCP(envFile, httpPort string, _ io.Writer) error {
	service, settings, err := NewJiraServiceFromEnvFile(envFile)
	if err != nil {
		return errors.Wrap(err, "create Jira service")
	}

	mcpServer := mcpadapter.NewServer(settings.App.Name, settings.App.Version, service)
	if strings.TrimSpace(httpPort) != "" {
		httpServer := server.NewStreamableHTTPServer(mcpServer, server.WithEndpointPath("/mcp"))
		if err := httpServer.Start(":" + httpPort); err != nil {
			return errors.Wrap(err, "start MCP HTTP server")
		}
		return nil
	}

	if err := server.ServeStdio(mcpServer); err != nil {
		return errors.Wrap(err, "serve MCP stdio")
	}
	return nil
}

func IsContextCanceled(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return true
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "context canceled") ||
		strings.Contains(errMsg, "operation was canceled") ||
		strings.Contains(errMsg, "context deadline exceeded")
}
