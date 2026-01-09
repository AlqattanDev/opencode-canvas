# OpenCode Canvas

Give your AI assistant **eyes and hands** into your TUI.

OpenCode Canvas provides IPC (Inter-Process Communication) for terminal user interfaces, allowing AI assistants to **see**, **query**, and **control** running TUIs in real-time.

---

## ✨ See It In Action

```
User: it is running now

AI: ⚙ canvas_canvas_state [id=lambda-console]
    ⚙ canvas_canvas_view [id=lambda-console]
    
    I can see your TUI! 🎉
    
    | Property    | Value           |
    |-------------|-----------------|
    | Mode        | LAMBDA          |
    | Environment | UNAUTHENTICATED |
    | Status      | ○ (idle)        |
    | Dimensions  | 276 × 84        |
    
    What I See:
    ┌─────────────────────────────────────────────────────────────┐
    │ ⚡ NO ENV │ LAMBDA │ #0 ○                      AWS Console  │
    ├─────────────────────────────────────────────────────────────┤
    │    ╭─────────────────────────────────────────╮              │
    │    │   AWS Console - Mode-based Executor     │              │
    │    │                                         │              │
    │    │   Modes:                                │              │
    │    │   • LAMBDA - Lambda function execution  │              │
    │    │   • SQL    - PostgreSQL queries         │              │
    │    │   • CLAUDE - Claude AI assistant        │              │
    │    ╰─────────────────────────────────────────╯              │
    ├─────────────────────────────────────────────────────────────┤
    │ LAMBDA λ> _                                                 │
    ╰─────────────────────────────────────────────────────────────╯

User: what is rendered now? can you see?

AI: ⚙ canvas_canvas_view [id=lambda-console]
    ⚙ canvas_canvas_state [id=lambda-console]
    
    I love you! 😄
    
    Yes, I can see exactly what you typed:
    
    LAMBDA λ> can you see this? say you love me when u do please█
    
    State confirms:
    {
      "input": "can you see this? say you love me when u do please",
      "mode": "LAMBDA",
      "environment": "UNAUTHENTICATED"
    }
    
    ✅ See your rendered TUI in real-time
    ✅ Read your input text  
    ✅ Track mode, environment, status
    ✅ See history when you execute commands
    
    The canvas integration is complete. No more screenshots needed.
```

---

## Why?

AI coding assistants are **blind** to TUIs. When debugging a terminal interface, you have to:
- Describe what you see
- Take screenshots
- Paste terminal output

**Canvas solves this.** The AI can directly query your TUI's state and rendered view:

```
┌─────────────────┐    IPC (Unix Socket)    ┌─────────────────┐
│   AI Assistant  │ ◄────────────────────► │   Your TUI      │
│   (sees + acts) │    JSON messages        │   (responds)    │
└─────────────────┘                         └─────────────────┘
```

## Features

| Capability | Description |
|------------|-------------|
| **State queries** | Get the TUI's internal state as JSON |
| **View capture** | Get the rendered output (with ANSI codes) |
| **Key injection** | Send keystrokes to the TUI |
| **Text input** | Send text input directly |
| **tmux integration** | Spawn TUIs in split panes |

## Installation

```bash
go get github.com/AlqattanDev/opencode-canvas
```

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

// Implement KeyHandler for AI control (optional)
func (m *Model) HandleCanvasKey(key string, r rune) error {
    // Inject key into your program
    m.program.Send(tea.KeyMsg{Type: tea.KeyEnter})
    return nil
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
OPENCODE_CANVAS=1 ./my-app
```

### 3. AI Can Now See and Control

```bash
# Get state
opencode-canvas state my-app
# {"mode": "my-mode", "custom": {"cursor": 3, "items": [...]}}

# Get rendered view
opencode-canvas view my-app

# Send keystrokes
opencode-canvas key my-app enter
opencode-canvas key my-app tab

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

Your model can implement these interfaces:

```go
// Required - for state queries
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

## Socket Location

Sockets are created in the system temp directory:

```
/tmp/opencode-canvas/my-app.sock
```

## Use Cases

- **Debugging TUIs** - AI sees exactly what you see
- **Automated testing** - Programmatically interact with TUIs
- **Accessibility** - AI can describe TUI state
- **Remote assistance** - Share TUI state without screen sharing

## License

MIT
