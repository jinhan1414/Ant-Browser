import { useEffect, useMemo, useState } from 'react'
import { Copy, Download, FolderPlus, Plus, RefreshCw, Trash2 } from 'lucide-react'
import { Button, Card, FormItem, Input, Select, Table, toast } from '../../../shared/components'
import type { TableColumn } from '../../../shared/components'
import { createRPAFlowGroup, deleteRPAFlow, encodeRPAFlowXML, fetchRPAFlowGroups, fetchRPAFlows, importRPAFlowByShareCode, saveRPAFlow, shareRPAFlow } from '../api'
import { FLOW_XML_PROMPT_TEMPLATE } from '../aiPrompt'
import { countRunnableNodes } from '../flowDocument'
import { FlowEditorModal } from '../components/FlowEditorModal'
import { FlowXMLModal } from '../components/FlowXMLModal'
import type { RPAFlow, RPAFlowGroup } from '../types'

const formatTime = (value: string) => value ? new Date(value).toLocaleString('zh-CN') : '-'

export function FlowManagementPage() {
  const [flows, setFlows] = useState<RPAFlow[]>([])
  const [groups, setGroups] = useState<RPAFlowGroup[]>([])
  const [keyword, setKeyword] = useState('')
  const [groupId, setGroupId] = useState('')
  const [editorOpen, setEditorOpen] = useState(false)
  const [xmlModalOpen, setXMLModalOpen] = useState(false)
  const [editingFlow, setEditingFlow] = useState<RPAFlow | null>(null)

  const loadData = async () => {
    const [groupItems, flowItems] = await Promise.all([
      fetchRPAFlowGroups(),
      fetchRPAFlows(keyword, groupId),
    ])
    setGroups(groupItems)
    setFlows(flowItems)
  }

  useEffect(() => {
    void loadData()
  }, [keyword, groupId])

  const groupMap = useMemo(() => new Map(groups.map(group => [group.groupId, group.groupName])), [groups])
  const groupOptions = [{ value: '', label: '全部分组' }, ...groups.map(group => ({ value: group.groupId, label: group.groupName }))]

  const columns: TableColumn<RPAFlow>[] = [
    { key: 'flowName', title: '流程名称' },
    { key: 'groupId', title: '分组', render: value => groupMap.get(value) || '未分组' },
    { key: 'document', title: '节点数', align: 'center', render: (_, record) => countRunnableNodes(record) },
    { key: 'updatedAt', title: '更新时间', render: value => formatTime(value) },
    {
      key: 'actions',
      title: '操作',
      align: 'right',
      render: (_, record) => (
        <div className="flex justify-end gap-1">
          <Button size="sm" variant="ghost" onClick={() => { setEditingFlow(record); setEditorOpen(true) }}>编辑</Button>
          <Button size="sm" variant="ghost" onClick={async () => {
            try {
              const xmlText = await encodeRPAFlowXML(record)
              if (xmlText) {
                await navigator.clipboard.writeText(xmlText)
                toast.success('流程 XML 已复制')
              }
            } catch (error: any) {
              toast.error(error?.message || 'XML 导出失败')
            }
          }}>
            <Download className="w-3.5 h-3.5" />XML
          </Button>
          <Button size="sm" variant="ghost" onClick={async () => {
            const code = await shareRPAFlow(record.flowId)
            if (code) {
              await navigator.clipboard.writeText(code)
              toast.success(`分享码已复制：${code}`)
            }
          }}>
            <Copy className="w-3.5 h-3.5" />分享
          </Button>
          <Button size="sm" variant="ghost" onClick={async () => {
            await deleteRPAFlow(record.flowId)
            toast.success('流程已删除')
            await loadData()
          }}>
            <Trash2 className="w-3.5 h-3.5" />
          </Button>
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-5 animate-fade-in">
      <Card
        title="流程管理"
        subtitle="管理 RPA 流程、分组、分享码导入和最小步骤集合"
        actions={
          <>
            <Button size="sm" variant="secondary" onClick={() => void loadData()}><RefreshCw className="w-4 h-4" />刷新</Button>
            <Button size="sm" variant="secondary" onClick={async () => {
              const name = prompt('请输入流程分组名称')
              if (!name?.trim()) return
              await createRPAFlowGroup(name.trim())
              toast.success('分组已创建')
              await loadData()
            }}><FolderPlus className="w-4 h-4" />新建分组</Button>
            <Button size="sm" variant="secondary" onClick={async () => {
              const shareCode = prompt('请输入分享码')
              if (!shareCode?.trim()) return
              const flow = await importRPAFlowByShareCode(shareCode.trim())
              if (flow) {
                toast.success(`流程已导入：${flow.flowName}`)
                await loadData()
              }
            }}><Download className="w-4 h-4" />导入分享码</Button>
            <Button size="sm" variant="secondary" onClick={async () => {
              await navigator.clipboard.writeText(FLOW_XML_PROMPT_TEMPLATE)
              toast.success('AI 提示词已复制')
            }}><Copy className="w-4 h-4" />复制 AI 提示词</Button>
            <Button size="sm" variant="secondary" onClick={() => setXMLModalOpen(true)}><Download className="w-4 h-4" />导入 XML</Button>
            <Button size="sm" onClick={() => { setEditingFlow(null); setEditorOpen(true) }}><Plus className="w-4 h-4" />新建流程</Button>
          </>
        }
      >
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormItem label="搜索流程">
            <Input value={keyword} onChange={e => setKeyword(e.target.value)} placeholder="按流程名称过滤" />
          </FormItem>
          <FormItem label="流程分组">
            <Select value={groupId} options={groupOptions} onChange={e => setGroupId(e.target.value)} />
          </FormItem>
        </div>
      </Card>

      <Card padding="none">
        <Table columns={columns} data={flows} rowKey="flowId" emptyText="暂无流程" />
      </Card>

      <FlowEditorModal
        open={editorOpen}
        groups={groups}
        initialFlow={editingFlow}
        onClose={() => setEditorOpen(false)}
        onSubmit={async (flow) => {
          try {
            const saved = await saveRPAFlow(flow)
            if (saved) {
              toast.success('流程已保存')
              setEditorOpen(false)
              await loadData()
            }
          } catch (error: any) {
            toast.error(error?.message || '流程保存失败')
          }
        }}
      />

      <FlowXMLModal
        open={xmlModalOpen}
        groups={groups}
        defaultGroupId={groupId}
        onClose={() => setXMLModalOpen(false)}
        onImported={async (flow) => {
          toast.success(`XML 流程已导入：${flow.flowName}`)
          await loadData()
        }}
      />
    </div>
  )
}
