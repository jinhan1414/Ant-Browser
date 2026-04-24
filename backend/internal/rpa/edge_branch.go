package rpa

import "strings"

func normalizeFlowEdge(edge FlowEdge) FlowEdge {
	edge.EdgeID = strings.TrimSpace(edge.EdgeID)
	edge.SourceNodeID = strings.TrimSpace(edge.SourceNodeID)
	edge.TargetNodeID = strings.TrimSpace(edge.TargetNodeID)
	edge.Label = strings.TrimSpace(edge.Label)
	edge.Condition = strings.TrimSpace(edge.Condition)
	edge.BranchType = normalizeFlowEdgeBranchType(edge.BranchType, edge.Condition)
	edge.Condition = flowEdgeCondition(edge.BranchType)
	return edge
}

func normalizeFlowEdgeBranchType(branchType FlowEdgeBranchType, condition string) FlowEdgeBranchType {
	value := strings.TrimSpace(string(branchType))
	if value == "" {
		value = strings.TrimSpace(condition)
	}
	switch FlowEdgeBranchType(value) {
	case FlowEdgeBranchTrue:
		return FlowEdgeBranchTrue
	case FlowEdgeBranchFalse:
		return FlowEdgeBranchFalse
	case FlowEdgeBranchOnError:
		return FlowEdgeBranchOnError
	default:
		return FlowEdgeBranchDefault
	}
}

func flowEdgeCondition(branchType FlowEdgeBranchType) string {
	switch branchType {
	case FlowEdgeBranchTrue:
		return string(FlowEdgeBranchTrue)
	case FlowEdgeBranchFalse:
		return string(FlowEdgeBranchFalse)
	case FlowEdgeBranchOnError:
		return string(FlowEdgeBranchOnError)
	default:
		return ""
	}
}

func isFlowEdgeConditionalBranch(branchType FlowEdgeBranchType) bool {
	return branchType == FlowEdgeBranchTrue || branchType == FlowEdgeBranchFalse
}
