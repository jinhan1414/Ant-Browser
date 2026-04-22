function Remove-BuildExecutableOnly {
    param(
        [Parameter(Mandatory = $true)]
        [string]$RepoRoot,
        [string]$RelativeBinaryPath = "build/bin/ant-chrome.exe"
    )

    $binaryPath = Join-Path $RepoRoot $RelativeBinaryPath
    $binaryDir = Split-Path -Parent $binaryPath
    if (-not (Test-Path -LiteralPath $binaryDir -PathType Container)) {
        New-Item -ItemType Directory -Path $binaryDir -Force | Out-Null
        return
    }

    if (-not (Test-Path -LiteralPath $binaryPath -PathType Leaf)) {
        return
    }

    Write-Host "预清理旧二进制: 仅删除 $RelativeBinaryPath"
    Remove-Item -LiteralPath $binaryPath -Force
}
