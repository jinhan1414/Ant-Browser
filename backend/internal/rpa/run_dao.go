package rpa

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type RunDAO interface {
	CreateRun(run *Run) error
	UpdateRun(run *Run) error
	ListRuns() ([]*Run, error)
	CreateRunTarget(target *RunTarget) error
	UpdateRunTarget(target *RunTarget) error
	ListRunTargets(runID string) ([]*RunTarget, error)
}

type SQLiteRunDAO struct {
	db *sql.DB
}

func NewSQLiteRunDAO(db *sql.DB) *SQLiteRunDAO {
	return &SQLiteRunDAO{db: db}
}

func (d *SQLiteRunDAO) CreateRun(run *Run) error {
	if run == nil {
		return fmt.Errorf("运行记录不能为空")
	}
	now := time.Now().Format(time.RFC3339)
	prepareRunForSave(run, now)
	_, err := d.db.Exec(`INSERT INTO rpa_runs (run_id, task_id, flow_id, trigger_type, status, summary, started_at, finished_at, error_message) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		run.RunID, run.TaskID, run.FlowID, run.TriggerType, run.Status, run.Summary, run.StartedAt, run.FinishedAt, run.ErrorMessage)
	if err != nil {
		return fmt.Errorf("创建运行记录失败: %w", err)
	}
	return nil
}

func (d *SQLiteRunDAO) UpdateRun(run *Run) error {
	if run == nil {
		return fmt.Errorf("运行记录不能为空")
	}
	_, err := d.db.Exec(`UPDATE rpa_runs SET status = ?, summary = ?, started_at = ?, finished_at = ?, error_message = ? WHERE run_id = ?`,
		run.Status, run.Summary, run.StartedAt, run.FinishedAt, run.ErrorMessage, strings.TrimSpace(run.RunID))
	if err != nil {
		return fmt.Errorf("更新运行记录失败: %w", err)
	}
	return nil
}

func (d *SQLiteRunDAO) ListRuns() ([]*Run, error) {
	rows, err := d.db.Query(`SELECT run_id, task_id, flow_id, trigger_type, status, summary, started_at, finished_at, error_message FROM rpa_runs ORDER BY started_at DESC, run_id DESC`)
	if err != nil {
		return nil, fmt.Errorf("查询运行记录失败: %w", err)
	}
	defer rows.Close()
	var items []*Run
	for rows.Next() {
		item := &Run{}
		if err := rows.Scan(&item.RunID, &item.TaskID, &item.FlowID, &item.TriggerType, &item.Status, &item.Summary, &item.StartedAt, &item.FinishedAt, &item.ErrorMessage); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (d *SQLiteRunDAO) CreateRunTarget(target *RunTarget) error {
	if target == nil {
		return fmt.Errorf("运行目标不能为空")
	}
	now := time.Now().Format(time.RFC3339)
	prepareRunTargetForSave(target, now)
	_, err := d.db.Exec(`INSERT INTO rpa_run_targets (run_target_id, run_id, profile_id, profile_name, status, step_index, started_at, finished_at, error_message, debug_port) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		target.RunTargetID, target.RunID, target.ProfileID, target.ProfileName, target.Status, target.StepIndex, target.StartedAt, target.FinishedAt, target.ErrorMessage, target.DebugPort)
	if err != nil {
		return fmt.Errorf("创建运行目标失败: %w", err)
	}
	return nil
}

func (d *SQLiteRunDAO) UpdateRunTarget(target *RunTarget) error {
	if target == nil {
		return fmt.Errorf("运行目标不能为空")
	}
	_, err := d.db.Exec(`UPDATE rpa_run_targets SET profile_name = ?, status = ?, step_index = ?, started_at = ?, finished_at = ?, error_message = ?, debug_port = ? WHERE run_target_id = ?`,
		target.ProfileName, target.Status, target.StepIndex, target.StartedAt, target.FinishedAt, target.ErrorMessage, target.DebugPort, strings.TrimSpace(target.RunTargetID))
	if err != nil {
		return fmt.Errorf("更新运行目标失败: %w", err)
	}
	return nil
}

func (d *SQLiteRunDAO) ListRunTargets(runID string) ([]*RunTarget, error) {
	rows, err := d.db.Query(`SELECT run_target_id, run_id, profile_id, profile_name, status, step_index, started_at, finished_at, error_message, debug_port FROM rpa_run_targets WHERE run_id = ? ORDER BY started_at ASC, run_target_id ASC`, strings.TrimSpace(runID))
	if err != nil {
		return nil, fmt.Errorf("查询运行目标失败: %w", err)
	}
	defer rows.Close()
	var items []*RunTarget
	for rows.Next() {
		item := &RunTarget{}
		if err := rows.Scan(&item.RunTargetID, &item.RunID, &item.ProfileID, &item.ProfileName, &item.Status, &item.StepIndex, &item.StartedAt, &item.FinishedAt, &item.ErrorMessage, &item.DebugPort); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func prepareRunForSave(run *Run, now string) {
	if run.RunID == "" {
		run.RunID = uuid.NewString()
	}
	if run.TriggerType == "" {
		run.TriggerType = RunTriggerManual
	}
	if run.Status == "" {
		run.Status = RunStatusPending
	}
	if run.StartedAt == "" {
		run.StartedAt = now
	}
}

func prepareRunTargetForSave(target *RunTarget, now string) {
	if target.RunTargetID == "" {
		target.RunTargetID = uuid.NewString()
	}
	if target.Status == "" {
		target.Status = RunStatusPending
	}
	if target.StartedAt == "" {
		target.StartedAt = now
	}
}
