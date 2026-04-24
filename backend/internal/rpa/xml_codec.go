package rpa

import (
	"encoding/xml"
	"fmt"
	"sort"
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
	Attrs    []xml.Attr
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
			node := xmlNode{XMLName: value.Name, Attrs: append([]xml.Attr{}, value.Attr...)}
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

func (n xmlNode) MarshalXML(encoder *xml.Encoder, _ xml.StartElement) error {
	start := xml.StartElement{
		Name: n.XMLName,
		Attr: n.Attrs,
	}
	if err := encoder.EncodeToken(start); err != nil {
		return err
	}
	if err := encoder.EncodeToken(start.End()); err != nil {
		return err
	}
	return encoder.Flush()
}

type xmlEdge struct {
	From      string `xml:"from,attr"`
	To        string `xml:"to,attr"`
	Condition string `xml:"condition,attr,omitempty"`
	Label     string `xml:"label,attr,omitempty"`
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
			node.Label = FlowNodeLabel(node.NodeType)
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
			Label:        strings.TrimSpace(item.Label),
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
		encodedNode, err := encodeFlowNode(node)
		if err != nil {
			return "", err
		}
		payload.Nodes = append(payload.Nodes, encodedNode)
	}
	for _, edge := range document.Edges {
		payload.Edges = append(payload.Edges, xmlEdge{
			From:      edge.SourceNodeID,
			To:        edge.TargetNodeID,
			Label:     edge.Label,
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
	catalogItem, ok := FindFlowNodeCatalogItem(nodeType)
	if !ok || !catalogItem.XMLSupported {
		return FlowNode{}, fmt.Errorf("不支持的 XML 节点类型: %s", nodeType)
	}
	node := FlowNode{
		NodeID:   strings.TrimSpace(item.attr("id")),
		NodeType: nodeType,
		Label:    FlowNodeLabel(nodeType),
		Position: FlowPosition{X: parseFloat(item.attr("x")), Y: parseFloat(item.attr("y"))},
		Config:   map[string]any{},
	}
	for _, field := range catalogItem.Fields {
		if field.XMLAttr == "" {
			continue
		}
		raw := strings.TrimSpace(item.attr(field.XMLAttr))
		if raw == "" {
			if field.Required {
				return FlowNode{}, fmt.Errorf("%s 节点 %s 不能为空", nodeType, field.XMLAttr)
			}
			continue
		}
		assignXMLFieldValue(node.Config, field, raw)
	}
	return node, nil
}

func encodeFlowNode(node FlowNode) (xmlNode, error) {
	catalogItem, ok := FindFlowNodeCatalogItem(node.NodeType)
	if !ok || !catalogItem.XMLSupported {
		return xmlNode{}, fmt.Errorf("节点类型不支持 XML 编码: %s", node.NodeType)
	}
	attrMap := map[string]string{
		"id": node.NodeID,
		"x":  formatFloat(node.Position.X),
		"y":  formatFloat(node.Position.Y),
	}
	for _, field := range catalogItem.Fields {
		if field.XMLAttr == "" {
			continue
		}
		value, ok := encodeXMLFieldValue(node.Config, field)
		if !ok || value == "" {
			continue
		}
		attrMap[field.XMLAttr] = value
	}
	attrs := make([]xml.Attr, 0, len(attrMap))
	for _, name := range sortedAttrKeys(attrMap) {
		if attrMap[name] == "" {
			continue
		}
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: name}, Value: attrMap[name]})
	}
	return xmlNode{
		XMLName: xml.Name{Local: string(node.NodeType)},
		Attrs:   attrs,
	}, nil
}

func (n xmlNode) attr(name string) string {
	for _, attr := range n.Attrs {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}

func assignXMLFieldValue(config map[string]any, field FlowNodeField, raw string) {
	switch field.Storage {
	case FieldStorageNumber:
		config[field.Key] = parseFloat(raw)
	case FieldStorageStringListFirst:
		config[field.Key] = []any{raw}
	default:
		config[field.Key] = raw
	}
}

func encodeXMLFieldValue(config map[string]any, field FlowNodeField) (string, bool) {
	switch field.Storage {
	case FieldStorageNumber:
		value, ok := numberConfig(config[field.Key])
		if !ok {
			return "", false
		}
		return formatFloat(value), true
	case FieldStorageStringListFirst:
		values := stringSliceConfig(config[field.Key])
		if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
			return "", false
		}
		return strings.TrimSpace(values[0]), true
	default:
		value := strings.TrimSpace(stringConfig(config[field.Key]))
		return value, value != ""
	}
}

func sortedAttrKeys(attrMap map[string]string) []string {
	keys := make([]string, 0, len(attrMap))
	for key := range attrMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
