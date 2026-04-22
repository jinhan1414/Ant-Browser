import { useEffect, useState } from 'react'
import { Button, FormItem, Input, Modal, Select, Textarea, toast } from '../../../shared/components'
import { importRPAFlowXML } from '../api'
import type { RPAFlow, RPAFlowGroup } from '../types'

interface FlowXMLModalProps {
  open: boolean
  groups: RPAFlowGroup[]
  defaultGroupId: string
  onClose: () => void
  submitText?: string
  submitXML?: (payload: { flowName: string; groupId: string; xmlText: string }) => Promise<RPAFlow | null>
  onImported: (flow: RPAFlow) => Promise<void>
}

export function FlowXMLModal({ open, groups, defaultGroupId, onClose, submitText = '校验并导入', submitXML, onImported }: FlowXMLModalProps) {
  const [flowName, setFlowName] = useState('')
  const [groupId, setGroupId] = useState('')
  const [xmlText, setXMLText] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!open) {
      return
    }
    setFlowName('')
    setGroupId(defaultGroupId)
    setXMLText('')
  }, [defaultGroupId, open])

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="导入 XML 流程"
      width="880px"
      footer={(
        <>
          <Button variant="secondary" onClick={onClose}>取消</Button>
          <Button
            loading={loading}
            onClick={async () => {
              setLoading(true)
              try {
                const flow = await (submitXML || importRPAFlowXML)({ flowName, groupId, xmlText })
                if (flow) {
                  await onImported(flow)
                  onClose()
                }
              } catch (error: any) {
                toast.error(error?.message || 'XML 导入失败')
              } finally {
                setLoading(false)
              }
            }}
          >
            {submitText}
          </Button>
        </>
      )}
    >
      <div className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormItem label="流程名称" hint="为空时后端会回退到默认名称">
            <Input value={flowName} onChange={e => setFlowName(e.target.value)} />
          </FormItem>
          <FormItem label="所属分组">
            <Select value={groupId} options={[{ value: '', label: '未分组' }, ...groups.map(group => ({ value: group.groupId, label: group.groupName }))]} onChange={e => setGroupId(e.target.value)} />
          </FormItem>
        </div>
        <FormItem label="AntRPA XML">
          <Textarea rows={20} value={xmlText} onChange={e => setXMLText(e.target.value)} placeholder="<flow schemaVersion=&quot;1&quot; name=&quot;打开站点&quot;>..." />
        </FormItem>
      </div>
    </Modal>
  )
}
