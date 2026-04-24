package rpa

import (
	"strings"
	"testing"
)

func TestExecutor_BrowserReadTextAndNotify(t *testing.T) {
	operator := &browserOperatorStub{}
	session := &automationSessionStub{
		texts: map[string]string{
			"#status": "需要验证",
		},
	}
	notifier := &notifierStub{}
	executor := NewExecutorWithDeps(
		operator,
		func(debugPort int) (AutomationSession, error) {
			return session, nil
		},
		notifier,
	)

	task := &Task{TaskID: "task-1", TaskName: "检查 Google", ExecutionOrder: TaskExecutionSequential}
	flow := &Flow{
		FlowID:   "flow-1",
		FlowName: "Google 状态巡检",
		Document: FlowDocument{
			SchemaVersion: 3,
			Nodes: []FlowNode{
				{NodeID: "start_1", NodeType: NodeTypeStart, Label: "开始"},
				{NodeID: "open_1", NodeType: NodeTypeBrowserOpenURL, Label: "打开页面", Config: map[string]any{"url": "https://accounts.google.com/"}},
				{NodeID: "read_1", NodeType: NodeTypeBrowserReadText, Label: "读取状态", Config: map[string]any{"selector": "#status", "saveAs": "accountStatus"}},
				{NodeID: "if_1", NodeType: NodeTypeConditionIf, Label: "判断是否异常", Config: map[string]any{"expression": `contains(accountStatus, "需要验证")`}},
				{NodeID: "notify_1", NodeType: NodeTypeSystemNotify, Label: "发送通知", Config: map[string]any{"title": "Google 账号异常", "body": "实例 ${profileId} 状态：${accountStatus}"}},
				{NodeID: "end_1", NodeType: NodeTypeEnd, Label: "结束"},
			},
			Edges: []FlowEdge{
				{EdgeID: "e1", SourceNodeID: "start_1", TargetNodeID: "open_1"},
				{EdgeID: "e2", SourceNodeID: "open_1", TargetNodeID: "read_1"},
				{EdgeID: "e3", SourceNodeID: "read_1", TargetNodeID: "if_1"},
				{EdgeID: "e4", SourceNodeID: "if_1", TargetNodeID: "notify_1", Condition: "true"},
				{EdgeID: "e5", SourceNodeID: "if_1", TargetNodeID: "end_1", Condition: "false"},
				{EdgeID: "e6", SourceNodeID: "notify_1", TargetNodeID: "end_1"},
			},
		},
	}

	run, runTargets, runSteps, err := executor.Execute(task, flow, []TaskTarget{{ProfileID: "profile-a"}})
	if err != nil {
		t.Fatalf("执行任务失败: %v", err)
	}
	if run.Status != RunStatusSuccess {
		t.Fatalf("运行状态错误: %+v", run)
	}
	if len(runTargets) != 1 {
		t.Fatalf("运行目标数错误: %d", len(runTargets))
	}
	if len(runSteps) < 3 {
		t.Fatalf("步骤记录过少: %d", len(runSteps))
	}
	if len(notifier.messages) != 1 {
		t.Fatalf("通知条数错误: %+v", notifier.messages)
	}
	if !strings.Contains(notifier.messages[0], "需要验证") {
		t.Fatalf("通知内容错误: %q", notifier.messages[0])
	}
}

type automationSessionStub struct {
	clicked []string
	inputs  map[string]string
	texts   map[string]string
}

func (s *automationSessionStub) QuerySelector(selector string) (string, error) {
	return selector, nil
}

func (s *automationSessionStub) WaitVisible(selector string, timeoutMs int) error {
	return nil
}

func (s *automationSessionStub) Click(selector string) error {
	s.clicked = append(s.clicked, selector)
	return nil
}

func (s *automationSessionStub) InputText(selector string, value string) error {
	if s.inputs == nil {
		s.inputs = map[string]string{}
	}
	s.inputs[selector] = value
	return nil
}

func (s *automationSessionStub) ReadText(selector string) (string, error) {
	return s.texts[selector], nil
}

func (s *automationSessionStub) ReadURL() (string, error) {
	return "https://accounts.google.com/", nil
}

func (s *automationSessionStub) Close() error {
	return nil
}

type notifierStub struct {
	messages []string
}

func (s *notifierStub) Notify(title string, body string) error {
	s.messages = append(s.messages, title+"|"+body)
	return nil
}
