package rpa

import (
	"fmt"
	"strings"
)

func validateRequiredNodeFields(document FlowDocument, nodeByID map[string]FlowNode) error {
	for nodeID, node := range nodeByID {
		item, ok := FindFlowNodeCatalogItem(node.NodeType)
		if !ok {
			continue
		}
		for _, field := range item.Fields {
			if !field.Required {
				continue
			}
			if hasNodeFieldValue(node, field) {
				continue
			}
			return fmt.Errorf("节点 %s 缺少必填字段: %s", nodeID, field.Key)
		}
	}
	return nil
}

func hasNodeFieldValue(node FlowNode, field FlowNodeField) bool {
	value := node.Config[field.Key]
	switch field.Storage {
	case FieldStorageStringListFirst:
		values := stringSliceConfig(value)
		return len(values) > 0 && strings.TrimSpace(values[0]) != ""
	case FieldStorageNumber:
		_, ok := numberConfig(value)
		return ok
	default:
		return strings.TrimSpace(stringConfig(value)) != ""
	}
}
