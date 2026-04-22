import { useEffect, useMemo, useState } from 'react'
import { Play, Plus, RefreshCw, Trash2 } from 'lucide-react'
import { Button, Card, Table, toast } from '../../../shared/components'
import type { TableColumn } from '../../../shared/components'
import { fetchBrowserProfiles } from '../../browser/api'
import type { BrowserProfile } from '../../browser/types'
import { deleteRPATask, executeRPATask, fetchRPAFlows, fetchRPATasks, getRPATask, saveRPATask } from '../api'
import { TaskEditorModal } from '../components/TaskEditorModal'
import type { RPAFlow, RPATask, RPATaskTarget } from '../types'

const formatTime = (value: string) => value ? new Date(value).toLocaleString('zh-CN') : '-'
const formatSchedule = (task: RPATask) => {
  if (task.taskType !== 'scheduled') {
    return '-'
  }
  const cron = typeof task.scheduleConfig?.cron === 'string' ? task.scheduleConfig.cron : ''
  const timezone = typeof task.scheduleConfig?.timezone === 'string' ? task.scheduleConfig.timezone : ''
  if (!cron) {
    return '未配置'
  }
  return timezone ? `${cron} (${timezone})` : cron
}

export function TaskManagementPage() {
  const [tasks, setTasks] = useState<RPATask[]>([])
  const [flows, setFlows] = useState<RPAFlow[]>([])
  const [profiles, setProfiles] = useState<BrowserProfile[]>([])
  const [editorOpen, setEditorOpen] = useState(false)
  const [editingTask, setEditingTask] = useState<RPATask | null>(null)
  const [editingTargets, setEditingTargets] = useState<RPATaskTarget[]>([])

  const loadData = async () => {
    const [taskItems, flowItems, profileItems] = await Promise.all([
      fetchRPATasks(),
      fetchRPAFlows(),
      fetchBrowserProfiles(),
    ])
    setTasks(taskItems)
    setFlows(flowItems)
    setProfiles(profileItems)
  }

  useEffect(() => {
    void loadData()
  }, [])

  const flowMap = useMemo(() => new Map(flows.map(flow => [flow.flowId, flow.flowName])), [flows])

  const columns: TableColumn<RPATask>[] = [
    { key: 'taskName', title: '任务名称' },
    { key: 'flowId', title: '运行流程', render: value => flowMap.get(value) || '-' },
    { key: 'executionOrder', title: '执行顺序', render: value => value === 'random' ? '随机执行' : '顺序执行' },
    { key: 'taskType', title: '执行类型', render: value => value === 'scheduled' ? '计划任务' : '普通任务' },
    { key: 'scheduleConfig', title: '定时信息', render: (_, record) => formatSchedule(record) },
    { key: 'lastRunAt', title: '上次运行', render: value => formatTime(value) },
    {
      key: 'actions',
      title: '操作',
      align: 'right',
      render: (_, record) => (
        <div className="flex justify-end gap-1">
          <Button size="sm" variant="ghost" onClick={async () => {
            const detail = await getRPATask(record.taskId)
            setEditingTask(detail.task)
            setEditingTargets(detail.targets)
            setEditorOpen(true)
          }}>编辑</Button>
          <Button size="sm" variant="ghost" onClick={async () => {
            try {
              const run = await executeRPATask(record.taskId)
              if (run) {
                const message = run.status === 'success'
                  ? `任务已执行，状态：${run.status}`
                  : `任务执行失败：${run.errorMessage || run.summary || run.status}`
                if (run.status === 'success') {
                  toast.success(message)
                } else {
                  toast.error(message)
                }
                await loadData()
              }
            } catch (error: any) {
              toast.error(error?.message || '任务执行失败')
            }
          }}><Play className="w-3.5 h-3.5" />执行</Button>
          <Button size="sm" variant="ghost" onClick={async () => {
            await deleteRPATask(record.taskId)
            toast.success('任务已删除')
            await loadData()
          }}><Trash2 className="w-3.5 h-3.5" /></Button>
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-5 animate-fade-in">
      <Card
        title="任务管理"
        subtitle="将流程绑定到指定浏览器实例列表，并控制顺序、普通任务或计划任务配置"
        actions={
          <>
            <Button size="sm" variant="secondary" onClick={() => void loadData()}><RefreshCw className="w-4 h-4" />刷新</Button>
            <Button size="sm" onClick={() => { setEditingTask(null); setEditingTargets([]); setEditorOpen(true) }}><Plus className="w-4 h-4" />新建任务</Button>
          </>
        }
      >
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
          <div className="rounded-lg border border-[var(--color-border-default)] p-4 bg-[var(--color-bg-secondary)]/50">
            <div className="text-[var(--color-text-muted)]">任务数</div>
            <div className="mt-2 text-2xl font-semibold text-[var(--color-text-primary)]">{tasks.length}</div>
          </div>
          <div className="rounded-lg border border-[var(--color-border-default)] p-4 bg-[var(--color-bg-secondary)]/50">
            <div className="text-[var(--color-text-muted)]">流程数</div>
            <div className="mt-2 text-2xl font-semibold text-[var(--color-text-primary)]">{flows.length}</div>
          </div>
          <div className="rounded-lg border border-[var(--color-border-default)] p-4 bg-[var(--color-bg-secondary)]/50">
            <div className="text-[var(--color-text-muted)]">可选环境</div>
            <div className="mt-2 text-2xl font-semibold text-[var(--color-text-primary)]">{profiles.length}</div>
          </div>
        </div>
      </Card>

      <Card padding="none">
        <Table columns={columns} data={tasks} rowKey="taskId" emptyText="暂无任务" />
      </Card>

      <TaskEditorModal
        open={editorOpen}
        flows={flows}
        profiles={profiles}
        initialTask={editingTask}
        initialTargets={editingTargets}
        onClose={() => setEditorOpen(false)}
        onSubmit={async (task, targets) => {
          try {
            const saved = await saveRPATask(task, targets)
            if (saved) {
              toast.success('任务已保存')
              setEditorOpen(false)
              await loadData()
            }
          } catch (error: any) {
            toast.error(error?.message || '任务保存失败')
          }
        }}
      />
    </div>
  )
}
