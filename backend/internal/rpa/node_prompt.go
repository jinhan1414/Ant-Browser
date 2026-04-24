package rpa

import (
	"fmt"
	"strings"
)

const flowXMLPromptExample = `<flow schemaVersion="1" name="检查 Google 账号">
  <variables>
    <var name="accountEmail" type="string" default="demo@example.com" />
  </variables>
  <nodes>
    <start id="start_1" x="80" y="120" />
    <browser.open_url id="open_1" x="280" y="120" url="https://accounts.google.com/" />
    <end id="end_1" x="520" y="120" />
  </nodes>
  <edges>
    <edge from="start_1" to="open_1" />
    <edge from="open_1" to="end_1" />
  </edges>
</flow>`

func BuildFlowXMLPromptTemplate() string {
	lines := []string{
		"你是 RPA 流程 XML 生成助手。请严格输出 AntRPA XML，且只输出 XML，不要附加解释。",
		"",
		"约束：",
		"1. 根节点固定为 <flow schemaVersion=\"1\" name=\"流程名称\">。",
		"2. 可选变量写在 <variables> 下，格式为 <var name=\"\" type=\"\" default=\"\" />。",
		"3. 允许的节点类型如下：",
	}
	for _, item := range flowNodeCatalog {
		if !item.PromptEnabled || !item.XMLSupported {
			continue
		}
		lines = append(lines, fmt.Sprintf("   - %s", formatPromptNodeLine(item)))
	}
	lines = append(lines,
		"4. 连线使用 <edge from=\"\" to=\"\" label=\"\" />；条件分支使用 condition=\"true\" 或 condition=\"false\"；异常分支使用 condition=\"on_error\"。",
		"5. 必须只有一个 start，至少一个 end，所有节点 id 唯一。",
		"6. x 和 y 必须是数字，edge 引用的节点必须存在。",
		"7. 仅输出合法 XML，不要输出 Markdown 代码块。",
		"",
		"示例：",
		flowXMLPromptExample,
	)
	return strings.Join(lines, "\n")
}

func formatPromptNodeLine(item FlowNodeCatalogItem) string {
	attrs := []string{`id=""`, `x=""`, `y=""`}
	required := make([]string, 0, len(item.Fields))
	for _, field := range item.Fields {
		if field.XMLAttr == "" {
			continue
		}
		sample := field.PromptSample
		if sample == "" {
			sample = ""
		}
		attrs = append(attrs, fmt.Sprintf(`%s="%s"`, field.XMLAttr, sample))
		if field.Required {
			required = append(required, field.XMLAttr)
		}
	}
	line := fmt.Sprintf("<%s %s />", item.NodeType, strings.Join(attrs, " "))
	if len(required) == 0 {
		return line
	}
	return fmt.Sprintf("%s 必填属性: %s", line, strings.Join(required, ", "))
}
