package rpa

import (
	"testing"

	"ant-chrome/backend/internal/browser"
)

func TestExecutor_RunTaskSequentially(t *testing.T) {
	operator := &browserOperatorStub{}
	executor := NewExecutor(operator)

	task := &Task{
		TaskID:         "task-1",
		TaskName:       "顺序任务",
		ExecutionOrder: TaskExecutionSequential,
	}
	flow := &Flow{
		FlowID:   "flow-1",
		FlowName: "流程一",
		Document: FlowDocument{
			SchemaVersion: 2,
			Nodes: []FlowNode{
				{NodeID: "start_1", NodeType: NodeTypeStart, Label: "开始"},
				{NodeID: "node_1", NodeType: NodeTypeBrowserStart, Label: "启动浏览器", Config: map[string]any{"startUrls": []string{"https://example.com"}}},
				{NodeID: "node_2", NodeType: NodeTypeDelay, Label: "等待", Config: map[string]any{"durationMs": float64(1)}},
				{NodeID: "node_3", NodeType: NodeTypeBrowserStop, Label: "关闭浏览器"},
				{NodeID: "end_1", NodeType: NodeTypeEnd, Label: "结束"},
			},
			Edges: []FlowEdge{
				{EdgeID: "e1", SourceNodeID: "start_1", TargetNodeID: "node_1"},
				{EdgeID: "e2", SourceNodeID: "node_1", TargetNodeID: "node_2"},
				{EdgeID: "e3", SourceNodeID: "node_2", TargetNodeID: "node_3"},
				{EdgeID: "e4", SourceNodeID: "node_3", TargetNodeID: "end_1"},
			},
		},
	}
	targets := []TaskTarget{
		{TaskID: task.TaskID, ProfileID: "profile-a", SortOrder: 1},
		{TaskID: task.TaskID, ProfileID: "profile-b", SortOrder: 2},
	}

	run, runTargets, runSteps, err := executor.Execute(task, flow, targets)
	if err != nil {
		t.Fatalf("执行任务失败: %v", err)
	}
	if run.Status != RunStatusSuccess || len(runTargets) != 2 {
		t.Fatalf("执行结果错误: run=%+v targets=%+v", run, runTargets)
	}
	if len(runSteps) == 0 {
		t.Fatal("步骤记录不能为空")
	}
	if len(operator.started) != 2 || len(operator.stopped) != 2 {
		t.Fatalf("浏览器操作次数错误: %+v", operator)
	}
	if operator.started[0] != "profile-a" || operator.started[1] != "profile-b" {
		t.Fatalf("顺序执行不正确: %+v", operator.started)
	}
}

func TestExecutor_UnknownStepFailsTarget(t *testing.T) {
	operator := &browserOperatorStub{}
	executor := NewExecutor(operator)

	task := &Task{TaskID: "task-2", ExecutionOrder: TaskExecutionSequential}
	flow := &Flow{
		FlowID: "flow-2",
		Document: FlowDocument{
			SchemaVersion: 2,
			Nodes: []FlowNode{
				{NodeID: "start_1", NodeType: NodeTypeStart, Label: "开始"},
				{NodeID: "node_1", NodeType: FlowNodeType("unknown"), Label: "未知"},
				{NodeID: "end_1", NodeType: NodeTypeEnd, Label: "结束"},
			},
			Edges: []FlowEdge{
				{EdgeID: "e1", SourceNodeID: "start_1", TargetNodeID: "node_1"},
				{EdgeID: "e2", SourceNodeID: "node_1", TargetNodeID: "end_1"},
			},
		},
	}
	targets := []TaskTarget{{TaskID: task.TaskID, ProfileID: "profile-a", SortOrder: 1}}

	run, runTargets, _, err := executor.Execute(task, flow, targets)
	if err == nil {
		t.Fatal("未知步骤应返回错误")
	}
	if run.Status != RunStatusFailed || len(runTargets) != 1 {
		t.Fatalf("失败状态错误: run=%+v targets=%+v", run, runTargets)
	}
	if runTargets[0].Status != RunStatusFailed || runTargets[0].ErrorMessage == "" {
		t.Fatalf("目标失败信息缺失: %+v", runTargets[0])
	}
}

func TestExecutor_ExecuteDocumentNodes(t *testing.T) {
	operator := &browserOperatorStub{}
	executor := NewExecutor(operator)

	task := &Task{
		TaskID:         "task-3",
		TaskName:       "文档任务",
		ExecutionOrder: TaskExecutionSequential,
	}
	flow := &Flow{
		FlowID:   "flow-3",
		FlowName: "文档流程",
		Document: FlowDocument{
			SchemaVersion: 2,
			Nodes: []FlowNode{
				{NodeID: "start_1", NodeType: NodeTypeStart, Label: "开始"},
				{NodeID: "open_1", NodeType: NodeTypeBrowserOpenURL, Label: "打开页面", Config: map[string]any{"url": "https://example.com"}},
				{NodeID: "delay_1", NodeType: NodeTypeDelay, Label: "等待", Config: map[string]any{"durationMs": float64(1)}},
				{NodeID: "end_1", NodeType: NodeTypeEnd, Label: "结束"},
			},
			Edges: []FlowEdge{
				{EdgeID: "e1", SourceNodeID: "start_1", TargetNodeID: "open_1"},
				{EdgeID: "e2", SourceNodeID: "open_1", TargetNodeID: "delay_1"},
				{EdgeID: "e3", SourceNodeID: "delay_1", TargetNodeID: "end_1"},
			},
		},
	}
	targets := []TaskTarget{{TaskID: task.TaskID, ProfileID: "profile-a", SortOrder: 1}}

	run, runTargets, runSteps, err := executor.Execute(task, flow, targets)
	if err != nil {
		t.Fatalf("执行文档流程失败: %v", err)
	}
	if run.Status != RunStatusSuccess || len(runTargets) != 1 {
		t.Fatalf("文档执行结果错误: run=%+v targets=%+v", run, runTargets)
	}
	if len(runSteps) < 2 {
		t.Fatalf("文档步骤记录过少: %d", len(runSteps))
	}
	if len(operator.started) != 1 || operator.started[0] != "profile-a" {
		t.Fatalf("文档流程未触发浏览器动作: %+v", operator.started)
	}
}

type browserOperatorStub struct {
	started []string
	stopped []string
}

func (s *browserOperatorStub) Start(profileID string, launchArgs []string, startURLs []string, skipDefault bool) (*browser.Profile, error) {
	s.started = append(s.started, profileID)
	return &browser.Profile{
		ProfileId:   profileID,
		ProfileName: profileID,
		Running:     true,
		DebugPort:   9222,
		DebugReady:  true,
	}, nil
}

func (s *browserOperatorStub) Stop(profileID string) (*browser.Profile, error) {
	s.stopped = append(s.stopped, profileID)
	return &browser.Profile{
		ProfileId:   profileID,
		ProfileName: profileID,
	}, nil
}
