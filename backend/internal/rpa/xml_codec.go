package rpa

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type xmlFlow struct {
	XMLName       xml.Name      `xml:"flow"`
	SchemaVersion int           `xml:"schemaVersion,attr"`
	Name          string        `xml:"name,attr"`
	Variables     []xmlVariable `xml:"variables>var"`
	Nodes         xmlNodes      `xml:"nodes"`
	Edges         []xmlEdge     `xml:"edges>edge"`
}

type xmlVariable struct {
	Name         string `xml:"name,attr"`
	Type         string `xml:"type,attr"`
	DefaultValue string `xml:"default,attr"`
}

type xmlNode struct {
	XMLName  xml.Name
	ID       string `xml:"id,attr"`
	X        string `xml:"x,attr"`
	Y        string `xml:"y,attr"`
	URL      string `xml:"url,attr"`
	Duration string `xml:"durationMs,attr"`
}

type xmlNodes []xmlNode

func (n *xmlNodes) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	items := make([]xmlNode, 0)
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch value := token.(type) {
		case xml.StartElement:
			node := xmlNode{XMLName: value.Name}
			for _, attr := range value.Attr {
				switch attr.Name.Local {
				case "id":
					node.ID = attr.Value
				case "x":
					node.X = attr.Value
				case "y":
					node.Y = attr.Value
				case "url":
					node.URL = attr.Value
				case "durationMs":
					node.Duration = attr.Value
				}
			}
			if err := decoder.Skip(); err != nil {
				return err
			}
			items = append(items, node)
		case xml.EndElement:
			if value.Name.Local == start.Name.Local {
				*n = items
				return nil
			}
		}
	}
}

type xmlEdge struct {
	From      string `xml:"from,attr"`
	To        string `xml:"to,attr"`
	Condition string `xml:"condition,attr"`
}

func ParseFlowXML(xmlText string) (*FlowDocument, string, error) {
	text := strings.TrimSpace(xmlText)
	if text == "" {
		return nil, "", fmt.Errorf("XML 不能为空")
	}

	var payload xmlFlow
	if err := xml.Unmarshal([]byte(text), &payload); err != nil {
		return nil, "", fmt.Errorf("XML 解析失败: %w", err)
	}

	document := FlowDocument{
		SchemaVersion: 2,
		Variables:     make([]FlowVariable, 0, len(payload.Variables)),
		Nodes:         make([]FlowNode, 0, len(payload.Nodes)),
		Edges:         make([]FlowEdge, 0, len(payload.Edges)),
	}

	for _, item := range payload.Variables {
		if strings.TrimSpace(item.Name) == "" {
			return nil, "", fmt.Errorf("变量 name 不能为空")
		}
		document.Variables = append(document.Variables, FlowVariable{
			Name:         strings.TrimSpace(item.Name),
			Type:         strings.TrimSpace(item.Type),
			DefaultValue: item.DefaultValue,
		})
	}

	for idx, item := range payload.Nodes {
		node, err := convertXMLNode(item)
		if err != nil {
			return nil, "", err
		}
		if node.Label == "" {
			node.Label = defaultNodeLabel(node.NodeType)
		}
		if node.NodeID == "" {
			return nil, "", fmt.Errorf("节点 id 不能为空")
		}
		if node.Position.X == 0 && node.Position.Y == 0 {
			node.Position = FlowPosition{X: float64(120 + idx*220), Y: 160}
		}
		document.Nodes = append(document.Nodes, node)
	}

	for idx, item := range payload.Edges {
		if strings.TrimSpace(item.From) == "" || strings.TrimSpace(item.To) == "" {
			return nil, "", fmt.Errorf("连线 from/to 不能为空")
		}
		document.Edges = append(document.Edges, FlowEdge{
			EdgeID:       fmt.Sprintf("edge_%d", idx+1),
			SourceNodeID: strings.TrimSpace(item.From),
			TargetNodeID: strings.TrimSpace(item.To),
			Condition:    strings.TrimSpace(item.Condition),
		})
	}

	document = normalizeDocument(document)
	if err := ValidateFlowDocument(document); err != nil {
		return nil, "", err
	}
	return &document, text, nil
}

func EncodeFlowXML(flow *Flow) (string, error) {
	if flow == nil {
		return "", fmt.Errorf("流程不能为空")
	}
	document := normalizeDocument(flow.Document)
	if err := ValidateFlowDocument(document); err != nil {
		return "", err
	}

	payload := xmlFlow{
		SchemaVersion: 1,
		Name:          strings.TrimSpace(flow.FlowName),
		Variables:     make([]xmlVariable, 0, len(document.Variables)),
		Nodes:         make([]xmlNode, 0, len(document.Nodes)),
		Edges:         make([]xmlEdge, 0, len(document.Edges)),
	}
	for _, item := range document.Variables {
		payload.Variables = append(payload.Variables, xmlVariable{
			Name:         item.Name,
			Type:         item.Type,
			DefaultValue: item.DefaultValue,
		})
	}
	for _, node := range document.Nodes {
		durationMs, _ := numberConfig(node.Config["durationMs"])
		payload.Nodes = append(payload.Nodes, xmlNode{
			XMLName:  xml.Name{Local: string(node.NodeType)},
			ID:       node.NodeID,
			X:        formatFloat(node.Position.X),
			Y:        formatFloat(node.Position.Y),
			URL:      stringConfig(node.Config["url"]),
			Duration: formatFloat(durationMs),
		})
	}
	for _, edge := range document.Edges {
		payload.Edges = append(payload.Edges, xmlEdge{
			From:      edge.SourceNodeID,
			To:        edge.TargetNodeID,
			Condition: edge.Condition,
		})
	}

	data, err := xml.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", fmt.Errorf("XML 编码失败: %w", err)
	}
	return xml.Header + string(data), nil
}

func convertXMLNode(item xmlNode) (FlowNode, error) {
	nodeType := FlowNodeType(strings.TrimSpace(item.XMLName.Local))
	node := FlowNode{
		NodeID:   strings.TrimSpace(item.ID),
		NodeType: nodeType,
		Label:    defaultNodeLabel(nodeType),
		Position: FlowPosition{X: parseFloat(item.X), Y: parseFloat(item.Y)},
		Config:   map[string]any{},
	}
	switch nodeType {
	case NodeTypeStart, NodeTypeEnd, NodeTypeBrowserStop:
		return node, nil
	case NodeTypeBrowserOpenURL:
		if strings.TrimSpace(item.URL) == "" {
			return FlowNode{}, fmt.Errorf("browser.open_url 节点 url 不能为空")
		}
		node.Config["url"] = strings.TrimSpace(item.URL)
		return node, nil
	case NodeTypeDelay:
		node.Config["durationMs"] = parseFloat(item.Duration)
		return node, nil
	case NodeTypeBrowserStart:
		if strings.TrimSpace(item.URL) != "" {
			node.Config["startUrls"] = []any{strings.TrimSpace(item.URL)}
		}
		return node, nil
	default:
		return FlowNode{}, fmt.Errorf("不支持的 XML 节点类型: %s", nodeType)
	}
}

func defaultNodeLabel(nodeType FlowNodeType) string {
	switch nodeType {
	case NodeTypeStart:
		return "开始"
	case NodeTypeEnd:
		return "结束"
	case NodeTypeBrowserStart:
		return "启动浏览器"
	case NodeTypeBrowserOpenURL:
		return "打开页面"
	case NodeTypeDelay:
		return "等待"
	case NodeTypeBrowserStop:
		return "关闭浏览器"
	default:
		return string(nodeType)
	}
}
