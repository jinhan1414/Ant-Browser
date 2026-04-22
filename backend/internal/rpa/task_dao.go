package rpa

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type TaskDAO interface {
	UpsertTask(task *Task) error
	ListTasks() ([]*Task, error)
	GetTask(taskID string) (*Task, error)
	DeleteTask(taskID string) error
	ReplaceTargets(taskID string, targets []TaskTarget) error
	ListTargets(taskID string) ([]*TaskTarget, error)
}

type SQLiteTaskDAO struct {
	db *sql.DB
}

func NewSQLiteTaskDAO(db *sql.DB) *SQLiteTaskDAO {
	return &SQLiteTaskDAO{db: db}
}

func (d *SQLiteTaskDAO) UpsertTask(task *Task) error {
	if task == nil {
		return fmt.Errorf("任务不能为空")
	}
	now := time.Now().Format(time.RFC3339)
	prepareTaskForSave(task, now)
	_, err := d.db.Exec(`
		INSERT INTO rpa_tasks (task_id, task_name, flow_id, execution_order, task_type, schedule_config_json, enabled, last_run_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(task_id) DO UPDATE SET
			task_name = excluded.task_name,
			flow_id = excluded.flow_id,
			execution_order = excluded.execution_order,
			task_type = excluded.task_type,
			schedule_config_json = excluded.schedule_config_json,
			enabled = excluded.enabled,
			last_run_at = excluded.last_run_at,
			updated_at = excluded.updated_at`,
		task.TaskID, task.TaskName, task.FlowID, task.ExecutionOrder, task.TaskType, mustJSON(task.ScheduleConfig, "{}"), boolToInt(task.Enabled), task.LastRunAt, task.CreatedAt, task.UpdatedAt)
	if err != nil {
		return fmt.Errorf("保存任务失败: %w", err)
	}
	return nil
}

func (d *SQLiteTaskDAO) ListTasks() ([]*Task, error) {
	rows, err := d.db.Query(`SELECT task_id, task_name, flow_id, execution_order, task_type, schedule_config_json, enabled, last_run_at, created_at, updated_at FROM rpa_tasks ORDER BY updated_at DESC, created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}
	defer rows.Close()
	var items []*Task
	for rows.Next() {
		item, scanErr := scanTask(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (d *SQLiteTaskDAO) GetTask(taskID string) (*Task, error) {
	row := d.db.QueryRow(`SELECT task_id, task_name, flow_id, execution_order, task_type, schedule_config_json, enabled, last_run_at, created_at, updated_at FROM rpa_tasks WHERE task_id = ?`, strings.TrimSpace(taskID))
	item, err := scanTask(row)
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}
	return item, nil
}

func (d *SQLiteTaskDAO) DeleteTask(taskID string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("开启删除任务事务失败: %w", err)
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`DELETE FROM rpa_task_targets WHERE task_id = ?`, strings.TrimSpace(taskID)); err != nil {
		return fmt.Errorf("删除任务目标失败: %w", err)
	}
	if _, err = tx.Exec(`DELETE FROM rpa_tasks WHERE task_id = ?`, strings.TrimSpace(taskID)); err != nil {
		return fmt.Errorf("删除任务失败: %w", err)
	}
	return tx.Commit()
}

func (d *SQLiteTaskDAO) ReplaceTargets(taskID string, targets []TaskTarget) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("开启保存任务目标事务失败: %w", err)
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`DELETE FROM rpa_task_targets WHERE task_id = ?`, strings.TrimSpace(taskID)); err != nil {
		return fmt.Errorf("清理任务目标失败: %w", err)
	}
	for _, target := range targets {
		if _, err = tx.Exec(`INSERT INTO rpa_task_targets (task_id, profile_id, sort_order) VALUES (?, ?, ?)`,
			strings.TrimSpace(taskID), strings.TrimSpace(target.ProfileID), target.SortOrder); err != nil {
			return fmt.Errorf("保存任务目标失败: %w", err)
		}
	}
	return tx.Commit()
}

func (d *SQLiteTaskDAO) ListTargets(taskID string) ([]*TaskTarget, error) {
	rows, err := d.db.Query(`SELECT task_id, profile_id, sort_order FROM rpa_task_targets WHERE task_id = ? ORDER BY sort_order ASC, profile_id ASC`, strings.TrimSpace(taskID))
	if err != nil {
		return nil, fmt.Errorf("查询任务目标失败: %w", err)
	}
	defer rows.Close()
	var items []*TaskTarget
	for rows.Next() {
		item := &TaskTarget{}
		if err := rows.Scan(&item.TaskID, &item.ProfileID, &item.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func prepareTaskForSave(task *Task, now string) {
	if task.TaskID == "" {
		task.TaskID = uuid.NewString()
	}
	task.TaskName = strings.TrimSpace(task.TaskName)
	task.FlowID = strings.TrimSpace(task.FlowID)
	if task.ExecutionOrder == "" {
		task.ExecutionOrder = TaskExecutionSequential
	}
	if task.TaskType == "" {
		task.TaskType = TaskTypeManual
	}
	normalizeTaskScheduleConfig(task)
	if task.CreatedAt == "" {
		task.CreatedAt = now
	}
	task.UpdatedAt = now
}

func scanTask(s rowScanner) (*Task, error) {
	var item Task
	var enabled int
	var scheduleJSON string
	if err := s.Scan(&item.TaskID, &item.TaskName, &item.FlowID, &item.ExecutionOrder, &item.TaskType, &scheduleJSON, &enabled, &item.LastRunAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	item.Enabled = enabled == 1
	decodeJSON(scheduleJSON, &item.ScheduleConfig)
	if item.ScheduleConfig == nil {
		item.ScheduleConfig = map[string]any{}
	}
	return &item, nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
