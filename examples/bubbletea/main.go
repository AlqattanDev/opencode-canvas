// Example: A simple counter TUI with canvas support
package main

import (
	"fmt"
	"os"

	"github.com/AlqattanDev/opencode-canvas/canvas"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))
	
	counterStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Padding(1, 2)
	
	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
)

type model struct {
	count    int
	quitting bool
}

func newModel() model {
	return model{count: 0}
}

// CanvasState implements canvas.StateProvider
func (m model) CanvasState() canvas.StatePayload {
	return canvas.StatePayload{
		Mode: "counter",
		Custom: map[string]any{
			"count":    m.count,
			"quitting": m.quitting,
		},
	}
}

// CanvasView implements canvas.ViewProvider (optional, View() is used by default)
func (m model) CanvasView() string {
	return m.View()
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k", "+":
			m.count++
		case "down", "j", "-":
			m.count--
		case "r":
			m.count = 0
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}
	
	s := titleStyle.Render("🎨 Canvas Counter Example") + "\n\n"
	s += counterStyle.Render(fmt.Sprintf("Count: %d", m.count)) + "\n\n"
	s += helpStyle.Render("↑/k/+ increment • ↓/j/- decrement • r reset • q quit")
	
	return s
}

func main() {
	// Create the base model
	m := newModel()
	
	// Wrap with canvas support (auto-detects canvas mode via env vars)
	wrapped := canvas.Wrap("counter-example", m)
	
	// Run the program
	p := tea.NewProgram(wrapped)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
