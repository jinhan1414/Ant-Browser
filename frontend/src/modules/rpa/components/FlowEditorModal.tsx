import { useEffect, useMemo, useState } from 'react'
import { Button, FormItem, Input, Modal, Select, toast } from '../../../shared/components'
import { Copy, Download } from 'lucide-react'
import { FLOW_XML_PROMPT_TEMPLATE } from '../aiPrompt'
import { encodeRPAFlowXML, parseRPAFlowXML } from '../api'
import { insertNodeIntoDocument, normalizeFlow } from '../flowDocument'
import type { RPAFlow, RPAFlowGroup } from '../types'
import { FlowCanvas } from './FlowCanvas'
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
  const [flow, setFlow] = useState<RPAFlow>(normalizeFlow(null))
  const [selectedNodeId, setSelectedNodeId] = useState('')
  const [saving, setSaving] = useState(false)
  const [xmlModalOpen, setXMLModalOpen] = useState(false)

  useEffect(() => {
    const next = normalizeFlow(initialFlow)
    setFlow(next)
    setSelectedNodeId(next.document.nodes[0]?.nodeId || '')
  }, [initialFlow, open])

  const groupOptions = [{ value: '', label: '未分组' }, ...groups.map(group => ({ value: group.groupId, label: group.groupName }))]
  const selectedNode = useMemo(() => flow.document.nodes.find(node => node.nodeId === selectedNodeId) || null, [flow.document.nodes, selectedNodeId])

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={flow.flowId ? '编辑流程' : '新建流程'}
      width="1480px"
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>取消</Button>
          <Button
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
        </>
      }
    >
      <div className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormItem label="流程名称" required>
            <Input value={flow.flowName} onChange={e => setFlow(prev => ({ ...prev, flowName: e.target.value }))} />
          </FormItem>
          <FormItem label="所属分组">
            <Select value={flow.groupId} options={groupOptions} onChange={e => setFlow(prev => ({ ...prev, groupId: e.target.value }))} />
          </FormItem>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button size="sm" variant="secondary" onClick={async () => {
            await navigator.clipboard.writeText(FLOW_XML_PROMPT_TEMPLATE)
            toast.success('AI 提示词已复制')
          }}>
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
        <div className="grid grid-cols-1 xl:grid-cols-[220px_minmax(0,1fr)_280px] gap-4">
          <FlowNodePalette
            onAddNode={(nodeType) => {
              const result = insertNodeIntoDocument(flow.document, nodeType)
              setFlow(prev => ({ ...prev, sourceType: 'visual', document: result.document }))
              setSelectedNodeId(result.nodeId)
            }}
          />
          <FlowCanvas
            document={flow.document}
            selectedNodeId={selectedNodeId}
            onSelectNode={setSelectedNodeId}
            onChange={document => setFlow(prev => ({ ...prev, sourceType: 'visual', document }))}
          />
          <FlowNodeInspector
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
            onDelete={(nodeId) => {
              setFlow(prev => ({
                ...prev,
                sourceType: 'visual',
                document: {
                  ...prev.document,
                  nodes: prev.document.nodes.filter(item => item.nodeId !== nodeId),
                  edges: prev.document.edges.filter(edge => edge.sourceNodeId !== nodeId && edge.targetNodeId !== nodeId),
                },
              }))
              setSelectedNodeId('start_1')
            }}
          />
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
          setSelectedNodeId(next.document.nodes[0]?.nodeId || '')
          toast.success('XML 已回填到当前编辑窗口')
        }}
      />
    </Modal>
  )
}
