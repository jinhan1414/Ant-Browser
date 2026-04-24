package rpa

type FlowNodeFieldKind string

const (
	FieldKindString FlowNodeFieldKind = "string"
	FieldKindNumber FlowNodeFieldKind = "number"
)

type FlowNodeFieldStorage string

const (
	FieldStorageString          FlowNodeFieldStorage = "string"
	FieldStorageNumber          FlowNodeFieldStorage = "number"
	FieldStorageStringListFirst FlowNodeFieldStorage = "string_list_first"
)

type FlowNodeField struct {
	Key          string               `json:"key"`
	Label        string               `json:"label"`
	Kind         FlowNodeFieldKind    `json:"kind"`
	Storage      FlowNodeFieldStorage `json:"storage"`
	Required     bool                 `json:"required"`
	Hint         string               `json:"hint"`
	Placeholder  string               `json:"placeholder"`
	XMLAttr      string               `json:"xmlAttr"`
	PromptSample string               `json:"promptSample"`
	DefaultValue any                  `json:"defaultValue"`
	MinValue     float64              `json:"minValue"`
	Multiline    bool                 `json:"multiline"`
}

type FlowNodeCatalogItem struct {
	NodeType         FlowNodeType    `json:"nodeType"`
	Label            string          `json:"label"`
	Category         string          `json:"category"`
	Description      string          `json:"description"`
	Palette          bool            `json:"palette"`
	Fixed            bool            `json:"fixed"`
	XMLSupported     bool            `json:"xmlSupported"`
	PromptEnabled    bool            `json:"promptEnabled"`
	RuntimeSupported bool            `json:"runtimeSupported"`
	AllowIncoming    bool            `json:"allowIncoming"`
	AllowOutgoing    bool            `json:"allowOutgoing"`
	MaxOutgoing      int             `json:"maxOutgoing"`
	SupportsIfBranch bool            `json:"supportsIfBranch"`
	SupportsOnError  bool            `json:"supportsOnError"`
	Fields           []FlowNodeField `json:"fields"`
}

type FlowNodeCatalogPayload struct {
	Items             []FlowNodeCatalogItem `json:"items"`
	XMLPromptTemplate string                `json:"xmlPromptTemplate"`
}

var flowNodeCatalog = []FlowNodeCatalogItem{
	{NodeType: NodeTypeStart, Label: "开始", Category: "base", Description: "流程入口", Fixed: true, XMLSupported: true, PromptEnabled: true, RuntimeSupported: true, AllowOutgoing: true, MaxOutgoing: 1},
	{NodeType: NodeTypeEnd, Label: "结束", Category: "base", Description: "流程结束", Fixed: true, XMLSupported: true, PromptEnabled: true, RuntimeSupported: true, AllowIncoming: true},
	{NodeType: NodeTypeBrowserStart, Label: "启动浏览器", Category: "browser", Description: "启动实例并可打开初始页面", Palette: true, XMLSupported: true, PromptEnabled: true, RuntimeSupported: true, AllowIncoming: true, AllowOutgoing: true, MaxOutgoing: 2, SupportsOnError: true, Fields: []FlowNodeField{
		{Key: "startUrls", Label: "启动地址", Kind: FieldKindString, Storage: FieldStorageStringListFirst, XMLAttr: "url", Placeholder: "https://accounts.google.com/", PromptSample: "https://accounts.google.com/"},
	}},
	{NodeType: NodeTypeBrowserOpenURL, Label: "打开页面", Category: "browser", Description: "在当前浏览器中打开页面", Palette: true, XMLSupported: true, PromptEnabled: true, RuntimeSupported: true, AllowIncoming: true, AllowOutgoing: true, MaxOutgoing: 2, SupportsOnError: true, Fields: []FlowNodeField{
		{Key: "url", Label: "页面地址", Kind: FieldKindString, Storage: FieldStorageString, Required: true, XMLAttr: "url", Placeholder: "https://accounts.google.com/", PromptSample: "https://example.com"},
	}},
	{NodeType: NodeTypeBrowserClick, Label: "点击元素", Category: "browser", Description: "等待元素可见后点击", Palette: true, XMLSupported: true, PromptEnabled: true, RuntimeSupported: true, AllowIncoming: true, AllowOutgoing: true, MaxOutgoing: 2, SupportsOnError: true, Fields: []FlowNodeField{
		{Key: "selector", Label: "元素选择器", Kind: FieldKindString, Storage: FieldStorageString, Required: true, XMLAttr: "selector", Placeholder: "#next", PromptSample: "#next"},
		{Key: "timeoutMs", Label: "等待超时毫秒", Kind: FieldKindNumber, Storage: FieldStorageNumber, XMLAttr: "timeoutMs", DefaultValue: 3000, MinValue: 0, PromptSample: "3000"},
		{Key: "maxAttempts", Label: "最大尝试次数", Kind: FieldKindNumber, Storage: FieldStorageNumber, XMLAttr: "maxAttempts", DefaultValue: 1, MinValue: 1, PromptSample: "2"},
		{Key: "intervalMs", Label: "重试间隔毫秒", Kind: FieldKindNumber, Storage: FieldStorageNumber, XMLAttr: "intervalMs", DefaultValue: 0, MinValue: 0, PromptSample: "300"},
	}},
	{NodeType: NodeTypeBrowserInput, Label: "输入文本", Category: "browser", Description: "向页面元素写入文本", Palette: true, XMLSupported: true, PromptEnabled: true, RuntimeSupported: true, AllowIncoming: true, AllowOutgoing: true, MaxOutgoing: 2, SupportsOnError: true, Fields: []FlowNodeField{
		{Key: "selector", Label: "元素选择器", Kind: FieldKindString, Storage: FieldStorageString, Required: true, XMLAttr: "selector", Placeholder: "#identifierId", PromptSample: "#identifierId"},
		{Key: "value", Label: "输入内容", Kind: FieldKindString, Storage: FieldStorageString, XMLAttr: "value", Hint: "支持 ${变量名} 模板", Placeholder: "${accountEmail}", PromptSample: "${accountEmail}"},
		{Key: "timeoutMs", Label: "等待超时毫秒", Kind: FieldKindNumber, Storage: FieldStorageNumber, XMLAttr: "timeoutMs", DefaultValue: 3000, MinValue: 0, PromptSample: "3000"},
		{Key: "maxAttempts", Label: "最大尝试次数", Kind: FieldKindNumber, Storage: FieldStorageNumber, XMLAttr: "maxAttempts", DefaultValue: 1, MinValue: 1, PromptSample: "2"},
		{Key: "intervalMs", Label: "重试间隔毫秒", Kind: FieldKindNumber, Storage: FieldStorageNumber, XMLAttr: "intervalMs", DefaultValue: 0, MinValue: 0, PromptSample: "300"},
	}},
	{NodeType: NodeTypeBrowserReadText, Label: "读取文本", Category: "browser", Description: "读取元素文本并保存到变量", Palette: true, XMLSupported: true, PromptEnabled: true, RuntimeSupported: true, AllowIncoming: true, AllowOutgoing: true, MaxOutgoing: 2, SupportsOnError: true, Fields: []FlowNodeField{
		{Key: "selector", Label: "元素选择器", Kind: FieldKindString, Storage: FieldStorageString, Required: true, XMLAttr: "selector", Placeholder: "#status", PromptSample: "#status"},
		{Key: "saveAs", Label: "保存变量名", Kind: FieldKindString, Storage: FieldStorageString, XMLAttr: "saveAs", Hint: "后续节点可通过 ${变量名} 使用", Placeholder: "accountStatus", PromptSample: "accountStatus"},
		{Key: "timeoutMs", Label: "等待超时毫秒", Kind: FieldKindNumber, Storage: FieldStorageNumber, XMLAttr: "timeoutMs", DefaultValue: 3000, MinValue: 0, PromptSample: "3000"},
		{Key: "maxAttempts", Label: "最大尝试次数", Kind: FieldKindNumber, Storage: FieldStorageNumber, XMLAttr: "maxAttempts", DefaultValue: 1, MinValue: 1, PromptSample: "2"},
		{Key: "intervalMs", Label: "重试间隔毫秒", Kind: FieldKindNumber, Storage: FieldStorageNumber, XMLAttr: "intervalMs", DefaultValue: 0, MinValue: 0, PromptSample: "300"},
	}},
	{NodeType: NodeTypeConditionIf, Label: "条件判断", Category: "control", Description: "按表达式走 true/false 分支", Palette: true, XMLSupported: true, PromptEnabled: true, RuntimeSupported: true, AllowIncoming: true, AllowOutgoing: true, MaxOutgoing: 2, SupportsIfBranch: true, Fields: []FlowNodeField{
		{Key: "expression", Label: "判断表达式", Kind: FieldKindString, Storage: FieldStorageString, Required: true, XMLAttr: "expression", Hint: "支持 ==、!=、contains()、empty()", Placeholder: `contains(accountStatus, "需要验证")`, PromptSample: `contains(accountStatus, "需要验证")`},
	}},
	{NodeType: NodeTypeDelay, Label: "等待", Category: "control", Description: "固定等待一段时间", Palette: true, XMLSupported: true, PromptEnabled: true, RuntimeSupported: true, AllowIncoming: true, AllowOutgoing: true, MaxOutgoing: 1, Fields: []FlowNodeField{
		{Key: "durationMs", Label: "等待毫秒", Kind: FieldKindNumber, Storage: FieldStorageNumber, XMLAttr: "durationMs", DefaultValue: 1000, MinValue: 1, PromptSample: "1000"},
	}},
	{NodeType: NodeTypeBrowserStop, Label: "关闭浏览器", Category: "browser", Description: "关闭当前实例", Palette: true, XMLSupported: true, PromptEnabled: true, RuntimeSupported: true, AllowIncoming: true, AllowOutgoing: true, MaxOutgoing: 2, SupportsOnError: true},
	{NodeType: NodeTypeSystemNotify, Label: "发送通知", Category: "system", Description: "发送 Windows 消息中心通知", Palette: true, XMLSupported: true, PromptEnabled: true, RuntimeSupported: true, AllowIncoming: true, AllowOutgoing: true, MaxOutgoing: 2, SupportsOnError: true, Fields: []FlowNodeField{
		{Key: "title", Label: "通知标题", Kind: FieldKindString, Storage: FieldStorageString, XMLAttr: "title", Placeholder: "Google 账号异常", PromptSample: "Google 账号异常"},
		{Key: "body", Label: "通知内容", Kind: FieldKindString, Storage: FieldStorageString, XMLAttr: "body", Hint: "支持 ${变量名} 模板", Placeholder: "实例 ${profileId} 状态：${accountStatus}", PromptSample: "实例 ${profileId} 状态：${accountStatus}", Multiline: true},
		{Key: "maxAttempts", Label: "最大尝试次数", Kind: FieldKindNumber, Storage: FieldStorageNumber, XMLAttr: "maxAttempts", DefaultValue: 1, MinValue: 1, PromptSample: "2"},
		{Key: "intervalMs", Label: "重试间隔毫秒", Kind: FieldKindNumber, Storage: FieldStorageNumber, XMLAttr: "intervalMs", DefaultValue: 0, MinValue: 0, PromptSample: "300"},
	}},
	{NodeType: NodeTypeFail, Label: "异常终止", Category: "control", Description: "主动抛出异常结束流程", Palette: true, XMLSupported: true, PromptEnabled: true, RuntimeSupported: true, AllowIncoming: true, AllowOutgoing: true, MaxOutgoing: 2, SupportsOnError: true, Fields: []FlowNodeField{
		{Key: "message", Label: "异常消息", Kind: FieldKindString, Storage: FieldStorageString, XMLAttr: "message", Placeholder: "Google 账号异常", PromptSample: "Google 账号异常", Multiline: true},
	}},
	{NodeType: NodeTypeRetry, Label: "重试块", Category: "control", Description: "后续用于结构化重试", XMLSupported: false, PromptEnabled: false, AllowIncoming: true, AllowOutgoing: true, MaxOutgoing: 1},
	{NodeType: NodeTypeParallel, Label: "并发块", Category: "control", Description: "后续用于并发执行", XMLSupported: false, PromptEnabled: false, AllowIncoming: true, AllowOutgoing: true, MaxOutgoing: 8},
}

func BuildFlowNodeCatalogPayload() FlowNodeCatalogPayload {
	return FlowNodeCatalogPayload{
		Items:             ListFlowNodeCatalog(),
		XMLPromptTemplate: BuildFlowXMLPromptTemplate(),
	}
}

func ListFlowNodeCatalog() []FlowNodeCatalogItem {
	items := make([]FlowNodeCatalogItem, 0, len(flowNodeCatalog))
	for _, item := range flowNodeCatalog {
		items = append(items, cloneFlowNodeCatalogItem(item))
	}
	return items
}

func FindFlowNodeCatalogItem(nodeType FlowNodeType) (FlowNodeCatalogItem, bool) {
	for _, item := range flowNodeCatalog {
		if item.NodeType == nodeType {
			return cloneFlowNodeCatalogItem(item), true
		}
	}
	return FlowNodeCatalogItem{}, false
}

func FlowNodeLabel(nodeType FlowNodeType) string {
	item, ok := FindFlowNodeCatalogItem(nodeType)
	if ok && item.Label != "" {
		return item.Label
	}
	return string(nodeType)
}

func cloneFlowNodeCatalogItem(item FlowNodeCatalogItem) FlowNodeCatalogItem {
	item.Fields = append([]FlowNodeField{}, item.Fields...)
	return item
}
