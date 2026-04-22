import { useEffect, useMemo, useState } from 'react'
import { Button, FormItem, Input, Modal, Select, Switch } from '../../../shared/components'
import type { BrowserProfile } from '../../browser/types'
import type { RPAFlow, RPATask, RPATaskScheduleConfig, RPATaskTarget } from '../types'

interface TaskEditorModalProps {
  open: boolean
  flows: RPAFlow[]
  profiles: BrowserProfile[]
  initialTask: RPATask | null
  initialTargets: RPATaskTarget[]
  onClose: () => void
  onSubmit: (task: RPATask, targets: RPATaskTarget[]) => Promise<void>
}

const emptyTask = (): RPATask => ({
  taskId: '',
  taskName: '',
  flowId: '',
  executionOrder: 'sequential',
  taskType: 'manual',
  scheduleConfig: {},
  enabled: true,
  lastRunAt: '',
  createdAt: '',
  updatedAt: '',
})

const DEFAULT_SCHEDULE_CONFIG: RPATaskScheduleConfig = {
  cron: '',
  timezone: 'Asia/Shanghai',
}

const TIMEZONE_OPTIONS = [
  { value: 'Asia/Shanghai', label: 'Asia/Shanghai' },
  { value: 'UTC', label: 'UTC' },
  { value: 'America/New_York', label: 'America/New_York' },
  { value: 'America/Los_Angeles', label: 'America/Los_Angeles' },
]

function normalizeScheduleConfig(raw: Record<string, any> | null | undefined): RPATaskScheduleConfig {
  return {
    cron: typeof raw?.cron === 'string' ? raw.cron : '',
    timezone: typeof raw?.timezone === 'string' && raw.timezone ? raw.timezone : DEFAULT_SCHEDULE_CONFIG.timezone,
  }
}

function normalizeTask(task: RPATask | null): RPATask {
  const next = task || emptyTask()
  return {
    ...next,
    scheduleConfig: next.taskType === 'scheduled'
      ? normalizeScheduleConfig(next.scheduleConfig)
      : {},
  }
}

export function TaskEditorModal({ open, flows, profiles, initialTask, initialTargets, onClose, onSubmit }: TaskEditorModalProps) {
  const [task, setTask] = useState<RPATask>(emptyTask())
  const [selectedProfileIds, setSelectedProfileIds] = useState<string[]>([])
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    setTask(normalizeTask(initialTask))
    setSelectedProfileIds(initialTargets.map(item => item.profileId))
  }, [initialTask, initialTargets, open])

  const flowOptions = useMemo(() => [{ value: '', label: '请选择流程' }, ...flows.map(flow => ({ value: flow.flowId, label: flow.flowName }))], [flows])

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={task.taskId ? '编辑任务' : '新建任务'}
      width="880px"
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>取消</Button>
          <Button
            loading={saving}
            onClick={async () => {
              setSaving(true)
              try {
                await onSubmit(task, selectedProfileIds.map((profileId, index) => ({ taskId: task.taskId, profileId, sortOrder: index + 1 })))
              } finally {
                setSaving(false)
              }
            }}
          >
            保存任务
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormItem label="任务名称" required>
            <Input value={task.taskName} onChange={e => setTask(prev => ({ ...prev, taskName: e.target.value }))} />
          </FormItem>
          <FormItem label="运行流程" required>
            <Select value={task.flowId} options={flowOptions} onChange={e => setTask(prev => ({ ...prev, flowId: e.target.value }))} />
          </FormItem>
          <FormItem label="执行顺序">
            <Select
              value={task.executionOrder}
              options={[{ value: 'sequential', label: '顺序执行' }, { value: 'random', label: '随机执行' }]}
              onChange={e => setTask(prev => ({ ...prev, executionOrder: e.target.value as RPATask['executionOrder'] }))}
            />
          </FormItem>
          <FormItem label="执行类型">
            <Select
              value={task.taskType}
              options={[{ value: 'manual', label: '普通任务' }, { value: 'scheduled', label: '计划任务' }]}
              onChange={e => setTask(prev => {
                const taskType = e.target.value as RPATask['taskType']
                return {
                  ...prev,
                  taskType,
                  scheduleConfig: taskType === 'scheduled' ? normalizeScheduleConfig(prev.scheduleConfig) : {},
                }
              })}
            />
          </FormItem>
        </div>
        {task.taskType === 'scheduled' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 rounded-lg border border-[var(--color-border-default)] p-4 bg-[var(--color-bg-secondary)]/40">
            <FormItem label="Cron 表达式" required hint="示例：0 9 * * * 表示每天 09:00">
              <Input
                value={normalizeScheduleConfig(task.scheduleConfig).cron}
                onChange={e => setTask(prev => ({
                  ...prev,
                  scheduleConfig: {
                    ...normalizeScheduleConfig(prev.scheduleConfig),
                    cron: e.target.value,
                  },
                }))}
                placeholder="0 9 * * *"
              />
            </FormItem>
            <FormItem label="时区">
              <Select
                value={normalizeScheduleConfig(task.scheduleConfig).timezone}
                options={TIMEZONE_OPTIONS}
                onChange={e => setTask(prev => ({
                  ...prev,
                  scheduleConfig: {
                    ...normalizeScheduleConfig(prev.scheduleConfig),
                    timezone: e.target.value,
                  },
                }))}
              />
            </FormItem>
          </div>
        )}
        <FormItem label="启用状态">
          <Switch checked={task.enabled} onChange={checked => setTask(prev => ({ ...prev, enabled: checked }))} />
        </FormItem>
        <FormItem label="执行环境">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-2 max-h-[260px] overflow-auto border border-[var(--color-border-default)] rounded-lg p-3">
            {profiles.map(profile => {
              const checked = selectedProfileIds.includes(profile.profileId)
              return (
                <label key={profile.profileId} className="flex items-center gap-2 text-sm text-[var(--color-text-secondary)]">
                  <input
                    type="checkbox"
                    checked={checked}
                    onChange={() => {
                      setSelectedProfileIds(prev => checked ? prev.filter(item => item !== profile.profileId) : [...prev, profile.profileId])
                    }}
                  />
                  <span>{profile.profileName}</span>
                </label>
              )
            })}
          </div>
        </FormItem>
      </div>
    </Modal>
  )
}
