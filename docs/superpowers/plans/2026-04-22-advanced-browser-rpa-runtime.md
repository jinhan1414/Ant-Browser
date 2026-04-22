# Advanced Browser RPA Runtime Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有 RPA MVP 基础上，把执行内核升级为可操作真实网页元素的浏览器 RPA，支持点击元素、输入文本、读取页面内容、条件判断、变量传递、重试、异常分支和并发执行。

**Architecture:** 后端继续以 `backend/internal/rpa` 为核心，新增基于现有 Chrome DevTools Protocol 调试端口的浏览器会话层，在执行器内部引入运行时上下文、变量存储、控制流编译和并发调度。前端沿用现有流程画布与节点编辑器，扩展节点类型、节点配置与运行记录展示，不引入第二套编辑模型。

**Tech Stack:** Go, SQLite, Wails, React, TypeScript, Chrome DevTools Protocol, gorilla/websocket

---

## File Structure

- `backend/internal/rpa/cdp_client.go`: 负责通过现有调试端口发现 target、建立 websocket 会话、发送 CDP 命令。
- `backend/internal/rpa/browser_session.go`: 封装页面级操作，如选择器查询、点击、输入、文本读取、等待元素。
- `backend/internal/rpa/runtime_context.go`: 保存变量、步骤输出、重试计数、并发分支结果。
- `backend/internal/rpa/expression.go`: 解析条件表达式和模板变量，如 `${profileId}`、`${steps.read_title.text}`。
- `backend/internal/rpa/compiler.go`: 将流程图编译为可执行控制流，支持条件、异常分支和并发块。
- `backend/internal/rpa/executor.go`: 负责任务级调度、目标实例执行、并发分支合并和运行状态落库。
- `backend/internal/rpa/executor_steps_browser.go`: 浏览器动作节点实现。
- `backend/internal/rpa/executor_steps_control.go`: 条件、重试、异常分支、并发节点实现。
- `backend/internal/rpa/run_dao.go`: 增加步骤级运行明细与并发分支记录。
- `backend/internal/database/sqlite.go`: 新增运行步骤表与迁移。
- `frontend/src/modules/rpa/types.ts`: 扩展节点类型、变量定义、运行步骤结构。
- `frontend/src/modules/rpa/flowDocument.ts`: 支持新节点默认配置和连线规则。
- `frontend/src/modules/rpa/components/FlowNodePalette.tsx`: 暴露浏览器动作和控制流节点。
- `frontend/src/modules/rpa/components/FlowNodeInspector.tsx`: 配置 selector、表达式、变量名、重试参数、并发分支。
- `frontend/src/modules/rpa/pages/RunRecordsPage.tsx`: 展示步骤级运行轨迹和分支状态。

### Task 1: 建立可复用的 CDP 浏览器会话层

**Files:**
- Create: `backend/internal/rpa/cdp_client.go`
- Create: `backend/internal/rpa/browser_session.go`
- Create: `backend/internal/rpa/browser_session_test.go`
- Modify: `backend/internal/rpa/executor.go`
- Reference: `backend/app_cookie.go`
- Reference: `backend/app_instance_errors.go`

- [ ] **Step 1: 先写失败测试，覆盖 target 发现、页面会话建立、元素查询和文本读取。**

```go
func TestBrowserSession_QueryClickAndReadText(t *testing.T) {
	server := newCDPStubServer(t, []stubTarget{{
		TargetID:    "page-1",
		Type:        "page",
		WebSocketURL: "ws://stub/page-1",
	}})
	client := NewCDPClient(server.HTTPURL(), server.Dialer())

	session, err := NewBrowserSession(client)
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}
	if err := session.AttachPage(); err != nil {
		t.Fatalf("附加页面失败: %v", err)
	}
	if _, err := session.QuerySelector("#login"); err != nil {
		t.Fatalf("查询元素失败: %v", err)
	}
	text, err := session.ReadText("#title")
	if err != nil {
		t.Fatalf("读取文本失败: %v", err)
	}
	if text != "控制台首页" {
		t.Fatalf("读取结果错误: %q", text)
	}
}
```

- [ ] **Step 2: 运行测试确认缺失实现时失败。**

Run: `go test ./backend/internal/rpa -run TestBrowserSession -count=1`
Expected: FAIL with `undefined: NewCDPClient` or `undefined: NewBrowserSession`

- [ ] **Step 3: 实现最小 CDP 客户端，复用现有 `/json/list`、`/json/version` 和 websocket 通信方式。**

```go
type CDPClient struct {
	baseURL string
	dialer  websocketDialer
	client  *http.Client
}

func (c *CDPClient) PageTarget() (*CDPTarget, error) {
	resp, err := c.client.Get(c.baseURL + "/json/list")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var targets []CDPTarget
	if err := json.NewDecoder(resp.Body).Decode(&targets); err != nil {
		return nil, err
	}
	for _, target := range targets {
		if target.Type == "page" && target.WebSocketDebuggerURL != "" {
			return &target, nil
		}
	}
	return nil, fmt.Errorf("page target not found")
}
```

- [ ] **Step 4: 实现页面会话接口，至少提供 `WaitVisible`、`Click`、`InputText`、`ReadText`。**

```go
type BrowserSession interface {
	AttachPage() error
	QuerySelector(selector string) (string, error)
	WaitVisible(selector string, timeout time.Duration) error
	Click(selector string) error
	InputText(selector string, value string) error
	ReadText(selector string) (string, error)
}
```

- [ ] **Step 5: 运行测试确认浏览器会话层通过。**

Run: `go test ./backend/internal/rpa -run TestBrowserSession -count=1`
Expected: PASS

- [ ] **Step 6: 提交本任务。**

```bash
git add backend/internal/rpa/cdp_client.go backend/internal/rpa/browser_session.go backend/internal/rpa/browser_session_test.go backend/internal/rpa/executor.go
git commit -m "feat: add cdp browser session for rpa"
```

### Task 2: 扩展流程节点模型和控制流编译器

**Files:**
- Modify: `backend/internal/rpa/document.go`
- Modify: `backend/internal/rpa/compiler.go`
- Modify: `backend/internal/rpa/xml_codec.go`
- Modify: `backend/internal/rpa/types.go`
- Create: `backend/internal/rpa/compiler_test.go`
- Modify: `frontend/src/modules/rpa/types.ts`
- Modify: `frontend/src/modules/rpa/flowDocument.ts`

- [ ] **Step 1: 先写编译器失败测试，覆盖新节点类型、条件边和并发块。**

```go
func TestBuildExecutionPlan_SupportsBranchRetryAndParallel(t *testing.T) {
	document := FlowDocument{
		SchemaVersion: 3,
		Nodes: []FlowNode{
			{NodeID: "start_1", NodeType: NodeTypeStart},
			{NodeID: "read_1", NodeType: NodeTypeBrowserReadText, Config: map[string]any{"selector": "#title", "saveAs": "pageTitle"}},
			{NodeID: "if_1", NodeType: NodeTypeConditionIf, Config: map[string]any{"expression": "pageTitle == \"控制台首页\""}},
			{NodeID: "parallel_1", NodeType: NodeTypeParallel},
			{NodeID: "end_1", NodeType: NodeTypeEnd},
		},
		Edges: []FlowEdge{
			{EdgeID: "e1", SourceNodeID: "start_1", TargetNodeID: "read_1"},
			{EdgeID: "e2", SourceNodeID: "read_1", TargetNodeID: "if_1"},
			{EdgeID: "e3", SourceNodeID: "if_1", TargetNodeID: "parallel_1", Condition: "true"},
			{EdgeID: "e4", SourceNodeID: "parallel_1", TargetNodeID: "end_1"},
		},
	}
	plan, err := BuildExecutionPlan(document)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	if len(plan.Nodes) == 0 {
		t.Fatal("执行计划不能为空")
	}
}
```

- [ ] **Step 2: 运行测试确认现有编译器因“不支持分支”而失败。**

Run: `go test ./backend/internal/rpa -run TestBuildExecutionPlan_SupportsBranchRetryAndParallel -count=1`
Expected: FAIL with `当前版本暂不支持`

- [ ] **Step 3: 增加节点类型常量和节点配置字段。**

```go
const (
	NodeTypeBrowserClick    FlowNodeType = "browser.click"
	NodeTypeBrowserInput    FlowNodeType = "browser.input"
	NodeTypeBrowserReadText FlowNodeType = "browser.read_text"
	NodeTypeConditionIf     FlowNodeType = "control.if"
	NodeTypeRetry           FlowNodeType = "control.retry"
	NodeTypeParallel        FlowNodeType = "control.parallel"
	NodeTypeFail            FlowNodeType = "control.fail"
)
```

- [ ] **Step 4: 重写编译器输出结构，从“线性节点数组”升级为“执行块 + 转移条件”。**

```go
type ExecutionPlan struct {
	EntryNodeID string
	Nodes       map[string]CompiledNode
}

type CompiledNode struct {
	Node    FlowNode
	Next    []CompiledEdge
	OnError []CompiledEdge
}
```

- [ ] **Step 5: 同步前端类型，确保画布能保存 `expression`、`saveAs`、`branches`、`maxAttempts`。**

```ts
export type RPAFlowNodeType =
  | 'start'
  | 'end'
  | 'browser.start'
  | 'browser.open_url'
  | 'browser.click'
  | 'browser.input'
  | 'browser.read_text'
  | 'delay'
  | 'browser.stop'
  | 'control.if'
  | 'control.retry'
  | 'control.parallel'
  | 'control.fail'
```

- [ ] **Step 6: 运行前后端测试确认节点模型和编译器通过。**

Run: `go test ./backend/internal/rpa -run TestBuildExecutionPlan -count=1`
Expected: PASS

Run: `npm run build`
Workdir: `frontend`
Expected: PASS

- [ ] **Step 7: 提交本任务。**

```bash
git add backend/internal/rpa/document.go backend/internal/rpa/compiler.go backend/internal/rpa/xml_codec.go backend/internal/rpa/types.go backend/internal/rpa/compiler_test.go frontend/src/modules/rpa/types.ts frontend/src/modules/rpa/flowDocument.ts
git commit -m "feat: extend rpa graph for browser actions and control flow"
```

### Task 3: 增加运行时上下文、变量系统和表达式求值

**Files:**
- Create: `backend/internal/rpa/runtime_context.go`
- Create: `backend/internal/rpa/expression.go`
- Create: `backend/internal/rpa/runtime_context_test.go`
- Modify: `backend/internal/rpa/executor.go`

- [ ] **Step 1: 先写失败测试，覆盖变量写入、模板插值和布尔表达式判断。**

```go
func TestRuntimeContext_ResolveTemplateAndExpression(t *testing.T) {
	ctx := NewRuntimeContext("profile-a")
	ctx.Set("pageTitle", "控制台首页")
	ctx.Set("username", "alice")

	text, err := ResolveTemplate("欢迎 ${username}", ctx)
	if err != nil {
		t.Fatalf("模板解析失败: %v", err)
	}
	if text != "欢迎 alice" {
		t.Fatalf("模板解析错误: %q", text)
	}
	ok, err := EvalBoolExpression(`pageTitle == "控制台首页"`, ctx)
	if err != nil {
		t.Fatalf("表达式执行失败: %v", err)
	}
	if !ok {
		t.Fatal("条件判断应为 true")
	}
}
```

- [ ] **Step 2: 运行测试确认运行时上下文尚未实现。**

Run: `go test ./backend/internal/rpa -run TestRuntimeContext -count=1`
Expected: FAIL with `undefined: NewRuntimeContext`

- [ ] **Step 3: 实现上下文和变量存储，默认注入 `profileId`、`taskId`、`flowId`。**

```go
type RuntimeContext struct {
	profileID string
	values    map[string]any
}

func NewRuntimeContext(profileID string) *RuntimeContext {
	return &RuntimeContext{
		profileID: profileID,
		values: map[string]any{
			"profileId": profileID,
		},
	}
}
```

- [ ] **Step 4: 实现最小表达式求值器，只支持本期需要的 `==`、`!=`、`contains`、`empty`。**

```go
func EvalBoolExpression(expr string, ctx *RuntimeContext) (bool, error) {
	switch {
	case strings.Contains(expr, "=="):
		left, right, err := splitBinary(expr, "==")
		if err != nil {
			return false, err
		}
		return resolveValue(left, ctx) == trimQuoted(right), nil
	case strings.HasPrefix(expr, "empty("):
		name := strings.TrimSuffix(strings.TrimPrefix(expr, "empty("), ")")
		return fmt.Sprint(ctx.Get(strings.TrimSpace(name))) == "", nil
	default:
		return false, fmt.Errorf("unsupported expression: %s", expr)
	}
}
```

- [ ] **Step 5: 在执行器中为每个目标实例创建独立上下文，后续步骤统一从上下文取值。**

```go
func (e *Executor) executeTarget(runID string, task *Task, flow *Flow, target TaskTarget) *RunTarget {
	ctx := NewRuntimeContext(target.ProfileID)
	ctx.Set("taskId", task.TaskID)
	ctx.Set("flowId", flow.FlowID)
	return e.runPlan(runID, target, ctx)
}
```

- [ ] **Step 6: 运行测试确认变量系统通过。**

Run: `go test ./backend/internal/rpa -run TestRuntimeContext -count=1`
Expected: PASS

- [ ] **Step 7: 提交本任务。**

```bash
git add backend/internal/rpa/runtime_context.go backend/internal/rpa/expression.go backend/internal/rpa/runtime_context_test.go backend/internal/rpa/executor.go
git commit -m "feat: add rpa runtime context and expressions"
```

### Task 4: 落地浏览器动作节点和步骤级运行明细

**Files:**
- Create: `backend/internal/rpa/executor_steps_browser.go`
- Modify: `backend/internal/rpa/executor.go`
- Modify: `backend/internal/rpa/run_dao.go`
- Modify: `backend/internal/rpa/types.go`
- Modify: `backend/internal/database/sqlite.go`
- Create: `backend/internal/rpa/executor_browser_test.go`

- [ ] **Step 1: 先写失败测试，覆盖点击、输入、读取文本和步骤级记录。**

```go
func TestExecutor_BrowserActionNodesPersistRunSteps(t *testing.T) {
	executor := NewExecutor(newBrowserOperatorStub())
	flow := newBrowserActionFlow()
	task := &Task{TaskID: "task-1", ExecutionOrder: TaskExecutionSequential}

	run, targets, err := executor.Execute(task, flow, []TaskTarget{{ProfileID: "profile-a"}})
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if run.Status != RunStatusSuccess {
		t.Fatalf("运行状态错误: %+v", run)
	}
	if len(targets) != 1 {
		t.Fatalf("目标数错误: %d", len(targets))
	}
}
```

- [ ] **Step 2: 运行测试确认浏览器动作节点尚未接入。**

Run: `go test ./backend/internal/rpa -run TestExecutor_BrowserActionNodesPersistRunSteps -count=1`
Expected: FAIL with `不支持的 RPA 节点类型`

- [ ] **Step 3: 增加运行步骤模型和数据库表。**

```go
type RunStep struct {
	RunStepID     string `json:"runStepId"`
	RunID         string `json:"runId"`
	RunTargetID   string `json:"runTargetId"`
	NodeID        string `json:"nodeId"`
	NodeType      string `json:"nodeType"`
	Status        string `json:"status"`
	Attempt       int    `json:"attempt"`
	OutputJSON    string `json:"outputJson"`
	ErrorMessage  string `json:"errorMessage"`
	StartedAt     string `json:"startedAt"`
	FinishedAt    string `json:"finishedAt"`
}
```

- [ ] **Step 4: 实现 `browser.click`、`browser.input`、`browser.read_text`，并支持 `${var}` 模板替换。**

```go
case NodeTypeBrowserClick:
	selector := mustString(node.Config["selector"])
	return nil, session.Click(selector)
case NodeTypeBrowserInput:
	selector := mustString(node.Config["selector"])
	value, err := ResolveTemplate(mustString(node.Config["value"]), ctx)
	if err != nil {
		return nil, err
	}
	return nil, session.InputText(selector, value)
case NodeTypeBrowserReadText:
	text, err := session.ReadText(mustString(node.Config["selector"]))
	if err != nil {
		return nil, err
	}
	ctx.Set(mustString(node.Config["saveAs"]), text)
	return map[string]any{"text": text}, nil
```

- [ ] **Step 5: 每执行一个节点都写一条 `RunStep`，失败时写明 selector、attempt、errorMessage。**

- [ ] **Step 6: 运行测试确认浏览器动作节点和步骤记录通过。**

Run: `go test ./backend/internal/rpa -run TestExecutor_BrowserActionNodesPersistRunSteps -count=1`
Expected: PASS

- [ ] **Step 7: 提交本任务。**

```bash
git add backend/internal/rpa/executor_steps_browser.go backend/internal/rpa/executor.go backend/internal/rpa/run_dao.go backend/internal/rpa/types.go backend/internal/database/sqlite.go backend/internal/rpa/executor_browser_test.go
git commit -m "feat: add browser action nodes and run step logs"
```

### Task 5: 实现条件判断、重试和异常分支

**Files:**
- Create: `backend/internal/rpa/executor_steps_control.go`
- Modify: `backend/internal/rpa/executor.go`
- Create: `backend/internal/rpa/executor_control_test.go`
- Modify: `backend/internal/rpa/compiler.go`

- [ ] **Step 1: 先写失败测试，覆盖 if 判断、重试成功和异常分支跳转。**

```go
func TestExecutor_ControlNodes_IfRetryAndOnError(t *testing.T) {
	executor := NewExecutor(newFlakyBrowserOperatorStub(2))
	flow := newControlFlowDocument()
	task := &Task{TaskID: "task-1", ExecutionOrder: TaskExecutionSequential}

	run, _, err := executor.Execute(task, flow, []TaskTarget{{ProfileID: "profile-a"}})
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if run.Status != RunStatusSuccess {
		t.Fatalf("运行状态错误: %+v", run)
	}
}
```

- [ ] **Step 2: 运行测试确认控制流节点尚未实现。**

Run: `go test ./backend/internal/rpa -run TestExecutor_ControlNodes -count=1`
Expected: FAIL with `unsupported expression` or `不支持的 RPA 节点类型`

- [ ] **Step 3: 实现 `control.if`，根据表达式选择 `true` / `false` 分支。**

```go
func (e *Executor) runIfNode(node FlowNode, ctx *RuntimeContext) (string, error) {
	ok, err := EvalBoolExpression(mustString(node.Config["expression"]), ctx)
	if err != nil {
		return "", err
	}
	if ok {
		return "true", nil
	}
	return "false", nil
}
```

- [ ] **Step 4: 实现 `control.retry`，支持 `maxAttempts`、`intervalMs`、`retryOn`。**

```go
for attempt := 1; attempt <= maxAttempts; attempt++ {
	err = e.executeChild(node, ctx, target)
	if err == nil {
		return nil
	}
	lastErr = err
	time.Sleep(time.Duration(intervalMs) * time.Millisecond)
}
return fmt.Errorf("retry exhausted: %w", lastErr)
```

- [ ] **Step 5: 实现异常边 `on_error`，允许失败后进入补偿节点而不是直接终止。**

- [ ] **Step 6: 运行测试确认控制流节点通过。**

Run: `go test ./backend/internal/rpa -run TestExecutor_ControlNodes -count=1`
Expected: PASS

- [ ] **Step 7: 提交本任务。**

```bash
git add backend/internal/rpa/executor_steps_control.go backend/internal/rpa/executor.go backend/internal/rpa/executor_control_test.go backend/internal/rpa/compiler.go
git commit -m "feat: add retry and error branch control nodes"
```

### Task 6: 实现同一目标实例内的并发分支执行

**Files:**
- Modify: `backend/internal/rpa/executor.go`
- Modify: `backend/internal/rpa/compiler.go`
- Create: `backend/internal/rpa/parallel_executor.go`
- Create: `backend/internal/rpa/parallel_executor_test.go`
- Modify: `backend/internal/rpa/types.go`

- [ ] **Step 1: 先写失败测试，覆盖两个并发分支都成功、其中一个失败以及变量隔离。**

```go
func TestParallelExecutor_RunsBranchesAndMergesOutputs(t *testing.T) {
	executor := NewExecutor(newBrowserOperatorStub())
	flow := newParallelFlowDocument()
	task := &Task{TaskID: "task-parallel", ExecutionOrder: TaskExecutionSequential}

	run, _, err := executor.Execute(task, flow, []TaskTarget{{ProfileID: "profile-a"}})
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if run.Status != RunStatusSuccess {
		t.Fatalf("并发执行结果错误: %+v", run)
	}
}
```

- [ ] **Step 2: 运行测试确认并发块尚未实现。**

Run: `go test ./backend/internal/rpa -run TestParallelExecutor -count=1`
Expected: FAIL with `control.parallel not implemented`

- [ ] **Step 3: 新增并发执行器，使用 `errgroup` 或 `sync.WaitGroup` 跑分支，并复制上下文避免分支互相覆盖。**

```go
func (e *Executor) runParallel(node FlowNode, ctx *RuntimeContext) error {
	branches := parseParallelBranches(node)
	results := make(chan branchResult, len(branches))
	var wg sync.WaitGroup
	for _, branch := range branches {
		wg.Add(1)
		go func(branch ParallelBranch) {
			defer wg.Done()
			branchCtx := ctx.Clone()
			results <- e.runBranch(branch, branchCtx)
		}(branch)
	}
	wg.Wait()
	close(results)
	return mergeBranchResults(ctx, results)
}
```

- [ ] **Step 4: 规定并发合并规则，只允许写入 `saveAs` 前缀不同的变量；命名冲突直接报错。**

- [ ] **Step 5: 运行测试确认并发节点通过。**

Run: `go test ./backend/internal/rpa -run TestParallelExecutor -count=1`
Expected: PASS

- [ ] **Step 6: 提交本任务。**

```bash
git add backend/internal/rpa/executor.go backend/internal/rpa/compiler.go backend/internal/rpa/parallel_executor.go backend/internal/rpa/parallel_executor_test.go backend/internal/rpa/types.go
git commit -m "feat: add parallel branches for rpa runtime"
```

### Task 7: 扩展 App/Wails 接口和前端流程编辑器

**Files:**
- Modify: `backend/app_rpa.go`
- Modify: `backend/app_rpa_test.go`
- Modify: `frontend/src/modules/rpa/types.ts`
- Modify: `frontend/src/modules/rpa/api.ts`
- Modify: `frontend/src/modules/rpa/components/FlowNodePalette.tsx`
- Modify: `frontend/src/modules/rpa/components/FlowNodeInspector.tsx`
- Modify: `frontend/src/modules/rpa/components/FlowEditorModal.tsx`
- Modify: `frontend/src/modules/rpa/pages/RunRecordsPage.tsx`
- Modify: `frontend/src/wailsjs/go/models.ts`
- Modify: `frontend/src/wailsjs/go/main/App.d.ts`
- Modify: `frontend/src/wailsjs/go/main/App.js`

- [ ] **Step 1: 先写 App 层测试，覆盖步骤级记录查询和新节点保存。**

```go
func TestAppRPARunStepList(t *testing.T) {
	app := newTestRPAApp(t)
	flow := mustCreateAdvancedFlow(t, app)
	task := mustCreateAdvancedTask(t, app, flow.FlowID)

	run, err := app.RPATaskExecute(task.TaskID)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	steps, err := app.RPARunStepList(run.RunID)
	if err != nil {
		t.Fatalf("查询步骤失败: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("步骤记录不能为空")
	}
}
```

- [ ] **Step 2: 运行测试确认 App/Wails 暴露尚未覆盖新能力。**

Run: `go test ./backend -run TestAppRPARunStepList -count=1`
Expected: FAIL with `app.RPARunStepList undefined`

- [ ] **Step 3: 暴露步骤记录查询接口，并保证前端拿到 `outputJson`、`attempt`、`nodeType`。**

```go
func (a *App) RPARunStepList(runID string) ([]*rpa.RunStep, error) {
	if a.rpaSvc == nil {
		return nil, fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.ListRunSteps(runID)
}
```

- [ ] **Step 4: 扩展前端节点面板和属性面板，新增 selector、value 模板、saveAs、expression、retry 参数、并发分支配置。**

```ts
{ value: 'browser.click', label: '点击元素' },
{ value: 'browser.input', label: '输入文本' },
{ value: 'browser.read_text', label: '读取文本' },
{ value: 'control.if', label: '条件判断' },
{ value: 'control.retry', label: '重试块' },
{ value: 'control.parallel', label: '并发分支' }
```

- [ ] **Step 5: 在运行记录页增加步骤明细表，显示节点类型、尝试次数、输出、错误信息。**

- [ ] **Step 6: 运行测试和构建确认前后端接口对齐。**

Run: `go test ./backend -run TestAppRPARunStepList -count=1`
Expected: PASS

Run: `npm run build`
Workdir: `frontend`
Expected: PASS

- [ ] **Step 7: 提交本任务。**

```bash
git add backend/app_rpa.go backend/app_rpa_test.go frontend/src/modules/rpa/types.ts frontend/src/modules/rpa/api.ts frontend/src/modules/rpa/components/FlowNodePalette.tsx frontend/src/modules/rpa/components/FlowNodeInspector.tsx frontend/src/modules/rpa/components/FlowEditorModal.tsx frontend/src/modules/rpa/pages/RunRecordsPage.tsx frontend/src/wailsjs/go/models.ts frontend/src/wailsjs/go/main/App.d.ts frontend/src/wailsjs/go/main/App.js
git commit -m "feat: expose advanced rpa runtime in app and editor"
```

### Task 8: 整体验证和回归

**Files:**
- Test: `backend/internal/rpa/browser_session_test.go`
- Test: `backend/internal/rpa/compiler_test.go`
- Test: `backend/internal/rpa/runtime_context_test.go`
- Test: `backend/internal/rpa/executor_browser_test.go`
- Test: `backend/internal/rpa/executor_control_test.go`
- Test: `backend/internal/rpa/parallel_executor_test.go`
- Test: `backend/app_rpa_test.go`

- [ ] **Step 1: 运行 RPA 内核测试。**

Run: `go test ./backend/internal/rpa -count=1`
Expected: PASS

- [ ] **Step 2: 运行 App 层 RPA 测试。**

Run: `go test ./backend -run RPA -count=1`
Expected: PASS

- [ ] **Step 3: 运行前端构建。**

Run: `npm run build`
Workdir: `frontend`
Expected: PASS

- [ ] **Step 4: 运行仓库级后端回归。**

Run: `go test ./backend/... -count=1`
Expected: PASS

- [ ] **Step 5: 手动验证一条完整流程。**

Run: `wails dev`
Expected:
- 能创建包含 `browser.open_url -> browser.read_text -> control.if -> browser.click` 的流程
- 能保存变量并在后续 `browser.input` 中使用 `${pageTitle}`
- 失败节点进入异常分支并生成步骤级记录
- 并发分支能完成并在运行记录里看到两个分支结果

- [ ] **Step 6: 提交整体验证结论。**

```bash
git add .
git commit -m "test: verify advanced browser rpa runtime"
```

## Self-Review

- 本计划覆盖了用户明确要求的 8 项能力：点击元素、输入文本、读取页面内容、条件判断、变量传递、重试、异常分支、并发执行。
- 本计划没有新增第二套执行模型，仍然沿用当前 `backend/internal/rpa` 和前端流程图模型，避免重复建设。
- 本计划默认复用现有调试端口和 websocket 基础设施，不强行引入 `chromedp`，先把仓库里已经存在的 CDP 能力抽象出来。
