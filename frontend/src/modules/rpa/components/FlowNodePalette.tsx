import type { DragEvent as ReactDragEvent } from 'react'
import { getPaletteNodeItems } from '../nodeCatalog'
import { writeDraggedNodeType } from './flowCanvasDnD'
import type { RPAFlowNodeCatalogItem } from '../types'

const CATEGORY_LABELS: Record<string, string> = {
  browser: '浏览器',
  control: '流程控制',
  system: '系统',
}

function handleDragStart(event: ReactDragEvent<HTMLDivElement>, nodeType: string) {
  writeDraggedNodeType(event, nodeType)
}

interface FlowNodePaletteProps {
  items: RPAFlowNodeCatalogItem[]
}

export function FlowNodePalette({ items }: FlowNodePaletteProps) {
  const groups = getPaletteNodeItems(items).reduce<Record<string, RPAFlowNodeCatalogItem[]>>((result, item) => {
    const key = item.category || 'default'
    result[key] = result[key] || []
    result[key].push(item)
    return result
  }, {})

  return (
    <div className="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-secondary)]/40 p-4 space-y-3">
      <div>
        <div className="text-sm font-semibold text-[var(--color-text-primary)]">节点面板</div>
        <div className="mt-1 text-xs text-[var(--color-text-muted)]">拖拽节点到画布中创建。开始和结束节点固定存在。</div>
      </div>
      {Object.entries(groups).map(([category, categoryItems]) => (
        <div key={category} className="space-y-2">
          <div className="text-xs font-semibold uppercase tracking-[0.08em] text-[var(--color-text-muted)]">
            {CATEGORY_LABELS[category] || category}
          </div>
          {categoryItems.map(item => (
            <div
              key={item.nodeType}
              draggable
              className="cursor-grab rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-3 py-2 text-sm text-[var(--color-text-primary)] shadow-sm transition hover:border-[var(--color-border-strong)] hover:bg-[var(--color-bg-surface-hover)] active:cursor-grabbing"
              onDragStart={(event) => handleDragStart(event, item.nodeType)}
              title={item.description}
            >
              <div>{item.label}</div>
              <div className="mt-1 text-[11px] text-[var(--color-text-muted)]">{item.description}</div>
            </div>
          ))}
        </div>
      ))}
    </div>
  )
}
