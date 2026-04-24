package rpa

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type BrowserSession struct {
	client *CDPClient
	conn   *websocket.Conn
	nextID int
}

func NewBrowserSession(client *CDPClient) (*BrowserSession, error) {
	if client == nil {
		return nil, fmt.Errorf("cdp client is nil")
	}
	return &BrowserSession{client: client, nextID: 1}, nil
}

func (s *BrowserSession) AttachPage() error {
	target, err := s.client.PageTarget()
	if err != nil {
		return err
	}
	conn, err := s.client.Dial(target.WebSocketDebuggerURL)
	if err != nil {
		return err
	}
	s.conn = conn
	if _, err := s.call("Page.enable", nil); err != nil {
		return err
	}
	if _, err := s.call("Runtime.enable", nil); err != nil {
		return err
	}
	return nil
}

func (s *BrowserSession) QuerySelector(selector string) (string, error) {
	value, err := s.evaluate(fmt.Sprintf(
		`(() => { const el = document.querySelector(%s); return el ? %s : null; })()`,
		jsStringLiteral(selector),
		jsStringLiteral(selector),
	))
	if err != nil {
		return "", err
	}
	text, ok := value.(string)
	if !ok || text == "" {
		return "", fmt.Errorf("selector not found: %s", selector)
	}
	return text, nil
}

func (s *BrowserSession) WaitVisible(selector string, timeoutMs int) error {
	if timeoutMs <= 0 {
		timeoutMs = 5000
	}
	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	for {
		if _, err := s.QuerySelector(selector); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("wait visible timeout: %s", selector)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *BrowserSession) Click(selector string) error {
	value, err := s.evaluate(fmt.Sprintf(
		`(() => { const el = document.querySelector(%s); if (!el) return false; el.click(); return true; })()`,
		jsStringLiteral(selector),
	))
	if err != nil {
		return err
	}
	if ok, _ := value.(bool); !ok {
		return fmt.Errorf("click failed: %s", selector)
	}
	return nil
}

func (s *BrowserSession) InputText(selector string, value string) error {
	result, err := s.evaluate(fmt.Sprintf(
		`(() => { const el = document.querySelector(%s); if (!el) return false; el.value = %s; el.dispatchEvent(new Event('input', { bubbles: true })); el.dispatchEvent(new Event('change', { bubbles: true })); return true; })()`,
		jsStringLiteral(selector),
		jsStringLiteral(value),
	))
	if err != nil {
		return err
	}
	if ok, _ := result.(bool); !ok {
		return fmt.Errorf("input failed: %s", selector)
	}
	return nil
}

func (s *BrowserSession) ReadText(selector string) (string, error) {
	value, err := s.evaluate(fmt.Sprintf(
		`(() => { const el = document.querySelector(%s); return el ? (el.innerText || el.textContent || '') : null; })()`,
		jsStringLiteral(selector),
	))
	if err != nil {
		return "", err
	}
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("text not found: %s", selector)
	}
	return text, nil
}

func (s *BrowserSession) ReadURL() (string, error) {
	value, err := s.evaluate(`(() => window.location.href || "")()`)
	if err != nil {
		return "", err
	}
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("read url failed")
	}
	return text, nil
}

func (s *BrowserSession) Close() error {
	if s.conn == nil {
		return nil
	}
	err := s.conn.Close()
	s.conn = nil
	return err
}

func (s *BrowserSession) evaluate(expression string) (any, error) {
	raw, err := s.call("Runtime.evaluate", map[string]any{
		"expression":    expression,
		"returnByValue": true,
	})
	if err != nil {
		return nil, err
	}
	payload := struct {
		Result struct {
			Value any `json:"value"`
		} `json:"result"`
	}{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("cdp evaluate decode failed: %w", err)
	}
	return payload.Result.Value, nil
}

func (s *BrowserSession) call(method string, params map[string]any) (json.RawMessage, error) {
	if s.conn == nil {
		return nil, fmt.Errorf("browser session not attached")
	}
	requestID := s.nextID
	s.nextID++

	if err := s.conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return nil, err
	}
	if err := s.conn.WriteJSON(cdpMessage{
		ID:     requestID,
		Method: method,
		Params: params,
	}); err != nil {
		return nil, fmt.Errorf("cdp command write failed: %w", err)
	}

	for {
		response := cdpResponse{}
		if err := s.conn.ReadJSON(&response); err != nil {
			return nil, fmt.Errorf("cdp response read failed: %w", err)
		}
		if response.ID != requestID {
			continue
		}
		if response.Error != nil {
			return nil, fmt.Errorf("cdp error: %s", response.Error.Message)
		}
		return response.Result, nil
	}
}

func jsStringLiteral(value string) string {
	return "'" + escapeJSString(value) + "'"
}

func escapeJSString(value string) string {
	quoted := strconv.Quote(value)
	quoted = quoted[1 : len(quoted)-1]
	quoted = replaceAll(quoted, "\\'", "'")
	quoted = replaceAll(quoted, "'", "\\'")
	return quoted
}

func replaceAll(input string, old string, new string) string {
	for old != "" {
		next := input
		if idx := indexOf(next, old); idx >= 0 {
			input = next[:idx] + new + next[idx+len(old):]
			continue
		}
		break
	}
	return input
}

func indexOf(input string, token string) int {
	limit := len(input) - len(token)
	for idx := 0; idx <= limit; idx++ {
		if input[idx:idx+len(token)] == token {
			return idx
		}
	}
	return -1
}
