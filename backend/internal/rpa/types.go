package rpa

type StepType string

const (
	StepTypeStartBrowser StepType = "start_browser"
	StepTypeOpenURLs     StepType = "open_urls"
	StepTypeWait         StepType = "wait"
	StepTypeStopBrowser  StepType = "stop_browser"
)

type TaskExecutionOrder string

const (
	TaskExecutionSequential TaskExecutionOrder = "sequential"
	TaskExecutionRandom     TaskExecutionOrder = "random"
)

type TaskType string

const (
	TaskTypeManual    TaskType = "manual"
	TaskTypeScheduled TaskType = "scheduled"
)

type RunStatus string

const (
	RunStatusPending   RunStatus = "pending"
	RunStatusRunning   RunStatus = "running"
	RunStatusSuccess   RunStatus = "success"
	RunStatusFailed    RunStatus = "failed"
	RunStatusCancelled RunStatus = "cancelled"
)

type RunTriggerType string

const (
	RunTriggerManual    RunTriggerType = "manual"
	RunTriggerScheduled RunTriggerType = "scheduled"
)

type FlowGroup struct {
	GroupID   string `json:"groupId"`
	GroupName string `json:"groupName"`
	SortOrder int    `json:"sortOrder"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type FlowGroupInput struct {
	GroupName string `json:"groupName"`
	SortOrder int    `json:"sortOrder"`
}

type FlowStep struct {
	StepID   string         `json:"stepId"`
	StepName string         `json:"stepName"`
	StepType StepType       `json:"stepType"`
	Config   map[string]any `json:"config"`
}

type Flow struct {
	FlowID     string         `json:"flowId"`
	FlowName   string         `json:"flowName"`
	GroupID    string         `json:"groupId"`
	Steps      []FlowStep     `json:"steps"`
	Document   FlowDocument   `json:"document"`
	SourceType FlowSourceType `json:"sourceType"`
	SourceXML  string         `json:"sourceXml"`
	ShareCode  string         `json:"shareCode"`
	Version    int            `json:"version"`
	CreatedAt  string         `json:"createdAt"`
	UpdatedAt  string         `json:"updatedAt"`
}

type Task struct {
	TaskID         string             `json:"taskId"`
	TaskName       string             `json:"taskName"`
	FlowID         string             `json:"flowId"`
	ExecutionOrder TaskExecutionOrder `json:"executionOrder"`
	TaskType       TaskType           `json:"taskType"`
	ScheduleConfig map[string]any     `json:"scheduleConfig"`
	Enabled        bool               `json:"enabled"`
	LastRunAt      string             `json:"lastRunAt"`
	CreatedAt      string             `json:"createdAt"`
	UpdatedAt      string             `json:"updatedAt"`
}

type TaskTarget struct {
	TaskID    string `json:"taskId"`
	ProfileID string `json:"profileId"`
	SortOrder int    `json:"sortOrder"`
}

type TaskDetail struct {
	Task    *Task         `json:"task"`
	Targets []*TaskTarget `json:"targets"`
}

type Run struct {
	RunID        string         `json:"runId"`
	TaskID       string         `json:"taskId"`
	FlowID       string         `json:"flowId"`
	TriggerType  RunTriggerType `json:"triggerType"`
	Status       RunStatus      `json:"status"`
	Summary      string         `json:"summary"`
	StartedAt    string         `json:"startedAt"`
	FinishedAt   string         `json:"finishedAt"`
	ErrorMessage string         `json:"errorMessage"`
}

type RunTarget struct {
	RunTargetID  string    `json:"runTargetId"`
	RunID        string    `json:"runId"`
	ProfileID    string    `json:"profileId"`
	ProfileName  string    `json:"profileName"`
	Status       RunStatus `json:"status"`
	StepIndex    int       `json:"stepIndex"`
	StartedAt    string    `json:"startedAt"`
	FinishedAt   string    `json:"finishedAt"`
	ErrorMessage string    `json:"errorMessage"`
	DebugPort    int       `json:"debugPort"`
}

type Template struct {
	TemplateID   string   `json:"templateId"`
	TemplateName string   `json:"templateName"`
	Description  string   `json:"description"`
	Tags         []string `json:"tags"`
	FlowSnapshot Flow     `json:"flowSnapshot"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
}
