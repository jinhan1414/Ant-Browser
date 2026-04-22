package rpa

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type TemplateDAO interface {
	UpsertTemplate(template *Template) error
	ListTemplates() ([]*Template, error)
	GetTemplate(templateID string) (*Template, error)
	DeleteTemplate(templateID string) error
}

type SQLiteTemplateDAO struct {
	db *sql.DB
}

func NewSQLiteTemplateDAO(db *sql.DB) *SQLiteTemplateDAO {
	return &SQLiteTemplateDAO{db: db}
}

func (d *SQLiteTemplateDAO) UpsertTemplate(template *Template) error {
	if template == nil {
		return fmt.Errorf("模板不能为空")
	}
	now := time.Now().Format(time.RFC3339)
	prepareTemplateForSave(template, now)
	_, err := d.db.Exec(`
		INSERT INTO rpa_templates (template_id, template_name, description, tags_json, flow_snapshot_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(template_id) DO UPDATE SET
			template_name = excluded.template_name,
			description = excluded.description,
			tags_json = excluded.tags_json,
			flow_snapshot_json = excluded.flow_snapshot_json,
			updated_at = excluded.updated_at`,
		template.TemplateID, template.TemplateName, template.Description, mustJSON(template.Tags, "[]"), mustJSON(template.FlowSnapshot, "{}"), template.CreatedAt, template.UpdatedAt)
	if err != nil {
		return fmt.Errorf("保存模板失败: %w", err)
	}
	return nil
}

func (d *SQLiteTemplateDAO) ListTemplates() ([]*Template, error) {
	rows, err := d.db.Query(`SELECT template_id, template_name, description, tags_json, flow_snapshot_json, created_at, updated_at FROM rpa_templates ORDER BY updated_at DESC, created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("查询模板失败: %w", err)
	}
	defer rows.Close()
	var items []*Template
	for rows.Next() {
		item, scanErr := scanTemplate(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (d *SQLiteTemplateDAO) GetTemplate(templateID string) (*Template, error) {
	row := d.db.QueryRow(`SELECT template_id, template_name, description, tags_json, flow_snapshot_json, created_at, updated_at FROM rpa_templates WHERE template_id = ?`, strings.TrimSpace(templateID))
	item, err := scanTemplate(row)
	if err != nil {
		return nil, fmt.Errorf("查询模板失败: %w", err)
	}
	return item, nil
}

func (d *SQLiteTemplateDAO) DeleteTemplate(templateID string) error {
	_, err := d.db.Exec(`DELETE FROM rpa_templates WHERE template_id = ?`, strings.TrimSpace(templateID))
	if err != nil {
		return fmt.Errorf("删除模板失败: %w", err)
	}
	return nil
}

func prepareTemplateForSave(template *Template, now string) {
	if template.TemplateID == "" {
		template.TemplateID = uuid.NewString()
	}
	template.TemplateName = strings.TrimSpace(template.TemplateName)
	template.Description = strings.TrimSpace(template.Description)
	if template.Tags == nil {
		template.Tags = []string{}
	}
	if template.CreatedAt == "" {
		template.CreatedAt = now
	}
	template.UpdatedAt = now
}

func scanTemplate(s rowScanner) (*Template, error) {
	var item Template
	var tagsJSON string
	var flowSnapshotJSON string
	if err := s.Scan(&item.TemplateID, &item.TemplateName, &item.Description, &tagsJSON, &flowSnapshotJSON, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	decodeJSON(tagsJSON, &item.Tags)
	decodeJSON(flowSnapshotJSON, &item.FlowSnapshot)
	if item.Tags == nil {
		item.Tags = []string{}
	}
	return &item, nil
}
