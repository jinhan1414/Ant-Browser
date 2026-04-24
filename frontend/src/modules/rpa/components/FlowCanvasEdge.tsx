import { BaseEdge, EdgeLabelRenderer, getBezierPath, type EdgeProps } from 'reactflow'
import type { RPAFlowEdgeBranchType } from '../types'

type FlowCanvasEdgeData = {
  onDelete?: (edgeId: string) => void
  branchType?: RPAFlowEdgeBranchType
  labelText?: string
  displayLabel?: string
}

export function FlowCanvasEdge({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  selected,
  markerEnd,
  label,
  data,
}: EdgeProps<FlowCanvasEdgeData>) {
  const [edgePath, labelX, labelY] = getBezierPath({
    sourceX,
    sourceY,
    targetX,
    targetY,
    sourcePosition,
    targetPosition,
  })

  return (
    <>
      <BaseEdge
        path={edgePath}
        markerEnd={markerEnd}
        interactionWidth={24}
        style={{
          stroke: selected ? 'var(--color-accent)' : 'var(--color-border-strong)',
          strokeWidth: selected ? 2.5 : 1.5,
        }}
      />
      <EdgeLabelRenderer>
        <>
          {label ? (
            <div
              className="absolute rounded-md border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-2 py-1 text-[11px] text-[var(--color-text-secondary)] shadow-sm"
              style={{
                transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY - 18}px)`,
                pointerEvents: 'none',
              }}
            >
              {String(label)}
            </div>
          ) : null}
          {selected ? (
            <button
              type="button"
              className="absolute flex h-6 w-6 items-center justify-center rounded-full border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] text-xs text-[var(--color-error)] shadow-sm"
              style={{
                transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY}px)`,
                pointerEvents: 'all',
              }}
              onClick={(event) => {
                event.preventDefault()
                event.stopPropagation()
                data?.onDelete?.(id)
              }}
            >
              x
            </button>
          ) : null}
        </>
      </EdgeLabelRenderer>
    </>
  )
}
