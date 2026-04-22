package rpa

import (
	"testing"
)

func TestService_ShareAndImportFlow(t *testing.T) {
	db := openRPATestDB(t)
	service := NewService(
		NewSQLiteFlowDAO(db.GetConn()),
		NewSQLiteTaskDAO(db.GetConn()),
		NewSQLiteRunDAO(db.GetConn()),
		NewSQLiteTemplateDAO(db.GetConn()),
	)

	group, err := service.CreateFlowGroup(FlowGroupInput{GroupName: "流程组"})
	if err != nil {
		t.Fatalf("创建流程组失败: %v", err)
	}

	flow, err := service.SaveFlow(&Flow{
		FlowName: "分享流程",
		GroupID:  group.GroupID,
		Steps: []FlowStep{
			{StepID: "s1", StepName: "启动", StepType: StepTypeStartBrowser},
		},
	})
	if err != nil {
		t.Fatalf("保存流程失败: %v", err)
	}

	shareCode, err := service.ShareFlow(flow.FlowID)
	if err != nil {
		t.Fatalf("生成分享码失败: %v", err)
	}
	if shareCode == "" {
		t.Fatal("分享码不能为空")
	}

	imported, err := service.ImportFlowByShareCode(shareCode)
	if err != nil {
		t.Fatalf("导入流程失败: %v", err)
	}
	if imported.FlowID == flow.FlowID {
		t.Fatalf("导入流程应生成新 ID: %+v", imported)
	}
	if imported.FlowName != flow.FlowName || len(imported.Steps) != 1 {
		t.Fatalf("导入流程内容不正确: %+v", imported)
	}
}

func TestService_SaveTaskWithTargets(t *testing.T) {
	db := openRPATestDB(t)
	service := NewService(
		NewSQLiteFlowDAO(db.GetConn()),
		NewSQLiteTaskDAO(db.GetConn()),
		NewSQLiteRunDAO(db.GetConn()),
		NewSQLiteTemplateDAO(db.GetConn()),
	)

	flow, err := service.SaveFlow(&Flow{FlowName: "任务流程"})
	if err != nil {
		t.Fatalf("保存流程失败: %v", err)
	}

	task, err := service.SaveTask(&Task{
		TaskName:       "批量任务",
		FlowID:         flow.FlowID,
		ExecutionOrder: TaskExecutionRandom,
		TaskType:       TaskTypeScheduled,
		ScheduleConfig: map[string]any{
			"cron":     "0 9 * * *",
			"timezone": "Asia/Shanghai",
		},
		Enabled:        true,
	}, []TaskTarget{
		{ProfileID: "profile-a", SortOrder: 1},
		{ProfileID: "profile-b", SortOrder: 2},
	})
	if err != nil {
		t.Fatalf("保存任务失败: %v", err)
	}

	gotTask, gotTargets, err := service.GetTask(task.TaskID)
	if err != nil {
		t.Fatalf("查询任务失败: %v", err)
	}
	if gotTask.ExecutionOrder != TaskExecutionRandom || gotTask.TaskType != TaskTypeScheduled {
		t.Fatalf("任务字段不正确: %+v", gotTask)
	}
	if gotTask.ScheduleConfig["cron"] != "0 9 * * *" || gotTask.ScheduleConfig["timezone"] != "Asia/Shanghai" {
		t.Fatalf("计划任务定时配置未持久化: %+v", gotTask.ScheduleConfig)
	}
	if len(gotTargets) != 2 || gotTargets[0].TaskID != task.TaskID {
		t.Fatalf("任务目标不正确: %+v", gotTargets)
	}
}

func TestService_SaveTaskRejectsInvalidConfig(t *testing.T) {
	db := openRPATestDB(t)
	service := NewService(
		NewSQLiteFlowDAO(db.GetConn()),
		NewSQLiteTaskDAO(db.GetConn()),
		NewSQLiteRunDAO(db.GetConn()),
		NewSQLiteTemplateDAO(db.GetConn()),
	)

	_, err := service.SaveTask(&Task{
		TaskName:       "无流程任务",
		ExecutionOrder: TaskExecutionSequential,
		TaskType:       TaskTypeManual,
		Enabled:        true,
	}, []TaskTarget{
		{ProfileID: "profile-a"},
	})
	if err == nil {
		t.Fatal("缺少 flowId 的任务应保存失败")
	}

	flow, err := service.SaveFlow(&Flow{FlowName: "有效流程"})
	if err != nil {
		t.Fatalf("保存流程失败: %v", err)
	}

	_, err = service.SaveTask(&Task{
		TaskName:       "无目标任务",
		FlowID:         flow.FlowID,
		ExecutionOrder: TaskExecutionSequential,
		TaskType:       TaskTypeManual,
		Enabled:        true,
	}, nil)
	if err == nil {
		t.Fatal("缺少执行环境的任务应保存失败")
	}

	_, err = service.SaveTask(&Task{
		TaskName:       "缺少定时配置",
		FlowID:         flow.FlowID,
		ExecutionOrder: TaskExecutionSequential,
		TaskType:       TaskTypeScheduled,
		ScheduleConfig: map[string]any{},
		Enabled:        true,
	}, []TaskTarget{
		{ProfileID: "profile-a"},
	})
	if err == nil {
		t.Fatal("计划任务缺少 cron 应保存失败")
	}
}

func TestService_CreateFlowFromTemplate(t *testing.T) {
	db := openRPATestDB(t)
	service := NewService(
		NewSQLiteFlowDAO(db.GetConn()),
		NewSQLiteTaskDAO(db.GetConn()),
		NewSQLiteRunDAO(db.GetConn()),
		NewSQLiteTemplateDAO(db.GetConn()),
	)

	template, err := service.SaveTemplate(&Template{
		TemplateName: "模板一",
		Description:  "模板描述",
		Tags:         []string{"模板"},
		FlowSnapshot: Flow{
			FlowName: "模板流程",
			Steps: []FlowStep{
				{StepID: "s1", StepName: "等待", StepType: StepTypeWait},
			},
		},
	})
	if err != nil {
		t.Fatalf("保存模板失败: %v", err)
	}

	flow, err := service.CreateFlowFromTemplate(template.TemplateID)
	if err != nil {
		t.Fatalf("从模板创建流程失败: %v", err)
	}
	if flow.FlowID == "" || flow.FlowName != "模板流程" || len(flow.Steps) != 1 {
		t.Fatalf("模板创建流程结果错误: %+v", flow)
	}
}

func TestService_ImportFlowXMLBuildsDocument(t *testing.T) {
	db := openRPATestDB(t)
	service := NewService(
		NewSQLiteFlowDAO(db.GetConn()),
		NewSQLiteTaskDAO(db.GetConn()),
		NewSQLiteRunDAO(db.GetConn()),
		NewSQLiteTemplateDAO(db.GetConn()),
	)

	xmlText := `<flow schemaVersion="1" name="搜索流程">
  <nodes>
    <start id="start_1" x="80" y="120" />
    <browser.open_url id="open_1" x="260" y="120" url="https://example.com" />
    <end id="end_1" x="520" y="120" />
  </nodes>
  <edges>
    <edge from="start_1" to="open_1" />
    <edge from="open_1" to="end_1" />
  </edges>
</flow>`

	flow, err := service.ImportFlowXML(FlowXMLImportInput{
		FlowName: "搜索流程",
		XMLText:  xmlText,
	})
	if err != nil {
		t.Fatalf("导入 XML 失败: %v", err)
	}
	if flow.SourceType != FlowSourceXMLImport || len(flow.Document.Nodes) != 3 {
		t.Fatalf("XML 未转为 FlowDocument: %+v", flow)
	}
}

func TestService_ImportFlowXMLRejectsInvalidXML(t *testing.T) {
	db := openRPATestDB(t)
	service := NewService(
		NewSQLiteFlowDAO(db.GetConn()),
		NewSQLiteTaskDAO(db.GetConn()),
		NewSQLiteRunDAO(db.GetConn()),
		NewSQLiteTemplateDAO(db.GetConn()),
	)

	_, err := service.ImportFlowXML(FlowXMLImportInput{
		FlowName: "非法流程",
		XMLText:  `<flow schemaVersion="1"><nodes><browser.open_url id="" /></nodes></flow>`,
	})
	if err == nil {
		t.Fatal("非法 XML 应导入失败")
	}
}
