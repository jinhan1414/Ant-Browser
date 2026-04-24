import { useEffect, useMemo, useRef, useState, type DragEvent as ReactDragEvent, type MouseEvent as ReactMouseEvent } from 'react'
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  MarkerType,
  addEdge,
  applyEdgeChanges,
  applyNodeChanges,
  reconnectEdge,
  useEdgesState,
  useNodesState,
  type Connection,
  type Edge,
  type EdgeChange,
  type EdgeTypes,
  type Node,
  type NodeChange,
  type NodeTypes,
  type OnEdgeUpdateFunc,
  type ReactFlowInstance,
} from 'reactflow'
import 'reactflow/dist/style.css'
import { canvasToDocument, documentToCanvas, insertNodeIntoDocument, normalizeDocument } from '../flowDocument'
import { buildCanvasEdgeData, createEdgeDraft, getNextBranchType, normalizeDocumentEdge } from '../flowEdge'
import { buildNodeConfig, getNodeConnectionRules, getNodeLabel } from '../nodeCatalog'
import type { RPAFlowDocument, RPAFlowEdge, RPAFlowNodeCatalogItem } from '../types'
import { FlowCanvasEdge } from './FlowCanvasEdge'
import { FlowCanvasNode } from './FlowCanvasNode'
import { hasDraggedNodeType, readDraggedNodeType } from './flowCanvasDnD'

interface FlowCanvasProps {
  catalog: RPAFlowNodeCatalogItem[]
  document: RPAFlowDocument
  selectedNodeId: string
  selectedEdgeId: string
  onChange: (document: RPAFlowDocument) => void
  onDeleteNode: (nodeId: string) => void
  onSelectNode: (nodeId: string) => void
  onSelectEdge: (edgeId: string) => void
}

const nodeTypes: NodeTypes = {
  flowNode: FlowCanvasNode,
}

const edgeTypes: EdgeTypes = {
  editable: FlowCanvasEdge,
}

function shouldPersistNodeChanges(changes: NodeChange[]) {
  return changes.some(change => ['add', 'remove', 'replace', 'position'].includes(change.type))
}

function shouldPersistEdgeChanges(changes: EdgeChange[]) {
  return changes.some(change => ['add', 'remove', 'replace'].includes(change.type))
}

function countOutgoing(edges: Edge[], nodeId: string, excludedEdgeId?: string) {
  return edges.filter(edge => edge.source === nodeId && edge.id !== excludedEdgeId).length
}

function getNodeType(nodeId: string, nodes: Node[]) {
  return nodes.find(node => node.id === nodeId)?.data?.nodeType || ''
}

function isValidCanvasConnection(
  catalog: RPAFlowNodeCatalogItem[],
  connection: Connection,
  nodes: Node[],
  edges: Edge[],
  currentEdgeId?: string,
) {
  const source = connection.source || ''
  const target = connection.target || ''
  const sourceNodeType = getNodeType(source, nodes)
  const targetNodeType = getNodeType(target, nodes)
  const sourceRules = getNodeConnectionRules(catalog, sourceNodeType)
  const targetRules = getNodeConnectionRules(catalog, targetNodeType)
  if (!source || !target || source === target) {
    return false
  }
  if (!sourceRules.allowOutgoing || !targetRules.allowIncoming) {
    return false
  }
  if (edges.some(edge => edge.source === source && edge.target === target && edge.id !== currentEdgeId)) {
    return false
  }
  const outgoingCount = countOutgoing(edges, source, currentEdgeId)
  return outgoingCount < sourceRules.maxOutgoing
}

function decorateEdges(edges: Edge[], onDeleteEdge: (edgeId: string) => void) {
  return edges.map(edge => ({
    ...edge,
    type: 'editable',
    markerEnd: {
      type: MarkerType.ArrowClosed,
      color: 'var(--color-border-strong)',
    },
    data: {
      ...(edge.data || {}),
      onDelete: onDeleteEdge,
    },
  }))
}

function buildNewEdge(
  catalog: RPAFlowNodeCatalogItem[],
  connection: Connection,
  document: RPAFlowDocument,
): RPAFlowEdge {
  const sourceNode = document.nodes.find(node => node.nodeId === connection.source)
  const siblingEdges = document.edges.filter(edge => edge.sourceNodeId === connection.source)
  const branchType = getNextBranchType(catalog, sourceNode?.nodeType || '', siblingEdges)
  return normalizeDocumentEdge({
    edgeId: `edge_${connection.source}_${connection.target}_${Date.now()}`,
    ...createEdgeDraft(connection.source || '', connection.target || '', branchType),
  })
}

export function FlowCanvas({
  catalog,
  document,
  selectedNodeId,
  selectedEdgeId,
  onChange,
  onDeleteNode,
  onSelectNode,
  onSelectEdge,
}: FlowCanvasProps) {
  const normalized = useMemo(() => normalizeDocument(document), [document])
  const canvas = useMemo(() => documentToCanvas(normalized), [normalized])
  const [nodes, setNodes] = useNodesState(canvas.nodes)
  const [edges, setEdges] = useEdgesState(canvas.edges)
  const [reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null)
  const [contextMenu, setContextMenu] = useState<{ nodeId: string; x: number; y: number } | null>(null)
  const wrapperRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    setNodes(canvas.nodes)
    setEdges(canvas.edges)
  }, [canvas, setEdges, setNodes])

  const emit = (nextNodes = nodes, nextEdges = edges) => {
    onChange(canvasToDocument(nextNodes, nextEdges, normalized))
  }

  const handleDeleteEdge = (edgeId: string) => {
    setContextMenu(null)
    const nextEdges = edges.filter(edge => edge.id !== edgeId)
    setEdges(nextEdges)
    onSelectEdge('')
    emit(nodes, nextEdges)
  }

  const handleNodesChange = (changes: NodeChange[]) => {
    const nextNodes = applyNodeChanges(changes, nodes)
    setNodes(nextNodes)
    if (shouldPersistNodeChanges(changes)) {
      emit(nextNodes, edges)
    }
  }

  const handleEdgesChange = (changes: EdgeChange[]) => {
    const nextEdges = applyEdgeChanges(changes, edges)
    setEdges(nextEdges)
    if (shouldPersistEdgeChanges(changes)) {
      emit(nodes, nextEdges)
    }
  }

  const handleConnect = (connection: Connection) => {
    if (!isValidCanvasConnection(catalog, connection, nodes, edges)) {
      return
    }
    const draft = buildNewEdge(catalog, connection, normalized)
    const nextEdges = addEdge({
      source: draft.sourceNodeId,
      target: draft.targetNodeId,
      id: draft.edgeId,
      data: buildCanvasEdgeData(draft),
      label: buildCanvasEdgeData(draft).displayLabel || undefined,
    }, edges)
    setEdges(nextEdges)
    onSelectNode('')
    onSelectEdge(draft.edgeId)
    emit(nodes, nextEdges)
  }

  const handleReconnect: OnEdgeUpdateFunc = (oldEdge, newConnection) => {
    if (!isValidCanvasConnection(catalog, newConnection, nodes, edges, oldEdge.id)) {
      return
    }
    const nextEdges = reconnectEdge(oldEdge, newConnection, edges)
    setEdges(nextEdges)
    emit(nodes, nextEdges)
  }

  const handleDragOver = (event: ReactDragEvent<HTMLDivElement>) => {
    if (!hasDraggedNodeType(event)) {
      return
    }
    event.preventDefault()
    event.dataTransfer.dropEffect = 'copy'
  }

  const handleDrop = (event: ReactDragEvent<HTMLDivElement>) => {
    const nodeType = readDraggedNodeType(event)
    if (!nodeType || !reactFlowInstance) {
      return
    }
    event.preventDefault()
    setContextMenu(null)
    const position = reactFlowInstance.screenToFlowPosition({
      x: event.clientX,
      y: event.clientY,
    })
    const result = insertNodeIntoDocument(
      normalized,
      nodeType,
      getNodeLabel(catalog, nodeType),
      buildNodeConfig(catalog, nodeType),
      position,
    )
    onChange(result.document)
    onSelectEdge('')
    onSelectNode(result.nodeId)
  }

  const handleNodeContextMenu = (event: ReactMouseEvent, node: Node) => {
    event.preventDefault()
    if (!wrapperRef.current) {
      return
    }
    const nodeType = getNodeType(node.id, nodes)
    if (getNodeConnectionRules(catalog, nodeType).fixed) {
      setContextMenu(null)
      return
    }
    const bounds = wrapperRef.current.getBoundingClientRect()
    onSelectEdge('')
    onSelectNode(node.id)
    setContextMenu({
      nodeId: node.id,
      x: event.clientX - bounds.left,
      y: event.clientY - bounds.top,
    })
  }

  const handleDeleteNodeFromMenu = () => {
    if (!contextMenu) {
      return
    }
    setContextMenu(null)
    onDeleteNode(contextMenu.nodeId)
  }

  return (
    <div
      ref={wrapperRef}
      className="relative h-full min-h-[680px] rounded-lg border border-[var(--color-border-default)] overflow-hidden bg-[var(--color-bg-surface)]"
    >
      <ReactFlow
        fitView
        onInit={setReactFlowInstance}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        deleteKeyCode={['Delete', 'Backspace']}
        elementsSelectable
        nodesDraggable
        nodesConnectable
        nodesFocusable
        edgesFocusable
        edgesUpdatable
        reconnectRadius={14}
        nodes={nodes.map(node => ({
          ...node,
          type: 'flowNode',
          selected: node.id === selectedNodeId,
          data: {
            ...node.data,
            label: `${node.data?.label || ''}`,
            typeLabel: getNodeLabel(catalog, String(node.data?.nodeType || '')),
            allowIncoming: getNodeConnectionRules(catalog, String(node.data?.nodeType || '')).allowIncoming,
            allowOutgoing: getNodeConnectionRules(catalog, String(node.data?.nodeType || '')).allowOutgoing,
          },
        }))}
        edges={decorateEdges(edges, handleDeleteEdge).map(edge => ({
          ...edge,
          selected: edge.id === selectedEdgeId,
        }))}
        isValidConnection={(connection) => isValidCanvasConnection(catalog, connection, nodes, edges)}
        onNodesChange={handleNodesChange}
        onEdgesChange={handleEdgesChange}
        onConnect={handleConnect}
        onReconnect={handleReconnect}
        onNodeClick={(_, node) => {
          setContextMenu(null)
          onSelectEdge('')
          onSelectNode(node.id)
        }}
        onNodeContextMenu={handleNodeContextMenu}
        onEdgeClick={(_, edge) => {
          setContextMenu(null)
          onSelectNode('')
          onSelectEdge(edge.id)
        }}
        onPaneClick={() => {
          setContextMenu(null)
          onSelectNode('')
          onSelectEdge('')
        }}
        onPaneContextMenu={() => {
          setContextMenu(null)
          onSelectEdge('')
        }}
        onDragOver={handleDragOver}
        onDrop={handleDrop}
      >
        <MiniMap />
        <Controls />
        <Background gap={16} size={1} />
      </ReactFlow>
      {contextMenu ? (
        <button
          type="button"
          className="absolute z-20 rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-3 py-2 text-sm text-[var(--color-error)] shadow-lg"
          style={{
            left: contextMenu.x,
            top: contextMenu.y,
          }}
          onClick={handleDeleteNodeFromMenu}
        >
          删除节点
        </button>
      ) : null}
    </div>
  )
}
