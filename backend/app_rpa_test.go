package backend

import (
	"path/filepath"
	"testing"

	dbpkg "ant-chrome/backend/internal/database"
	"ant-chrome/backend/internal/rpa"
)

func TestAppRPAFlowTaskAndRunLifecycle(t *testing.T) {
	app := newTestRPAApp(t)

	group, err := app.RPAFlowGroupCreate(rpa.FlowGroupInput{GroupName: "RPA 分组"})
	if err != nil {
		t.Fatalf("创建流程分组失败: %v", err)
	}

	flow, err := app.RPAFlowSave(rpa.Flow{
		FlowName: "打开站点",
		GroupID:  group.GroupID,
		Steps: []rpa.FlowStep{
			{StepID: "s1", StepName: "启动浏览器", StepType: rpa.StepTypeStartBrowser},
		},
	})
	if err != nil {
		t.Fatalf("保存流程失败: %v", err)
	}

	task, err := app.RPATaskSave(rpa.Task{
		TaskName:       "任务一",
		FlowID:         flow.FlowID,
		ExecutionOrder: rpa.TaskExecutionSequential,
		TaskType:       rpa.TaskTypeScheduled,
		ScheduleConfig: map[string]any{
			"cron":     "0 9 * * *",
			"timezone": "Asia/Shanghai",
		},
		Enabled:        true,
	}, []rpa.TaskTarget{
		{ProfileID: "profile-a"},
		{ProfileID: "profile-b"},
	})
	if err != nil {
		t.Fatalf("保存任务失败: %v", err)
	}

	detail, err := app.RPATaskGet(task.TaskID)
	if err != nil {
		t.Fatalf("查询任务失败: %v", err)
	}
	if detail == nil || detail.Task == nil {
		t.Fatalf("任务详情为空: %+v", detail)
	}
	if detail.Task.TaskName != "任务一" || len(detail.Targets) != 2 {
		t.Fatalf("任务数据错误: detail=%+v", detail)
	}
	if detail.Task.ScheduleConfig["cron"] != "0 9 * * *" {
		t.Fatalf("任务定时配置未回填: %+v", detail.Task.ScheduleConfig)
	}

	run, err := app.RPATaskExecute(task.TaskID)
	if err != nil {
		t.Fatalf("执行任务失败: %v", err)
	}
	if run.Status != rpa.RunStatusSuccess {
		t.Fatalf("执行结果错误: %+v", run)
	}

	runs, err := app.RPARunList()
	if err != nil {
		t.Fatalf("查询运行记录失败: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("期望 1 条运行记录，实际 %d", len(runs))
	}

	runTargets, err := app.RPARunTargetList(run.RunID)
	if err != nil {
		t.Fatalf("查询运行目标失败: %v", err)
	}
	if len(runTargets) != 2 {
		t.Fatalf("期望 2 条运行目标，实际 %d", len(runTargets))
	}
}

func TestAppRPAImportFlowXML(t *testing.T) {
	app := newTestRPAApp(t)

	flow, err := app.RPAFlowImportXML(rpa.FlowXMLImportInput{
		FlowName: "XML 流程",
		XMLText: `<flow schemaVersion="1" name="XML 流程">
  <nodes>
    <start id="start_1" x="0" y="0" />
    <end id="end_1" x="100" y="0" />
  </nodes>
  <edges>
    <edge from="start_1" to="end_1" />
  </edges>
</flow>`,
	})
	if err != nil {
		t.Fatalf("导入 XML 失败: %v", err)
	}
	if flow.SourceType != rpa.FlowSourceXMLImport || len(flow.Document.Nodes) != 2 {
		t.Fatalf("导入结果错误: %+v", flow)
	}
}

func TestAppRPAParseAndEncodeFlowXML(t *testing.T) {
	app := newTestRPAApp(t)

	draft, err := app.RPAFlowParseXML(rpa.FlowXMLImportInput{
		FlowName: "解析流程",
		XMLText: `<flow schemaVersion="1" name="解析流程">
  <nodes>
    <start id="start_1" x="0" y="0" />
    <browser.open_url id="open_1" x="120" y="0" url="https://example.com" />
    <end id="end_1" x="240" y="0" />
  </nodes>
  <edges>
    <edge from="start_1" to="open_1" />
    <edge from="open_1" to="end_1" />
  </edges>
</flow>`,
	})
	if err != nil {
		t.Fatalf("解析 XML 失败: %v", err)
	}
	if draft.FlowID != "" || len(draft.Document.Nodes) != 3 {
		t.Fatalf("解析结果错误: %+v", draft)
	}

	xmlText, err := app.RPAFlowEncodeXML(*draft)
	if err != nil {
		t.Fatalf("编码 XML 失败: %v", err)
	}
	if xmlText == "" {
		t.Fatal("编码 XML 结果不能为空")
	}
}

func TestAppExecuteScheduledTaskSetsScheduledTrigger(t *testing.T) {
	app := newTestRPAApp(t)

	flow, err := app.RPAFlowSave(rpa.Flow{FlowName: "定时流程"})
	if err != nil {
		t.Fatalf("保存流程失败: %v", err)
	}

	task, err := app.RPATaskSave(rpa.Task{
		TaskName:       "定时任务",
		FlowID:         flow.FlowID,
		ExecutionOrder: rpa.TaskExecutionSequential,
		TaskType:       rpa.TaskTypeScheduled,
		ScheduleConfig: map[string]any{
			"cron":     "0 * * * * ?",
			"timezone": "Asia/Shanghai",
		},
		Enabled: true,
	}, []rpa.TaskTarget{{ProfileID: "profile-a"}})
	if err != nil {
		t.Fatalf("保存任务失败: %v", err)
	}

	run, err := app.executeRPATask(task.TaskID, rpa.RunTriggerScheduled)
	if err != nil {
		t.Fatalf("执行定时任务失败: %v", err)
	}
	if run.TriggerType != rpa.RunTriggerScheduled {
		t.Fatalf("定时触发类型错误: %+v", run)
	}

	runs, err := app.RPARunList()
	if err != nil {
		t.Fatalf("查询运行记录失败: %v", err)
	}
	if len(runs) != 1 || runs[0].TriggerType != rpa.RunTriggerScheduled {
		t.Fatalf("定时运行记录未持久化: %+v", runs)
	}
}

func newTestRPAApp(t *testing.T) *App {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "app-rpa.db")
	db, err := dbpkg.NewDB(dbPath)
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	if err := db.Migrate(); err != nil {
		t.Fatalf("执行数据库迁移失败: %v", err)
	}

	flowDAO := rpa.NewSQLiteFlowDAO(db.GetConn())
	taskDAO := rpa.NewSQLiteTaskDAO(db.GetConn())
	runDAO := rpa.NewSQLiteRunDAO(db.GetConn())
	templateDAO := rpa.NewSQLiteTemplateDAO(db.GetConn())

	app := NewApp(".")
	app.db = db
	app.rpaSvc = rpa.NewService(flowDAO, taskDAO, runDAO, templateDAO)
	app.rpaExecutor = &rpaExecutorStub{}
	return app
}

type rpaExecutorStub struct{}

func (s *rpaExecutorStub) Execute(task *rpa.Task, flow *rpa.Flow, targets []rpa.TaskTarget) (*rpa.Run, []*rpa.RunTarget, error) {
	run := &rpa.Run{
		TaskID:      task.TaskID,
		FlowID:      flow.FlowID,
		TriggerType: rpa.RunTriggerManual,
		Status:      rpa.RunStatusSuccess,
		Summary:     "执行成功",
		StartedAt:   "2026-04-15T10:00:00Z",
		FinishedAt:  "2026-04-15T10:00:02Z",
	}
	items := make([]*rpa.RunTarget, 0, len(targets))
	for _, target := range targets {
		items = append(items, &rpa.RunTarget{
			ProfileID:   target.ProfileID,
			ProfileName: target.ProfileID,
			Status:      rpa.RunStatusSuccess,
			StepIndex:   len(flow.Steps),
			StartedAt:   "2026-04-15T10:00:00Z",
			FinishedAt:  "2026-04-15T10:00:01Z",
			DebugPort:   9222,
		})
	}
	return run, items, nil
}
