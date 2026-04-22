import { useEffect, useState } from 'react'
import { CopyPlus, Plus, RefreshCw, Trash2 } from 'lucide-react'
import { Button, Card, Table, toast } from '../../../shared/components'
import type { TableColumn } from '../../../shared/components'
import { createFlowFromTemplate, deleteRPATemplate, fetchRPAFlows, fetchRPATemplates, saveRPATemplate } from '../api'
import { TemplateEditorModal } from '../components/TemplateEditorModal'
import type { RPAFlow, RPATemplate } from '../types'

const formatTime = (value: string) => value ? new Date(value).toLocaleString('zh-CN') : '-'

export function TemplateCenterPage() {
  const [templates, setTemplates] = useState<RPATemplate[]>([])
  const [flows, setFlows] = useState<RPAFlow[]>([])
  const [editorOpen, setEditorOpen] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<RPATemplate | null>(null)

  const loadData = async () => {
    const [templateItems, flowItems] = await Promise.all([
      fetchRPATemplates(),
      fetchRPAFlows(),
    ])
    setTemplates(templateItems)
    setFlows(flowItems)
  }

  useEffect(() => {
    void loadData()
  }, [])

  const columns: TableColumn<RPATemplate>[] = [
    { key: 'templateName', title: '模板名称' },
    { key: 'description', title: '描述' },
    { key: 'tags', title: '标签', render: value => Array.isArray(value) ? value.join(' / ') : '' },
    { key: 'updatedAt', title: '更新时间', render: value => formatTime(value) },
    {
      key: 'actions',
      title: '操作',
      align: 'right',
      render: (_, record) => (
        <div className="flex justify-end gap-1">
          <Button size="sm" variant="ghost" onClick={() => { setEditingTemplate(record); setEditorOpen(true) }}>编辑</Button>
          <Button size="sm" variant="ghost" onClick={async () => {
            const flow = await createFlowFromTemplate(record.templateId)
            if (flow) {
              toast.success(`已从模板创建流程：${flow.flowName}`)
            }
          }}><CopyPlus className="w-3.5 h-3.5" />创建流程</Button>
          <Button size="sm" variant="ghost" onClick={async () => {
            await deleteRPATemplate(record.templateId)
            toast.success('模板已删除')
            await loadData()
          }}><Trash2 className="w-3.5 h-3.5" /></Button>
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-5 animate-fade-in">
      <Card
        title="模板中心"
        subtitle="沉淀可复用的流程模板，并快速从模板生成新流程"
        actions={
          <>
            <Button size="sm" variant="secondary" onClick={() => void loadData()}><RefreshCw className="w-4 h-4" />刷新</Button>
            <Button size="sm" onClick={() => { setEditingTemplate(null); setEditorOpen(true) }}><Plus className="w-4 h-4" />新建模板</Button>
          </>
        }
      >
        <div className="text-sm text-[var(--color-text-secondary)]">
          模板当前基于已有流程快照保存，适合沉淀常用执行套路。
        </div>
      </Card>

      <Card padding="none">
        <Table columns={columns} data={templates} rowKey="templateId" emptyText="暂无模板" />
      </Card>

      <TemplateEditorModal
        open={editorOpen}
        flows={flows}
        initialTemplate={editingTemplate}
        onClose={() => setEditorOpen(false)}
        onSubmit={async (template) => {
          const saved = await saveRPATemplate(template)
          if (saved) {
            toast.success('模板已保存')
            setEditorOpen(false)
            await loadData()
          }
        }}
      />
    </div>
  )
}
