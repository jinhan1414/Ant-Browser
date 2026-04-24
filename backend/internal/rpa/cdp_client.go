package rpa

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type cdpDialFunc func(urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error)

type CDPTarget struct {
	ID                   string `json:"id"`
	Type                 string `json:"type"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

type cdpBrowserVersion struct {
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

type cdpMessage struct {
	ID     int            `json:"id"`
	Method string         `json:"method,omitempty"`
	Params map[string]any `json:"params,omitempty"`
}

type cdpResponse struct {
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type CDPClient struct {
	baseURL    string
	httpClient *http.Client
	dial       cdpDialFunc
}

func NewCDPClient(baseURL string) *CDPClient {
	return &CDPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		dial: websocket.DefaultDialer.Dial,
	}
}

func NewCDPClientForDebugPort(debugPort int) *CDPClient {
	return NewCDPClient(fmt.Sprintf("http://127.0.0.1:%d", debugPort))
}

func (c *CDPClient) PageTarget() (*CDPTarget, error) {
	targets, err := c.listTargets()
	if err != nil {
		return nil, err
	}
	for _, target := range targets {
		if target.Type == "page" && target.WebSocketDebuggerURL != "" {
			item := target
			return &item, nil
		}
	}
	for _, target := range targets {
		if target.WebSocketDebuggerURL != "" {
			item := target
			return &item, nil
		}
	}
	return nil, fmt.Errorf("page target not found")
}

func (c *CDPClient) BrowserTarget() (*cdpBrowserVersion, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/json/version")
	if err != nil {
		return nil, fmt.Errorf("cdp /json/version request failed: %w", err)
	}
	defer resp.Body.Close()

	item := &cdpBrowserVersion{}
	if err := json.NewDecoder(resp.Body).Decode(item); err != nil {
		return nil, fmt.Errorf("cdp browser target decode failed: %w", err)
	}
	if strings.TrimSpace(item.WebSocketDebuggerURL) == "" {
		return nil, fmt.Errorf("browser websocket target not found")
	}
	return item, nil
}

func (c *CDPClient) listTargets() ([]CDPTarget, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/json/list")
	if err != nil {
		return nil, fmt.Errorf("cdp /json/list request failed: %w", err)
	}
	defer resp.Body.Close()

	items := make([]CDPTarget, 0, 4)
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("cdp targets decode failed: %w", err)
	}
	return items, nil
}

func (c *CDPClient) Dial(targetURL string) (*websocket.Conn, error) {
	conn, _, err := c.dial(targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("cdp websocket dial failed: %w", err)
	}
	return conn, nil
}
