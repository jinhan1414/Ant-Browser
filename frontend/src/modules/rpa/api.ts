import type {
  RPAFlow,
  RPAFlowNodeCatalogPayload,
  RPAFlowGroup,
  RPAFlowXMLImportInput,
  RPARun,
  RPARunStep,
  RPARunTarget,
  RPATask,
  RPATaskTarget,
  RPATemplate,
} from './types'

const getBindings = async () => {
  try {
    return await import('../../wailsjs/go/main/App')
  } catch {
    return null
  }
}

async function getRPAAppBinding(): Promise<any> {
  const bindings: any = await getBindings()
  if (bindings) {
    return bindings
  }
  return (window as any).go?.main?.App || null
}

export async function fetchRPAFlowGroups(): Promise<RPAFlowGroup[]> {
  const app = await getRPAAppBinding()
  return (await app?.RPAFlowGroupList?.()) || []
}

export async function createRPAFlowGroup(groupName: string): Promise<RPAFlowGroup | null> {
  const app = await getRPAAppBinding()
  return (await app?.RPAFlowGroupCreate?.({ groupName, sortOrder: 0 })) || null
}

export async function fetchRPAFlows(keyword = '', groupId = ''): Promise<RPAFlow[]> {
  const app = await getRPAAppBinding()
  return (await app?.RPAFlowList?.(keyword, groupId)) || []
}

export async function saveRPAFlow(flow: RPAFlow): Promise<RPAFlow | null> {
  const app = await getRPAAppBinding()
  return (await app?.RPAFlowSave?.(flow)) || null
}

export async function deleteRPAFlow(flowId: string): Promise<void> {
  const app = await getRPAAppBinding()
  await app?.RPAFlowDelete?.(flowId)
}

export async function shareRPAFlow(flowId: string): Promise<string> {
  const app = await getRPAAppBinding()
  return (await app?.RPAFlowShare?.(flowId)) || ''
}

export async function importRPAFlowByShareCode(shareCode: string): Promise<RPAFlow | null> {
  const app = await getRPAAppBinding()
  return (await app?.RPAFlowImportByShareCode?.(shareCode)) || null
}

export async function importRPAFlowXML(input: RPAFlowXMLImportInput): Promise<RPAFlow | null> {
  const app = await getRPAAppBinding()
  return (await app?.RPAFlowImportXML?.(input)) || null
}

export async function parseRPAFlowXML(input: RPAFlowXMLImportInput): Promise<RPAFlow | null> {
  const app = await getRPAAppBinding()
  return (await app?.RPAFlowParseXML?.(input)) || null
}

export async function exportRPAFlowXML(flowId: string): Promise<string> {
  const app = await getRPAAppBinding()
  return (await app?.RPAFlowExportXML?.(flowId)) || ''
}

export async function encodeRPAFlowXML(flow: RPAFlow): Promise<string> {
  const app = await getRPAAppBinding()
  return (await app?.RPAFlowEncodeXML?.(flow)) || ''
}

export async function fetchRPAFlowNodeCatalog(): Promise<RPAFlowNodeCatalogPayload | null> {
  const app = await getRPAAppBinding()
  return (await app?.RPAFlowNodeCatalog?.()) || null
}

export async function fetchRPATasks(): Promise<RPATask[]> {
  const app = await getRPAAppBinding()
  return (await app?.RPATaskList?.()) || []
}

export async function getRPATask(taskId: string): Promise<{ task: RPATask | null; targets: RPATaskTarget[] }> {
  const app = await getRPAAppBinding()
  const result = await app?.RPATaskGet?.(taskId)
  if (result && typeof result === 'object' && !Array.isArray(result)) {
    return {
      task: result.task || null,
      targets: Array.isArray(result.targets) ? result.targets : [],
    }
  }
  if (Array.isArray(result)) {
    return {
      task: result[0] || null,
      targets: result[1] || [],
    }
  }
  return { task: null, targets: [] }
}

export async function saveRPATask(task: RPATask, targets: RPATaskTarget[]): Promise<RPATask | null> {
  const app = await getRPAAppBinding()
  return (await app?.RPATaskSave?.(task, targets)) || null
}

export async function deleteRPATask(taskId: string): Promise<void> {
  const app = await getRPAAppBinding()
  await app?.RPATaskDelete?.(taskId)
}

export async function executeRPATask(taskId: string): Promise<RPARun | null> {
  const app = await getRPAAppBinding()
  return (await app?.RPATaskExecute?.(taskId)) || null
}

export async function fetchRPARuns(): Promise<RPARun[]> {
  const app = await getRPAAppBinding()
  return (await app?.RPARunList?.()) || []
}

export async function fetchRPARunTargets(runId: string): Promise<RPARunTarget[]> {
  const app = await getRPAAppBinding()
  return (await app?.RPARunTargetList?.(runId)) || []
}

export async function fetchRPARunSteps(runId: string): Promise<RPARunStep[]> {
  const app = await getRPAAppBinding()
  return (await app?.RPARunStepList?.(runId)) || []
}

export async function fetchRPATemplates(): Promise<RPATemplate[]> {
  const app = await getRPAAppBinding()
  return (await app?.RPATemplateList?.()) || []
}

export async function saveRPATemplate(template: RPATemplate): Promise<RPATemplate | null> {
  const app = await getRPAAppBinding()
  return (await app?.RPATemplateSave?.(template)) || null
}

export async function deleteRPATemplate(templateId: string): Promise<void> {
  const app = await getRPAAppBinding()
  await app?.RPATemplateDelete?.(templateId)
}

export async function createFlowFromTemplate(templateId: string): Promise<RPAFlow | null> {
  const app = await getRPAAppBinding()
  return (await app?.RPATemplateCreateFlow?.(templateId)) || null
}
