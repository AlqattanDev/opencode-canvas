package canvas

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// BubbleTeaAdapter wraps a Bubble Tea model with canvas IPC support
type BubbleTeaAdapter struct {
	server *Server
	model  tea.Model
}

// Wrap creates a canvas-enabled wrapper around a Bubble Tea model.
// The model should implement StateProvider and/or ViewProvider for full functionality.
//
// Usage:
//
//	func main() {
//	    model := NewMyModel()
//	    wrapped := canvas.Wrap("my-canvas", model)
//	    p := tea.NewProgram(wrapped)
//	    p.Run()
//	}
func Wrap(canvasID string, model tea.Model) tea.Model {
	// Check if running with canvas support
	if os.Getenv("OPENCODE_CANVAS") == "" && os.Getenv("CANVAS_ID") == "" {
		// No canvas mode - return model as-is
		return model
	}

	// Use provided ID or from env
	id := canvasID
	if envID := os.Getenv("CANVAS_ID"); envID != "" {
		id = envID
	}

	server, err := NewServer(id)
	if err != nil {
		// Fall back to non-canvas mode
		return model
	}

	adapter := &BubbleTeaAdapter{
		server: server,
		model:  model,
	}

	server.SetModel(adapter)
	server.Start()

	return adapter
}

// Init implements tea.Model
func (a *BubbleTeaAdapter) Init() tea.Cmd {
	return a.model.Init()
}

// Update implements tea.Model
func (a *BubbleTeaAdapter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.QuitMsg); ok {
		a.server.Stop()
	}

	newModel, cmd := a.model.Update(msg)
	a.model = newModel

	return a, cmd
}

// View implements tea.Model
func (a *BubbleTeaAdapter) View() string {
	return a.model.View()
}

// CanvasState implements StateProvider by delegating to the wrapped model
func (a *BubbleTeaAdapter) CanvasState() StatePayload {
	if sp, ok := a.model.(StateProvider); ok {
		return sp.CanvasState()
	}
	return StatePayload{}
}

// CanvasView implements ViewProvider
func (a *BubbleTeaAdapter) CanvasView() string {
	return a.model.View()
}
