package mcp

import (
	"context"
	"encoding/json"
	"testing"

	markmcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/require"
)

func newTestServer() *server.MCPServer {
	return server.NewMCPServer(
		"test-server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
		server.WithResourceCapabilities(true, true),
		server.WithRecovery(),
	)
}

func mustGetTool(t *testing.T, srv *server.MCPServer, name string) *server.ServerTool {
	t.Helper()

	tool := srv.GetTool(name)
	require.NotNil(t, tool)
	return tool
}

func callTool(t *testing.T, srv *server.MCPServer, name string, args map[string]any) *markmcp.CallToolResult {
	t.Helper()

	tool := mustGetTool(t, srv, name)
	result, err := tool.Handler(context.Background(), markmcp.CallToolRequest{
		Params: markmcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	return result
}

func toolText(t *testing.T, result *markmcp.CallToolResult) string {
	t.Helper()

	require.NotEmpty(t, result.Content)
	content, ok := markmcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	return content.Text
}

func handleMessage(t *testing.T, srv *server.MCPServer, message any) markmcp.JSONRPCMessage {
	t.Helper()

	payload, err := json.Marshal(message)
	require.NoError(t, err)
	return srv.HandleMessage(context.Background(), payload)
}

func initializeServer(t *testing.T, srv *server.MCPServer) markmcp.InitializeResult {
	t.Helper()

	response := handleMessage(t, srv, markmcp.JSONRPCRequest{
		JSONRPC: markmcp.JSONRPC_VERSION,
		ID:      markmcp.NewRequestId(int64(1)),
		Request: markmcp.Request{Method: string(markmcp.MethodInitialize)},
	})

	resp, ok := response.(markmcp.JSONRPCResponse)
	require.True(t, ok)
	result, ok := resp.Result.(markmcp.InitializeResult)
	require.True(t, ok)
	return result
}

func listPrompts(t *testing.T, srv *server.MCPServer) markmcp.ListPromptsResult {
	t.Helper()

	response := handleMessage(t, srv, map[string]any{
		"jsonrpc": markmcp.JSONRPC_VERSION,
		"id":      1,
		"method":  string(markmcp.MethodPromptsList),
	})

	resp, ok := response.(markmcp.JSONRPCResponse)
	require.True(t, ok)
	result, ok := resp.Result.(markmcp.ListPromptsResult)
	require.True(t, ok)
	return result
}

func getPrompt(t *testing.T, srv *server.MCPServer, name string, args map[string]string) markmcp.GetPromptResult {
	t.Helper()

	response := handleMessage(t, srv, map[string]any{
		"jsonrpc": markmcp.JSONRPC_VERSION,
		"id":      1,
		"method":  string(markmcp.MethodPromptsGet),
		"params": map[string]any{
			"name":      name,
			"arguments": args,
		},
	})

	resp, ok := response.(markmcp.JSONRPCResponse)
	require.True(t, ok)
	result, ok := resp.Result.(markmcp.GetPromptResult)
	require.True(t, ok)
	return result
}

func promptText(t *testing.T, result markmcp.GetPromptResult) string {
	t.Helper()

	require.NotEmpty(t, result.Messages)
	content, ok := result.Messages[0].Content.(markmcp.TextContent)
	require.True(t, ok)
	return content.Text
}
