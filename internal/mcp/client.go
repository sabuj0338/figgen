package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
)

type Client struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader

	reqID uint64
	mu    sync.Mutex
	pending map[uint64]chan []byte
}

type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      uint64      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      uint64          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewClient(command string, args []string, env []string) (*Client, error) {
	cmd := exec.Command(command, args...)
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	
	// Stderr can just go to our stderr for debugging
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	c := &Client{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  bufio.NewReader(stdout),
		pending: make(map[uint64]chan []byte),
	}

	go c.listen()

	return c, nil
}

func (c *Client) listen() {
	for {
		line, err := c.stdout.ReadBytes('\n')
		if err != nil {
			return
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		var resp Response
		if err := json.Unmarshal(line, &resp); err == nil && resp.ID != 0 {
			c.mu.Lock()
			ch, ok := c.pending[resp.ID]
			if ok {
				delete(c.pending, resp.ID)
			}
			c.mu.Unlock()

			if ok {
				ch <- line
			}
		}
	}
}

func (c *Client) Call(ctx context.Context, method string, params interface{}) ([]byte, error) {
	id := atomic.AddUint64(&c.reqID, 1)

	req := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	ch := make(chan []byte, 1)
	c.mu.Lock()
	c.pending[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	b = append(b, '\n')
	if _, err := c.stdin.Write(b); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resBytes := <-ch:
		var res Response
		if err := json.Unmarshal(resBytes, &res); err != nil {
			return nil, err
		}
		if res.Error != nil {
			return nil, fmt.Errorf("rpc error %d: %s", res.Error.Code, res.Error.Message)
		}
		return res.Result, nil
	}
}

func (c *Client) Initialize(ctx context.Context) error {
	params := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "figgen",
			"version": "1.0.0",
		},
	}

	_, err := c.Call(ctx, "initialize", params)
	if err != nil {
		return err
	}
	
	// Send initialized notification
	initNotif := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	b, _ := json.Marshal(initNotif)
	b = append(b, '\n')
	_, err = c.stdin.Write(b)
	return err
}

func (c *Client) Close() error {
	c.stdin.Close()
	return c.cmd.Wait()
}
