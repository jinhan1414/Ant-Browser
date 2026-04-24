import type { DragEvent as ReactDragEvent } from 'react'
import type { RPAFlowNodeType } from '../types'

export const FLOW_NODE_DND_MIME = 'application/x-ant-browser-flow-node'

export function writeDraggedNodeType(event: ReactDragEvent<HTMLElement>, nodeType: RPAFlowNodeType) {
  event.dataTransfer.setData(FLOW_NODE_DND_MIME, nodeType)
  event.dataTransfer.setData('text/plain', nodeType)
  event.dataTransfer.effectAllowed = 'copy'
}

export function hasDraggedNodeType(event: ReactDragEvent<HTMLElement>) {
  return Array.from(event.dataTransfer.types).includes(FLOW_NODE_DND_MIME)
}

export function readDraggedNodeType(event: ReactDragEvent<HTMLElement>) {
  const nodeType = event.dataTransfer.getData(FLOW_NODE_DND_MIME) || event.dataTransfer.getData('text/plain')
  return nodeType ? nodeType as RPAFlowNodeType : ''
}
