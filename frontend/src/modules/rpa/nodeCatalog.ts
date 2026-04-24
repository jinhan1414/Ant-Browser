import { useEffect, useMemo, useState } from 'react'
import { fetchRPAFlowNodeCatalog } from './api'
import type {
  RPAFlowNodeCatalogItem,
  RPAFlowNodeCatalogPayload,
  RPAFlowNodeField,
  RPAFlowNodeType,
} from './types'

const EMPTY_PAYLOAD: RPAFlowNodeCatalogPayload = {
  items: [],
  xmlPromptTemplate: '',
}

let cachedPayload: RPAFlowNodeCatalogPayload | null = null
let cachedPromise: Promise<RPAFlowNodeCatalogPayload> | null = null

export async function loadFlowNodeCatalog(force = false): Promise<RPAFlowNodeCatalogPayload> {
  if (!force && cachedPayload) {
    return cachedPayload
  }
  if (!force && cachedPromise) {
    return cachedPromise
  }
  cachedPromise = fetchRPAFlowNodeCatalog()
    .then(payload => {
      cachedPayload = payload || EMPTY_PAYLOAD
      return cachedPayload
    })
    .finally(() => {
      cachedPromise = null
    })
  return cachedPromise
}

export function useFlowNodeCatalog() {
  const [payload, setPayload] = useState<RPAFlowNodeCatalogPayload>(cachedPayload || EMPTY_PAYLOAD)
  const [loading, setLoading] = useState(!cachedPayload)

  useEffect(() => {
    let active = true
    void loadFlowNodeCatalog().then(next => {
      if (!active) {
        return
      }
      setPayload(next)
      setLoading(false)
    })
    return () => {
      active = false
    }
  }, [])

  return useMemo(() => ({
    loading,
    payload,
    items: payload.items,
    xmlPromptTemplate: payload.xmlPromptTemplate,
  }), [loading, payload])
}

export function findNodeCatalogItem(items: RPAFlowNodeCatalogItem[], nodeType: RPAFlowNodeType) {
  return items.find(item => item.nodeType === nodeType) || null
}

export function getNodeLabel(items: RPAFlowNodeCatalogItem[], nodeType: RPAFlowNodeType) {
  return findNodeCatalogItem(items, nodeType)?.label || nodeType
}

export function getPaletteNodeItems(items: RPAFlowNodeCatalogItem[]) {
  return items.filter(item => item.palette)
}

export function getNodeTypeOptions(items: RPAFlowNodeCatalogItem[]) {
  return items.map(item => ({ value: item.nodeType, label: item.label }))
}

export function getNodeTypeLabelMap(items: RPAFlowNodeCatalogItem[]) {
  return items.reduce<Record<string, string>>((result, item) => {
    result[item.nodeType] = item.label
    return result
  }, {})
}

export function getNodeConnectionRules(items: RPAFlowNodeCatalogItem[], nodeType: RPAFlowNodeType) {
  const item = findNodeCatalogItem(items, nodeType)
  return {
    allowIncoming: item?.allowIncoming ?? nodeType !== 'start',
    allowOutgoing: item?.allowOutgoing ?? nodeType !== 'end',
    maxOutgoing: item?.maxOutgoing ?? 1,
    supportsIfBranch: item?.supportsIfBranch ?? false,
    supportsOnError: item?.supportsOnError ?? false,
    fixed: item?.fixed ?? (nodeType === 'start' || nodeType === 'end'),
  }
}

export function buildNodeConfig(items: RPAFlowNodeCatalogItem[], nodeType: RPAFlowNodeType, current: Record<string, any> = {}) {
  const item = findNodeCatalogItem(items, nodeType)
  if (!item) {
    return { ...current }
  }
  return item.fields.reduce<Record<string, any>>((result, field) => {
    if (current[field.key] !== undefined) {
      result[field.key] = current[field.key]
      return result
    }
    if (field.defaultValue !== undefined && field.defaultValue !== null && field.defaultValue !== '') {
      result[field.key] = field.defaultValue
    }
    return result
  }, {})
}

export function readNodeFieldValue(config: Record<string, any>, field: RPAFlowNodeField) {
  const value = config[field.key]
  if (field.storage === 'string_list_first') {
    return Array.isArray(value) ? String(value[0] || '') : ''
  }
  if (field.storage === 'number') {
    return value === undefined || value === null ? '' : String(value)
  }
  return String(value || '')
}

export function writeNodeFieldValue(config: Record<string, any>, field: RPAFlowNodeField, raw: string) {
  const next = { ...config }
  if (!raw.trim()) {
    delete next[field.key]
    return next
  }
  if (field.storage === 'number') {
    next[field.key] = Number(raw)
    return next
  }
  if (field.storage === 'string_list_first') {
    next[field.key] = [raw]
    return next
  }
  next[field.key] = raw
  return next
}
