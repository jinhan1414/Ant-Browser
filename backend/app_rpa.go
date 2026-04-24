package backend

import (
	"fmt"
	"time"

	"ant-chrome/backend/internal/browser"
	"ant-chrome/backend/internal/rpa"
)

type rpaExecutionEngine interface {
	Execute(task *rpa.Task, flow *rpa.Flow, targets []rpa.TaskTarget) (*rpa.Run, []*rpa.RunTarget, []*rpa.RunStep, error)
}

type appRPAOperator struct {
	app *App
}

func (o *appRPAOperator) Start(profileID string, launchArgs []string, startURLs []string, skipDefault bool) (*browser.Profile, error) {
	if o.app == nil {
		return nil, fmt.Errorf("app is nil")
	}
	return o.app.BrowserInstanceStartWithParams(profileID, launchArgs, startURLs, skipDefault)
}

func (o *appRPAOperator) Stop(profileID string) (*browser.Profile, error) {
	if o.app == nil {
		return nil, fmt.Errorf("app is nil")
	}
	return o.app.BrowserInstanceStop(profileID)
}

func (a *App) RPAFlowGroupCreate(input rpa.FlowGroupInput) (*rpa.FlowGroup, error) {
	if a.rpaSvc == nil {
		return nil, fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.CreateFlowGroup(input)
}

func (a *App) RPAFlowGroupList() ([]*rpa.FlowGroup, error) {
	if a.rpaSvc == nil {
		return nil, fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.ListFlowGroups()
}

func (a *App) RPAFlowSave(flow rpa.Flow) (*rpa.Flow, error) {
	if a.rpaSvc == nil {
		return nil, fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.SaveFlow(&flow)
}

func (a *App) RPAFlowList(keyword string, groupID string) ([]*rpa.Flow, error) {
	if a.rpaSvc == nil {
		return nil, fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.ListFlows(keyword, groupID)
}

func (a *App) RPAFlowDelete(flowID string) error {
	if a.rpaSvc == nil {
		return fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.DeleteFlow(flowID)
}

func (a *App) RPAFlowShare(flowID string) (string, error) {
	if a.rpaSvc == nil {
		return "", fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.ShareFlow(flowID)
}

func (a *App) RPAFlowImportByShareCode(shareCode string) (*rpa.Flow, error) {
	if a.rpaSvc == nil {
		return nil, fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.ImportFlowByShareCode(shareCode)
}

func (a *App) RPAFlowImportXML(input rpa.FlowXMLImportInput) (*rpa.Flow, error) {
	if a.rpaSvc == nil {
		return nil, fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.ImportFlowXML(input)
}

func (a *App) RPAFlowParseXML(input rpa.FlowXMLImportInput) (*rpa.Flow, error) {
	document, normalizedXML, err := rpa.ParseFlowXML(input.XMLText)
	if err != nil {
		return nil, err
	}
	return &rpa.Flow{
		FlowName:   input.FlowName,
		GroupID:    input.GroupID,
		Document:   *document,
		SourceType: rpa.FlowSourceXMLImport,
		SourceXML:  normalizedXML,
		Version:    1,
	}, nil
}

func (a *App) RPAFlowExportXML(flowID string) (string, error) {
	if a.rpaSvc == nil {
		return "", fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.ExportFlowXML(flowID)
}

func (a *App) RPAFlowEncodeXML(flow rpa.Flow) (string, error) {
	return rpa.EncodeFlowXML(&flow)
}

func (a *App) RPAFlowNodeCatalog() (*rpa.FlowNodeCatalogPayload, error) {
	payload := rpa.BuildFlowNodeCatalogPayload()
	return &payload, nil
}

func (a *App) RPATaskSave(task rpa.Task, targets []rpa.TaskTarget) (*rpa.Task, error) {
	if a.rpaSvc == nil {
		return nil, fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.SaveTask(&task, targets)
}

func (a *App) RPATaskGet(taskID string) (*rpa.TaskDetail, error) {
	if a.rpaSvc == nil {
		return nil, fmt.Errorf("rpa service not initialized")
	}
	task, targets, err := a.rpaSvc.GetTask(taskID)
	if err != nil {
		return nil, err
	}
	return &rpa.TaskDetail{
		Task:    task,
		Targets: targets,
	}, nil
}

func (a *App) RPATaskList() ([]*rpa.Task, error) {
	if a.rpaSvc == nil {
		return nil, fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.ListTasks()
}

func (a *App) RPATaskDelete(taskID string) error {
	if a.rpaSvc == nil {
		return fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.DeleteTask(taskID)
}

func (a *App) RPATaskExecute(taskID string) (*rpa.Run, error) {
	return a.executeRPATask(taskID, rpa.RunTriggerManual)
}

func (a *App) executeRPATask(taskID string, triggerType rpa.RunTriggerType) (*rpa.Run, error) {
	if a.rpaSvc == nil || a.rpaExecutor == nil {
		return nil, fmt.Errorf("rpa runtime not initialized")
	}

	task, targets, err := a.rpaSvc.GetTask(taskID)
	if err != nil {
		return nil, err
	}
	flow, err := a.rpaSvc.GetFlow(task.FlowID)
	if err != nil {
		return nil, err
	}

	run, runTargets, runSteps, execErr := a.rpaExecutor.Execute(task, flow, derefTaskTargets(targets))
	if run == nil {
		return nil, fmt.Errorf("rpa executor returned nil run")
	}
	run.TriggerType = triggerType
	if run.StartedAt == "" {
		run.StartedAt = time.Now().Format(time.RFC3339)
	}
	task.LastRunAt = run.StartedAt
	if _, err = a.rpaSvc.SaveTask(task, derefTaskTargets(targets)); err != nil {
		return nil, err
	}
	if err = a.rpaSvc.SaveRun(run, runTargets, runSteps); err != nil {
		return nil, err
	}
	if execErr != nil {
		return run, nil
	}
	return run, nil
}

func (a *App) RPARunList() ([]*rpa.Run, error) {
	if a.rpaSvc == nil {
		return nil, fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.ListRuns()
}

func (a *App) RPARunTargetList(runID string) ([]*rpa.RunTarget, error) {
	if a.rpaSvc == nil {
		return nil, fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.ListRunTargets(runID)
}

func (a *App) RPARunStepList(runID string) ([]*rpa.RunStep, error) {
	if a.rpaSvc == nil {
		return nil, fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.ListRunSteps(runID)
}

func (a *App) RPATemplateSave(template rpa.Template) (*rpa.Template, error) {
	if a.rpaSvc == nil {
		return nil, fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.SaveTemplate(&template)
}

func (a *App) RPATemplateList() ([]*rpa.Template, error) {
	if a.rpaSvc == nil {
		return nil, fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.ListTemplates()
}

func (a *App) RPATemplateDelete(templateID string) error {
	if a.rpaSvc == nil {
		return fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.DeleteTemplate(templateID)
}

func (a *App) RPATemplateCreateFlow(templateID string) (*rpa.Flow, error) {
	if a.rpaSvc == nil {
		return nil, fmt.Errorf("rpa service not initialized")
	}
	return a.rpaSvc.CreateFlowFromTemplate(templateID)
}

func derefTaskTargets(items []*rpa.TaskTarget) []rpa.TaskTarget {
	result := make([]rpa.TaskTarget, 0, len(items))
	for _, item := range items {
		if item != nil {
			result = append(result, *item)
		}
	}
	return result
}
