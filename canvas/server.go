package canvas

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
)

// StateProvider is implemented by TUI models to expose their state
type StateProvider interface {
	// CanvasState returns the current state for IPC queries
	CanvasState() StatePayload
}

// ViewProvider is implemented by TUI models to expose their rendered view
type ViewProvider interface {
	// CanvasView returns the current rendered view
	CanvasView() string
}

// KeyHandler is implemented by TUI models to receive key events
type KeyHandler interface {
	// HandleCanvasKey processes a key sent via IPC
	HandleCanvasKey(key string, r rune) error
}

// InputHandler is implemented by TUI models to receive text input
type InputHandler interface {
	// HandleCanvasInput processes text input sent via IPC
	HandleCanvasInput(text string) error
}

// Server handles IPC communication for a TUI
type Server struct {
	id       string
	socket   string
	listener net.Listener
	
	mu       sync.RWMutex
	model    any // The TUI model
	onClose  func()
	
	done     chan struct{}
}

// DefaultSocketDir returns the default directory for canvas sockets
func DefaultSocketDir() string {
	return filepath.Join(os.TempDir(), "opencode-canvas")
}

// SocketPath returns the socket path for a canvas ID
func SocketPath(id string) string {
	return filepath.Join(DefaultSocketDir(), fmt.Sprintf("%s.sock", id))
}

// NewServer creates a new IPC server for the given canvas ID
func NewServer(id string) (*Server, error) {
	socketDir := DefaultSocketDir()
	if err := os.MkdirAll(socketDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create socket dir: %w", err)
	}
	
	socketPath := SocketPath(id)
	
	// Remove existing socket if present
	os.Remove(socketPath)
	
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on socket: %w", err)
	}
	
	return &Server{
		id:       id,
		socket:   socketPath,
		listener: listener,
		done:     make(chan struct{}),
	}, nil
}

// SetModel sets the TUI model for state queries
func (s *Server) SetModel(model any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.model = model
}

// OnClose sets a callback for when a close message is received
func (s *Server) OnClose(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onClose = fn
}

// Start begins accepting connections
func (s *Server) Start() {
	go s.acceptLoop()
}

// Stop closes the server
func (s *Server) Stop() {
	close(s.done)
	s.listener.Close()
	os.Remove(s.socket)
}

// SocketPath returns the path to the Unix socket
func (s *Server) SocketPath() string {
	return s.socket
}

// ID returns the canvas ID
func (s *Server) ID() string {
	return s.id
}

// SendEvent sends an async event to any connected client
func (s *Server) SendEvent(msgType MessageType, payload any) error {
	msg, err := NewMessage(msgType, payload)
	if err != nil {
		return err
	}
	// For now, events are logged but not broadcast
	// (would need connection tracking for proper broadcast)
	_ = msg
	return nil
}

func (s *Server) acceptLoop() {
	for {
		select {
		case <-s.done:
			return
		default:
		}
		
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				continue
			}
		}
		
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	
	reader := bufio.NewReader(conn)
	encoder := json.NewEncoder(conn)
	
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}
		
		var msg Message
		if err := json.Unmarshal(line, &msg); err != nil {
			s.sendError(encoder, "parse_error", err.Error())
			continue
		}
		
		s.handleMessage(&msg, encoder)
	}
}

func (s *Server) handleMessage(msg *Message, enc *json.Encoder) {
	s.mu.RLock()
	model := s.model
	onClose := s.onClose
	s.mu.RUnlock()
	
	switch msg.Type {
	case MsgGetState:
		if sp, ok := model.(StateProvider); ok {
			state := sp.CanvasState()
			resp, _ := NewMessage(MsgState, state)
			enc.Encode(resp)
		} else {
			s.sendError(enc, "not_supported", "model does not implement StateProvider")
		}
		
	case MsgGetView:
		if vp, ok := model.(ViewProvider); ok {
			view := ViewPayload{Content: vp.CanvasView(), ANSI: true}
			resp, _ := NewMessage(MsgView, view)
			enc.Encode(resp)
		} else {
			s.sendError(enc, "not_supported", "model does not implement ViewProvider")
		}
		
	case MsgSendKey:
		if kh, ok := model.(KeyHandler); ok {
			var payload KeyPayload
			msg.ParsePayload(&payload)
			if err := kh.HandleCanvasKey(payload.Key, payload.Rune); err != nil {
				s.sendError(enc, "key_error", err.Error())
			} else {
				resp, _ := NewMessage(MsgAck, nil)
				enc.Encode(resp)
			}
		} else {
			s.sendError(enc, "not_supported", "model does not implement KeyHandler")
		}
		
	case MsgSendInput:
		if ih, ok := model.(InputHandler); ok {
			var payload InputPayload
			msg.ParsePayload(&payload)
			if err := ih.HandleCanvasInput(payload.Text); err != nil {
				s.sendError(enc, "input_error", err.Error())
			} else {
				resp, _ := NewMessage(MsgAck, nil)
				enc.Encode(resp)
			}
		} else {
			s.sendError(enc, "not_supported", "model does not implement InputHandler")
		}
		
	case MsgClose:
		if onClose != nil {
			onClose()
		}
		resp, _ := NewMessage(MsgAck, nil)
		enc.Encode(resp)
		
	default:
		s.sendError(enc, "unknown_type", fmt.Sprintf("unknown message type: %s", msg.Type))
	}
}

func (s *Server) sendError(enc *json.Encoder, code, message string) {
	resp, _ := NewMessage(MsgError, ErrorPayload{Code: code, Message: message})
	enc.Encode(resp)
}
