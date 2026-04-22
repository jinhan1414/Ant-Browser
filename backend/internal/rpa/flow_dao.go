package rpa

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type FlowDAO interface {
	CreateGroup(input FlowGroupInput) (*FlowGroup, error)
	ListGroups() ([]*FlowGroup, error)
	UpsertFlow(flow *Flow) error
	ListFlows(keyword string, groupID string) ([]*Flow, error)
	GetFlow(flowID string) (*Flow, error)
	GetFlowByShareCode(shareCode string) (*Flow, error)
	DeleteFlow(flowID string) error
}

type SQLiteFlowDAO struct {
	db *sql.DB
}

func NewSQLiteFlowDAO(db *sql.DB) *SQLiteFlowDAO {
	return &SQLiteFlowDAO{db: db}
}

func (d *SQLiteFlowDAO) CreateGroup(input FlowGroupInput) (*FlowGroup, error) {
	now := time.Now().Format(time.RFC3339)
	group := &FlowGroup{
		GroupID:   uuid.NewString(),
		GroupName: strings.TrimSpace(input.GroupName),
		SortOrder: input.SortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if group.GroupName == "" {
		return nil, fmt.Errorf("流程分组名称不能为空")
	}
	_, err := d.db.Exec(`INSERT INTO rpa_flow_groups (group_id, group_name, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		group.GroupID, group.GroupName, group.SortOrder, group.CreatedAt, group.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("创建流程分组失败: %w", err)
	}
	return group, nil
}

func (d *SQLiteFlowDAO) ListGroups() ([]*FlowGroup, error) {
	rows, err := d.db.Query(`SELECT group_id, group_name, sort_order, created_at, updated_at FROM rpa_flow_groups ORDER BY sort_order ASC, created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("查询流程分组失败: %w", err)
	}
	defer rows.Close()
	var items []*FlowGroup
	for rows.Next() {
		item, scanErr := scanFlowGroup(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (d *SQLiteFlowDAO) UpsertFlow(flow *Flow) error {
	if flow == nil {
		return fmt.Errorf("流程不能为空")
	}
	now := time.Now().Format(time.RFC3339)
	prepareFlowForSave(flow, now)
	_, err := d.db.Exec(`
		INSERT INTO rpa_flows (flow_id, flow_name, group_id, steps_json, document_json, source_type, source_xml, share_code, version, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(flow_id) DO UPDATE SET
			flow_name = excluded.flow_name,
			group_id = excluded.group_id,
			steps_json = excluded.steps_json,
			document_json = excluded.document_json,
			source_type = excluded.source_type,
			source_xml = excluded.source_xml,
			share_code = excluded.share_code,
			version = excluded.version,
			updated_at = excluded.updated_at`,
		flow.FlowID, flow.FlowName, flow.GroupID, mustJSON(flow.Steps, "[]"), mustJSON(flow.Document, "{}"), flow.SourceType, flow.SourceXML, flow.ShareCode, flow.Version, flow.CreatedAt, flow.UpdatedAt)
	if err != nil {
		return fmt.Errorf("保存流程失败: %w", err)
	}
	return nil
}

func (d *SQLiteFlowDAO) ListFlows(keyword string, groupID string) ([]*Flow, error) {
	query := `SELECT flow_id, flow_name, group_id, steps_json, document_json, source_type, source_xml, share_code, version, created_at, updated_at FROM rpa_flows WHERE 1=1`
	args := make([]any, 0, 2)
	if strings.TrimSpace(groupID) != "" {
		query += ` AND group_id = ?`
		args = append(args, strings.TrimSpace(groupID))
	}
	if strings.TrimSpace(keyword) != "" {
		query += ` AND flow_name LIKE ?`
		args = append(args, "%"+strings.TrimSpace(keyword)+"%")
	}
	query += ` ORDER BY updated_at DESC, created_at DESC`
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询流程失败: %w", err)
	}
	defer rows.Close()
	var items []*Flow
	for rows.Next() {
		item, scanErr := scanFlow(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (d *SQLiteFlowDAO) GetFlow(flowID string) (*Flow, error) {
	row := d.db.QueryRow(`SELECT flow_id, flow_name, group_id, steps_json, document_json, source_type, source_xml, share_code, version, created_at, updated_at FROM rpa_flows WHERE flow_id = ?`, strings.TrimSpace(flowID))
	item, err := scanFlow(row)
	if err != nil {
		return nil, fmt.Errorf("查询流程失败: %w", err)
	}
	return item, nil
}

func (d *SQLiteFlowDAO) GetFlowByShareCode(shareCode string) (*Flow, error) {
	row := d.db.QueryRow(`SELECT flow_id, flow_name, group_id, steps_json, document_json, source_type, source_xml, share_code, version, created_at, updated_at FROM rpa_flows WHERE share_code = ?`, strings.TrimSpace(shareCode))
	item, err := scanFlow(row)
	if err != nil {
		return nil, fmt.Errorf("查询分享流程失败: %w", err)
	}
	return item, nil
}

func (d *SQLiteFlowDAO) DeleteFlow(flowID string) error {
	_, err := d.db.Exec(`DELETE FROM rpa_flows WHERE flow_id = ?`, strings.TrimSpace(flowID))
	if err != nil {
		return fmt.Errorf("删除流程失败: %w", err)
	}
	return nil
}

func prepareFlowForSave(flow *Flow, now string) {
	if flow.FlowID == "" {
		flow.FlowID = uuid.NewString()
	}
	flow.FlowName = strings.TrimSpace(flow.FlowName)
	flow.GroupID = strings.TrimSpace(flow.GroupID)
	flow.SourceXML = strings.TrimSpace(flow.SourceXML)
	flow.ShareCode = strings.TrimSpace(flow.ShareCode)
	if flow.SourceType == "" {
		flow.SourceType = FlowSourceVisual
	}
	if flow.Version <= 0 {
		flow.Version = 1
	}
	flow.Document = normalizeDocument(flow.Document)
	if len(flow.Document.Nodes) == 0 {
		flow.Document = defaultFlowDocument()
	}
	if flow.Steps == nil {
		flow.Steps = []FlowStep{}
	}
	if flow.CreatedAt == "" {
		flow.CreatedAt = now
	}
	flow.UpdatedAt = now
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanFlowGroup(s rowScanner) (*FlowGroup, error) {
	var item FlowGroup
	if err := s.Scan(&item.GroupID, &item.GroupName, &item.SortOrder, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	return &item, nil
}

func scanFlow(s rowScanner) (*Flow, error) {
	var item Flow
	var stepsJSON string
	var documentJSON string
	if err := s.Scan(&item.FlowID, &item.FlowName, &item.GroupID, &stepsJSON, &documentJSON, &item.SourceType, &item.SourceXML, &item.ShareCode, &item.Version, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	decodeJSON(stepsJSON, &item.Steps)
	decodeJSON(documentJSON, &item.Document)
	if item.Steps == nil {
		item.Steps = []FlowStep{}
	}
	item.Document = normalizeDocument(item.Document)
	return &item, nil
}
