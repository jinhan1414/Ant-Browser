package rpa

import (
	"sync"
	"testing"
	"time"
)

func TestParseCronExpressionSupportsQuartz(t *testing.T) {
	expr, err := parseCronExpression("0 * * * * ?")
	if err != nil {
		t.Fatalf("解析 Quartz Cron 失败: %v", err)
	}

	matchedTime := time.Date(2026, 4, 16, 10, 5, 0, 0, time.UTC)
	if !expr.matchesTime(matchedTime) {
		t.Fatal("Quartz Cron 在整分钟应匹配成功")
	}

	unmatchedTime := matchedTime.Add(time.Second)
	if expr.matchesTime(unmatchedTime) {
		t.Fatal("Quartz Cron 在非 0 秒不应匹配")
	}
}

func TestTaskSchedulerTriggersScheduledTaskOncePerMinute(t *testing.T) {
	task := &Task{
		TaskID:   "task-1",
		TaskType: TaskTypeScheduled,
		Enabled:  true,
		ScheduleConfig: map[string]any{
			"cron":     "0 * * * * ?",
			"timezone": "UTC",
		},
	}
	var mu sync.Mutex
	triggered := 0
	taskStore := map[string]*Task{task.TaskID: task}

	scheduler := NewTaskScheduler(
		func() ([]*Task, error) { return []*Task{taskStore[task.TaskID]}, nil },
		func(taskID string) (*Task, error) { return taskStore[taskID], nil },
		func(taskID string) error {
			mu.Lock()
			triggered++
			mu.Unlock()
			return nil
		},
		time.Second,
	)

	now := time.Date(2026, 4, 16, 10, 5, 0, 0, time.UTC)
	scheduler.nowFunc = func() time.Time { return now }
	scheduler.RunDueTasks()
	waitSchedulerWorker()
	assertTriggeredCount(t, &mu, &triggered, 1)

	scheduler.RunDueTasks()
	waitSchedulerWorker()
	assertTriggeredCount(t, &mu, &triggered, 1)

	now = now.Add(time.Minute)
	scheduler.RunDueTasks()
	waitSchedulerWorker()
	assertTriggeredCount(t, &mu, &triggered, 2)
}

func TestTaskSchedulerUsesUpdatedCronExpression(t *testing.T) {
	task := &Task{
		TaskID:   "task-1",
		TaskType: TaskTypeScheduled,
		Enabled:  true,
		ScheduleConfig: map[string]any{
			"cron":     "0 * * * * ?",
			"timezone": "UTC",
		},
	}
	taskStore := map[string]*Task{task.TaskID: task}
	var mu sync.Mutex
	triggered := 0

	scheduler := NewTaskScheduler(
		func() ([]*Task, error) { return []*Task{taskStore[task.TaskID]}, nil },
		func(taskID string) (*Task, error) { return taskStore[taskID], nil },
		func(taskID string) error {
			mu.Lock()
			triggered++
			mu.Unlock()
			return nil
		},
		time.Second,
	)

	now := time.Date(2026, 4, 16, 10, 5, 0, 0, time.UTC)
	scheduler.nowFunc = func() time.Time { return now }
	scheduler.RunDueTasks()
	waitSchedulerWorker()
	assertTriggeredCount(t, &mu, &triggered, 1)

	taskStore[task.TaskID] = &Task{
		TaskID:   task.TaskID,
		TaskType: TaskTypeScheduled,
		Enabled:  true,
		ScheduleConfig: map[string]any{
			"cron":     "0 */10 * * * ?",
			"timezone": "UTC",
		},
	}

	now = time.Date(2026, 4, 16, 10, 6, 0, 0, time.UTC)
	scheduler.RunDueTasks()
	waitSchedulerWorker()
	assertTriggeredCount(t, &mu, &triggered, 1)

	now = time.Date(2026, 4, 16, 10, 10, 0, 0, time.UTC)
	scheduler.RunDueTasks()
	waitSchedulerWorker()
	assertTriggeredCount(t, &mu, &triggered, 2)
}

func TestTaskSchedulerSkipsStaleReservationAfterTaskEdited(t *testing.T) {
	task := &Task{
		TaskID:   "task-1",
		TaskType: TaskTypeScheduled,
		Enabled:  true,
		ScheduleConfig: map[string]any{
			"cron":     "0 * * * * ?",
			"timezone": "UTC",
		},
	}
	taskStore := map[string]*Task{task.TaskID: task}
	startCh := make(chan struct{})
	releaseCh := make(chan struct{})
	var mu sync.Mutex
	triggered := 0

	scheduler := NewTaskScheduler(
		func() ([]*Task, error) { return []*Task{taskStore[task.TaskID]}, nil },
		func(taskID string) (*Task, error) {
			close(startCh)
			<-releaseCh
			return taskStore[taskID], nil
		},
		func(taskID string) error {
			mu.Lock()
			triggered++
			mu.Unlock()
			return nil
		},
		time.Second,
	)

	now := time.Date(2026, 4, 16, 10, 5, 0, 0, time.UTC)
	scheduler.nowFunc = func() time.Time { return now }
	scheduler.RunDueTasks()
	<-startCh

	taskStore[task.TaskID] = &Task{
		TaskID:   task.TaskID,
		TaskType: TaskTypeScheduled,
		Enabled:  true,
		ScheduleConfig: map[string]any{
			"cron":     "0 */10 * * * ?",
			"timezone": "UTC",
		},
	}

	close(releaseCh)
	waitSchedulerWorker()
	assertTriggeredCount(t, &mu, &triggered, 0)
}

func assertTriggeredCount(t *testing.T, mu *sync.Mutex, count *int, expected int) {
	t.Helper()
	mu.Lock()
	defer mu.Unlock()
	if *count != expected {
		t.Fatalf("触发次数错误: got=%d want=%d", *count, expected)
	}
}

func waitSchedulerWorker() {
	time.Sleep(20 * time.Millisecond)
}
