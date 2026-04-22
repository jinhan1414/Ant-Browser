package rpa

import (
	"fmt"
	"math/rand"
	"time"

	"ant-chrome/backend/internal/browser"
)

type BrowserOperator interface {
	Start(profileID string, launchArgs []string, startURLs []string, skipDefault bool) (*browser.Profile, error)
	Stop(profileID string) (*browser.Profile, error)
}

type Executor struct {
	operator BrowserOperator
}

func NewExecutor(operator BrowserOperator) *Executor {
	return &Executor{operator: operator}
}

func (e *Executor) Execute(task *Task, flow *Flow, targets []TaskTarget) (*Run, []*RunTarget, error) {
	run := &Run{
		TaskID:      task.TaskID,
		FlowID:      flow.FlowID,
		TriggerType: RunTriggerManual,
		Status:      RunStatusRunning,
		Summary:     "执行中",
		StartedAt:   time.Now().Format(time.RFC3339),
	}
	ordered := orderTargets(task.ExecutionOrder, targets)
	plan, err := BuildExecutionPlan(flow.Document)
	if err != nil {
		run.Status = RunStatusFailed
		run.Summary = "执行失败"
		run.ErrorMessage = err.Error()
		run.FinishedAt = time.Now().Format(time.RFC3339)
		return run, nil, err
	}
	runTargets := make([]*RunTarget, 0, len(ordered))
	for _, target := range ordered {
		item := e.executeTarget(run.RunID, target, plan)
		runTargets = append(runTargets, item)
		if item.Status == RunStatusFailed {
			run.Status = RunStatusFailed
			run.Summary = "执行失败"
			run.FinishedAt = time.Now().Format(time.RFC3339)
			run.ErrorMessage = item.ErrorMessage
			return run, runTargets, fmt.Errorf(item.ErrorMessage)
		}
	}
	run.Status = RunStatusSuccess
	run.Summary = "执行成功"
	run.FinishedAt = time.Now().Format(time.RFC3339)
	return run, runTargets, nil
}

func (e *Executor) executeTarget(runID string, target TaskTarget, plan []FlowNode) *RunTarget {
	item := &RunTarget{
		RunID:       runID,
		ProfileID:   target.ProfileID,
		ProfileName: target.ProfileID,
		Status:      RunStatusRunning,
		StartedAt:   time.Now().Format(time.RFC3339),
	}
	for idx, node := range plan {
		item.StepIndex = idx + 1
		profile, err := e.executeNode(target.ProfileID, node)
		if profile != nil && profile.DebugPort > 0 {
			item.DebugPort = profile.DebugPort
			if profile.ProfileName != "" {
				item.ProfileName = profile.ProfileName
			}
		}
		if err != nil {
			item.Status = RunStatusFailed
			item.ErrorMessage = err.Error()
			item.FinishedAt = time.Now().Format(time.RFC3339)
			return item
		}
	}
	item.Status = RunStatusSuccess
	item.FinishedAt = time.Now().Format(time.RFC3339)
	return item
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
