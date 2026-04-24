package rpa

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

func TestBrowserSession_QueryClickInputAndReadText(t *testing.T) {
	server := newCDPStubServer(t)
	client := NewCDPClient(server.baseURL)

	session, err := NewBrowserSession(client)
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}
	if err := session.AttachPage(); err != nil {
		t.Fatalf("附加页面失败: %v", err)
	}
	if _, err := session.QuerySelector("#login"); err != nil {
		t.Fatalf("查询元素失败: %v", err)
	}
	if err := session.Click("#login"); err != nil {
		t.Fatalf("点击元素失败: %v", err)
	}
	if err := session.InputText("#email", "alice@example.com"); err != nil {
		t.Fatalf("输入文本失败: %v", err)
	}
	text, err := session.ReadText("#title")
	if err != nil {
		t.Fatalf("读取文本失败: %v", err)
	}
	if text != "控制台首页" {
		t.Fatalf("读取结果错误: %q", text)
	}
	if !server.wasClicked("#login") {
		t.Fatal("点击记录未写入")
	}
	if got := server.inputValue("#email"); got != "alice@example.com" {
		t.Fatalf("输入结果错误: %q", got)
	}
}

type cdpStubServer struct {
	baseURL   string
	server    *httptest.Server
	upgrader  websocket.Upgrader
	mu        sync.Mutex
	clicked   map[string]bool
	inputs    map[string]string
	texts     map[string]string
	available map[string]bool
}

func newCDPStubServer(t *testing.T) *cdpStubServer {
	t.Helper()

	stub := &cdpStubServer{
		upgrader: websocket.Upgrader{},
		clicked: map[string]bool{},
		inputs: map[string]string{},
		texts: map[string]string{
			"#title": "控制台首页",
		},
		available: map[string]bool{
			"#login": true,
			"#email": true,
			"#title": true,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		wsURL := "ws" + strings.TrimPrefix(stub.baseURL, "http") + "/page/1"
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"id":                   "page-1",
			"type":                 "page",
			"webSocketDebuggerUrl": wsURL,
		}})
	})
	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		wsURL := "ws" + strings.TrimPrefix(stub.baseURL, "http") + "/browser"
		_ = json.NewEncoder(w).Encode(map[string]any{
			"webSocketDebuggerUrl": wsURL,
		})
	})
	mux.HandleFunc("/page/1", stub.handleWebSocket)
	server := httptest.NewServer(mux)
	stub.server = server
	stub.baseURL = server.URL
	t.Cleanup(server.Close)
	return stub
}

func (s *cdpStubServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		var message map[string]any
		if err := conn.ReadJSON(&message); err != nil {
			return
		}
		id, _ := message["id"].(float64)
		method, _ := message["method"].(string)
		params, _ := message["params"].(map[string]any)
		result, callErr := s.handleMethod(method, params)
		response := map[string]any{"id": int(id)}
		if callErr != nil {
			response["error"] = map[string]any{"message": callErr.Error()}
		} else {
			response["result"] = result
		}
		if err := conn.WriteJSON(response); err != nil {
			return
		}
	}
}

func (s *cdpStubServer) handleMethod(method string, params map[string]any) (map[string]any, error) {
	switch method {
	case "Page.enable", "Runtime.enable":
		return map[string]any{}, nil
	case "Runtime.evaluate":
		expression, _ := params["expression"].(string)
		return s.handleEvaluate(expression)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

func (s *cdpStubServer) handleEvaluate(expression string) (map[string]any, error) {
	switch {
	case strings.Contains(expression, "querySelector(") && strings.Contains(expression, ".click()"):
		selector := quotedArgument(expression)
		s.mu.Lock()
		s.clicked[selector] = true
		s.mu.Unlock()
		return remoteValue(true), nil
	case strings.Contains(expression, "querySelector(") && strings.Contains(expression, ".value = "):
		values := quotedArguments(expression)
		if len(values) < 2 {
			return nil, fmt.Errorf("input arguments invalid")
		}
		s.mu.Lock()
		s.inputs[values[0]] = values[1]
		s.mu.Unlock()
		return remoteValue(true), nil
	case strings.Contains(expression, "querySelector(") && strings.Contains(expression, "textContent"):
		selector := quotedArgument(expression)
		s.mu.Lock()
		text := s.texts[selector]
		s.mu.Unlock()
		return remoteValue(text), nil
	case strings.Contains(expression, "querySelector("):
		selector := quotedArgument(expression)
		s.mu.Lock()
		available := s.available[selector]
		s.mu.Unlock()
		if !available {
			return remoteValue(nil), nil
		}
		return remoteValue(selector), nil
	default:
		return nil, fmt.Errorf("unsupported expression: %s", expression)
	}
}

func (s *cdpStubServer) wasClicked(selector string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.clicked[selector]
}

func (s *cdpStubServer) inputValue(selector string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.inputs[selector]
}

func remoteValue(value any) map[string]any {
	return map[string]any{
		"result": map[string]any{
			"value": value,
		},
	}
}

func quotedArgument(input string) string {
	values := quotedArguments(input)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func quotedArguments(input string) []string {
	values := make([]string, 0, 2)
	start := -1
	for idx, char := range input {
		if char == '\'' {
			if start >= 0 {
				values = append(values, input[start:idx])
				start = -1
				continue
			}
			start = idx + 1
		}
	}
	return values
}
