package rpa

import "fmt"

func BuildExecutionPlan(document FlowDocument) ([]FlowNode, error) {
	document = normalizeDocument(document)
	if err := ValidateFlowDocument(document); err != nil {
		return nil, err
	}

	nodeByID := make(map[string]FlowNode, len(document.Nodes))
	outgoing := make(map[string][]FlowEdge, len(document.Nodes))
	for _, node := range document.Nodes {
		nodeByID[node.NodeID] = node
	}
	for _, edge := range document.Edges {
		outgoing[edge.SourceNodeID] = append(outgoing[edge.SourceNodeID], edge)
	}

	startNode, err := findStartNode(document.Nodes)
	if err != nil {
		return nil, err
	}

	plan := make([]FlowNode, 0, len(document.Nodes))
	visited := map[string]bool{}
	current := startNode
	for {
		if visited[current.NodeID] {
			return nil, fmt.Errorf("流程包含循环，当前版本暂不支持")
		}
		visited[current.NodeID] = true
		plan = append(plan, current)
		if current.NodeType == NodeTypeEnd {
			return plan, nil
		}

		nextEdges := outgoing[current.NodeID]
		if len(nextEdges) == 0 {
			return nil, fmt.Errorf("节点 %s 缺少后续连线", current.NodeID)
		}
		if len(nextEdges) > 1 {
			return nil, fmt.Errorf("节点 %s 存在分支，当前版本暂不支持", current.NodeID)
		}

		nextNode, ok := nodeByID[nextEdges[0].TargetNodeID]
		if !ok {
			return nil, fmt.Errorf("连线目标节点不存在: %s", nextEdges[0].TargetNodeID)
		}
		current = nextNode
	}
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
