package rpa

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
