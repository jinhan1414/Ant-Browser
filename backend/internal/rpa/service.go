package rpa

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	flows     FlowDAO
	tasks     TaskDAO
	runs      RunDAO
	templates TemplateDAO
}

func NewService(flows FlowDAO, tasks TaskDAO, runs RunDAO, templates TemplateDAO) *Service {
	return &Service{
		flows:     flows,
		tasks:     tasks,
		runs:      runs,
		templates: templates,
	}
}

func (s *Service) CreateFlowGroup(input FlowGroupInput) (*FlowGroup, error) {
	return s.flows.CreateGroup(input)
}

func (s *Service) ListFlowGroups() ([]*FlowGroup, error) {
	return s.flows.ListGroups()
}

func (s *Service) SaveFlow(flow *Flow) (*Flow, error) {
	if err := validateFlow(flow); err != nil {
		return nil, err
	}
	if err := s.flows.UpsertFlow(flow); err != nil {
		return nil, err
	}
	return s.flows.GetFlow(flow.FlowID)
}

func (s *Service) ImportFlowXML(input FlowXMLImportInput) (*Flow, error) {
	document, normalizedXML, err := ParseFlowXML(input.XMLText)
	if err != nil {
		return nil, err
	}
	flowName := strings.TrimSpace(input.FlowName)
	if flowName == "" {
		flowName = "XML 导入流程"
	}
	return s.SaveFlow(&Flow{
		FlowName:   flowName,
		GroupID:    strings.TrimSpace(input.GroupID),
		Document:   *document,
		SourceType: FlowSourceXMLImport,
		SourceXML:  normalizedXML,
	})
}

func (s *Service) ExportFlowXML(flowID string) (string, error) {
	flow, err := s.flows.GetFlow(flowID)
	if err != nil {
		return "", err
	}
	return EncodeFlowXML(flow)
}

func (s *Service) ListFlows(keyword string, groupID string) ([]*Flow, error) {
	return s.flows.ListFlows(keyword, groupID)
}

func (s *Service) GetFlow(flowID string) (*Flow, error) {
	return s.flows.GetFlow(flowID)
}

func (s *Service) DeleteFlow(flowID string) error {
	return s.flows.DeleteFlow(flowID)
}

func (s *Service) ShareFlow(flowID string) (string, error) {
	flow, err := s.flows.GetFlow(flowID)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(flow.ShareCode) != "" {
		return flow.ShareCode, nil
	}
	shareCode, err := nextShareCode()
	if err != nil {
		return "", err
	}
	flow.ShareCode = shareCode
	if err := s.flows.UpsertFlow(flow); err != nil {
		return "", err
	}
	return flow.ShareCode, nil
}

func (s *Service) ImportFlowByShareCode(shareCode string) (*Flow, error) {
	source, err := s.flows.GetFlowByShareCode(shareCode)
	if err != nil {
		return nil, err
	}
	clone := cloneFlow(source)
	if err := s.flows.UpsertFlow(clone); err != nil {
		return nil, err
	}
	return s.flows.GetFlow(clone.FlowID)
}

func (s *Service) SaveTask(task *Task, targets []TaskTarget) (*Task, error) {
	if err := validateTask(task, targets); err != nil {
		return nil, err
	}
	if err := s.tasks.UpsertTask(task); err != nil {
		return nil, err
	}
	normalized := normalizeTargets(task.TaskID, targets)
	if err := s.tasks.ReplaceTargets(task.TaskID, normalized); err != nil {
		return nil, err
	}
	gotTask, _, err := s.GetTask(task.TaskID)
	return gotTask, err
}

func (s *Service) GetTask(taskID string) (*Task, []*TaskTarget, error) {
	task, err := s.tasks.GetTask(taskID)
	if err != nil {
		return nil, nil, err
	}
	targets, err := s.tasks.ListTargets(taskID)
	if err != nil {
		return nil, nil, err
	}
	return task, targets, nil
}

func (s *Service) ListTasks() ([]*Task, error) {
	return s.tasks.ListTasks()
}

func (s *Service) DeleteTask(taskID string) error {
	return s.tasks.DeleteTask(taskID)
}

func (s *Service) SaveTemplate(template *Template) (*Template, error) {
	if err := s.templates.UpsertTemplate(template); err != nil {
		return nil, err
	}
	return s.templates.GetTemplate(template.TemplateID)
}

func (s *Service) ListTemplates() ([]*Template, error) {
	return s.templates.ListTemplates()
}

func (s *Service) GetTemplate(templateID string) (*Template, error) {
	return s.templates.GetTemplate(templateID)
}

func (s *Service) DeleteTemplate(templateID string) error {
	return s.templates.DeleteTemplate(templateID)
}

func (s *Service) ListRuns() ([]*Run, error) {
	return s.runs.ListRuns()
}

func (s *Service) ListRunTargets(runID string) ([]*RunTarget, error) {
	return s.runs.ListRunTargets(runID)
}

func (s *Service) ListRunSteps(runID string) ([]*RunStep, error) {
	return s.runs.ListRunSteps(runID)
}

func (s *Service) SaveRun(run *Run, targets []*RunTarget, steps []*RunStep) error {
	if err := s.runs.CreateRun(run); err != nil {
		return err
	}
	for _, target := range targets {
		target.RunID = run.RunID
		if err := s.runs.CreateRunTarget(target); err != nil {
			return err
		}
	}
	for _, step := range steps {
		step.RunID = run.RunID
		if err := s.runs.CreateRunStep(step); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) CreateFlowFromTemplate(templateID string) (*Flow, error) {
	template, err := s.templates.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}
	flow := cloneFlow(&template.FlowSnapshot)
	flow.FlowName = strings.TrimSpace(template.FlowSnapshot.FlowName)
	if flow.FlowName == "" {
		flow.FlowName = strings.TrimSpace(template.TemplateName)
	}
	return s.SaveFlow(flow)
}

func cloneFlow(source *Flow) *Flow {
	if source == nil {
		return &Flow{Document: defaultFlowDocument()}
	}
	steps := make([]FlowStep, 0, len(source.Steps))
	for _, step := range source.Steps {
		steps = append(steps, FlowStep{
			StepID:   uuid.NewString(),
			StepName: step.StepName,
			StepType: step.StepType,
			Config:   cloneMap(step.Config),
		})
	}
	return &Flow{
		FlowID:     uuid.NewString(),
		FlowName:   source.FlowName,
		GroupID:    source.GroupID,
		Steps:      steps,
		Document:   cloneDocument(source.Document),
		SourceType: source.SourceType,
		SourceXML:  source.SourceXML,
		ShareCode:  "",
		Version:    1,
	}
}

func normalizeTargets(taskID string, targets []TaskTarget) []TaskTarget {
	items := make([]TaskTarget, 0, len(targets))
	for idx, target := range targets {
		items = append(items, TaskTarget{
			TaskID:    taskID,
			ProfileID: strings.TrimSpace(target.ProfileID),
			SortOrder: idx + 1,
		})
	}
	return items
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func nextShareCode() (string, error) {
	code := strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", ""))
	if len(code) < 8 {
		return "", fmt.Errorf("分享码生成失败")
	}
	return code[:8], nil
}

func validateFlow(flow *Flow) error {
	if flow == nil {
		return fmt.Errorf("流程不能为空")
	}
	if strings.TrimSpace(flow.FlowName) == "" {
		return fmt.Errorf("流程名称不能为空")
	}
	if len(flow.Document.Nodes) == 0 {
		flow.Document = defaultFlowDocument()
	}
	if err := ValidateFlowDocument(normalizeDocument(flow.Document)); err != nil {
		return err
	}
	return nil
}

func validateTask(task *Task, targets []TaskTarget) error {
	if task == nil {
		return fmt.Errorf("任务不能为空")
	}
	if strings.TrimSpace(task.TaskName) == "" {
		return fmt.Errorf("任务名称不能为空")
	}
	if strings.TrimSpace(task.FlowID) == "" {
		return fmt.Errorf("运行流程不能为空")
	}
	if len(targets) == 0 {
		return fmt.Errorf("执行环境不能为空")
	}
	for _, target := range targets {
		if strings.TrimSpace(target.ProfileID) == "" {
			return fmt.Errorf("执行环境包含空实例")
		}
	}
	normalizeTaskScheduleConfig(task)
	if err := validateTaskScheduleConfig(task); err != nil {
		return err
	}
	return nil
}
