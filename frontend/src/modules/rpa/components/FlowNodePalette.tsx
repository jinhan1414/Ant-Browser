import { Button } from '../../../shared/components'
import { FLOW_NODE_OPTIONS } from '../flowDocument'
import type { RPAFlowNodeType } from '../types'

interface FlowNodePaletteProps {
  onAddNode: (nodeType: RPAFlowNodeType) => void
}

export function FlowNodePalette({ onAddNode }: FlowNodePaletteProps) {
  return (
    <div className="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-secondary)]/40 p-4 space-y-3">
      <div>
        <div className="text-sm font-semibold text-[var(--color-text-primary)]">节点面板</div>
        <div className="mt-1 text-xs text-[var(--color-text-muted)]">开始和结束节点固定存在，其他节点从这里添加。</div>
      </div>
      <div className="space-y-2">
        {FLOW_NODE_OPTIONS.map(option => (
          <Button
            key={option.value}
            className="w-full justify-start"
            size="sm"
            variant="secondary"
            onClick={() => onAddNode(option.value)}
          >
            {option.label}
          </Button>
        ))}
      </div>
    </div>
  )
}
