package rpa

import (
	"fmt"
	"sync"
	"time"
)

type TaskListFunc func() ([]*Task, error)
type TaskGetFunc func(taskID string) (*Task, error)
type TaskTriggerFunc func(taskID string) error

type TaskScheduler struct {
	listTasks   TaskListFunc
	getTask     TaskGetFunc
	triggerTask TaskTriggerFunc
	interval    time.Duration
	nowFunc     func() time.Time
	stopCh      chan struct{}

	mu            sync.Mutex
	running       bool
	lastTriggered map[string]string
	activeTasks   map[string]bool
	fingerprints  map[string]string
}

func NewTaskScheduler(listTasks TaskListFunc, getTask TaskGetFunc, triggerTask TaskTriggerFunc, interval time.Duration) *TaskScheduler {
	if interval <= 0 {
		interval = time.Second
	}
	return &TaskScheduler{
		listTasks:      listTasks,
		getTask:        getTask,
		triggerTask:    triggerTask,
		interval:       interval,
		nowFunc:        time.Now,
		stopCh:         make(chan struct{}),
		lastTriggered:  make(map[string]string),
		activeTasks:    make(map[string]bool),
		fingerprints:   make(map[string]string),
	}
}

func (s *TaskScheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return
	}
	s.running = true
	go s.loop()
}

func (s *TaskScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	s.running = false
	close(s.stopCh)
}

func (s *TaskScheduler) RunDueTasks() {
	if s.listTasks == nil || s.triggerTask == nil {
		return
	}
	tasks, err := s.listTasks()
	if err != nil {
		return
	}
	now := s.nowFunc()
	for _, task := range tasks {
		s.evaluateTask(now, task)
	}
}

func (s *TaskScheduler) loop() {
	s.RunDueTasks()
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.RunDueTasks()
		case <-s.stopCh:
			return
		}
	}
}

func (s *TaskScheduler) evaluateTask(now time.Time, task *Task) {
	fingerprint := taskScheduleFingerprint(task)
	s.syncTaskFingerprint(task.TaskID, fingerprint)

	dueKey, ok := dueTaskKey(task, now)
	if !ok {
		return
	}
	if !s.reserveTask(task.TaskID, dueKey, fingerprint) {
		return
	}
	go s.runTask(task.TaskID, dueKey, fingerprint)
}

func dueTaskKey(task *Task, now time.Time) (string, bool) {
	if task == nil || !task.Enabled || task.TaskType != TaskTypeScheduled {
		return "", false
	}
	location, err := taskScheduleLocation(task)
	if err != nil {
		return "", false
	}
	expr, err := parseCronExpression(stringConfig(task.ScheduleConfig[scheduleConfigKeyCron]))
	if err != nil {
		return "", false
	}
	scheduledTime := now.In(location)
	if !expr.matchesTime(scheduledTime) {
		return "", false
	}
	return formatDueKey(expr, scheduledTime), true
}

func taskScheduleLocation(task *Task) (*time.Location, error) {
	if task == nil {
		return nil, fmt.Errorf("任务不能为空")
	}
	timezone := stringConfig(task.ScheduleConfig[scheduleConfigKeyTimezone])
	return time.LoadLocation(timezone)
}

func formatDueKey(expr *cronExpression, now time.Time) string {
	if expr == nil || !expr.second.wildcard {
		return now.Format(time.RFC3339)
	}
	return now.Format("2006-01-02T15:04")
}

func (s *TaskScheduler) reserveTask(taskID string, dueKey string, fingerprint string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activeTasks[taskID] || s.lastTriggered[taskID] == dueKey {
		return false
	}
	s.activeTasks[taskID] = true
	s.lastTriggered[taskID] = dueKey
	s.fingerprints[taskID] = fingerprint
	return true
}

func (s *TaskScheduler) runTask(taskID string, dueKey string, fingerprint string) {
	defer s.releaseTask(taskID)
	if !s.shouldTriggerTask(taskID, dueKey, fingerprint) {
		return
	}
	_ = s.triggerTask(taskID)
}

func (s *TaskScheduler) releaseTask(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.activeTasks, taskID)
}

func (s *TaskScheduler) shouldTriggerTask(taskID string, dueKey string, fingerprint string) bool {
	if s.getTask == nil {
		return true
	}
	task, err := s.getTask(taskID)
	if err != nil || task == nil {
		return false
	}
	if taskScheduleFingerprint(task) != fingerprint {
		s.resetTaskCache(taskID)
		return false
	}
	currentDueKey, ok := dueTaskKey(task, s.nowFunc())
	if !ok || currentDueKey != dueKey {
		s.resetTaskCache(taskID)
		return false
	}
	return true
}

func (s *TaskScheduler) syncTaskFingerprint(taskID string, fingerprint string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if fingerprint == "" {
		return
	}
	if previous, ok := s.fingerprints[taskID]; ok && previous != fingerprint {
		delete(s.lastTriggered, taskID)
	}
	s.fingerprints[taskID] = fingerprint
}

func (s *TaskScheduler) resetTaskCache(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.lastTriggered, taskID)
}
