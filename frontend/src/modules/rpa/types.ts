export type RPAStepType = 'start_browser' | 'open_urls' | 'wait' | 'stop_browser'
export type RPATaskExecutionOrder = 'sequential' | 'random'
export type RPATaskType = 'manual' | 'scheduled'
export type RPARunStatus = 'pending' | 'running' | 'success' | 'failed' | 'cancelled'
export type RPAFlowSourceType = 'visual' | 'xml_import'
export type RPAFlowNodeType = string
export type RPAFlowEdgeBranchType = 'default' | 'true' | 'false' | 'on_error'

export interface RPAFlowNodeField {
  key: string
  label: string
  kind: 'string' | 'number'
  storage: 'string' | 'number' | 'string_list_first'
  required: boolean
  hint: string
  placeholder: string
  xmlAttr: string
  promptSample: string
  defaultValue: any
  minValue: number
  multiline: boolean
}

export interface RPAFlowNodeCatalogItem {
  nodeType: RPAFlowNodeType
  label: string
  category: string
  description: string
  palette: boolean
  fixed: boolean
  xmlSupported: boolean
  promptEnabled: boolean
  runtimeSupported: boolean
  allowIncoming: boolean
  allowOutgoing: boolean
  maxOutgoing: number
  supportsIfBranch: boolean
  supportsOnError: boolean
  fields: RPAFlowNodeField[]
}

export interface RPAFlowNodeCatalogPayload {
  items: RPAFlowNodeCatalogItem[]
  xmlPromptTemplate: string
}

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
  label: string
  branchType: RPAFlowEdgeBranchType
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

export interface RPARunStep {
  runStepId: string
  runId: string
  runTargetId: string
  profileId: string
  nodeId: string
  nodeType: string
  nodeLabel: string
  status: RPARunStatus
  attempt: number
  outputJson: string
  errorMessage: string
  startedAt: string
  finishedAt: string
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
