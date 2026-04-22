package rpa

import (
	"fmt"
	"time"

	"ant-chrome/backend/internal/browser"
)

func (e *Executor) executeNode(profileID string, node FlowNode) (*browser.Profile, error) {
	switch node.NodeType {
	case NodeTypeStart, NodeTypeEnd:
		return nil, nil
	case NodeTypeBrowserStart:
		return e.operator.Start(profileID, nodeLaunchArgs(node), nodeStartURLs(node), nodeSkipDefault(node))
	case NodeTypeBrowserOpenURL:
		return e.operator.Start(profileID, nil, nodeStartURLs(node), true)
	case NodeTypeDelay:
		waitDuration(node)
		return nil, nil
	case NodeTypeBrowserStop:
		return e.operator.Stop(profileID)
	default:
		return nil, fmt.Errorf("不支持的 RPA 节点类型: %s", node.NodeType)
	}
}

func nodeLaunchArgs(node FlowNode) []string {
	values := stringSliceConfig(node.Config["launchArgs"])
	if len(values) == 0 {
		return nil
	}
	return values
}

func nodeStartURLs(node FlowNode) []string {
	if url := stringConfig(node.Config["url"]); url != "" {
		return []string{url}
	}
	values := stringSliceConfig(node.Config["startUrls"])
	if len(values) == 0 {
		return nil
	}
	return values
}

func nodeSkipDefault(node FlowNode) bool {
	value, ok := node.Config["skipDefaultStartUrls"].(bool)
	return ok && value
}

func waitDuration(node FlowNode) {
	durationMs, ok := numberConfig(node.Config["durationMs"])
	if !ok || durationMs <= 0 {
		return
	}
	time.Sleep(time.Duration(durationMs) * time.Millisecond)
}

func stringSliceConfig(raw any) []string {
	if list, ok := raw.([]string); ok {
		values := make([]string, 0, len(list))
		for _, item := range list {
			if item != "" {
				values = append(values, item)
			}
		}
		return values
	}
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	values := make([]string, 0, len(list))
	for _, item := range list {
		text, ok := item.(string)
		if ok && text != "" {
			values = append(values, text)
		}
	}
	return values
}

func numberConfig(raw any) (float64, bool) {
	switch value := raw.(type) {
	case float64:
		return value, true
	case int:
		return float64(value), true
	case int64:
		return float64(value), true
	default:
		return 0, false
	}
}

func stringConfig(raw any) string {
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return value
}
