package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ali/opencode-canvas/canvas"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "state":
		cmdState(args)
	case "view":
		cmdView(args)
	case "key":
		cmdKey(args)
	case "input":
		cmdInput(args)
	case "close":
		cmdClose(args)
	case "list":
		cmdList()
	case "spawn":
		cmdSpawn(args)
	case "ping":
		cmdPing(args)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`opencode-canvas - IPC tool for TUI canvases

USAGE:
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

EXAMPLES:
    # Query a canvas
    opencode-canvas state my-tui
    opencode-canvas view my-tui

    # Send input
    opencode-canvas key my-tui enter
    opencode-canvas input my-tui "hello world"

    # Spawn a TUI in tmux
    opencode-canvas spawn my-tui ./my-app --flag

ENVIRONMENT:
    CANVAS_ID               Default canvas ID
    OPENCODE_CANVAS=1       Enable canvas mode in wrapped TUIs`)
}

func getID(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	if id := os.Getenv("CANVAS_ID"); id != "" {
		return id
	}
	fmt.Fprintln(os.Stderr, "Error: canvas ID required")
	os.Exit(1)
	return ""
}

func cmdState(args []string) {
	id := getID(args)
	client := canvas.NewClient(id)
	
	state, err := client.GetState()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(state)
}

func cmdView(args []string) {
	id := getID(args)
	client := canvas.NewClient(id)
	
	view, err := client.GetView()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Print(view)
}

func cmdKey(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: opencode-canvas key <id> <key>")
		os.Exit(1)
	}
	
	id := args[0]
	key := args[1]
	client := canvas.NewClient(id)
	
	if err := client.SendKey(key); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("OK")
}

func cmdInput(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: opencode-canvas input <id> <text>")
		os.Exit(1)
	}
	
	id := args[0]
	text := strings.Join(args[1:], " ")
	client := canvas.NewClient(id)
	
	if err := client.SendInput(text); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("OK")
}

func cmdClose(args []string) {
	id := getID(args)
	client := canvas.NewClient(id)
	
	if err := client.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("OK")
}

func cmdList() {
	socketDir := canvas.DefaultSocketDir()
	entries, err := os.ReadDir(socketDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No active canvases")
			return
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	found := false
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".sock") {
			id := strings.TrimSuffix(entry.Name(), ".sock")
			client := canvas.NewClient(id)
			status := "dead"
			if client.Ping() {
				status = "alive"
			}
			fmt.Printf("%s\t%s\n", id, status)
			found = true
		}
	}
	
	if !found {
		fmt.Println("No active canvases")
	}
}

func cmdSpawn(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: opencode-canvas spawn <id> <command...>")
		os.Exit(1)
	}
	
	id := args[0]
	command := args[1:]
	
	// Check if we're in tmux
	if os.Getenv("TMUX") == "" {
		fmt.Fprintln(os.Stderr, "Error: spawn requires tmux session")
		os.Exit(1)
	}
	
	// Build the command with canvas environment
	cmdStr := fmt.Sprintf("OPENCODE_CANVAS=1 CANVAS_ID=%s %s", 
		id, strings.Join(command, " "))
	
	// Spawn in a new tmux pane
	tmuxArgs := []string{
		"split-window", "-h",
		"-p", "50",          // 50% width
		"-P", "-F", "#{pane_id}",
		cmdStr,
	}
	
	tmux := exec.Command("tmux", tmuxArgs...)
	output, err := tmux.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error spawning tmux pane: %v\n", err)
		os.Exit(1)
	}
	
	paneID := strings.TrimSpace(string(output))
	
	// Save pane ID for later reference
	paneFile := filepath.Join(canvas.DefaultSocketDir(), fmt.Sprintf("%s.pane", id))
	os.MkdirAll(canvas.DefaultSocketDir(), 0755)
	os.WriteFile(paneFile, []byte(paneID), 0644)
	
	fmt.Printf("Spawned canvas '%s' in pane %s\n", id, paneID)
}

func cmdPing(args []string) {
	id := getID(args)
	client := canvas.NewClient(id)
	
	if client.Ping() {
		fmt.Println("OK")
	} else {
		fmt.Println("FAIL")
		os.Exit(1)
	}
}
