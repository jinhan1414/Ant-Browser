package rpa

import (
	"path/filepath"
	"testing"

	dbpkg "ant-chrome/backend/internal/database"
)

func TestSQLiteFlowDAO_CreateAndList(t *testing.T) {
	db := openRPATestDB(t)
	dao := NewSQLiteFlowDAO(db.GetConn())

	group, err := dao.CreateGroup(FlowGroupInput{
		GroupName: "默认分组",
		SortOrder: 1,
	})
	if err != nil {
		t.Fatalf("创建流程分组失败: %v", err)
	}

	flow := &Flow{
		FlowName: "打开首页",
		GroupID:  group.GroupID,
		Steps: []FlowStep{
			{StepID: "s1", StepName: "启动浏览器", StepType: StepTypeStartBrowser},
			{
				StepID:   "s2",
				StepName: "等待",
				StepType: StepTypeWait,
				Config: map[string]any{
					"durationMs": float64(300),
				},
			},
		},
	}
	if err := dao.UpsertFlow(flow); err != nil {
		t.Fatalf("保存流程失败: %v", err)
	}

	flows, err := dao.ListFlows("", group.GroupID)
	if err != nil {
		t.Fatalf("查询流程失败: %v", err)
	}
	if len(flows) != 1 {
		t.Fatalf("期望 1 条流程，实际 %d", len(flows))
	}
	if flows[0].FlowName != "打开首页" || len(flows[0].Steps) != 2 {
		t.Fatalf("流程数据不正确: %+v", flows[0])
	}
}

func TestSQLiteFlowDAO_PersistsDocumentFields(t *testing.T) {
	db := openRPATestDB(t)
	dao := NewSQLiteFlowDAO(db.GetConn())

	flow := &Flow{
		FlowName:   "文档流程",
		SourceType: FlowSourceXMLImport,
		SourceXML:  `<flow schemaVersion="1" name="文档流程"></flow>`,
		Document: FlowDocument{
			SchemaVersion: 2,
			Nodes: []FlowNode{
				{NodeID: "start_1", NodeType: NodeTypeStart, Label: "开始"},
			},
		},
	}

	if err := dao.UpsertFlow(flow); err != nil {
		t.Fatalf("保存流程失败: %v", err)
	}

	got, err := dao.GetFlow(flow.FlowID)
	if err != nil {
		t.Fatalf("查询流程失败: %v", err)
	}
	if got.SourceType != FlowSourceXMLImport || got.SourceXML == "" || len(got.Document.Nodes) != 1 {
		t.Fatalf("流程文档字段未正确持久化: %+v", got)
	}
}

func TestSQLiteTaskDAO_SaveTaskAndTargets(t *testing.T) {
	db := openRPATestDB(t)
	flowDAO := NewSQLiteFlowDAO(db.GetConn())
	taskDAO := NewSQLiteTaskDAO(db.GetConn())

	group, err := flowDAO.CreateGroup(FlowGroupInput{GroupName: "任务分组"})
	if err != nil {
		t.Fatalf("创建流程分组失败: %v", err)
	}
	flow := &Flow{FlowName: "任务流程", GroupID: group.GroupID}
	if err := flowDAO.UpsertFlow(flow); err != nil {
		t.Fatalf("保存流程失败: %v", err)
	}

	task := &Task{
		TaskName:       "批量打开",
		FlowID:         flow.FlowID,
		ExecutionOrder: TaskExecutionSequential,
		TaskType:       TaskTypeManual,
		ScheduleConfig: map[string]any{},
		Enabled:        true,
	}
	if err := taskDAO.UpsertTask(task); err != nil {
		t.Fatalf("保存任务失败: %v", err)
	}

	targets := []TaskTarget{
		{TaskID: task.TaskID, ProfileID: "p1", SortOrder: 1},
		{TaskID: task.TaskID, ProfileID: "p2", SortOrder: 2},
	}
	if err := taskDAO.ReplaceTargets(task.TaskID, targets); err != nil {
		t.Fatalf("保存任务目标失败: %v", err)
	}

	list, err := taskDAO.ListTasks()
	if err != nil {
		t.Fatalf("查询任务失败: %v", err)
	}
	if len(list) != 1 || list[0].TaskName != "批量打开" {
		t.Fatalf("任务列表错误: %+v", list)
	}

	gotTargets, err := taskDAO.ListTargets(task.TaskID)
	if err != nil {
		t.Fatalf("查询任务目标失败: %v", err)
	}
	if len(gotTargets) != 2 || gotTargets[0].ProfileID != "p1" || gotTargets[1].ProfileID != "p2" {
		t.Fatalf("任务目标错误: %+v", gotTargets)
	}
}

func TestSQLiteRunAndTemplateDAO_BasicPersistence(t *testing.T) {
	db := openRPATestDB(t)
	runDAO := NewSQLiteRunDAO(db.GetConn())
	templateDAO := NewSQLiteTemplateDAO(db.GetConn())

	run := &Run{
		TaskID:       "task-1",
		FlowID:       "flow-1",
		TriggerType:  RunTriggerManual,
		Status:       RunStatusRunning,
		Summary:      "执行中",
		StartedAt:    "2026-04-15T10:00:00Z",
		ErrorMessage: "",
	}
	if err := runDAO.CreateRun(run); err != nil {
		t.Fatalf("创建运行记录失败: %v", err)
	}

	target := &RunTarget{
		RunID:        run.RunID,
		ProfileID:    "p1",
		ProfileName:  "实例一",
		Status:       RunStatusSuccess,
		StepIndex:    2,
		DebugPort:    9222,
		StartedAt:    "2026-04-15T10:00:01Z",
		FinishedAt:   "2026-04-15T10:00:03Z",
		ErrorMessage: "",
	}
	if err := runDAO.CreateRunTarget(target); err != nil {
		t.Fatalf("创建运行目标失败: %v", err)
	}

	tpl := &Template{
		TemplateName: "默认打开站点",
		Description:  "模板描述",
		Tags:         []string{"入门", "站点"},
		FlowSnapshot: Flow{FlowName: "模板流程"},
	}
	if err := templateDAO.UpsertTemplate(tpl); err != nil {
		t.Fatalf("保存模板失败: %v", err)
	}

	runs, err := runDAO.ListRuns()
	if err != nil {
		t.Fatalf("查询运行记录失败: %v", err)
	}
	if len(runs) != 1 || runs[0].Summary != "执行中" {
		t.Fatalf("运行记录错误: %+v", runs)
	}

	targets, err := runDAO.ListRunTargets(run.RunID)
	if err != nil {
		t.Fatalf("查询运行目标失败: %v", err)
	}
	if len(targets) != 1 || targets[0].ProfileName != "实例一" {
		t.Fatalf("运行目标错误: %+v", targets)
	}

	templates, err := templateDAO.ListTemplates()
	if err != nil {
		t.Fatalf("查询模板失败: %v", err)
	}
	if len(templates) != 1 || templates[0].TemplateName != "默认打开站点" {
		t.Fatalf("模板列表错误: %+v", templates)
	}
}

func openRPATestDB(t *testing.T) *dbpkg.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "rpa-test.db")
	db, err := dbpkg.NewDB(dbPath)
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	if err := db.Migrate(); err != nil {
		t.Fatalf("执行迁移失败: %v", err)
	}
	return db
}
