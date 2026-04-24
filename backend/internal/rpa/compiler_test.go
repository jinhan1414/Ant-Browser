package rpa

import "testing"

func TestBuildExecutionPlan_SupportsBranchRetryAndParallel(t *testing.T) {
	document := FlowDocument{
		SchemaVersion: 3,
		Nodes: []FlowNode{
			{NodeID: "start_1", NodeType: NodeTypeStart},
			{NodeID: "read_1", NodeType: NodeTypeBrowserReadText, Config: map[string]any{"selector": "#title", "saveAs": "pageTitle"}},
			{NodeID: "if_1", NodeType: NodeTypeConditionIf, Config: map[string]any{"expression": `pageTitle == "控制台首页"`}},
			{NodeID: "parallel_1", NodeType: NodeTypeParallel},
			{NodeID: "end_1", NodeType: NodeTypeEnd},
		},
		Edges: []FlowEdge{
			{EdgeID: "e1", SourceNodeID: "start_1", TargetNodeID: "read_1"},
			{EdgeID: "e2", SourceNodeID: "read_1", TargetNodeID: "if_1"},
			{EdgeID: "e3", SourceNodeID: "if_1", TargetNodeID: "parallel_1", BranchType: FlowEdgeBranchTrue},
			{EdgeID: "e4", SourceNodeID: "if_1", TargetNodeID: "end_1", BranchType: FlowEdgeBranchFalse},
			{EdgeID: "e5", SourceNodeID: "parallel_1", TargetNodeID: "end_1"},
		},
	}

	plan, err := BuildExecutionPlan(document)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	if len(plan.Nodes) == 0 {
		t.Fatal("执行计划不能为空")
	}
	if plan.EntryNodeID != "start_1" {
		t.Fatalf("入口节点错误: %s", plan.EntryNodeID)
	}
	if len(plan.Nodes["if_1"].Next) != 2 {
		t.Fatalf("条件节点分支数错误: %+v", plan.Nodes["if_1"])
	}
}

func TestValidateFlowDocument_RejectsInvalidBranchSemantics(t *testing.T) {
	document := FlowDocument{
		SchemaVersion: 3,
		Nodes: []FlowNode{
			{NodeID: "start_1", NodeType: NodeTypeStart},
			{NodeID: "if_1", NodeType: NodeTypeConditionIf, Config: map[string]any{"expression": `ok == true`}},
			{NodeID: "end_1", NodeType: NodeTypeEnd},
		},
		Edges: []FlowEdge{
			{EdgeID: "e1", SourceNodeID: "start_1", TargetNodeID: "if_1"},
			{EdgeID: "e2", SourceNodeID: "if_1", TargetNodeID: "end_1", BranchType: FlowEdgeBranchDefault},
		},
	}

	err := ValidateFlowDocument(document)
	if err == nil {
		t.Fatal("条件节点缺少 true/false 分支时应校验失败")
	}
}

func TestValidateFlowDocument_RejectsMissingRequiredNodeField(t *testing.T) {
	document := FlowDocument{
		SchemaVersion: 3,
		Nodes: []FlowNode{
			{NodeID: "start_1", NodeType: NodeTypeStart},
			{NodeID: "if_1", NodeType: NodeTypeConditionIf, Config: map[string]any{}},
			{NodeID: "end_true", NodeType: NodeTypeEnd},
			{NodeID: "end_false", NodeType: NodeTypeEnd},
		},
		Edges: []FlowEdge{
			{EdgeID: "e1", SourceNodeID: "start_1", TargetNodeID: "if_1"},
			{EdgeID: "e2", SourceNodeID: "if_1", TargetNodeID: "end_true", BranchType: FlowEdgeBranchTrue},
			{EdgeID: "e3", SourceNodeID: "if_1", TargetNodeID: "end_false", BranchType: FlowEdgeBranchFalse},
		},
	}

	err := ValidateFlowDocument(document)
	if err == nil {
		t.Fatal("条件节点缺少必填表达式时应校验失败")
	}
}
