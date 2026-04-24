import type { ChangeEvent } from 'react'
import { Button, FormItem, Input, Select, Textarea } from '../../../shared/components'
import { buildNodeConfig, findNodeCatalogItem, getNodeLabel, getNodeTypeOptions, readNodeFieldValue, writeNodeFieldValue } from '../nodeCatalog'
import type { RPAFlowNode, RPAFlowNodeCatalogItem, RPAFlowNodeField, RPAFlowNodeType } from '../types'

interface FlowNodeInspectorProps {
  catalog: RPAFlowNodeCatalogItem[]
  node: RPAFlowNode | null
  onChange: (node: RPAFlowNode) => void
  onDelete: (nodeId: string) => void
}

function shouldReplaceNodeLabel(node: RPAFlowNode, catalog: RPAFlowNodeCatalogItem[]) {
  const current = findNodeCatalogItem(catalog, node.nodeType)
  return !node.label || node.label === current?.label || node.label === node.nodeType
}

function renderField(
  node: RPAFlowNode,
  field: RPAFlowNodeField,
  onChange: (nextNode: RPAFlowNode) => void,
) {
  const commonProps = {
    value: readNodeFieldValue(node.config || {}, field),
    placeholder: field.placeholder || undefined,
    onChange: (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => onChange({
      ...node,
      config: writeNodeFieldValue(node.config || {}, field, e.target.value),
    }),
  }

  return (
    <FormItem key={field.key} label={field.label} required={field.required} hint={field.hint || undefined}>
      {field.multiline ? (
        <Textarea rows={3} {...commonProps} />
      ) : (
        <Input type={field.kind === 'number' ? 'number' : 'text'} min={field.kind === 'number' ? field.minValue : undefined} {...commonProps} />
      )}
    </FormItem>
  )
}

export function FlowNodeInspector({ catalog, node, onChange, onDelete }: FlowNodeInspectorProps) {
  if (!node) {
    return (
      <div className="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-secondary)]/40 p-4 text-sm text-[var(--color-text-muted)]">
        选择一个节点后，在这里编辑标签和配置。
      </div>
    )
  }

  const item = findNodeCatalogItem(catalog, node.nodeType)
  const isFixedNode = item?.fixed ?? (node.nodeType === 'start' || node.nodeType === 'end')

  return (
    <div className="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-secondary)]/40 p-4 space-y-4">
      <div>
        <div className="text-sm font-semibold text-[var(--color-text-primary)]">节点属性</div>
        <div className="mt-1 text-xs text-[var(--color-text-muted)]">节点 ID：{node.nodeId}</div>
      </div>
      <FormItem label="节点名称">
        <Input value={node.label} onChange={e => onChange({ ...node, label: e.target.value })} />
      </FormItem>
      <FormItem label="节点类型">
        <Select
          value={node.nodeType}
          disabled={isFixedNode}
          options={getNodeTypeOptions(catalog)}
          onChange={e => {
            const nextType = e.target.value as RPAFlowNodeType
            onChange({
              ...node,
              nodeType: nextType,
              label: shouldReplaceNodeLabel(node, catalog) ? getNodeLabel(catalog, nextType) : node.label,
              config: buildNodeConfig(catalog, nextType, node.config || {}),
            })
          }}
        />
      </FormItem>
      {(item?.fields || []).map(field => renderField(node, field, onChange))}
      {!isFixedNode && (
        <Button variant="danger" size="sm" onClick={() => onDelete(node.nodeId)}>
          删除节点
        </Button>
      )}
    </div>
  )
}
