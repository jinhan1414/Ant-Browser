import { useEffect, useMemo, useState } from 'react'
import { Button, Input } from '../../../shared/components'
import { TagFilterBar } from '../../browser/components/TagFilterBar'
import type { BrowserProfile } from '../../browser/types'
import { collectTaskProfileTags, filterTaskProfiles, type TaskProfileFilters } from '../taskProfileFilter'

const EMPTY_PROFILE_FILTERS: TaskProfileFilters = {
  keyword: '',
  tags: [],
}

interface TaskProfileSelectorProps {
  profiles: BrowserProfile[]
  selectedProfileIds: string[]
  resetKey: string
  onChange: (profileIds: string[]) => void
}

interface ProfileFilterToolbarProps {
  filters: TaskProfileFilters
  filteredCount: number
  visibleSelectedCount: number
  onChange: (filters: TaskProfileFilters) => void
  onSelectFiltered: () => void
  onClearFiltered: () => void
}

interface ProfileSelectionGridProps {
  profiles: BrowserProfile[]
  selectedIds: Set<string>
  onToggle: (profileId: string) => void
}

function appendMissingIds(current: string[], nextIds: string[]): string[] {
  return [...current, ...nextIds.filter(id => !current.includes(id))]
}

function removeIds(current: string[], idsToRemove: string[]): string[] {
  return current.filter(id => !idsToRemove.includes(id))
}

function ProfileFilterToolbar({
  filters,
  filteredCount,
  visibleSelectedCount,
  onChange,
  onSelectFiltered,
  onClearFiltered,
}: ProfileFilterToolbarProps) {
  const hasFilter = filters.keyword.trim().length > 0 || filters.tags.length > 0
  return (
    <div className="flex flex-col gap-2 md:flex-row md:items-center">
      <Input
        value={filters.keyword}
        onChange={e => onChange({ ...filters, keyword: e.target.value })}
        placeholder="按实例名称筛选"
        className="md:max-w-xs"
      />
      <div className="flex flex-wrap gap-2">
        <Button size="sm" variant="secondary" onClick={onSelectFiltered} disabled={filteredCount === 0}>全选当前结果</Button>
        <Button size="sm" variant="secondary" onClick={onClearFiltered} disabled={visibleSelectedCount === 0}>取消当前结果</Button>
        <Button size="sm" variant="ghost" onClick={() => onChange(EMPTY_PROFILE_FILTERS)} disabled={!hasFilter}>清空筛选</Button>
      </div>
    </div>
  )
}

function ProfileSelectionGrid({ profiles, selectedIds, onToggle }: ProfileSelectionGridProps) {
  if (profiles.length === 0) {
    return <div className="col-span-full py-8 text-center text-sm text-[var(--color-text-muted)]">没有匹配的执行环境实例</div>
  }

  return profiles.map(profile => (
    <label key={profile.profileId} className="flex items-start gap-2 rounded-lg border border-transparent px-2 py-2 text-sm text-[var(--color-text-secondary)] hover:border-[var(--color-border-default)] hover:bg-[var(--color-bg-secondary)]/40">
      <input type="checkbox" checked={selectedIds.has(profile.profileId)} onChange={() => onToggle(profile.profileId)} className="mt-1" />
      <div className="min-w-0 space-y-1">
        <div className="break-all text-sm text-[var(--color-text-primary)]">{profile.profileName}</div>
        <div className="break-all text-xs text-[var(--color-text-muted)]">{profile.profileId}</div>
        {profile.tags?.length ? (
          <div className="flex flex-wrap gap-1">
            {profile.tags.map(tag => (
              <span key={`${profile.profileId}-${tag}`} className="rounded-full bg-[var(--color-accent)]/10 px-2 py-0.5 text-[11px] text-[var(--color-accent)]">
                {tag}
              </span>
            ))}
          </div>
        ) : null}
      </div>
    </label>
  ))
}

export function TaskProfileSelector({ profiles, selectedProfileIds, resetKey, onChange }: TaskProfileSelectorProps) {
  const [filters, setFilters] = useState<TaskProfileFilters>(EMPTY_PROFILE_FILTERS)

  useEffect(() => {
    setFilters(EMPTY_PROFILE_FILTERS)
  }, [resetKey])

  const allTags = useMemo(() => collectTaskProfileTags(profiles), [profiles])
  const filteredProfiles = useMemo(() => filterTaskProfiles(profiles, filters), [profiles, filters])
  const selectedIds = useMemo(() => new Set(selectedProfileIds), [selectedProfileIds])
  const visibleSelectedCount = useMemo(
    () => filteredProfiles.filter(profile => selectedIds.has(profile.profileId)).length,
    [filteredProfiles, selectedIds],
  )
  const hiddenSelectedCount = selectedProfileIds.length - visibleSelectedCount

  return (
    <div className="space-y-3">
      <div className="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-secondary)]/40 p-3 space-y-3">
        <ProfileFilterToolbar
          filters={filters}
          filteredCount={filteredProfiles.length}
          visibleSelectedCount={visibleSelectedCount}
          onChange={setFilters}
          onSelectFiltered={() => onChange(appendMissingIds(selectedProfileIds, filteredProfiles.map(profile => profile.profileId)))}
          onClearFiltered={() => onChange(removeIds(selectedProfileIds, filteredProfiles.map(profile => profile.profileId)))}
        />
        <TagFilterBar tags={allTags} selected={new Set(filters.tags)} onChange={tags => setFilters(prev => ({ ...prev, tags: Array.from(tags) }))} />
        <p className="text-xs text-[var(--color-text-muted)]">
          共 {profiles.length} 个实例，当前显示 {filteredProfiles.length} 个，已选 {selectedProfileIds.length} 个。
          {hiddenSelectedCount > 0 ? ` 其中 ${hiddenSelectedCount} 个已选实例被当前筛选隐藏。` : ''}
        </p>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-2 max-h-[260px] overflow-auto border border-[var(--color-border-default)] rounded-lg p-3">
        <ProfileSelectionGrid
          profiles={filteredProfiles}
          selectedIds={selectedIds}
          onToggle={profileId => onChange(selectedIds.has(profileId) ? removeIds(selectedProfileIds, [profileId]) : [...selectedProfileIds, profileId])}
        />
      </div>
    </div>
  )
}
