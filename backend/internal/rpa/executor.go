package rpa

import (
	"fmt"
	"math/rand"
	"time"

	"ant-chrome/backend/internal/browser"

	"github.com/google/uuid"
)

type BrowserOperator interface {
	Start(profileID string, launchArgs []string, startURLs []string, skipDefault bool) (*browser.Profile, error)
	Stop(profileID string) (*browser.Profile, error)
}

type AutomationSession interface {
	QuerySelector(selector string) (string, error)
	WaitVisible(selector string, timeoutMs int) error
	Click(selector string) error
	InputText(selector string, value string) error
	ReadText(selector string) (string, error)
	ReadURL() (string, error)
	Close() error
}

type AutomationSessionFactory func(debugPort int) (AutomationSession, error)

type Executor struct {
	operator       BrowserOperator
	sessionFactory AutomationSessionFactory
	notifier       Notifier
}

func NewExecutor(operator BrowserOperator) *Executor {
	return NewExecutorWithDeps(operator, defaultAutomationSessionFactory, NewUnsupportedNotifier())
}

func NewExecutorWithDeps(operator BrowserOperator, sessionFactory AutomationSessionFactory, notifier Notifier) *Executor {
	if sessionFactory == nil {
		sessionFactory = defaultAutomationSessionFactory
	}
	if notifier == nil {
		notifier = NewUnsupportedNotifier()
	}
	return &Executor{operator: operator, sessionFactory: sessionFactory, notifier: notifier}
}

func (e *Executor) Execute(task *Task, flow *Flow, targets []TaskTarget) (*Run, []*RunTarget, []*RunStep, error) {
	run := newRun(task, flow)
	plan, err := BuildExecutionPlan(flow.Document)
	if err != nil {
		return failRun(run, err), nil, nil, err
	}
	runTargets, runSteps, execErr := e.executeTargets(run, task, flow, plan, targets)
	if execErr != nil {
		return failRun(run, execErr), runTargets, runSteps, execErr
	}
	run.Status = RunStatusSuccess
	run.Summary = "执行成功"
	run.FinishedAt = time.Now().Format(time.RFC3339)
	return run, runTargets, runSteps, nil
}

func defaultAutomationSessionFactory(debugPort int) (AutomationSession, error) {
	session, err := NewBrowserSession(NewCDPClientForDebugPort(debugPort))
	if err != nil {
		return nil, err
	}
	if err := session.AttachPage(); err != nil {
		return nil, err
	}
	return session, nil
}

func newRun(task *Task, flow *Flow) *Run {
	return &Run{
		RunID:        uuid.NewString(),
		TaskID:       task.TaskID,
		FlowID:       flow.FlowID,
		TriggerType:  RunTriggerManual,
		Status:       RunStatusRunning,
		Summary:      "执行中",
		StartedAt:    time.Now().Format(time.RFC3339),
		ErrorMessage: "",
	}
}

func failRun(run *Run, err error) *Run {
	run.Status = RunStatusFailed
	run.Summary = "执行失败"
	run.ErrorMessage = err.Error()
	run.FinishedAt = time.Now().Format(time.RFC3339)
	return run
}

func orderTargets(order TaskExecutionOrder, targets []TaskTarget) []TaskTarget {
	items := append([]TaskTarget{}, targets...)
	if order != TaskExecutionRandom {
		return items
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})
	return items
}

func unsupportedNodeError(nodeType FlowNodeType) error {
	return fmt.Errorf("不支持的 RPA 节点类型: %s", nodeType)
}
