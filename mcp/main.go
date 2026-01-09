// MCP server for OpenCode Canvas - allows AI assistants to query TUIs
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlqattanDev/opencode-canvas/canvas"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"opencode-canvas",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register all canvas tools
	registerTools(s)

	// Start stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func registerTools(s *server.MCPServer) {
	// canvas_list - List active canvases
	s.AddTool(
		mcp.NewTool("canvas_list",
			mcp.WithDescription("List all active canvas TUIs. Returns canvas IDs and their status (alive/dead)."),
		),
		handleList,
	)

	// canvas_ping - Check if canvas is responsive
	s.AddTool(
		mcp.NewTool("canvas_ping",
			mcp.WithDescription("Check if a canvas TUI is responsive and accepting connections."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Canvas ID to ping"),
			),
		),
		handlePing,
	)

	// canvas_state - Get canvas state as JSON
	s.AddTool(
		mcp.NewTool("canvas_state",
			mcp.WithDescription("Get the internal state of a canvas TUI as JSON. Includes mode, cursor position, custom state, and any other state the TUI exposes."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Canvas ID to query"),
			),
		),
		handleState,
	)

	// canvas_view - Get rendered view
	s.AddTool(
		mcp.NewTool("canvas_view",
			mcp.WithDescription("Get the current rendered view of a canvas TUI. Returns the terminal output as the user would see it (may include ANSI codes)."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Canvas ID to query"),
			),
		),
		handleView,
	)

	// canvas_key - Send a key press
	s.AddTool(
		mcp.NewTool("canvas_key",
			mcp.WithDescription("Send a key press to a canvas TUI. Supported keys: enter, tab, space, backspace, delete, escape, up, down, left, right, home, end, pageup, pagedown, ctrl+c, ctrl+d, ctrl+z, or any single character."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Canvas ID to send key to"),
			),
			mcp.WithString("key",
				mcp.Required(),
				mcp.Description("Key to send (e.g., 'enter', 'tab', 'ctrl+c', 'a')"),
			),
		),
		handleKey,
	)

	// canvas_input - Send text input
	s.AddTool(
		mcp.NewTool("canvas_input",
			mcp.WithDescription("Send text input to a canvas TUI. The text is sent as if the user typed it."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Canvas ID to send input to"),
			),
			mcp.WithString("text",
				mcp.Required(),
				mcp.Description("Text to send"),
			),
		),
		handleInput,
	)

	// canvas_close - Request canvas to close
	s.AddTool(
		mcp.NewTool("canvas_close",
			mcp.WithDescription("Request a canvas TUI to close gracefully."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Canvas ID to close"),
			),
		),
		handleClose,
	)
}

// Tool handlers

func handleList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	socketDir := canvas.DefaultSocketDir()
	entries, err := os.ReadDir(socketDir)
	if err != nil {
		if os.IsNotExist(err) {
			return mcp.NewToolResultText("No active canvases"), nil
		}
		return nil, fmt.Errorf("failed to read socket directory: %w", err)
	}

	type canvasInfo struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}

	var canvases []canvasInfo
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".sock") {
			id := strings.TrimSuffix(entry.Name(), ".sock")
			client := canvas.NewClient(id)
			status := "dead"
			if client.Ping() {
				status = "alive"
			}
			canvases = append(canvases, canvasInfo{ID: id, Status: status})
		}
	}

	if len(canvases) == 0 {
		return mcp.NewToolResultText("No active canvases"), nil
	}

	data, _ := json.MarshalIndent(canvases, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handlePing(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := request.RequireString("id")
	if err != nil {
		return nil, err
	}

	client := canvas.NewClient(id)
	if client.Ping() {
		return mcp.NewToolResultText(fmt.Sprintf("Canvas '%s' is alive and responsive", id)), nil
	}

	// Check if socket exists but not responding
	socketPath := filepath.Join(canvas.DefaultSocketDir(), id+".sock")
	if _, err := os.Stat(socketPath); err == nil {
		return mcp.NewToolResultText(fmt.Sprintf("Canvas '%s' socket exists but is not responding", id)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Canvas '%s' not found", id)), nil
}

func handleState(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := request.RequireString("id")
	if err != nil {
		return nil, err
	}

	client := canvas.NewClient(id)
	state, err := client.GetState()
	if err != nil {
		return nil, fmt.Errorf("failed to get state from canvas '%s': %w", id, err)
	}

	data, _ := json.MarshalIndent(state, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleView(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := request.RequireString("id")
	if err != nil {
		return nil, err
	}

	client := canvas.NewClient(id)
	view, err := client.GetView()
	if err != nil {
		return nil, fmt.Errorf("failed to get view from canvas '%s': %w", id, err)
	}

	return mcp.NewToolResultText(view), nil
}

func handleKey(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := request.RequireString("id")
	if err != nil {
		return nil, err
	}

	key, err := request.RequireString("key")
	if err != nil {
		return nil, err
	}

	client := canvas.NewClient(id)
	if err := client.SendKey(key); err != nil {
		return nil, fmt.Errorf("failed to send key to canvas '%s': %w", id, err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Sent key '%s' to canvas '%s'", key, id)), nil
}

func handleInput(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := request.RequireString("id")
	if err != nil {
		return nil, err
	}

	text, err := request.RequireString("text")
	if err != nil {
		return nil, err
	}

	client := canvas.NewClient(id)
	if err := client.SendInput(text); err != nil {
		return nil, fmt.Errorf("failed to send input to canvas '%s': %w", id, err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Sent input to canvas '%s': %s", id, text)), nil
}

func handleClose(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := request.RequireString("id")
	if err != nil {
		return nil, err
	}

	client := canvas.NewClient(id)
	if err := client.Close(); err != nil {
		return nil, fmt.Errorf("failed to close canvas '%s': %w", id, err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Requested canvas '%s' to close", id)), nil
}
