import { useEffect, useState } from 'react'
import { RefreshCw } from 'lucide-react'
import { Button, Card, Table } from '../../../shared/components'
import type { TableColumn } from '../../../shared/components'
import { fetchRPARuns, fetchRPARunTargets } from '../api'
import type { RPARun, RPARunTarget } from '../types'

const formatTime = (value: string) => value ? new Date(value).toLocaleString('zh-CN') : '-'

export function RunRecordsPage() {
  const [runs, setRuns] = useState<RPARun[]>([])
  const [runTargets, setRunTargets] = useState<RPARunTarget[]>([])
  const [activeRunId, setActiveRunId] = useState('')

  const loadRuns = async () => {
    const items = await fetchRPARuns()
    setRuns(items)
    const nextRunId = activeRunId || items[0]?.runId || ''
    setActiveRunId(nextRunId)
    if (nextRunId) {
      setRunTargets(await fetchRPARunTargets(nextRunId))
    } else {
      setRunTargets([])
    }
  }

  useEffect(() => {
    void loadRuns()
  }, [])

  const runColumns: TableColumn<RPARun>[] = [
    { key: 'taskId', title: '任务 ID' },
    { key: 'status', title: '状态' },
    { key: 'summary', title: '摘要' },
    { key: 'startedAt', title: '开始时间', render: value => formatTime(value) },
    { key: 'finishedAt', title: '结束时间', render: value => formatTime(value) },
  ]

  const targetColumns: TableColumn<RPARunTarget>[] = [
    { key: 'profileName', title: '实例' },
    { key: 'status', title: '状态' },
    { key: 'stepIndex', title: '执行到步骤', align: 'center' },
    { key: 'debugPort', title: '调试端口', align: 'center' },
    { key: 'errorMessage', title: '失败原因' },
  ]

  return (
    <div className="space-y-5 animate-fade-in">
      <Card
        title="运行记录"
        subtitle="查看任务运行结果和每个浏览器实例的执行明细"
        actions={<Button size="sm" variant="secondary" onClick={() => void loadRuns()}><RefreshCw className="w-4 h-4" />刷新</Button>}
      >
        <div className="text-sm text-[var(--color-text-secondary)]">
          当前共 {runs.length} 条运行记录，点击某条记录可查看对应实例明细。
        </div>
      </Card>

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-5">
        <Card title="任务运行">
          <Table
            columns={runColumns}
            data={runs}
            rowKey="runId"
            maxHeight="420px"
            onRowClick={async (record) => {
              setActiveRunId(record.runId)
              setRunTargets(await fetchRPARunTargets(record.runId))
            }}
          />
        </Card>
        <Card title="实例运行明细" subtitle={activeRunId ? `Run ID: ${activeRunId}` : '暂无选中运行'}>
          <Table columns={targetColumns} data={runTargets} rowKey="runTargetId" maxHeight="420px" emptyText="暂无实例明细" />
        </Card>
      </div>
    </div>
  )
}
