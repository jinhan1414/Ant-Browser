const GITHUB_RELEASES_PAGE_SIZE = 30
const GITHUB_RELEASES_API_URL = `https://api.github.com/repos/adryfish/fingerprint-chromium/releases?per_page=${GITHUB_RELEASES_PAGE_SIZE}`

interface GithubReleaseAsset {
  name?: string
  browser_download_url?: string
}

interface GithubReleasePayload {
  tag_name?: string
  name?: string
  draft?: boolean
  published_at?: string
  assets?: GithubReleaseAsset[]
}

export interface CoreReleaseOption {
  value: string
  label: string
  version: string
  assetName: string
  url: string
  suggestedCoreName: string
  publishedAt: string
}

interface GithubErrorPayload {
  message?: string
}

function isZipAsset(asset: GithubReleaseAsset): asset is Required<GithubReleaseAsset> {
  return Boolean(asset.name && asset.browser_download_url && /\.zip$/i.test(asset.name))
}

function normalizeReleaseVersion(release: GithubReleasePayload): string {
  const rawVersion = String(release.tag_name || release.name || '').trim()
  const normalizedVersion = rawVersion.replace(/^v/i, '')
  if (!normalizedVersion) {
    throw new Error('GitHub release 缺少可识别的版本号')
  }
  return normalizedVersion
}

function buildReleaseLabel(version: string, assetName: string, assetCount: number): string {
  if (assetCount <= 1) return version
  return `${version} (${assetName})`
}

function buildGithubHttpError(status: number, payload: unknown): Error {
  const errorMessage = typeof payload === 'object' && payload !== null
    ? String((payload as GithubErrorPayload).message || '')
    : ''
  if (errorMessage) {
    return new Error(`GitHub Releases 请求失败（HTTP ${status}）：${errorMessage}`)
  }
  return new Error(`GitHub Releases 请求失败（HTTP ${status}）`)
}

export function buildSuggestedCoreName(version: string): string {
  const majorVersion = version.split('.')[0]?.trim()
  if (!majorVersion || !/^\d+$/.test(majorVersion)) {
    throw new Error(`无法从版本号生成内核名称：${version}`)
  }
  return `chrome-${majorVersion}`
}

export function mapGithubReleasesToCoreOptions(releases: GithubReleasePayload[]): CoreReleaseOption[] {
  const options: CoreReleaseOption[] = []

  for (const release of releases) {
    if (release.draft) continue

    const version = normalizeReleaseVersion(release)
    const zipAssets = Array.isArray(release.assets)
      ? release.assets.filter(isZipAsset)
      : []
    if (zipAssets.length === 0) continue

    for (const asset of zipAssets) {
      options.push({
        value: `${version}::${asset.name}`,
        label: buildReleaseLabel(version, asset.name, zipAssets.length),
        version,
        assetName: asset.name,
        url: asset.browser_download_url,
        suggestedCoreName: buildSuggestedCoreName(version),
        publishedAt: String(release.published_at || ''),
      })
    }
  }

  if (options.length === 0) {
    throw new Error('fingerprint-chromium Releases 中未找到可下载的 ZIP 版本')
  }

  return options
}

export async function fetchFingerprintChromiumReleaseOptions(signal?: AbortSignal): Promise<CoreReleaseOption[]> {
  const response = await fetch(GITHUB_RELEASES_API_URL, {
    headers: {
      Accept: 'application/vnd.github+json',
    },
    signal,
  })

  let payload: unknown = null
  try {
    payload = await response.json()
  } catch {
    payload = null
  }

  if (!response.ok) {
    throw buildGithubHttpError(response.status, payload)
  }

  if (!Array.isArray(payload)) {
    throw new Error('GitHub Releases 返回格式异常')
  }

  return mapGithubReleasesToCoreOptions(payload as GithubReleasePayload[])
}
