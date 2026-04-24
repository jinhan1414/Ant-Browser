package rpa

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type executionState struct {
	target  *RunTarget
	ctx     *RuntimeContext
	session AutomationSession
	steps   []*RunStep
}

func (e *Executor) executeTargets(run *Run, task *Task, flow *Flow, plan *ExecutionPlan, targets []TaskTarget) ([]*RunTarget, []*RunStep, error) {
	ordered := orderTargets(task.ExecutionOrder, targets)
	runTargets := make([]*RunTarget, 0, len(ordered))
	runSteps := make([]*RunStep, 0, len(ordered)*4)
	for _, target := range ordered {
		item, steps, err := e.executeTarget(run.RunID, task, flow, target, plan)
		runTargets = append(runTargets, item)
		runSteps = append(runSteps, steps...)
		if err != nil {
			return runTargets, runSteps, err
		}
	}
	return runTargets, runSteps, nil
}

func (e *Executor) executeTarget(runID string, task *Task, flow *Flow, target TaskTarget, plan *ExecutionPlan) (*RunTarget, []*RunStep, error) {
	state := &executionState{
		target: &RunTarget{
			RunTargetID: uuid.NewString(),
			RunID:       runID,
			ProfileID:   target.ProfileID,
			ProfileName: target.ProfileID,
			Status:      RunStatusRunning,
			StartedAt:   time.Now().Format(time.RFC3339),
		},
		ctx: NewRuntimeContext(target.ProfileID),
	}
	state.ctx.Set("taskId", task.TaskID)
	state.ctx.Set("flowId", flow.FlowID)
	err := e.runPlan(state, plan)
	_ = closeSession(state.session)
	if err != nil {
		state.target.Status = RunStatusFailed
		state.target.ErrorMessage = err.Error()
		state.target.FinishedAt = time.Now().Format(time.RFC3339)
		return state.target, state.steps, err
	}
	state.target.Status = RunStatusSuccess
	state.target.FinishedAt = time.Now().Format(time.RFC3339)
	return state.target, state.steps, nil
}

func (e *Executor) runPlan(state *executionState, plan *ExecutionPlan) error {
	currentID := plan.EntryNodeID
	visited := map[string]bool{}
	for currentID != "" {
		if visited[currentID] {
			return fmt.Errorf("流程包含循环: %s", currentID)
		}
		visited[currentID] = true

		current, ok := plan.Nodes[currentID]
		if !ok {
			return fmt.Errorf("执行节点不存在: %s", currentID)
		}
		state.target.StepIndex++

		nextID, err := e.executeCompiledNode(state, current)
		if err != nil {
			onErrorID, ok := onErrorTargetID(current)
			if !ok {
				return err
			}
			state.ctx.Set("lastError", err.Error())
			currentID = onErrorID
			continue
		}
		if current.Node.NodeType == NodeTypeEnd {
			return nil
		}
		currentID = nextID
	}
	return nil
}

func onErrorTargetID(node CompiledNode) (string, bool) {
	if len(node.OnError) == 0 {
		return "", false
	}
	return node.OnError[0].TargetNodeID, true
}

func nextNodeID(node CompiledNode, ctx *RuntimeContext) (string, error) {
	if len(node.Next) == 0 {
		return "", nil
	}
	if len(node.Next) == 1 && node.Next[0].Condition == "" {
		return node.Next[0].TargetNodeID, nil
	}
	if node.Node.NodeType != NodeTypeConditionIf {
		return "", fmt.Errorf("节点 %s 存在分支，当前执行器暂不支持", node.Node.NodeID)
	}
	ok, err := EvalBoolExpression(stringConfig(node.Node.Config["expression"]), ctx)
	if err != nil {
		return "", err
	}
	expected := "false"
	if ok {
		expected = "true"
	}
	for _, edge := range node.Next {
		if edge.Condition == expected {
			return edge.TargetNodeID, nil
		}
	}
	return "", fmt.Errorf("条件节点缺少 %s 分支", expected)
}

func closeSession(session AutomationSession) error {
	if session == nil {
		return nil
	}
	return session.Close()
}
