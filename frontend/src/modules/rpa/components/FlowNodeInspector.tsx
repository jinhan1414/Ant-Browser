import { Button, FormItem, Input, Select } from '../../../shared/components'
import type { RPAFlowNode, RPAFlowNodeType } from '../types'

const NODE_TYPE_OPTIONS: { value: RPAFlowNodeType; label: string }[] = [
  { value: 'start', label: '开始' },
  { value: 'end', label: '结束' },
  { value: 'browser.start', label: '启动浏览器' },
  { value: 'browser.open_url', label: '打开页面' },
  { value: 'delay', label: '等待' },
  { value: 'browser.stop', label: '关闭浏览器' },
]

interface FlowNodeInspectorProps {
  node: RPAFlowNode | null
  onChange: (node: RPAFlowNode) => void
  onDelete: (nodeId: string) => void
}

export function FlowNodeInspector({ node, onChange, onDelete }: FlowNodeInspectorProps) {
  if (!node) {
    return (
      <div className="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-secondary)]/40 p-4 text-sm text-[var(--color-text-muted)]">
        选择一个节点后，在这里编辑标签和配置。
      </div>
    )
  }

  const updateConfig = (patch: Record<string, any>) => {
    onChange({
      ...node,
      config: {
        ...node.config,
        ...patch,
      },
    })
  }

  return (
    <div className="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-secondary)]/40 p-4 space-y-4">
      <div>
        <div className="text-sm font-semibold text-[var(--color-text-primary)]">节点属性</div>
        <div className="mt-1 text-xs text-[var(--color-text-muted)]">节点 ID：{node.nodeId}</div>
      </div>
      <FormItem label="节点名称">
        <Input value={node.label} onChange={e => onChange({ ...node, label: e.target.value })} />
      </FormItem>
      <FormItem label="节点类型">
        <Select
          value={node.nodeType}
          disabled={node.nodeType === 'start' || node.nodeType === 'end'}
          options={NODE_TYPE_OPTIONS}
          onChange={e => onChange({ ...node, nodeType: e.target.value as RPAFlowNodeType })}
        />
      </FormItem>
      {node.nodeType === 'browser.open_url' && (
        <FormItem label="页面地址" required>
          <Input value={String(node.config.url || '')} onChange={e => updateConfig({ url: e.target.value })} placeholder="https://example.com" />
        </FormItem>
      )}
      {node.nodeType === 'browser.start' && (
        <FormItem label="启动地址" hint="多个地址用换行分隔时，先填写第一个常用入口">
          <Input value={String((node.config.startUrls && node.config.startUrls[0]) || '')} onChange={e => updateConfig({ startUrls: e.target.value ? [e.target.value] : [] })} placeholder="https://example.com" />
        </FormItem>
      )}
      {node.nodeType === 'delay' && (
        <FormItem label="等待毫秒">
          <Input type="number" min={1} value={String(node.config.durationMs || 1000)} onChange={e => updateConfig({ durationMs: Number(e.target.value) || 1000 })} />
        </FormItem>
      )}
      {node.nodeType !== 'start' && node.nodeType !== 'end' && (
        <Button variant="danger" size="sm" onClick={() => onDelete(node.nodeId)}>
          删除节点
        </Button>
      )}
    </div>
  )
}
