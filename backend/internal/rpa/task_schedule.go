package rpa

import (
	"fmt"
	"strings"
	"time"
)

const (
	defaultTaskScheduleTimezone = "Asia/Shanghai"
	scheduleConfigKeyCron       = "cron"
	scheduleConfigKeyTimezone   = "timezone"
)

func normalizeTaskScheduleConfig(task *Task) {
	if task.ScheduleConfig == nil {
		task.ScheduleConfig = map[string]any{}
	}
	if task.TaskType != TaskTypeScheduled {
		task.ScheduleConfig = map[string]any{}
		return
	}

	cron := strings.TrimSpace(stringConfig(task.ScheduleConfig[scheduleConfigKeyCron]))
	timezone := strings.TrimSpace(stringConfig(task.ScheduleConfig[scheduleConfigKeyTimezone]))
	if timezone == "" {
		timezone = defaultTaskScheduleTimezone
	}

	task.ScheduleConfig = map[string]any{
		scheduleConfigKeyCron:     cron,
		scheduleConfigKeyTimezone: timezone,
	}
}

func validateTaskScheduleConfig(task *Task) error {
	if task == nil || task.TaskType != TaskTypeScheduled {
		return nil
	}
	cronExpr := strings.TrimSpace(stringConfig(task.ScheduleConfig[scheduleConfigKeyCron]))
	if cronExpr == "" {
		return fmt.Errorf("计划任务的 Cron 表达式不能为空")
	}
	if _, err := parseCronExpression(cronExpr); err != nil {
		return err
	}
	if _, err := time.LoadLocation(strings.TrimSpace(stringConfig(task.ScheduleConfig[scheduleConfigKeyTimezone]))); err != nil {
		return fmt.Errorf("计划任务时区无效: %w", err)
	}
	return nil
}

func taskScheduleFingerprint(task *Task) string {
	if task == nil {
		return ""
	}
	return strings.Join([]string{
		task.TaskID,
		string(task.TaskType),
		fmt.Sprintf("%t", task.Enabled),
		strings.TrimSpace(stringConfig(task.ScheduleConfig[scheduleConfigKeyCron])),
		strings.TrimSpace(stringConfig(task.ScheduleConfig[scheduleConfigKeyTimezone])),
	}, "|")
}
