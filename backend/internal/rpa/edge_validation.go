package rpa

import "fmt"

func validateEdgeSemantics(document FlowDocument, nodeByID map[string]FlowNode) error {
	outgoing := buildOutgoingEdges(document.Edges)
	for nodeID, edges := range outgoing {
		node, ok := nodeByID[nodeID]
		if !ok {
			continue
		}
		if err := validateOutgoingEdges(node, edges); err != nil {
			return err
		}
	}
	return nil
}

func buildOutgoingEdges(edges []FlowEdge) map[string][]FlowEdge {
	result := make(map[string][]FlowEdge, len(edges))
	for _, edge := range edges {
		result[edge.SourceNodeID] = append(result[edge.SourceNodeID], edge)
	}
	return result
}

func validateOutgoingEdges(node FlowNode, edges []FlowEdge) error {
	item, _ := FindFlowNodeCatalogItem(node.NodeType)
	defaultCount := 0
	onErrorCount := 0
	trueCount := 0
	falseCount := 0
	for _, edge := range edges {
		switch edge.BranchType {
		case FlowEdgeBranchDefault:
			defaultCount++
		case FlowEdgeBranchOnError:
			onErrorCount++
		case FlowEdgeBranchTrue:
			trueCount++
		case FlowEdgeBranchFalse:
			falseCount++
		default:
			return fmt.Errorf("节点 %s 存在未知分支类型: %s", node.NodeID, edge.BranchType)
		}
	}
	if item.SupportsIfBranch {
		if defaultCount > 0 || onErrorCount > 0 {
			return fmt.Errorf("条件节点 %s 只允许 true/false 分支", node.NodeID)
		}
		if trueCount != 1 || falseCount != 1 {
			return fmt.Errorf("条件节点 %s 必须包含一条 true 和一条 false 分支", node.NodeID)
		}
		return nil
	}
	if trueCount > 0 || falseCount > 0 {
		return fmt.Errorf("节点 %s 不支持 true/false 分支", node.NodeID)
	}
	if defaultCount > 1 {
		return fmt.Errorf("节点 %s 默认分支不能超过一条", node.NodeID)
	}
	if onErrorCount > 1 {
		return fmt.Errorf("节点 %s 异常分支不能超过一条", node.NodeID)
	}
	if onErrorCount > 0 && !item.SupportsOnError {
		return fmt.Errorf("节点 %s 不支持 on_error 分支", node.NodeID)
	}
	if item.MaxOutgoing > 0 && len(edges) > item.MaxOutgoing {
		return fmt.Errorf("节点 %s 出边数超过限制", node.NodeID)
	}
	return nil
}
