package canvas

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Client connects to a canvas server to query/control it
type Client struct {
	id     string
	socket string
}

// NewClient creates a client for the given canvas ID
func NewClient(id string) *Client {
	return &Client{
		id:     id,
		socket: SocketPath(id),
	}
}

// NewClientWithSocket creates a client with a custom socket path
func NewClientWithSocket(socket string) *Client {
	return &Client{socket: socket}
}

// GetState queries the canvas for its current state
func (c *Client) GetState() (*StatePayload, error) {
	resp, err := c.send(MsgGetState, nil)
	if err != nil {
		return nil, err
	}

	if resp.Type == MsgError {
		var errPayload ErrorPayload
		resp.ParsePayload(&errPayload)
		return nil, fmt.Errorf("%s: %s", errPayload.Code, errPayload.Message)
	}

	var state StatePayload
	if err := resp.ParsePayload(&state); err != nil {
		return nil, err
	}

	return &state, nil
}

// GetView queries the canvas for its rendered view
func (c *Client) GetView() (string, error) {
	resp, err := c.send(MsgGetView, nil)
	if err != nil {
		return "", err
	}

	if resp.Type == MsgError {
		var errPayload ErrorPayload
		resp.ParsePayload(&errPayload)
		return "", fmt.Errorf("%s: %s", errPayload.Code, errPayload.Message)
	}

	var view ViewPayload
	if err := resp.ParsePayload(&view); err != nil {
		return "", err
	}

	return view.Content, nil
}

// SendKey sends a key press to the canvas
func (c *Client) SendKey(key string) error {
	resp, err := c.send(MsgSendKey, KeyPayload{Key: key})
	if err != nil {
		return err
	}

	if resp.Type == MsgError {
		var errPayload ErrorPayload
		resp.ParsePayload(&errPayload)
		return fmt.Errorf("%s: %s", errPayload.Code, errPayload.Message)
	}

	return nil
}

// SendInput sends text input to the canvas
func (c *Client) SendInput(text string) error {
	resp, err := c.send(MsgSendInput, InputPayload{Text: text})
	if err != nil {
		return err
	}

	if resp.Type == MsgError {
		var errPayload ErrorPayload
		resp.ParsePayload(&errPayload)
		return fmt.Errorf("%s: %s", errPayload.Code, errPayload.Message)
	}

	return nil
}

// Close requests the canvas to close
func (c *Client) Close() error {
	_, err := c.send(MsgClose, nil)
	return err
}

// Ping checks if the canvas is responsive
func (c *Client) Ping() bool {
	_, err := c.GetState()
	return err == nil
}

func (c *Client) send(msgType MessageType, payload any) (*Message, error) {
	conn, err := net.DialTimeout("unix", c.socket, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to canvas: %w", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(10 * time.Second))

	// Send request
	msg, err := NewMessage(msgType, payload)
	if err != nil {
		return nil, err
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Read response
	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resp Message
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}
