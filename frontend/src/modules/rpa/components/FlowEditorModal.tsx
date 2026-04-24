import { useEffect, useMemo, useState } from 'react'
import { Button, FormItem, Input, Modal, Select, toast } from '../../../shared/components'
import { Copy, Download, PanelLeftClose, PanelLeftOpen, PanelRightClose, Settings2 } from 'lucide-react'
import { encodeRPAFlowXML, parseRPAFlowXML } from '../api'
import { normalizeFlow } from '../flowDocument'
import { removeEdgeFromDocument, updateEdgeInDocument } from '../flowEdge'
import { useFlowNodeCatalog } from '../nodeCatalog'
import type { RPAFlow, RPAFlowGroup } from '../types'
import { FlowCanvas } from './FlowCanvas'
import { FlowEdgeInspector } from './FlowEdgeInspector'
import { FlowNodeInspector } from './FlowNodeInspector'
import { FlowNodePalette } from './FlowNodePalette'
import { FlowXMLModal } from './FlowXMLModal'

interface FlowEditorModalProps {
  open: boolean
  groups: RPAFlowGroup[]
  initialFlow: RPAFlow | null
  onClose: () => void
  onSubmit: (flow: RPAFlow) => Promise<void>
}

export function FlowEditorModal({ open, groups, initialFlow, onClose, onSubmit }: FlowEditorModalProps) {
  const { items: catalog, xmlPromptTemplate } = useFlowNodeCatalog()
  const [flow, setFlow] = useState<RPAFlow>(normalizeFlow(null))
  const [selectedNodeId, setSelectedNodeId] = useState('')
  const [selectedEdgeId, setSelectedEdgeId] = useState('')
  const [saving, setSaving] = useState(false)
  const [xmlModalOpen, setXMLModalOpen] = useState(false)
  const [showPalette, setShowPalette] = useState(false)
  const [showInspector, setShowInspector] = useState(true)
  const [showUtilities, setShowUtilities] = useState(false)

  useEffect(() => {
    const next = normalizeFlow(initialFlow)
    setFlow(next)
    setSelectedNodeId('')
    setSelectedEdgeId('')
    setShowInspector(false)
    setShowUtilities(false)
  }, [initialFlow, open])

  const groupOptions = [{ value: '', label: '未分组' }, ...groups.map(group => ({ value: group.groupId, label: group.groupName }))]
  const selectedNode = useMemo(() => flow.document.nodes.find(node => node.nodeId === selectedNodeId) || null, [flow.document.nodes, selectedNodeId])
  const selectedEdge = useMemo(() => flow.document.edges.find(edge => edge.edgeId === selectedEdgeId) || null, [flow.document.edges, selectedEdgeId])

  useEffect(() => {
    if (selectedNodeId || selectedEdgeId) {
      setShowInspector(true)
    }
  }, [selectedEdgeId, selectedNodeId])

  const handleDeleteNode = (nodeId: string) => {
    setFlow(prev => ({
      ...prev,
      sourceType: 'visual',
      document: {
        ...prev.document,
        nodes: prev.document.nodes.filter(item => item.nodeId !== nodeId),
        edges: prev.document.edges.filter(edge => edge.sourceNodeId !== nodeId && edge.targetNodeId !== nodeId),
      },
    }))
    setSelectedNodeId('')
    setSelectedEdgeId('')
    setShowInspector(false)
  }

  const handleDeleteEdge = (edgeId: string) => {
    setFlow(prev => ({
      ...prev,
      sourceType: 'visual',
      document: removeEdgeFromDocument(prev.document, edgeId),
    }))
    setSelectedEdgeId('')
    setShowInspector(false)
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={flow.flowId ? '编辑流程' : '新建流程'}
      width="90vw"
    >
      <div className="flex h-[calc(90vh-180px)] min-h-[760px] flex-col gap-3">
        <div className="relative flex flex-wrap items-center gap-2 rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-secondary)]/50 px-3 py-3">
          <Button
            size="sm"
            variant="secondary"
            onClick={() => setShowPalette(prev => !prev)}
          >
            {showPalette ? <PanelLeftClose className="w-4 h-4" /> : <PanelLeftOpen className="w-4 h-4" />}
            节点面板
          </Button>
          <div className="min-w-[220px] flex-1 max-w-[420px]">
            <Input value={flow.flowName} onChange={e => setFlow(prev => ({ ...prev, flowName: e.target.value }))} placeholder="请输入流程名称" />
          </div>
          <div className="hidden xl:block text-xs text-[var(--color-text-muted)]">
            画布优先模式：拖拽节点到画布，选中节点后再编辑属性。
          </div>
          <div className="ml-auto flex items-center gap-2">
            <Button size="sm" variant="secondary" onClick={() => setShowUtilities(prev => !prev)}>
              <Settings2 className="w-4 h-4" />
              更多
            </Button>
            <Button size="sm" variant="secondary" onClick={onClose}>关闭</Button>
            <Button
              size="sm"
              loading={saving}
              onClick={async () => {
                setSaving(true)
                try {
                  await onSubmit(flow)
                } finally {
                  setSaving(false)
                }
              }}
            >
              保存流程
            </Button>
          </div>
          {showUtilities ? (
            <div className="absolute right-0 top-full z-30 mt-2 w-[320px] rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-elevated)] p-4 shadow-xl">
              <div className="space-y-3">
                <FormItem label="所属分组">
                  <Select value={flow.groupId} options={groupOptions} onChange={e => setFlow(prev => ({ ...prev, groupId: e.target.value }))} />
                </FormItem>
                <div className="grid grid-cols-1 gap-2">
                  <Button size="sm" variant="secondary" onClick={async () => {
                    await navigator.clipboard.writeText(xmlPromptTemplate)
                    toast.success('AI 提示词已复制')
                  }} disabled={!xmlPromptTemplate}>
                    <Copy className="w-4 h-4" />复制 AI 提示词
                  </Button>
                  <Button size="sm" variant="secondary" onClick={() => setXMLModalOpen(true)}>
                    <Download className="w-4 h-4" />导入 XML
                  </Button>
                  <Button size="sm" variant="secondary" onClick={async () => {
                    try {
                      const xmlText = await encodeRPAFlowXML(flow)
                      if (xmlText) {
                        await navigator.clipboard.writeText(xmlText)
                        toast.success('当前流程 XML 已复制')
                      }
                    } catch (error: any) {
                      toast.error(error?.message || 'XML 导出失败')
                    }
                  }}>
                    <Copy className="w-4 h-4" />复制当前 XML
                  </Button>
                </div>
              </div>
            </div>
          ) : null}
        </div>
        <div className="relative flex-1 min-h-0">
          <FlowCanvas
            catalog={catalog}
            document={flow.document}
            selectedNodeId={selectedNodeId}
            selectedEdgeId={selectedEdgeId}
            onDeleteNode={handleDeleteNode}
            onSelectNode={setSelectedNodeId}
            onSelectEdge={setSelectedEdgeId}
            onChange={document => setFlow(prev => ({ ...prev, sourceType: 'visual', document }))}
          />
          {showPalette ? (
            <div className="absolute left-4 top-4 bottom-4 z-20 w-[240px] overflow-y-auto rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-elevated)]/95 p-3 shadow-xl backdrop-blur">
              <div className="mb-3 flex items-center justify-between">
                <div className="text-sm font-semibold text-[var(--color-text-primary)]">节点</div>
                <Button size="sm" variant="ghost" onClick={() => setShowPalette(false)}>
                  <PanelLeftClose className="w-4 h-4" />
                </Button>
              </div>
              <FlowNodePalette items={catalog} />
            </div>
          ) : (
            <Button
              size="sm"
              variant="secondary"
              className="absolute left-4 top-4 z-20 shadow-lg"
              onClick={() => setShowPalette(true)}
            >
              <PanelLeftOpen className="w-4 h-4" />
              节点
            </Button>
          )}
          {(selectedNode || selectedEdge) && showInspector ? (
            <div className="absolute right-4 top-4 bottom-4 z-20 w-[320px] overflow-y-auto rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-elevated)]/95 p-3 shadow-xl backdrop-blur">
              <div className="mb-3 flex items-center justify-between">
                <div className="text-sm font-semibold text-[var(--color-text-primary)]">{selectedEdge ? '连线属性' : '节点属性'}</div>
                <Button size="sm" variant="ghost" onClick={() => setShowInspector(false)}>
                  <PanelRightClose className="w-4 h-4" />
                </Button>
              </div>
              {selectedEdge ? (
                <FlowEdgeInspector
                  catalog={catalog}
                  document={flow.document}
                  edge={selectedEdge}
                  onChange={edge => {
                    setFlow(prev => ({
                      ...prev,
                      sourceType: 'visual',
                      document: updateEdgeInDocument(prev.document, edge),
                    }))
                  }}
                  onDelete={handleDeleteEdge}
                />
              ) : (
                <FlowNodeInspector
                  catalog={catalog}
                  node={selectedNode}
                  onChange={node => {
                    setFlow(prev => ({
                      ...prev,
                      sourceType: 'visual',
                      document: {
                        ...prev.document,
                        nodes: prev.document.nodes.map(item => item.nodeId === node.nodeId ? node : item),
                      },
                    }))
                  }}
                  onDelete={handleDeleteNode}
                />
              )}
            </div>
          ) : (selectedNode || selectedEdge) ? (
            <Button
              size="sm"
              variant="secondary"
              className="absolute right-4 top-4 z-20 shadow-lg"
              onClick={() => setShowInspector(true)}
            >
              属性
            </Button>
          ) : null}
        </div>
      </div>
      <FlowXMLModal
        open={xmlModalOpen}
        groups={groups}
        defaultGroupId={flow.groupId}
        submitText="解析并回填"
        submitXML={parseRPAFlowXML}
        onClose={() => setXMLModalOpen(false)}
        onImported={async (draft) => {
          const next = normalizeFlow({
            ...flow,
            flowName: draft.flowName || flow.flowName,
            groupId: draft.groupId || flow.groupId,
            document: draft.document,
            sourceType: 'xml_import',
            sourceXml: draft.sourceXml,
          })
          setFlow(next)
          setSelectedEdgeId('')
          setSelectedNodeId(next.document.nodes[0]?.nodeId || '')
          toast.success('XML 已回填到当前编辑窗口')
        }}
      />
    </Modal>
  )
}
