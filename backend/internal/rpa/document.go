package rpa

import "strings"

type FlowSourceType string

const (
	FlowSourceVisual    FlowSourceType = "visual"
	FlowSourceXMLImport FlowSourceType = "xml_import"
)

type FlowNodeType string

const (
	NodeTypeStart           FlowNodeType = "start"
	NodeTypeEnd             FlowNodeType = "end"
	NodeTypeBrowserStart    FlowNodeType = "browser.start"
	NodeTypeBrowserOpenURL  FlowNodeType = "browser.open_url"
	NodeTypeBrowserClick    FlowNodeType = "browser.click"
	NodeTypeBrowserInput    FlowNodeType = "browser.input"
	NodeTypeBrowserReadText FlowNodeType = "browser.read_text"
	NodeTypeDelay           FlowNodeType = "delay"
	NodeTypeBrowserStop     FlowNodeType = "browser.stop"
	NodeTypeConditionIf     FlowNodeType = "control.if"
	NodeTypeRetry           FlowNodeType = "control.retry"
	NodeTypeParallel        FlowNodeType = "control.parallel"
	NodeTypeFail            FlowNodeType = "control.fail"
	NodeTypeSystemNotify    FlowNodeType = "system.notify"
)

type FlowXMLImportInput struct {
	FlowName string `json:"flowName"`
	XMLText  string `json:"xmlText"`
	GroupID  string `json:"groupId"`
}

type FlowPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type FlowVariable struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	DefaultValue string `json:"defaultValue"`
}

type FlowNode struct {
	NodeID   string         `json:"nodeId"`
	NodeType FlowNodeType   `json:"nodeType"`
	Label    string         `json:"label"`
	Position FlowPosition   `json:"position"`
	Config   map[string]any `json:"config"`
}

type FlowEdgeBranchType string

const (
	FlowEdgeBranchDefault FlowEdgeBranchType = "default"
	FlowEdgeBranchTrue    FlowEdgeBranchType = "true"
	FlowEdgeBranchFalse   FlowEdgeBranchType = "false"
	FlowEdgeBranchOnError FlowEdgeBranchType = "on_error"
)

type FlowEdge struct {
	EdgeID       string             `json:"edgeId"`
	SourceNodeID string             `json:"sourceNodeId"`
	TargetNodeID string             `json:"targetNodeId"`
	Label        string             `json:"label"`
	BranchType   FlowEdgeBranchType `json:"branchType"`
	Condition    string             `json:"condition"`
}

type FlowDocument struct {
	SchemaVersion int            `json:"schemaVersion"`
	Variables     []FlowVariable `json:"variables"`
	Nodes         []FlowNode     `json:"nodes"`
	Edges         []FlowEdge     `json:"edges"`
}

func defaultFlowDocument() FlowDocument {
	return FlowDocument{
		SchemaVersion: 2,
		Variables:     []FlowVariable{},
		Nodes: []FlowNode{
			newFlowNode("start_1", NodeTypeStart, "开始", 120, 160, nil),
			newFlowNode("end_1", NodeTypeEnd, "结束", 520, 160, nil),
		},
		Edges: []FlowEdge{
			{
				EdgeID:       "edge_start_end",
				SourceNodeID: "start_1",
				TargetNodeID: "end_1",
			},
		},
	}
}

func newFlowNode(nodeID string, nodeType FlowNodeType, label string, x float64, y float64, config map[string]any) FlowNode {
	if config == nil {
		config = map[string]any{}
	}
	return FlowNode{
		NodeID:   nodeID,
		NodeType: nodeType,
		Label:    label,
		Position: FlowPosition{X: x, Y: y},
		Config:   config,
	}
}

func cloneDocument(input FlowDocument) FlowDocument {
	out := FlowDocument{
		SchemaVersion: input.SchemaVersion,
		Variables:     append([]FlowVariable{}, input.Variables...),
		Nodes:         make([]FlowNode, 0, len(input.Nodes)),
		Edges:         append([]FlowEdge{}, input.Edges...),
	}
	for _, node := range input.Nodes {
		out.Nodes = append(out.Nodes, FlowNode{
			NodeID:   node.NodeID,
			NodeType: node.NodeType,
			Label:    node.Label,
			Position: node.Position,
			Config:   cloneMap(node.Config),
		})
	}
	return out
}

func normalizeDocument(document FlowDocument) FlowDocument {
	if document.SchemaVersion <= 0 {
		document.SchemaVersion = 2
	}
	if document.Variables == nil {
		document.Variables = []FlowVariable{}
	}
	if document.Nodes == nil {
		document.Nodes = []FlowNode{}
	}
	if document.Edges == nil {
		document.Edges = []FlowEdge{}
	}
	for idx := range document.Nodes {
		document.Nodes[idx].NodeID = strings.TrimSpace(document.Nodes[idx].NodeID)
		document.Nodes[idx].Label = strings.TrimSpace(document.Nodes[idx].Label)
		if document.Nodes[idx].Config == nil {
			document.Nodes[idx].Config = map[string]any{}
		}
	}
	for idx := range document.Edges {
		document.Edges[idx] = normalizeFlowEdge(document.Edges[idx])
	}
	return document
}
