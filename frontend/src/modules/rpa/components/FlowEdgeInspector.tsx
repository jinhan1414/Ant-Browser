import { Button, FormItem, Input, Select } from '../../../shared/components'
import { getBranchTypeOptions, getBranchTypeOptionLabel, normalizeDocumentEdge } from '../flowEdge'
import { findNodeCatalogItem } from '../nodeCatalog'
import type { RPAFlowDocument, RPAFlowEdge, RPAFlowEdgeBranchType, RPAFlowNodeCatalogItem } from '../types'

interface FlowEdgeInspectorProps {
  catalog: RPAFlowNodeCatalogItem[]
  document: RPAFlowDocument
  edge: RPAFlowEdge | null
  onChange: (edge: RPAFlowEdge) => void
  onDelete: (edgeId: string) => void
}

export function FlowEdgeInspector({ catalog, document, edge, onChange, onDelete }: FlowEdgeInspectorProps) {
  if (!edge) {
    return (
      <div className="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-secondary)]/40 p-4 text-sm text-[var(--color-text-muted)]">
        选择一条连线后，在这里编辑分支类型和名称。
      </div>
    )
  }

  const sourceNode = document.nodes.find(node => node.nodeId === edge.sourceNodeId) || null
  const targetNode = document.nodes.find(node => node.nodeId === edge.targetNodeId) || null
  const sourceCatalog = sourceNode ? findNodeCatalogItem(catalog, sourceNode.nodeType) : null
  const siblingEdges = document.edges.filter(item => item.sourceNodeId === edge.sourceNodeId)
  const branchOptions = getBranchTypeOptions(catalog, sourceNode?.nodeType || '', siblingEdges, edge.edgeId)

  return (
    <div className="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-secondary)]/40 p-4 space-y-4">
      <div>
        <div className="text-sm font-semibold text-[var(--color-text-primary)]">连线属性</div>
        <div className="mt-1 text-xs text-[var(--color-text-muted)]">连线 ID：{edge.edgeId}</div>
      </div>
      <FormItem label="起点节点">
        <Input value={sourceNode?.label || edge.sourceNodeId} disabled />
      </FormItem>
      <FormItem label="终点节点">
        <Input value={targetNode?.label || edge.targetNodeId} disabled />
      </FormItem>
      <FormItem
        label="分支类型"
        hint={sourceCatalog?.supportsIfBranch ? '条件节点必须分别配置 TRUE 和 FALSE。' : sourceCatalog?.supportsOnError ? '普通节点可配置默认分支或异常分支。' : '当前节点只支持默认分支。'}
      >
        <Select
          value={edge.branchType}
          options={branchOptions.length > 0 ? branchOptions : [{ value: edge.branchType, label: getBranchTypeOptionLabel(edge.branchType) }]}
          onChange={e => onChange(normalizeDocumentEdge({ ...edge, branchType: e.target.value as RPAFlowEdgeBranchType }))}
        />
      </FormItem>
      <FormItem label="连线名称" hint="用于画布显示，可为空">
        <Input
          value={edge.label}
          placeholder="例如：命中异常"
          onChange={e => onChange(normalizeDocumentEdge({ ...edge, label: e.target.value }))}
        />
      </FormItem>
      <Button variant="danger" size="sm" onClick={() => onDelete(edge.edgeId)}>
        删除连线
      </Button>
    </div>
  )
}
