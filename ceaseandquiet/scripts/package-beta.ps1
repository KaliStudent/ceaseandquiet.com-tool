param([Parameter(Mandatory=$true)][string]$OutputDirectory)
$ErrorActionPreference = 'Stop'
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
if (-not [IO.Path]::IsPathRooted($OutputDirectory)) { throw 'Use an absolute output directory.' }
Push-Location $repoRoot
try {
    if (git status --porcelain) { throw 'Commit the reviewed source before packaging.' }
    $revision = (git rev-parse HEAD).Trim()
    $extensionVersion = (Get-Content -Raw -LiteralPath (Join-Path $repoRoot 'ceaseandquiet/extension/manifest.json') | ConvertFrom-Json).version
    if ($extensionVersion -notmatch '^\d+\.\d+\.\d+$') { throw 'Invalid extension version.' }
    $packageName = 'ceaseandquiet-beta-' + $extensionVersion + '-' + $revision.Substring(0,7)
    $packageRoot = Join-Path $OutputDirectory $packageName
    if (Test-Path -LiteralPath $packageRoot) { throw 'This package already exists. Choose a different output directory.' }
    New-Item -ItemType Directory -Path $packageRoot | Out-Null
    $extensionDirectory = New-Item -ItemType Directory -Path (Join-Path $packageRoot 'extension')
    $previewDirectory = New-Item -ItemType Directory -Path (Join-Path $packageRoot 'preview')
    foreach ($file in @('manifest.json','popup.html','popup.css','popup.js','vault.mjs','fill.mjs','address.mjs')) {
        Copy-Item -LiteralPath (Join-Path $repoRoot ('ceaseandquiet/extension/' + $file)) -Destination $extensionDirectory.FullName
    }
    foreach ($file in @('README.md','SECURITY-REVIEW.md','BETA-CHECKLIST.md','STAGING-CHECKLIST.md','ADDRESS-AUTOCOMPLETE.md','NATIVE-CHROME-RESULTS.md','ROLLOUT.md')) {
        Copy-Item -LiteralPath (Join-Path $repoRoot ('ceaseandquiet/' + $file)) -Destination $packageRoot
    }
    Copy-Item -LiteralPath (Join-Path $repoRoot 'LICENSE') -Destination $packageRoot
    & go build -trimpath -ldflags='-s -w' -o (Join-Path $previewDirectory.FullName 'cq-preview.exe') ./cmd/cq-preview
    if ($LASTEXITCODE -ne 0) { throw 'Preview build failed.' }
    @{ version=$extensionVersion; sourceCommit=$revision; workingTree='clean'; target='windows-beta'; nativeChromeSmoke='pending-for-this-build'; earlierNativeEvidence='NATIVE-CHROME-RESULTS.md'; nativeChromeChecklist='pending'; stagingOrigin='https://abcw.xyz:8443'; stagingChromeCheck='pending'; finalReview='pending' } | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $packageRoot 'package-manifest.json') -Encoding utf8
    $zipPath = Join-Path $OutputDirectory ($packageName + '.zip')
    if (Test-Path -LiteralPath $zipPath) { throw 'ZIP already exists.' }
    Compress-Archive -LiteralPath $packageRoot -DestinationPath $zipPath
    $hash = Get-FileHash -LiteralPath $zipPath -Algorithm SHA256
    ($hash.Hash.ToLowerInvariant() + '  ' + [IO.Path]::GetFileName($zipPath)) | Set-Content -LiteralPath ($zipPath + '.sha256') -Encoding ascii
    Write-Output $zipPath
} finally { Pop-Location }
