# OpenCode Canvas

Give your AI assistant eyes into your TUI. 

OpenCode Canvas provides IPC (Inter-Process Communication) for terminal user interfaces, allowing AI assistants like Claude, GPT, or any OpenCode-compatible tool to query and interact with running TUIs.

## Why?

AI coding assistants are blind to TUIs. When you're debugging a terminal interface, you have to describe what you see, take screenshots, or paste terminal output. Canvas solves this by letting the AI directly query your TUI's state.

```
┌─────────────────┐    IPC (Unix Socket)    ┌─────────────────┐
│   AI Assistant  │ ◄────────────────────► │   Your TUI      │
│   (queries)     │    JSON messages        │   (responds)    │
└─────────────────┘                         └─────────────────┘
```

## Features

- **State queries** - Get the TUI's internal state as JSON
- **View capture** - Get the rendered output (with ANSI codes)
- **Key injection** - Send keystrokes to the TUI
- **Text input** - Send text input directly
- **tmux integration** - Spawn TUIs in split panes

## Installation

```bash
# Install CLI tool
go install github.com/AlqattanDev/opencode-canvas/cli@latest

# Install MCP server (for OpenCode integration)
go install github.com/AlqattanDev/opencode-canvas/mcp@latest
```

Or clone and build:

```bash
git clone https://github.com/AlqattanDev/opencode-canvas
cd opencode-canvas
go build -o opencode-canvas ./cli
go build -o opencode-canvas-mcp ./mcp
```

## OpenCode Integration

Add to your OpenCode config (`.opencode/config.json` or `~/.config/opencode/config.json`):

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "canvas": {
      "type": "local",
      "command": ["opencode-canvas-mcp"],
      "enabled": true
    }
  }
}
```

This gives your AI assistant these tools:
- `canvas_list` - List all active canvas TUIs
- `canvas_state` - Get TUI state as JSON
- `canvas_view` - Get rendered terminal output
- `canvas_key` - Send keystrokes
- `canvas_input` - Send text input
- `canvas_ping` - Check if TUI is responsive
- `canvas_close` - Request TUI to close

## Quick Start

### 1. Add Canvas Support to Your Bubble Tea App

```go
import (
    "github.com/AlqattanDev/opencode-canvas/canvas"
    tea "github.com/charmbracelet/bubbletea"
)

// Implement StateProvider for your model
func (m Model) CanvasState() canvas.StatePayload {
    return canvas.StatePayload{
        Mode: "my-mode",
        Custom: map[string]any{
            "cursor": m.cursor,
            "items":  m.items,
        },
    }
}

func main() {
    m := NewModel()
    
    // Wrap your model with canvas support
    wrapped := canvas.Wrap("my-app", m)
    
    p := tea.NewProgram(wrapped)
    p.Run()
}
```

### 2. Run Your TUI with Canvas Enabled

```bash
# Enable canvas mode
OPENCODE_CANVAS=1 ./my-app

# Or spawn in tmux
opencode-canvas spawn my-app ./my-app
```

### 3. Query from AI/CLI

```bash
# Get state
opencode-canvas state my-app
# {"mode": "my-mode", "custom": {"cursor": 3, "items": [...]}}

# Get rendered view
opencode-canvas view my-app

# Send keystrokes
opencode-canvas key my-app enter
opencode-canvas key my-app ctrl+c

# Send text input
opencode-canvas input my-app "hello world"
```

## CLI Reference

```
opencode-canvas <command> [arguments]

COMMANDS:
    state <id>              Get canvas state as JSON
    view <id>               Get rendered view (with ANSI codes)
    key <id> <key>          Send a key press (e.g., "enter", "tab", "ctrl+c")
    input <id> <text>       Send text input
    close <id>              Request canvas to close
    list                    List active canvases
    spawn <id> <cmd...>     Spawn a command as a canvas in tmux
    ping <id>               Check if canvas is responsive
```

## Protocol

Canvas uses a simple JSON protocol over Unix domain sockets:

### Queries (AI → TUI)

```json
{"type": "get_state"}
{"type": "get_view"}
{"type": "send_key", "payload": {"key": "enter"}}
{"type": "send_input", "payload": {"text": "hello"}}
{"type": "close"}
```

### Responses (TUI → AI)

```json
{"type": "state", "payload": {"mode": "...", "custom": {...}}}
{"type": "view", "payload": {"content": "...", "ansi": true}}
{"type": "ack"}
{"type": "error", "payload": {"code": "...", "message": "..."}}
```

## Interfaces

Your model can implement these interfaces for full functionality:

```go
// Required for state queries
type StateProvider interface {
    CanvasState() StatePayload
}

// Optional - defaults to View()
type ViewProvider interface {
    CanvasView() string
}

// Optional - for key injection
type KeyHandler interface {
    HandleCanvasKey(key string, r rune) error
}

// Optional - for text input
type InputHandler interface {
    HandleCanvasInput(text string) error
}
```

## Example

See [examples/bubbletea](./examples/bubbletea) for a complete example.

## Socket Location

Sockets are created in `/tmp/opencode-canvas/`:

```
/tmp/opencode-canvas/my-app.sock
/tmp/opencode-canvas/my-app.pane  (tmux pane ID)
```

## Integrating with AI Assistants

For AI assistants to use canvas, they can:

1. Check if a canvas is running: `opencode-canvas ping <id>`
2. Query state: `opencode-canvas state <id>`
3. Get view: `opencode-canvas view <id>`
4. Interact: `opencode-canvas key <id> <key>`

This allows the AI to "see" and interact with your TUI without screenshots or descriptions.

## License

MIT
