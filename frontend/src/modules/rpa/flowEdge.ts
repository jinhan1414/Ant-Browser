import type { Edge } from 'reactflow'
import type { RPAFlowDocument, RPAFlowEdge, RPAFlowEdgeBranchType, RPAFlowNodeCatalogItem, RPAFlowNodeType } from './types'
import { findNodeCatalogItem } from './nodeCatalog'

export function normalizeEdgeBranchType(branchType?: string, condition?: string): RPAFlowEdgeBranchType {
  const value = (branchType || condition || '').trim()
  if (value === 'true' || value === 'false' || value === 'on_error') {
    return value
  }
  return 'default'
}

export function branchTypeToCondition(branchType: RPAFlowEdgeBranchType) {
  return branchType === 'default' ? '' : branchType
}

export function normalizeDocumentEdge(edge: Partial<RPAFlowEdge>): RPAFlowEdge {
  const branchType = normalizeEdgeBranchType(edge.branchType, edge.condition)
  return {
    edgeId: edge.edgeId || '',
    sourceNodeId: edge.sourceNodeId || '',
    targetNodeId: edge.targetNodeId || '',
    label: String(edge.label || '').trim(),
    branchType,
    condition: branchTypeToCondition(branchType),
  }
}

export function getBranchTypeLabel(branchType: RPAFlowEdgeBranchType) {
  switch (branchType) {
    case 'true':
      return 'TRUE'
    case 'false':
      return 'FALSE'
    case 'on_error':
      return '异常'
    default:
      return ''
  }
}

export function buildEdgeDisplayLabel(edge: Pick<RPAFlowEdge, 'label' | 'branchType'>) {
  const branchLabel = getBranchTypeLabel(edge.branchType)
  const customLabel = (edge.label || '').trim()
  if (branchLabel && customLabel) {
    return `${branchLabel} · ${customLabel}`
  }
  return branchLabel || customLabel
}

export function buildCanvasEdgeData(edge: RPAFlowEdge) {
  return {
    branchType: edge.branchType,
    labelText: edge.label,
    displayLabel: buildEdgeDisplayLabel(edge),
  }
}

export function getAllowedBranchTypes(
  catalog: RPAFlowNodeCatalogItem[],
  sourceNodeType: RPAFlowNodeType,
): RPAFlowEdgeBranchType[] {
  const item = findNodeCatalogItem(catalog, sourceNodeType)
  if (item?.supportsIfBranch) {
    return ['true', 'false']
  }
  if (item?.supportsOnError) {
    return ['default', 'on_error']
  }
  return ['default']
}

export function getNextBranchType(
  catalog: RPAFlowNodeCatalogItem[],
  sourceNodeType: RPAFlowNodeType,
  edges: RPAFlowEdge[],
  currentEdgeId = '',
): RPAFlowEdgeBranchType {
  const allowed = getAllowedBranchTypes(catalog, sourceNodeType)
  for (const branchType of allowed) {
    const used = edges.some(edge => edge.edgeId !== currentEdgeId && edge.branchType === branchType)
    if (!used) {
      return branchType
    }
  }
  return allowed[0] || 'default'
}

export function getBranchTypeOptions(
  catalog: RPAFlowNodeCatalogItem[],
  sourceNodeType: RPAFlowNodeType,
  edges: RPAFlowEdge[],
  currentEdgeId = '',
) {
  const allowed = getAllowedBranchTypes(catalog, sourceNodeType)
  return allowed
    .filter(branchType => {
      if (currentEdgeId && edges.find(edge => edge.edgeId === currentEdgeId)?.branchType === branchType) {
        return true
      }
      return !edges.some(edge => edge.edgeId !== currentEdgeId && edge.branchType === branchType)
    })
    .map(branchType => ({
      value: branchType,
      label: getBranchTypeOptionLabel(branchType),
    }))
}

export function getBranchTypeOptionLabel(branchType: RPAFlowEdgeBranchType) {
  switch (branchType) {
    case 'true':
      return 'TRUE 分支'
    case 'false':
      return 'FALSE 分支'
    case 'on_error':
      return '异常分支'
    default:
      return '默认分支'
  }
}

export function createEdgeDraft(
  source: string,
  target: string,
  branchType: RPAFlowEdgeBranchType,
): Pick<RPAFlowEdge, 'sourceNodeId' | 'targetNodeId' | 'branchType' | 'condition' | 'label'> {
  return {
    sourceNodeId: source,
    targetNodeId: target,
    branchType,
    condition: branchTypeToCondition(branchType),
    label: '',
  }
}

export function updateEdgeInDocument(document: RPAFlowDocument, nextEdge: RPAFlowEdge): RPAFlowDocument {
  return {
    ...document,
    edges: document.edges.map(edge => edge.edgeId === nextEdge.edgeId ? normalizeDocumentEdge(nextEdge) : edge),
  }
}

export function removeEdgeFromDocument(document: RPAFlowDocument, edgeId: string): RPAFlowDocument {
  return {
    ...document,
    edges: document.edges.filter(edge => edge.edgeId !== edgeId),
  }
}

export function edgeFromCanvas(edge: Edge, existingEdge?: RPAFlowEdge): RPAFlowEdge {
  const branchType = normalizeEdgeBranchType(
    String(edge.data?.branchType || existingEdge?.branchType || ''),
    String(existingEdge?.condition || ''),
  )
  return normalizeDocumentEdge({
    edgeId: edge.id,
    sourceNodeId: edge.source,
    targetNodeId: edge.target,
    label: String(edge.data?.labelText || existingEdge?.label || ''),
    branchType,
  })
}
