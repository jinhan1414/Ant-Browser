export type RPAStepType = 'start_browser' | 'open_urls' | 'wait' | 'stop_browser'
export type RPATaskExecutionOrder = 'sequential' | 'random'
export type RPATaskType = 'manual' | 'scheduled'
export type RPARunStatus = 'pending' | 'running' | 'success' | 'failed' | 'cancelled'
export type RPAFlowSourceType = 'visual' | 'xml_import'
export type RPAFlowNodeType =
  | 'start'
  | 'end'
  | 'browser.start'
  | 'browser.open_url'
  | 'delay'
  | 'browser.stop'

export interface RPAFlowGroup {
  groupId: string
  groupName: string
  sortOrder: number
  createdAt: string
  updatedAt: string
}

export interface RPAFlowStep {
  stepId: string
  stepName: string
  stepType: RPAStepType
  config: Record<string, any>
}

export interface RPAFlowPosition {
  x: number
  y: number
}

export interface RPAFlowVariable {
  name: string
  type: string
  defaultValue: string
}

export interface RPAFlowNode {
  nodeId: string
  nodeType: RPAFlowNodeType
  label: string
  position: RPAFlowPosition
  config: Record<string, any>
}

export interface RPAFlowEdge {
  edgeId: string
  sourceNodeId: string
  targetNodeId: string
  condition: string
}

export interface RPAFlowDocument {
  schemaVersion: number
  variables: RPAFlowVariable[]
  nodes: RPAFlowNode[]
  edges: RPAFlowEdge[]
}

export interface RPAFlow {
  flowId: string
  flowName: string
  groupId: string
  steps: RPAFlowStep[]
  document: RPAFlowDocument
  sourceType: RPAFlowSourceType
  sourceXml: string
  shareCode: string
  version: number
  createdAt: string
  updatedAt: string
}

export interface RPAFlowXMLImportInput {
  flowName: string
  groupId: string
  xmlText: string
}

export interface RPATask {
  taskId: string
  taskName: string
  flowId: string
  executionOrder: RPATaskExecutionOrder
  taskType: RPATaskType
  scheduleConfig: Record<string, any>
  enabled: boolean
  lastRunAt: string
  createdAt: string
  updatedAt: string
}

export interface RPATaskScheduleConfig {
  cron: string
  timezone: string
}

export interface RPATaskTarget {
  taskId: string
  profileId: string
  sortOrder: number
}

export interface RPARun {
  runId: string
  taskId: string
  flowId: string
  triggerType: 'manual' | 'scheduled'
  status: RPARunStatus
  summary: string
  startedAt: string
  finishedAt: string
  errorMessage: string
}

export interface RPARunTarget {
  runTargetId: string
  runId: string
  profileId: string
  profileName: string
  status: RPARunStatus
  stepIndex: number
  startedAt: string
  finishedAt: string
  errorMessage: string
  debugPort: number
}

export interface RPATemplate {
  templateId: string
  templateName: string
  description: string
  tags: string[]
  flowSnapshot: RPAFlow
  createdAt: string
  updatedAt: string
}
