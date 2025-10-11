Param(
  [string]$IconPath = "assets/icon.ico",
  [string]$Out = "survey-gui.exe",
  [string]$Webhook = "",
  [switch]$RegenIcon,
  [switch]$UseGoversioninfo
)

# Ensure Go in PATH for this session
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
  $env:Path = 'C:\\Program Files\\Go\\bin;' + $env:Path
}

# If no webhook provided, try configs/config.json
if (-not $Webhook -or $Webhook.Trim() -eq '') {
  $cfg = Join-Path $PSScriptRoot 'configs/config.json'
  if (Test-Path $cfg) {
    try {
      $json = Get-Content $cfg -Raw | ConvertFrom-Json
      if ($json.webhook_url) { $Webhook = $json.webhook_url }
    } catch {}
  }
}

$ld = "-s -w"
if ($Webhook -and $Webhook.Trim() -ne '') {
  $ld = "$ld -X 'customer-survey/internal/survey.DefaultWebhookURL=$Webhook'"
}

# Helper: convert PNG to ICO by wrapping PNG in ICO container (Vista+ supports PNG-in-ICO)
function Convert-PngToIco {
  param(
    [Parameter(Mandatory=$true)][string]$PngPath,
    [Parameter(Mandatory=$true)][string]$IcoPath
  )
  if (-not (Test-Path $PngPath)) { throw "PNG not found: $PngPath" }
  # Read PNG bytes and dimensions
  [byte[]]$pngBytes = [System.IO.File]::ReadAllBytes($PngPath)
  $width = 0; $height = 0
  try {
    Add-Type -AssemblyName System.Drawing -ErrorAction SilentlyContinue | Out-Null
    $img = [System.Drawing.Image]::FromStream([System.IO.MemoryStream]::new($pngBytes))
    $width = [byte]([Math]::Min($img.Width,255))
    $height = [byte]([Math]::Min($img.Height,255))
    $img.Dispose()
  } catch { $width = 0; $height = 0 }
  if ($width -eq 0 -or $height -eq 0) { $width = 256; $height = 256 }

  $ms = New-Object System.IO.MemoryStream
  $bw = New-Object System.IO.BinaryWriter($ms)
  # ICONDIR
  $bw.Write([UInt16]0)       # reserved
  $bw.Write([UInt16]1)       # type: 1=icon
  $bw.Write([UInt16]1)       # count
  # ICONDIRENTRY
  $bw.Write([byte]$width)    # width (0 means 256)
  $bw.Write([byte]$height)   # height (0 means 256)
  $bw.Write([byte]0)         # color count
  $bw.Write([byte]0)         # reserved
  $bw.Write([UInt16]1)       # planes
  $bw.Write([UInt16]32)      # bit count
  $bw.Write([UInt32]$pngBytes.Length) # bytes in resource
  $bw.Write([UInt32]22)      # offset to image data (6+16)
  $bw.Write($pngBytes)
  $bw.Flush()
  [System.IO.File]::WriteAllBytes($IcoPath, $ms.ToArray())
  $bw.Dispose(); $ms.Dispose()
}

# Resolve icon: support .ico or fall back to root icon.png
$ver = Join-Path $PSScriptRoot 'versioninfo.json'
$rootPng = Join-Path $PSScriptRoot 'icon.png'
if (((-not (Test-Path $IconPath)) -and (Test-Path $rootPng)) -or ($RegenIcon -and (Test-Path $rootPng))) {
  $assetsDir = Join-Path $PSScriptRoot 'assets'
  if (-not (Test-Path $assetsDir)) { New-Item -ItemType Directory -Path $assetsDir | Out-Null }
  $IconPath = Join-Path $assetsDir 'icon.ico'
  try {
    Convert-PngToIco -PngPath $rootPng -IcoPath $IconPath
    Write-Output "Converted icon.png -> $IconPath"
  } catch { Write-Warning "Icon conversion failed: $($_.Exception.Message)" }
}

# If explicitly requested, use goversioninfo (optional)
if ($UseGoversioninfo -and (Test-Path $IconPath) -and (Test-Path $ver)) {
  try {
    if (-not (Get-Command goversioninfo -ErrorAction SilentlyContinue)) {
      Write-Output 'goversioninfo not found. Attempting to install...'
      go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
    }
    if (Get-Command goversioninfo -ErrorAction SilentlyContinue) {
      Push-Location $PSScriptRoot
      $iconFullPath = (Resolve-Path $IconPath).Path
      goversioninfo -icon="$iconFullPath" $ver
      Pop-Location
      Write-Output 'Embedded icon and version via resource.syso (goversioninfo)'
    } else { Write-Warning 'goversioninfo still not available; skipping.' }
  } catch { Write-Output $_.Exception.Message }
}

# Build GUI subsystem (no console window) and trimpath
$ld = "$ld -H=windowsgui"

# Always regenerate resource.syso using rsrc for reliable icon embedding
if (Test-Path $IconPath) {
  Write-Output "Creating resource with rsrc tool..."
  $iconFullPath = (Resolve-Path $IconPath).Path
  Remove-Item "resource.syso" -ErrorAction SilentlyContinue
  & rsrc -ico "$iconFullPath" -o resource.syso
}

Write-Output "Building $Out with ldflags: $ld"
go build -trimpath -ldflags $ld -o $Out ./cmd/survey/main.go

if (Test-Path $Out) {
  Get-Item $Out | Select-Object Name,Length,LastWriteTime | Format-Table -AutoSize | Out-String | Write-Output
} else {
  Write-Error 'Build failed.'
}
