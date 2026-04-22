import { useEffect, useMemo } from 'react'
import ReactFlow, { Background, Controls, MiniMap, addEdge, applyEdgeChanges, applyNodeChanges, useEdgesState, useNodesState, type Connection, type EdgeChange, type NodeChange } from 'reactflow'
import 'reactflow/dist/style.css'
import { canvasToDocument, documentToCanvas, normalizeDocument } from '../flowDocument'
import type { RPAFlowDocument } from '../types'

interface FlowCanvasProps {
  document: RPAFlowDocument
  selectedNodeId: string
  onChange: (document: RPAFlowDocument) => void
  onSelectNode: (nodeId: string) => void
}

function shouldPersistNodeChanges(changes: NodeChange[]) {
  return changes.some(change => ['add', 'remove', 'replace', 'position'].includes(change.type))
}

function shouldPersistEdgeChanges(changes: EdgeChange[]) {
  return changes.some(change => ['add', 'remove', 'replace'].includes(change.type))
}

export function FlowCanvas({ document, selectedNodeId, onChange, onSelectNode }: FlowCanvasProps) {
  const normalized = useMemo(() => normalizeDocument(document), [document])
  const canvas = useMemo(() => documentToCanvas(normalized), [normalized])
  const [nodes, setNodes] = useNodesState(canvas.nodes)
  const [edges, setEdges] = useEdgesState(canvas.edges)

  useEffect(() => {
    setNodes(canvas.nodes)
    setEdges(canvas.edges)
  }, [canvas, setEdges, setNodes])

  const emit = (nextNodes = nodes, nextEdges = edges) => {
    onChange(canvasToDocument(nextNodes, nextEdges, normalized))
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
    const nextEdges = addEdge({
      ...connection,
      id: `edge_${connection.source}_${connection.target}_${Date.now()}`,
    }, edges)
    setEdges(nextEdges)
    emit(nodes, nextEdges)
  }

  return (
    <div className="h-[560px] rounded-lg border border-[var(--color-border-default)] overflow-hidden bg-[var(--color-bg-surface)]">
      <ReactFlow
        fitView
        nodes={nodes.map(node => ({
          ...node,
          selected: node.id === selectedNodeId,
          data: {
            ...node.data,
            label: `${node.data?.label || ''}`,
          },
        }))}
        edges={edges}
        onNodesChange={handleNodesChange}
        onEdgesChange={handleEdgesChange}
        onConnect={handleConnect}
        onNodeClick={(_, node) => onSelectNode(node.id)}
      >
        <MiniMap />
        <Controls />
        <Background gap={16} size={1} />
      </ReactFlow>
    </div>
  )
}
