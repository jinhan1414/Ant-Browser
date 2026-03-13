<#
.SYNOPSIS
Publish a sanitized snapshot to public release/dev branches.

.DESCRIPTION
- Default mode: publish only release/<version> from master using a safe bot identity.
- Optional dev mode: append one snapshot commit per publish to the public dev branch.
- Public release/<version> branch: always a single-commit branch (orphan).
- Private development history is never pushed directly.

.EXAMPLE
.\bat\publish-public.ps1

.EXAMPLE
.\bat\publish-public.ps1 -Version 1.0.0 -PublicRemote github

.EXAMPLE
.\bat\publish-public.ps1 -Version 1.0.1 -PublicRemote https://github.com/org/repo.git -PublishDev -DryRun

.EXAMPLE
.\bat\publish-public.ps1 -Version 1.0.0 -PublicRemote github -ForceOverwriteRelease
#>
param(
    [string]$Version,
    [string]$PublicRemote,
    [string]$SourceRef = "master",
    [string]$DevBranch = "dev",
    [string]$ReleasePrefix = "release/",

    [switch]$PublishDev,
    [switch]$ForceOverwriteRelease,
    [switch]$AllowDirtyWorkingTree,
    [switch]$SkipRuntimeCheck,
    [switch]$DryRun,
    [switch]$KeepTempDir,
    [string]$CommitterName,
    [string]$CommitterEmail,
    [switch]$IncludeSourceCommit
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
if (Get-Variable -Name PSNativeCommandUseErrorActionPreference -ErrorAction SilentlyContinue) {
    $PSNativeCommandUseErrorActionPreference = $false
}

$defaultCommitterName = "Ant Browser Release Bot"
$defaultCommitterEmail = "release-bot@ant-browser.local"

function Get-TrimmedText {
    param([AllowNull()][string]$Value)

    if ($null -eq $Value) {
        return ""
    }
    return $Value.Trim()
}

function Write-Step {
    param([string]$Message)
    Write-Host ""
    Write-Host "==> $Message" -ForegroundColor Cyan
}

function Invoke-Git {
    param(
        [Parameter(Mandatory = $true)]
        [string[]]$Args,
        [switch]$AllowFailure
    )

    $tempToken = [guid]::NewGuid().ToString("N")
    $stdoutPath = Join-Path $env:TEMP "git-stdout-$tempToken.log"
    $stderrPath = Join-Path $env:TEMP "git-stderr-$tempToken.log"
    $previousErrorActionPreference = $ErrorActionPreference

    try {
        $ErrorActionPreference = "Continue"
        & git @Args 1> $stdoutPath 2> $stderrPath
        $code = $LASTEXITCODE
        $output = @()
        if (Test-Path -LiteralPath $stdoutPath) {
            $output += Get-Content -LiteralPath $stdoutPath
        }
        if (Test-Path -LiteralPath $stderrPath) {
            $output += Get-Content -LiteralPath $stderrPath
        }
    }
    finally {
        $ErrorActionPreference = $previousErrorActionPreference
        Remove-Item -LiteralPath $stdoutPath -Force -ErrorAction SilentlyContinue
        Remove-Item -LiteralPath $stderrPath -Force -ErrorAction SilentlyContinue
    }

    if ($code -ne 0 -and -not $AllowFailure) {
        $argText = $Args -join " "
        $outText = $output -join [Environment]::NewLine
        throw "git $argText failed with exit code $code.`n$outText"
    }
    return @{
        Code   = $code
        Output = $output
    }
}

function Get-FirstOutputLine {
    param([string[]]$Lines)
    if (($Lines | Measure-Object).Count -eq 0) {
        return ""
    }
    return $Lines[0].Trim()
}

function Resolve-PublicRemoteUrl {
    param(
        [Parameter(Mandatory = $true)]
        [string]$RemoteOrUrl
    )

    $knownRemotes = (Invoke-Git -Args @("remote")).Output
    foreach ($item in $knownRemotes) {
        if ($item.Trim() -eq $RemoteOrUrl) {
            return Get-FirstOutputLine -Lines (Invoke-Git -Args @("remote", "get-url", $RemoteOrUrl)).Output
        }
    }
    return $RemoteOrUrl
}

function Resolve-DefaultPublicRemote {
    $knownRemotes = @((Invoke-Git -Args @("remote")).Output | ForEach-Object { $_.Trim() } | Where-Object { $_ -ne "" })
    foreach ($candidate in @("github", "public")) {
        if ($knownRemotes -contains $candidate) {
            return $candidate
        }
    }
    throw "PublicRemote was not provided and no default public remote (github/public) was found. Pass -PublicRemote explicitly."
}

function Resolve-VersionValue {
    param([string]$ExplicitVersion)

    $explicit = Get-TrimmedText $ExplicitVersion
    if ($explicit -ne "") {
        return $explicit
    }

    $wailsConfigPath = Join-Path $repoRoot "wails.json"
    if (-not (Test-Path -LiteralPath $wailsConfigPath)) {
        throw "wails.json was not found and -Version was not provided."
    }

    $wailsConfig = Get-Content -LiteralPath $wailsConfigPath -Raw | ConvertFrom-Json
    $resolvedVersion = Get-TrimmedText ([string]$wailsConfig.info.productVersion)
    if ($resolvedVersion -eq "") {
        throw "Could not resolve productVersion from wails.json. Pass -Version explicitly."
    }
    return $resolvedVersion
}

function Resolve-CommitterIdentity {
    param(
        [string]$ExplicitName,
        [string]$ExplicitEmail
    )

    $resolvedName = Get-TrimmedText $ExplicitName
    if ($resolvedName -eq "") {
        $resolvedName = Get-TrimmedText ([string]$env:PUBLISH_COMMITTER_NAME)
    }
    if ($resolvedName -eq "") {
        $resolvedName = $defaultCommitterName
    }

    $resolvedEmail = Get-TrimmedText $ExplicitEmail
    if ($resolvedEmail -eq "") {
        $resolvedEmail = Get-TrimmedText ([string]$env:PUBLISH_COMMITTER_EMAIL)
    }
    if ($resolvedEmail -eq "") {
        $resolvedEmail = $defaultCommitterEmail
    }
    if ($resolvedEmail -notmatch "^[^@\s]+@[^@\s]+$") {
        throw "Invalid publish committer email: $resolvedEmail"
    }

    return @{
        Name  = $resolvedName
        Email = $resolvedEmail
    }
}

function Assert-CleanTrackedWorkingTree {
    $unstaged = Invoke-Git -Args @("diff", "--quiet", "--ignore-submodules", "--") -AllowFailure
    if ($unstaged.Code -gt 1) {
        throw "Unable to inspect unstaged tracked-file changes."
    }
    if ($unstaged.Code -eq 1) {
        throw "Tracked files have unstaged changes. Commit or stash them first, or use -AllowDirtyWorkingTree."
    }

    $staged = Invoke-Git -Args @("diff", "--cached", "--quiet", "--ignore-submodules", "--") -AllowFailure
    if ($staged.Code -gt 1) {
        throw "Unable to inspect staged tracked-file changes."
    }
    if ($staged.Code -eq 1) {
        throw "Tracked files have staged but uncommitted changes. Commit them first, or use -AllowDirtyWorkingTree."
    }
}

function Test-RemoteBranchExists {
    param(
        [Parameter(Mandatory = $true)]
        [string]$RemoteUrl,
        [Parameter(Mandatory = $true)]
        [string]$BranchName
    )

    $result = Invoke-Git -Args @("ls-remote", "--heads", $RemoteUrl, "refs/heads/$BranchName")
    if ($result.Code -ne 0) {
        throw "Unable to query remote branch refs/heads/$BranchName from $RemoteUrl."
    }
    return ($result.Output | Measure-Object).Count -gt 0
}

function Sync-SnapshotToRepo {
    param(
        [Parameter(Mandatory = $true)]
        [string]$SnapshotDir,
        [Parameter(Mandatory = $true)]
        [string]$RepoDir
    )

    Get-ChildItem -LiteralPath $RepoDir -Force |
        Where-Object { $_.Name -ne ".git" } |
        Remove-Item -Recurse -Force

    Get-ChildItem -LiteralPath $SnapshotDir -Force | ForEach-Object {
        $target = Join-Path $RepoDir $_.Name
        Copy-Item -LiteralPath $_.FullName -Destination $target -Recurse -Force
    }
}

function Build-CommitMessage {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Title,
        [Parameter(Mandatory = $true)]
        [string]$Channel,
        [Parameter(Mandatory = $true)]
        [string]$VersionValue,
        [Parameter(Mandatory = $true)]
        [string]$SourceRefValue,
        [string]$SourceCommit,
        [switch]$AppendSourceCommit
    )

    $publishedAtUtc = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    $lines = @(
        $Title,
        "",
        "channel: $Channel",
        "version: $VersionValue",
        "source-ref: $SourceRefValue",
        "published-at-utc: $publishedAtUtc"
    )
    if ($AppendSourceCommit -and ((Get-TrimmedText $SourceCommit) -ne "")) {
        $lines += "source-commit: $SourceCommit"
    }
    return $lines -join "`n"
}

function Invoke-GitCommit {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Message,
        [switch]$AllowEmpty
    )

    $messagePath = Join-Path $env:TEMP ("ant-chrome-commit-message-" + [guid]::NewGuid().ToString("N") + ".txt")
    try {
        $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
        [System.IO.File]::WriteAllText($messagePath, $Message, $utf8NoBom)

        $commitArgs = @("commit")
        if ($AllowEmpty) {
            $commitArgs += "--allow-empty"
        }
        $commitArgs += @("-F", $messagePath)

        Invoke-Git -Args $commitArgs | Out-Null
    }
    finally {
        Remove-Item -LiteralPath $messagePath -Force -ErrorAction SilentlyContinue
    }
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Set-Location $repoRoot

Write-Step "Validating repository state"
Invoke-Git -Args @("rev-parse", "--is-inside-work-tree") | Out-Null

if (-not $AllowDirtyWorkingTree) {
    Assert-CleanTrackedWorkingTree
}

$Version = Resolve-VersionValue -ExplicitVersion $Version
if ((Get-TrimmedText $PublicRemote) -eq "") {
    $PublicRemote = Resolve-DefaultPublicRemote
}
$committer = Resolve-CommitterIdentity -ExplicitName $CommitterName -ExplicitEmail $CommitterEmail

$sourceCommit = Get-FirstOutputLine -Lines (Invoke-Git -Args @("rev-parse", "--verify", "$SourceRef`^{commit}")).Output
$sourceShort = Get-FirstOutputLine -Lines (Invoke-Git -Args @("rev-parse", "--short", $sourceCommit)).Output
$releaseBranch = "$ReleasePrefix$Version"
$publicUrl = Resolve-PublicRemoteUrl -RemoteOrUrl $PublicRemote

Write-Host "Source ref: $SourceRef -> $sourceCommit"
Write-Host "Public remote: $publicUrl"
if ($PublishDev) {
    Write-Host "Dev branch: $DevBranch"
} else {
    Write-Host "Dev branch: skipped"
}
Write-Host "Release branch: $releaseBranch"
Write-Host "Committer: $($committer.Name) <$($committer.Email)>"
if ($DryRun) {
    Write-Host "Dry-run: enabled (no push will be performed)"
}

Write-Step "Checking remote branch existence"
$devExists = $false
if ($PublishDev) {
    $devExists = Test-RemoteBranchExists -RemoteUrl $publicUrl -BranchName $DevBranch
}
$releaseExists = Test-RemoteBranchExists -RemoteUrl $publicUrl -BranchName $releaseBranch

if ($releaseExists -and -not $ForceOverwriteRelease) {
    throw "Remote release branch $releaseBranch already exists. Use -ForceOverwriteRelease to replace it."
}

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$tempRoot = Join-Path $env:TEMP "ant-chrome-public-$Version-$timestamp"
$snapshotDir = Join-Path $tempRoot "snapshot"
$workRepoDir = Join-Path $tempRoot "workrepo"
$archivePath = Join-Path $tempRoot "snapshot.zip"

New-Item -ItemType Directory -Path $snapshotDir -Force | Out-Null
New-Item -ItemType Directory -Path $workRepoDir -Force | Out-Null

try {
    Write-Step "Exporting source snapshot"
    Invoke-Git -Args @("archive", "--format=zip", "-o", $archivePath, $sourceCommit) | Out-Null
    Expand-Archive -LiteralPath $archivePath -DestinationPath $snapshotDir -Force

    if (-not $SkipRuntimeCheck) {
        Write-Step "Checking required runtime files in snapshot"
        $requiredFiles = @(
            "bin/xray.exe",
            "bin/sing-box.exe"
        )
        foreach ($file in $requiredFiles) {
            $path = Join-Path $snapshotDir $file
            if (-not (Test-Path -LiteralPath $path)) {
                throw "Required runtime file missing in source snapshot: $file. Add it before publishing, or use -SkipRuntimeCheck."
            }
        }
    }

    Write-Step "Preparing temporary publish repository"
    Push-Location $workRepoDir
    try {
        Invoke-Git -Args @("init") | Out-Null
        Invoke-Git -Args @("remote", "add", "public", $publicUrl) | Out-Null
        Invoke-Git -Args @("config", "user.name", $committer.Name) | Out-Null
        Invoke-Git -Args @("config", "user.email", $committer.Email) | Out-Null

        if ($PublishDev) {
            if ($devExists) {
                Invoke-Git -Args @("fetch", "--no-tags", "public", "refs/heads/${DevBranch}:refs/remotes/public/$DevBranch") | Out-Null
                Invoke-Git -Args @("checkout", "-B", $DevBranch, "refs/remotes/public/$DevBranch") | Out-Null
            } else {
                Invoke-Git -Args @("checkout", "--orphan", $DevBranch) | Out-Null
            }
        }
        if ($releaseExists) {
            Invoke-Git -Args @("fetch", "--no-tags", "public", "refs/heads/${releaseBranch}:refs/remotes/public/$releaseBranch") | Out-Null
        }

        if ($PublishDev) {
            Write-Step "Publishing snapshot to $DevBranch"
            Sync-SnapshotToRepo -SnapshotDir $snapshotDir -RepoDir $workRepoDir
            Invoke-Git -Args @("add", "-A") | Out-Null
            $devMessage = Build-CommitMessage `
                -Title "publish: $Version snapshot ($sourceShort)" `
                -Channel "dev" `
                -VersionValue $Version `
                -SourceRefValue $SourceRef `
                -SourceCommit $sourceCommit `
                -AppendSourceCommit:$IncludeSourceCommit
            Invoke-GitCommit -Message $devMessage -AllowEmpty
            if ($DryRun) {
                Write-Host "DRY-RUN: skip push -> public $DevBranch"
            } else {
                Invoke-Git -Args @("push", "public", $DevBranch) | Out-Null
            }
        }

        Write-Step "Publishing snapshot to $releaseBranch (single-commit branch)"
        Invoke-Git -Args @("checkout", "--orphan", $releaseBranch) | Out-Null
        Sync-SnapshotToRepo -SnapshotDir $snapshotDir -RepoDir $workRepoDir
        Invoke-Git -Args @("add", "-A") | Out-Null
        $releaseMessage = Build-CommitMessage `
            -Title "release: $Version snapshot ($sourceShort)" `
            -Channel "release" `
            -VersionValue $Version `
            -SourceRefValue $SourceRef `
            -SourceCommit $sourceCommit `
            -AppendSourceCommit:$IncludeSourceCommit
        Invoke-GitCommit -Message $releaseMessage

        $releaseCount = Get-FirstOutputLine -Lines (Invoke-Git -Args @("rev-list", "--count", $releaseBranch)).Output
        if ($releaseCount -ne "1") {
            throw "Local release branch $releaseBranch is expected to have exactly 1 commit, got $releaseCount."
        }

        if ($releaseExists) {
            if ($DryRun) {
                Write-Host "DRY-RUN: skip push -> public $releaseBranch --force-with-lease"
            } else {
                Invoke-Git -Args @("push", "--force-with-lease", "public", $releaseBranch) | Out-Null
            }
        } else {
            if ($DryRun) {
                Write-Host "DRY-RUN: skip push -> public $releaseBranch"
            } else {
                Invoke-Git -Args @("push", "public", $releaseBranch) | Out-Null
            }
        }
    }
    finally {
        Pop-Location
    }

    Write-Step "Publish completed"
    Write-Host "Published source commit: $sourceCommit"
    if ($PublishDev) {
        Write-Host "Updated dev branch: $DevBranch (appends one snapshot commit per publish)"
    }
    Write-Host "Updated release branch: $releaseBranch (always single commit)"
}
finally {
    if ($KeepTempDir) {
        Write-Host ""
        Write-Host "Temporary directory kept: $tempRoot"
    } else {
        Remove-Item -LiteralPath $tempRoot -Recurse -Force -ErrorAction SilentlyContinue
    }
}
