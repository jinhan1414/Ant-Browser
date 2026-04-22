import { useEffect, useMemo, useState } from 'react'
import { Button, FormItem, Input, Modal, Select, Textarea } from '../../../shared/components'
import { normalizeFlow } from '../flowDocument'
import type { RPAFlow, RPATemplate } from '../types'

interface TemplateEditorModalProps {
  open: boolean
  flows: RPAFlow[]
  initialTemplate: RPATemplate | null
  onClose: () => void
  onSubmit: (template: RPATemplate) => Promise<void>
}

const emptyTemplate = (): RPATemplate => ({
  templateId: '',
  templateName: '',
  description: '',
  tags: [],
  flowSnapshot: normalizeFlow(null),
  createdAt: '',
  updatedAt: '',
})

export function TemplateEditorModal({ open, flows, initialTemplate, onClose, onSubmit }: TemplateEditorModalProps) {
  const [template, setTemplate] = useState<RPATemplate>(emptyTemplate())
  const [flowId, setFlowId] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    const current = initialTemplate || emptyTemplate()
    setTemplate(current)
    setFlowId(current.flowSnapshot?.flowId || '')
  }, [initialTemplate, open])

  const flowOptions = useMemo(() => [{ value: '', label: '请选择流程' }, ...flows.map(flow => ({ value: flow.flowId, label: flow.flowName }))], [flows])

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={template.templateId ? '编辑模板' : '新建模板'}
      width="720px"
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>取消</Button>
          <Button
            loading={saving}
            onClick={async () => {
              const sourceFlow = flows.find(flow => flow.flowId === flowId)
              setSaving(true)
              try {
                await onSubmit({
                  ...template,
                  flowSnapshot: sourceFlow || template.flowSnapshot,
                })
              } finally {
                setSaving(false)
              }
            }}
          >
            保存模板
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        <FormItem label="模板名称" required>
          <Input value={template.templateName} onChange={e => setTemplate(prev => ({ ...prev, templateName: e.target.value }))} />
        </FormItem>
        <FormItem label="来源流程" required>
          <Select value={flowId} options={flowOptions} onChange={e => setFlowId(e.target.value)} />
        </FormItem>
        <FormItem label="模板描述">
          <Textarea rows={4} value={template.description} onChange={e => setTemplate(prev => ({ ...prev, description: e.target.value }))} />
        </FormItem>
        <FormItem label="标签" hint="逗号分隔">
          <Input
            value={(template.tags || []).join(', ')}
            onChange={e => setTemplate(prev => ({ ...prev, tags: e.target.value.split(',').map(item => item.trim()).filter(Boolean) }))}
          />
        </FormItem>
      </div>
    </Modal>
  )
}
