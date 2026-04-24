package rpa

import (
	"fmt"
	"time"

	"ant-chrome/backend/internal/browser"

	"github.com/google/uuid"
)

type nodeResult struct {
	profile *browser.Profile
	output  any
}

func (e *Executor) executeCompiledNode(state *executionState, node CompiledNode) (string, error) {
	if _, err := e.executeNodeWithRetry(state, node.Node); err != nil {
		return "", err
	}
	return nextNodeID(node, state.ctx)
}

func (e *Executor) executeNodeWithRetry(state *executionState, node FlowNode) (*nodeResult, error) {
	maxAttempts := nodeMaxAttempts(node)
	intervalMs := nodeRetryIntervalMs(node)
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		startedAt := time.Now().Format(time.RFC3339)
		result, err := e.executeNodeAttempt(state, node)
		state.appendStep(newRunStep(state.target, node, attempt, startedAt, result, err))
		if err == nil {
			state.updateProfile(result)
			return result, nil
		}
		lastErr = err
		if attempt < maxAttempts && intervalMs > 0 {
			time.Sleep(time.Duration(intervalMs) * time.Millisecond)
		}
	}
	return nil, lastErr
}

func (e *Executor) executeNodeAttempt(state *executionState, node FlowNode) (*nodeResult, error) {
	switch node.NodeType {
	case NodeTypeStart, NodeTypeEnd:
		return &nodeResult{}, nil
	case NodeTypeBrowserStart:
		return startProfileResult(e.operator.Start(state.target.ProfileID, nodeLaunchArgs(node), nodeStartURLs(node), nodeSkipDefault(node)))
	case NodeTypeBrowserOpenURL:
		return startProfileResult(e.operator.Start(state.target.ProfileID, nil, nodeStartURLs(node), true))
	case NodeTypeBrowserClick:
		return e.executeClick(state, node)
	case NodeTypeBrowserInput:
		return e.executeInput(state, node)
	case NodeTypeBrowserReadText:
		return e.executeReadText(state, node)
	case NodeTypeSystemNotify:
		return e.executeNotify(state, node)
	case NodeTypeConditionIf:
		return e.executeIf(state, node)
	case NodeTypeDelay:
		waitDuration(node)
		return &nodeResult{output: map[string]any{"durationMs": durationMs(node)}}, nil
	case NodeTypeBrowserStop:
		_ = closeSession(state.session)
		state.session = nil
		return startProfileResult(e.operator.Stop(state.target.ProfileID))
	case NodeTypeFail:
		return nil, fmt.Errorf(stringConfig(node.Config["message"]))
	default:
		return nil, unsupportedNodeError(node.NodeType)
	}
}

func (e *Executor) executeClick(state *executionState, node FlowNode) (*nodeResult, error) {
	session, err := e.ensureSession(state)
	if err != nil {
		return nil, err
	}
	selector, err := resolveNodeString(node, "selector", state.ctx)
	if err != nil {
		return nil, err
	}
	if err := waitVisibleIfNeeded(session, node, selector); err != nil {
		return nil, err
	}
	return &nodeResult{output: map[string]any{"selector": selector}}, session.Click(selector)
}

func (e *Executor) executeInput(state *executionState, node FlowNode) (*nodeResult, error) {
	session, err := e.ensureSession(state)
	if err != nil {
		return nil, err
	}
	selector, err := resolveNodeString(node, "selector", state.ctx)
	if err != nil {
		return nil, err
	}
	value, err := resolveNodeString(node, "value", state.ctx)
	if err != nil {
		return nil, err
	}
	if err := waitVisibleIfNeeded(session, node, selector); err != nil {
		return nil, err
	}
	return &nodeResult{output: map[string]any{"selector": selector, "value": value}}, session.InputText(selector, value)
}

func (e *Executor) executeReadText(state *executionState, node FlowNode) (*nodeResult, error) {
	session, err := e.ensureSession(state)
	if err != nil {
		return nil, err
	}
	selector, err := resolveNodeString(node, "selector", state.ctx)
	if err != nil {
		return nil, err
	}
	if err := waitVisibleIfNeeded(session, node, selector); err != nil {
		return nil, err
	}
	text, err := session.ReadText(selector)
	if err != nil {
		return nil, err
	}
	saveAs := stringConfig(node.Config["saveAs"])
	if saveAs != "" {
		state.ctx.Set(saveAs, text)
	}
	return &nodeResult{output: map[string]any{"selector": selector, "text": text, "saveAs": saveAs}}, nil
}

func (e *Executor) executeNotify(state *executionState, node FlowNode) (*nodeResult, error) {
	title, err := resolveNodeString(node, "title", state.ctx)
	if err != nil {
		return nil, err
	}
	body, err := resolveNodeString(node, "body", state.ctx)
	if err != nil {
		return nil, err
	}
	if err := e.notifier.Notify(title, body); err != nil {
		return nil, err
	}
	return &nodeResult{output: map[string]any{"title": title, "body": body}}, nil
}

func (e *Executor) executeIf(state *executionState, node FlowNode) (*nodeResult, error) {
	expr := stringConfig(node.Config["expression"])
	ok, err := EvalBoolExpression(expr, state.ctx)
	if err != nil {
		return nil, err
	}
	return &nodeResult{output: map[string]any{"expression": expr, "matched": ok}}, nil
}

func (e *Executor) ensureSession(state *executionState) (AutomationSession, error) {
	if state.session != nil {
		return state.session, nil
	}
	if state.target.DebugPort <= 0 {
		return nil, fmt.Errorf("实例 %s 调试端口不可用", state.target.ProfileID)
	}
	session, err := e.sessionFactory(state.target.DebugPort)
	if err != nil {
		return nil, err
	}
	state.session = session
	return session, nil
}

func (s *executionState) updateProfile(result *nodeResult) {
	if result == nil || result.profile == nil || result.profile.DebugPort <= 0 {
		return
	}
	s.target.DebugPort = result.profile.DebugPort
	if result.profile.ProfileName != "" {
		s.target.ProfileName = result.profile.ProfileName
	}
}

func (s *executionState) appendStep(step *RunStep) {
	if step != nil {
		s.steps = append(s.steps, step)
	}
}

func newRunStep(target *RunTarget, node FlowNode, attempt int, startedAt string, result *nodeResult, err error) *RunStep {
	step := &RunStep{
		RunStepID:    uuid.NewString(),
		RunID:        target.RunID,
		RunTargetID:  target.RunTargetID,
		ProfileID:    target.ProfileID,
		NodeID:       node.NodeID,
		NodeType:     string(node.NodeType),
		NodeLabel:    node.Label,
		Status:       RunStatusSuccess,
		Attempt:      attempt,
		OutputJSON:   "",
		ErrorMessage: "",
		StartedAt:    startedAt,
		FinishedAt:   time.Now().Format(time.RFC3339),
	}
	if result != nil && result.output != nil {
		step.OutputJSON = mustJSON(result.output, "{}")
	}
	if err != nil {
		step.Status = RunStatusFailed
		step.ErrorMessage = err.Error()
	}
	return step
}

func startProfileResult(profile *browser.Profile, err error) (*nodeResult, error) {
	if err != nil {
		return nil, err
	}
	output := map[string]any{}
	if profile != nil {
		output["debugPort"] = profile.DebugPort
		output["profileName"] = profile.ProfileName
	}
	return &nodeResult{profile: profile, output: output}, nil
}

func resolveNodeString(node FlowNode, key string, ctx *RuntimeContext) (string, error) {
	raw := stringConfig(node.Config[key])
	if raw == "" {
		return "", nil
	}
	return ResolveTemplate(raw, ctx)
}

func waitVisibleIfNeeded(session AutomationSession, node FlowNode, selector string) error {
	timeout := int(numberConfigOrDefault(node.Config["timeoutMs"], 3000))
	if timeout <= 0 {
		return nil
	}
	return session.WaitVisible(selector, timeout)
}

func nodeMaxAttempts(node FlowNode) int {
	value := int(numberConfigOrDefault(node.Config["maxAttempts"], 1))
	if value <= 0 {
		return 1
	}
	return value
}

func nodeRetryIntervalMs(node FlowNode) int {
	value := int(numberConfigOrDefault(node.Config["intervalMs"], 0))
	if value < 0 {
		return 0
	}
	return value
}

func durationMs(node FlowNode) int {
	value := int(numberConfigOrDefault(node.Config["durationMs"], 0))
	if value < 0 {
		return 0
	}
	return value
}

func waitDuration(node FlowNode) {
	if value := durationMs(node); value > 0 {
		time.Sleep(time.Duration(value) * time.Millisecond)
	}
}

func numberConfigOrDefault(raw any, fallback float64) float64 {
	value, ok := numberConfig(raw)
	if !ok {
		return fallback
	}
	return value
}
