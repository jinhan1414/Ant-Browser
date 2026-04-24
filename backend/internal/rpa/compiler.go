package rpa

import "fmt"

type CompiledEdge struct {
	TargetNodeID string `json:"targetNodeId"`
	Condition    string `json:"condition"`
}

type CompiledNode struct {
	Node    FlowNode       `json:"node"`
	Next    []CompiledEdge `json:"next"`
	OnError []CompiledEdge `json:"onError"`
}

type ExecutionPlan struct {
	EntryNodeID string                  `json:"entryNodeId"`
	Nodes       map[string]CompiledNode `json:"nodes"`
}

func BuildExecutionPlan(document FlowDocument) (*ExecutionPlan, error) {
	document = normalizeDocument(document)
	if err := ValidateFlowDocument(document); err != nil {
		return nil, err
	}

	startNode, err := findStartNode(document.Nodes)
	if err != nil {
		return nil, err
	}

	plan := &ExecutionPlan{
		EntryNodeID: startNode.NodeID,
		Nodes:       make(map[string]CompiledNode, len(document.Nodes)),
	}
	for _, node := range document.Nodes {
		plan.Nodes[node.NodeID] = CompiledNode{
			Node:    node,
			Next:    []CompiledEdge{},
			OnError: []CompiledEdge{},
		}
	}
	for _, edge := range document.Edges {
		node := plan.Nodes[edge.SourceNodeID]
		compiled := CompiledEdge{
			TargetNodeID: edge.TargetNodeID,
			Condition:    edge.Condition,
		}
		if edge.BranchType == FlowEdgeBranchOnError {
			node.OnError = append(node.OnError, compiled)
		} else {
			node.Next = append(node.Next, compiled)
		}
		plan.Nodes[edge.SourceNodeID] = node
	}
	return plan, nil
}

func ValidateFlowDocument(document FlowDocument) error {
	document = normalizeDocument(document)
	if len(document.Nodes) == 0 {
		return fmt.Errorf("流程至少需要一个节点")
	}
	if _, err := findStartNode(document.Nodes); err != nil {
		return err
	}

	nodeByID := make(map[string]FlowNode, len(document.Nodes))
	endCount := 0
	for _, node := range document.Nodes {
		if node.NodeID == "" {
			return fmt.Errorf("节点 id 不能为空")
		}
		if _, exists := nodeByID[node.NodeID]; exists {
			return fmt.Errorf("节点 id 重复: %s", node.NodeID)
		}
		nodeByID[node.NodeID] = node
		if node.NodeType == NodeTypeEnd {
			endCount++
		}
	}
	if endCount == 0 {
		return fmt.Errorf("流程至少需要一个结束节点")
	}

	seenEdges := map[string]bool{}
	for _, edge := range document.Edges {
		if edge.SourceNodeID == "" || edge.TargetNodeID == "" {
			return fmt.Errorf("连线起点和终点不能为空")
		}
		if _, ok := nodeByID[edge.SourceNodeID]; !ok {
			return fmt.Errorf("连线起点不存在: %s", edge.SourceNodeID)
		}
		if _, ok := nodeByID[edge.TargetNodeID]; !ok {
			return fmt.Errorf("连线终点不存在: %s", edge.TargetNodeID)
		}
		key := edge.SourceNodeID + "|" + edge.TargetNodeID + "|" + edge.Condition
		if seenEdges[key] {
			return fmt.Errorf("重复连线: %s", key)
		}
		seenEdges[key] = true
	}
	if err := validateRequiredNodeFields(document, nodeByID); err != nil {
		return err
	}
	if err := validateEdgeSemantics(document, nodeByID); err != nil {
		return err
	}
	return nil
}

func findStartNode(nodes []FlowNode) (FlowNode, error) {
	var start *FlowNode
	for idx := range nodes {
		if nodes[idx].NodeType != NodeTypeStart {
			continue
		}
		if start != nil {
			return FlowNode{}, fmt.Errorf("流程只能包含一个开始节点")
		}
		start = &nodes[idx]
	}
	if start == nil {
		return FlowNode{}, fmt.Errorf("流程缺少开始节点")
	}
	return *start, nil
}
