# package-release.ps1
# Build Windows GUI exe and create a releases ZIP for distribution.
# Usage: .\package-release.ps1

$ErrorActionPreference = 'Stop'

$pwdRoot = (Get-Location).Path
$exeName = 'customer-survey.exe'
$releasesDir = Join-Path $pwdRoot 'releases'
if (-Not (Test-Path $releasesDir)) { New-Item -ItemType Directory -Path $releasesDir | Out-Null }

# Determine version from version.json if present
$version = 'v0.0.0'
$versionFile = Join-Path $pwdRoot 'version.json'
if (Test-Path $versionFile) {
    try {
        $v = Get-Content $versionFile -Raw | ConvertFrom-Json
        # Prefer StringFileInfo.FileVersion if present (e.g. "1.0.1.0")
        if ($v.StringFileInfo -and $v.StringFileInfo.FileVersion) {
            $verRaw = $v.StringFileInfo.FileVersion -as [string]
            # Normalize to semantic '1.0.1' by dropping trailing .0 if present
            $parts = $verRaw -split '\.'
            if ($parts.Length -ge 3) { $ver = "$($parts[0]).$($parts[1]).$($parts[2])" } else { $ver = $verRaw }
            $version = "v$ver"
        } elseif ($v.FixedFileInfo -and $v.FixedFileInfo.FileVersion) {
            $fv = $v.FixedFileInfo.FileVersion
            if ($fv.Major -ne $null) { $version = "v$($fv.Major).$($fv.Minor).$($fv.Patch)" }
        }
    } catch { }
}

# Build binary (Windows GUI mode)
Write-Host "Building $exeName (windowsgui)" -ForegroundColor Cyan
# Attempt to remove existing exe; ignore errors if in use
if (Test-Path $exeName) {
    try { Remove-Item $exeName -Force -ErrorAction Stop } catch { Write-Host "Warning: couldn't remove existing $exeName; continuing" -ForegroundColor Yellow }
}

go build -ldflags='-H windowsgui' -o $exeName .\cmd\survey\main.go
if ($LASTEXITCODE -ne 0) { throw "go build failed" }

# Files to include
$include = @(
    $exeName,
    'config.json',
    'README.md',
    'version.json'
)

# Create a small client README to include in the package
$clientReadme = "README_CLIENT.txt"
$clientContent = @"
Customer Survey - Client Package

1) Extract all files into a folder (e.g. C:\CustomerSurvey) on the client machine.
2) Edit config.json if you need to change webhook target. By default it uses the packaged config.json.
3) Ensure Microsoft Edge or Google Chrome is installed (required for automated hidden submissions).
4) Run customer-survey.exe (double-click). The app runs as a GUI app and starts a local server.
5) To trigger a submission for testing, POST to http://localhost:8080/submit or use the UI.

Log files are stored at %APPDATA%\\.customer-survey\\webhook.log for troubleshooting.
"@
Set-Content -Path $clientReadme -Value $clientContent -Encoding UTF8
if (Test-Path $clientReadme) { $include += $clientReadme }

# Copy to temp staging dir
$staging = Join-Path $pwdRoot ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $staging | Out-Null

foreach ($f in $include) {
    if (Test-Path $f) { Copy-Item -Path $f -Destination $staging -Recurse -Force }
}

# Copy assets and required folders (if exist)
$folders = @('assets','internal','ui','pkg')
foreach ($d in $folders) {
    if (Test-Path $d) { Copy-Item -Path $d -Destination $staging -Recurse -Force }
}

# Create zip
$zipName = "customer-survey-$version.zip"
$zipPath = Join-Path $releasesDir $zipName
if (Test-Path $zipPath) { Remove-Item $zipPath -Force }

Add-Type -AssemblyName System.IO.Compression.FileSystem
[System.IO.Compression.ZipFile]::CreateFromDirectory($staging, $zipPath)

# Clean up staging
Remove-Item $staging -Recurse -Force

Write-Host "Created release: $zipPath" -ForegroundColor Green
Write-Host "Files included: $($include -join ', ') and existing folders: $((Get-ChildItem -Path $releasesDir | Where-Object Name -like 'customer-survey-*').Name)" -ForegroundColor Gray

Write-Host "Done." -ForegroundColor Green
