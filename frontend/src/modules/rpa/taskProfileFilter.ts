export interface TaskProfileFilterCandidate {
  profileId: string
  profileName: string
  tags?: string[] | null
}

export interface TaskProfileFilters {
  keyword: string
  tags: string[]
}

function normalizeKeyword(value: string): string {
  return value.trim().toLowerCase()
}

function normalizeTags(tags: readonly string[]): string[] {
  return Array.from(
    new Set(tags.map(tag => tag.trim()).filter(Boolean)),
  )
}

function matchesKeyword(profile: TaskProfileFilterCandidate, keyword: string): boolean {
  if (!keyword) {
    return true
  }
  return profile.profileName.toLowerCase().includes(keyword)
}

function matchesTags(profile: TaskProfileFilterCandidate, tags: readonly string[]): boolean {
  if (tags.length === 0) {
    return true
  }
  return (profile.tags || []).some(tag => tags.includes(tag))
}

export function filterTaskProfiles<T extends TaskProfileFilterCandidate>(
  profiles: readonly T[],
  filters: TaskProfileFilters,
): T[] {
  const keyword = normalizeKeyword(filters.keyword)
  const selectedTags = normalizeTags(filters.tags)

  return profiles.filter(profile => {
    if (!matchesKeyword(profile, keyword)) {
      return false
    }
    return matchesTags(profile, selectedTags)
  })
}

export function collectTaskProfileTags<T extends TaskProfileFilterCandidate>(profiles: readonly T[]): string[] {
  const tags = new Set<string>()
  profiles.forEach(profile => {
    ;(profile.tags || []).forEach(tag => {
      const normalized = tag.trim()
      if (normalized) {
        tags.add(normalized)
      }
    })
  })
  return Array.from(tags).sort()
}
