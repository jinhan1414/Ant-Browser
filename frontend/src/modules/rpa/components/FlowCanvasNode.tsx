import { Handle, Position, type NodeProps } from 'reactflow'
import type { RPAFlowNodeType } from '../types'

type FlowCanvasNodeData = {
  label: string
  nodeType: RPAFlowNodeType
  typeLabel?: string
  allowIncoming?: boolean
  allowOutgoing?: boolean
}

export function FlowCanvasNode({ data, selected }: NodeProps<FlowCanvasNodeData>) {
  const label = data?.label || data?.typeLabel || '节点'
  const nodeType = data?.nodeType || 'delay'
  const typeLabel = data?.typeLabel || nodeType

  return (
    <div
      className={[
        'min-w-[136px] rounded-xl border bg-[var(--color-bg-surface)] px-4 py-3 shadow-sm transition-all',
        selected
          ? 'border-[var(--color-accent)] ring-2 ring-[var(--color-accent)]/20'
          : 'border-[var(--color-border-default)] hover:border-[var(--color-border-strong)]',
      ].join(' ')}
    >
      {data?.allowIncoming !== false && (
        <Handle
          type="target"
          position={Position.Left}
          className="!h-3 !w-3 !border-2 !border-[var(--color-bg-surface)] !bg-[var(--color-accent)]"
        />
      )}
      <div className="space-y-1">
        <div className="text-sm font-semibold text-[var(--color-text-primary)]">{label}</div>
        <div className="text-[11px] text-[var(--color-text-muted)]">{typeLabel || nodeType}</div>
      </div>
      {data?.allowOutgoing !== false && (
        <Handle
          type="source"
          position={Position.Right}
          className="!h-3 !w-3 !border-2 !border-[var(--color-bg-surface)] !bg-[var(--color-accent)]"
        />
      )}
    </div>
  )
}
