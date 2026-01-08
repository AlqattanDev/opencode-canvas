// Package canvas provides IPC communication for TUI applications.
// It allows AI assistants to query and control TUI state via Unix sockets.
package canvas

import "encoding/json"

// MessageType identifies the type of IPC message
type MessageType string

const (
	// Queries (AI → TUI)
	MsgGetState    MessageType = "get_state"
	MsgGetView     MessageType = "get_view"
	MsgSendKey     MessageType = "send_key"
	MsgSendInput   MessageType = "send_input"
	MsgClose       MessageType = "close"

	// Responses (TUI → AI)
	MsgState       MessageType = "state"
	MsgView        MessageType = "view"
	MsgAck         MessageType = "ack"
	MsgError       MessageType = "error"

	// Events (TUI → AI, async)
	MsgReady       MessageType = "ready"
	MsgUpdated     MessageType = "updated"
	MsgSelected    MessageType = "selected"
	MsgCancelled   MessageType = "cancelled"
)

// Message is the base IPC message structure
type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// StatePayload contains the TUI's current state
type StatePayload struct {
	// Custom fields from the TUI's model
	Custom map[string]any `json:"custom,omitempty"`
	
	// Standard fields
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Focused  bool   `json:"focused,omitempty"`
	Mode     string `json:"mode,omitempty"`
	Input    string `json:"input,omitempty"`
	Cursor   int    `json:"cursor,omitempty"`
}

// ViewPayload contains the rendered view
type ViewPayload struct {
	Content string `json:"content"`
	ANSI    bool   `json:"ansi"` // true if content contains ANSI codes
}

// KeyPayload contains a key to send
type KeyPayload struct {
	Key  string `json:"key"`            // e.g., "enter", "tab", "ctrl+c"
	Rune rune   `json:"rune,omitempty"` // for character input
}

// InputPayload contains text input
type InputPayload struct {
	Text string `json:"text"`
}

// ErrorPayload contains error information
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewMessage creates a new message with the given type and payload
func NewMessage(t MessageType, payload any) (*Message, error) {
	var raw json.RawMessage
	if payload != nil {
		var err error
		raw, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}
	return &Message{Type: t, Payload: raw}, nil
}

// ParsePayload unmarshals the payload into the given type
func (m *Message) ParsePayload(v any) error {
	if m.Payload == nil {
		return nil
	}
	return json.Unmarshal(m.Payload, v)
}
