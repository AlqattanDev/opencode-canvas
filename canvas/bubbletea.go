package canvas

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type BubbleTeaAdapter struct {
	server *Server
	model  tea.Model
}

func Wrap(canvasID string, model tea.Model) tea.Model {
	if os.Getenv("OPENCODE_CANVAS") == "" && os.Getenv("CANVAS_ID") == "" {
		return model
	}

	id := canvasID
	if envID := os.Getenv("CANVAS_ID"); envID != "" {
		id = envID
	}

	server, err := NewServer(id)
	if err != nil {
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

func (a *BubbleTeaAdapter) Init() tea.Cmd {
	return a.model.Init()
}

func (a *BubbleTeaAdapter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.QuitMsg); ok {
		a.server.Stop()
	}

	newModel, cmd := a.model.Update(msg)
	a.model = newModel

	return a, cmd
}

func (a *BubbleTeaAdapter) View() string {
	return a.model.View()
}

func (a *BubbleTeaAdapter) CanvasState() StatePayload {
	if sp, ok := a.model.(StateProvider); ok {
		return sp.CanvasState()
	}
	return StatePayload{}
}

func (a *BubbleTeaAdapter) CanvasView() string {
	return a.model.View()
}

func (a *BubbleTeaAdapter) HandleCanvasKey(key string, r rune) error {
	if kh, ok := a.model.(KeyHandler); ok {
		return kh.HandleCanvasKey(key, r)
	}
	return nil
}

func (a *BubbleTeaAdapter) HandleCanvasInput(text string) error {
	if ih, ok := a.model.(InputHandler); ok {
		return ih.HandleCanvasInput(text)
	}
	return nil
}
