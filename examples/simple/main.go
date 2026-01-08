// A simple non-TUI test server for canvas protocol testing
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ali/opencode-canvas/canvas"
)

type testModel struct {
	counter int
	mode    string
}

func (m *testModel) CanvasState() canvas.StatePayload {
	return canvas.StatePayload{
		Mode: m.mode,
		Custom: map[string]any{
			"counter": m.counter,
			"time":    time.Now().Format(time.RFC3339),
		},
	}
}

func (m *testModel) CanvasView() string {
	return fmt.Sprintf("Counter: %d\nMode: %s\n", m.counter, m.mode)
}

func main() {
	id := os.Getenv("CANVAS_ID")
	if id == "" {
		id = "test-server"
	}

	server, err := canvas.NewServer(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}

	model := &testModel{counter: 42, mode: "test"}
	server.SetModel(model)
	server.Start()

	fmt.Printf("Canvas server '%s' started at %s\n", id, server.SocketPath())

	// Handle shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	server.Stop()
	fmt.Println("Server stopped")
}
