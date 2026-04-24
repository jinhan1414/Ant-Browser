package rpa

import (
	"strings"
	"testing"
)

func TestBuildFlowNodeCatalogPayload_ContainsPromptAndVisibleNodes(t *testing.T) {
	payload := BuildFlowNodeCatalogPayload()
	if len(payload.Items) == 0 {
		t.Fatal("节点目录不能为空")
	}
	if strings.TrimSpace(payload.XMLPromptTemplate) == "" {
		t.Fatal("AI 提示词不能为空")
	}

	required := []FlowNodeType{
		NodeTypeStart,
		NodeTypeEnd,
		NodeTypeBrowserStart,
		NodeTypeBrowserOpenURL,
		NodeTypeBrowserClick,
		NodeTypeBrowserInput,
		NodeTypeBrowserReadText,
		NodeTypeConditionIf,
		NodeTypeDelay,
		NodeTypeBrowserStop,
		NodeTypeSystemNotify,
		NodeTypeFail,
	}
	for _, nodeType := range required {
		if _, ok := FindFlowNodeCatalogItem(nodeType); !ok {
			t.Fatalf("节点目录缺少类型: %s", nodeType)
		}
		if !strings.Contains(payload.XMLPromptTemplate, "<"+string(nodeType)) {
			t.Fatalf("AI 提示词未包含节点: %s", nodeType)
		}
	}
}

func TestParseAndEncodeFlowXML_SupportsAdvancedNodes(t *testing.T) {
	xmlText := `<flow schemaVersion="1" name="高级流程">
  <variables>
    <var name="accountEmail" type="string" default="demo@example.com" />
  </variables>
  <nodes>
    <start id="start_1" x="80" y="120" />
    <browser.start id="browser_start_1" x="220" y="120" url="https://accounts.google.com/" />
    <browser.click id="click_1" x="380" y="120" selector="#next" timeoutMs="1500" maxAttempts="2" intervalMs="300" />
    <browser.input id="input_1" x="560" y="120" selector="#identifierId" value="${accountEmail}" />
    <browser.read_text id="read_1" x="760" y="120" selector="#status" saveAs="accountStatus" />
    <control.if id="if_1" x="960" y="120" expression="contains(accountStatus, &quot;需要验证&quot;)" />
    <system.notify id="notify_1" x="1160" y="60" title="Google 账号异常" body="实例 ${accountEmail} 状态：${accountStatus}" />
    <control.fail id="fail_1" x="1160" y="180" message="Google 账号异常" />
    <end id="end_1" x="1360" y="120" />
  </nodes>
  <edges>
    <edge from="start_1" to="browser_start_1" label="启动实例" />
    <edge from="browser_start_1" to="click_1" />
    <edge from="click_1" to="input_1" />
    <edge from="input_1" to="read_1" />
    <edge from="read_1" to="if_1" />
    <edge from="if_1" to="notify_1" condition="true" label="命中异常" />
    <edge from="if_1" to="end_1" condition="false" label="状态正常" />
    <edge from="notify_1" to="fail_1" condition="on_error" label="通知失败" />
    <edge from="fail_1" to="end_1" />
  </edges>
</flow>`

	document, _, err := ParseFlowXML(xmlText)
	if err != nil {
		t.Fatalf("解析高级 XML 失败: %v", err)
	}
	if got := len(document.Nodes); got != 9 {
		t.Fatalf("节点数错误: %d", got)
	}
	if got := stringConfig(document.Nodes[2].Config["selector"]); got != "#next" {
		t.Fatalf("点击节点 selector 错误: %s", got)
	}
	if got := stringConfig(document.Nodes[3].Config["value"]); got != "${accountEmail}" {
		t.Fatalf("输入节点 value 错误: %s", got)
	}
	if got := stringConfig(document.Nodes[4].Config["saveAs"]); got != "accountStatus" {
		t.Fatalf("读取节点 saveAs 错误: %s", got)
	}
	if got := stringConfig(document.Nodes[5].Config["expression"]); got == "" {
		t.Fatal("条件节点 expression 不能为空")
	}
	if got := stringConfig(document.Nodes[6].Config["title"]); got != "Google 账号异常" {
		t.Fatalf("通知节点 title 错误: %s", got)
	}
	if got := stringConfig(document.Nodes[7].Config["message"]); got != "Google 账号异常" {
		t.Fatalf("失败节点 message 错误: %s", got)
	}
	if got := document.Edges[0].Label; got != "启动实例" {
		t.Fatalf("普通边 label 错误: %s", got)
	}
	if got := document.Edges[5].BranchType; got != FlowEdgeBranchTrue {
		t.Fatalf("条件真分支类型错误: %s", got)
	}
	if got := document.Edges[6].BranchType; got != FlowEdgeBranchFalse {
		t.Fatalf("条件假分支类型错误: %s", got)
	}
	if got := document.Edges[7].BranchType; got != FlowEdgeBranchOnError {
		t.Fatalf("异常分支类型错误: %s", got)
	}

	encoded, err := EncodeFlowXML(&Flow{
		FlowName: "高级流程",
		Document: *document,
	})
	if err != nil {
		t.Fatalf("编码高级 XML 失败: %v", err)
	}
	for _, token := range []string{
		"<browser.click",
		"selector=\"#next\"",
		"<browser.input",
		"value=\"${accountEmail}\"",
		"<browser.read_text",
		"saveAs=\"accountStatus\"",
		"<control.if",
		"expression=\"contains(accountStatus, &#34;需要验证&#34;)\"",
		"<system.notify",
		"title=\"Google 账号异常\"",
		"<control.fail",
		"message=\"Google 账号异常\"",
		"<edge from=\"start_1\" to=\"browser_start_1\" label=\"启动实例\"></edge>",
		"condition=\"true\" label=\"命中异常\"",
		"condition=\"on_error\" label=\"通知失败\"",
	} {
		if !strings.Contains(encoded, token) {
			t.Fatalf("编码结果缺少片段: %s\n%s", token, encoded)
		}
	}
}
