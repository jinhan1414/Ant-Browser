import { Button, FormItem, Input, Select, Textarea } from '../../../shared/components'
import type { RPAFlowStep, RPAStepType } from '../types'

const STEP_TYPE_OPTIONS: { value: RPAStepType; label: string }[] = [
  { value: 'start_browser', label: '启动浏览器' },
  { value: 'open_urls', label: '打开网址' },
  { value: 'wait', label: '等待' },
  { value: 'stop_browser', label: '关闭浏览器' },
]

interface FlowStepEditorProps {
  steps: RPAFlowStep[]
  onChange: (steps: RPAFlowStep[]) => void
}

function emptyStep(type: RPAStepType): RPAFlowStep {
  return {
    stepId: `step-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
    stepName: STEP_TYPE_OPTIONS.find(item => item.value === type)?.label || '新步骤',
    stepType: type,
    config: {},
  }
}

export function FlowStepEditor({ steps, onChange }: FlowStepEditorProps) {
  const updateStep = (stepId: string, updater: (step: RPAFlowStep) => RPAFlowStep) => {
    onChange(steps.map(step => (step.stepId === stepId ? updater(step) : step)))
  }

  const removeStep = (stepId: string) => {
    onChange(steps.filter(step => step.stepId !== stepId))
  }

  const addStep = (type: RPAStepType) => {
    onChange([...steps, emptyStep(type)])
  }

  return (
    <div className="space-y-3">
      {steps.map((step, index) => (
        <div key={step.stepId} className="rounded-lg border border-[var(--color-border-default)] p-4 bg-[var(--color-bg-secondary)]/60">
          <div className="flex items-center justify-between gap-3 mb-3">
            <div className="text-sm font-medium text-[var(--color-text-primary)]">步骤 {index + 1}</div>
            <Button size="sm" variant="ghost" onClick={() => removeStep(step.stepId)}>删除</Button>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <FormItem label="步骤名称">
              <Input value={step.stepName} onChange={e => updateStep(step.stepId, current => ({ ...current, stepName: e.target.value }))} />
            </FormItem>
            <FormItem label="步骤类型">
              <Select
                value={step.stepType}
                options={STEP_TYPE_OPTIONS}
                onChange={e => updateStep(step.stepId, current => ({ ...current, stepType: e.target.value as RPAStepType }))}
              />
            </FormItem>
          </div>
          {(step.stepType === 'start_browser' || step.stepType === 'open_urls') && (
            <FormItem label="网址列表" hint="每行一个 URL">
              <Textarea
                rows={3}
                value={Array.isArray(step.config.startUrls) ? step.config.startUrls.join('\n') : ''}
                onChange={e => updateStep(step.stepId, current => ({
                  ...current,
                  config: {
                    ...current.config,
                    startUrls: e.target.value.split('\n').map(item => item.trim()).filter(Boolean),
                  },
                }))}
              />
            </FormItem>
          )}
          {step.stepType === 'wait' && (
            <FormItem label="等待毫秒">
              <Input
                type="number"
                min={1}
                value={String(step.config.durationMs || 1000)}
                onChange={e => updateStep(step.stepId, current => ({
                  ...current,
                  config: {
                    ...current.config,
                    durationMs: Number(e.target.value) || 1000,
                  },
                }))}
              />
            </FormItem>
          )}
        </div>
      ))}
      <div className="flex flex-wrap gap-2">
        {STEP_TYPE_OPTIONS.map(option => (
          <Button key={option.value} size="sm" variant="secondary" onClick={() => addStep(option.value)}>
            添加{option.label}
          </Button>
        ))}
      </div>
    </div>
  )
}
