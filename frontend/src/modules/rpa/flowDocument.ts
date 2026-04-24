import type { Edge, Node } from 'reactflow'
import { buildCanvasEdgeData, edgeFromCanvas, normalizeDocumentEdge } from './flowEdge'
import type { RPAFlow, RPAFlowDocument, RPAFlowNode, RPAFlowNodeType } from './types'

const DEFAULT_NODE_Y = 180
const DEFAULT_START_X = 120
const DEFAULT_END_X = 520

export function createDefaultDocument(): RPAFlowDocument {
  return {
    schemaVersion: 3,
    variables: [],
    nodes: [
      createDocumentNode('start_1', 'start', '开始', DEFAULT_START_X, DEFAULT_NODE_Y),
      createDocumentNode('end_1', 'end', '结束', DEFAULT_END_X, DEFAULT_NODE_Y),
    ],
    edges: [
      normalizeDocumentEdge({
        edgeId: 'edge_start_end',
        sourceNodeId: 'start_1',
        targetNodeId: 'end_1',
      }),
    ],
  }
}

export function normalizeFlow(flow: RPAFlow | null): RPAFlow {
  if (flow) {
    return {
      ...flow,
      document: normalizeDocument(flow.document),
      steps: Array.isArray(flow.steps) ? flow.steps : [],
      sourceType: flow.sourceType || 'visual',
      sourceXml: flow.sourceXml || '',
    }
  }
  return {
    flowId: '',
    flowName: '',
    groupId: '',
    steps: [],
    document: createDefaultDocument(),
    sourceType: 'visual',
    sourceXml: '',
    shareCode: '',
    version: 1,
    createdAt: '',
    updatedAt: '',
  }
}

export function normalizeDocument(document?: Partial<RPAFlowDocument> | null): RPAFlowDocument {
  const fallback = createDefaultDocument()
  const nodes = Array.isArray(document?.nodes) ? document.nodes : []
  const edges = Array.isArray(document?.edges) ? document.edges : []
  const hasNodes = nodes.length > 0
  const hasEdges = edges.length > 0
  return {
    schemaVersion: document?.schemaVersion || 3,
    variables: Array.isArray(document?.variables) ? document!.variables : [],
    nodes: hasNodes
      ? nodes.map(node => ({
        ...node,
        config: node.config || {},
        position: node.position || { x: DEFAULT_START_X, y: DEFAULT_NODE_Y },
      }))
      : fallback.nodes,
    edges: hasEdges
      ? edges.map(edge => normalizeDocumentEdge(edge))
      : hasNodes
        ? []
        : fallback.edges,
  }
}

export function documentToCanvas(document: RPAFlowDocument): { nodes: Node[]; edges: Edge[] } {
  return {
    nodes: document.nodes.map(node => ({
      id: node.nodeId,
      type: 'default',
      position: node.position,
      data: {
        label: node.label,
        nodeType: node.nodeType,
      },
      draggable: true,
    })),
    edges: document.edges.map(edge => ({
      id: edge.edgeId,
      source: edge.sourceNodeId,
      target: edge.targetNodeId,
      label: buildCanvasEdgeData(edge).displayLabel || undefined,
      animated: false,
      data: buildCanvasEdgeData(edge),
    })),
  }
}

export function canvasToDocument(nodes: Node[], edges: Edge[], existing: RPAFlowDocument): RPAFlowDocument {
  const nodeMap = new Map(existing.nodes.map(node => [node.nodeId, node]))
  const edgeMap = new Map(existing.edges.map(edge => [edge.edgeId, edge]))
  return {
    schemaVersion: 3,
    variables: existing.variables || [],
    nodes: nodes.map(node => {
      const original = nodeMap.get(node.id)
      return {
        nodeId: node.id,
        nodeType: (node.data?.nodeType || original?.nodeType || 'delay') as RPAFlowNodeType,
        label: String(node.data?.label || original?.label || ''),
        position: { x: node.position.x, y: node.position.y },
        config: original?.config || {},
      }
    }),
    edges: edges.map(edge => {
      const original = edgeMap.get(edge.id)
      return edgeFromCanvas(edge, original)
    }),
  }
}

export function createDocumentNode(
  nodeId: string,
  nodeType: RPAFlowNodeType,
  label: string,
  x: number,
  y: number,
  config: Record<string, any> = {},
): RPAFlowNode {
  return {
    nodeId,
    nodeType,
    label,
    position: { x, y },
    config,
  }
}

export function nextNodePosition(document: RPAFlowDocument): { x: number; y: number } {
  const editableNodes = document.nodes.filter(node => node.nodeType !== 'start' && node.nodeType !== 'end')
  if (editableNodes.length === 0) {
    return { x: 300, y: DEFAULT_NODE_Y }
  }
  const maxX = Math.max(...editableNodes.map(node => node.position.x))
  return { x: maxX + 220, y: DEFAULT_NODE_Y }
}

export function countRunnableNodes(flow: RPAFlow): number {
  return normalizeDocument(flow.document).nodes.filter(node => node.nodeType !== 'start' && node.nodeType !== 'end').length
}

export function insertNodeIntoDocument(
  document: RPAFlowDocument,
  nodeType: RPAFlowNodeType,
  label: string,
  config: Record<string, any>,
  positionOverride?: { x: number; y: number },
): { document: RPAFlowDocument; nodeId: string } {
  const next = normalizeDocument(document)
  const nodeId = `${nodeType.replace('.', '_')}_${Date.now()}`
  const position = positionOverride
    ? {
      x: Math.round(positionOverride.x),
      y: Math.round(positionOverride.y),
    }
    : nextNodePosition(next)
  const newNode = createDocumentNode(nodeId, nodeType, label, position.x, position.y, config)

  return {
    nodeId,
    document: {
      ...next,
      nodes: [...next.nodes, newNode],
    },
  }
}
