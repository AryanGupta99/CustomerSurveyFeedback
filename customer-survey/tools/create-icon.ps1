Param(
    [string]$PngFile = "icon.png",
    [string]$IcoFile = "icon.ico"
)

if (-not (Test-Path $PngFile)) { Write-Error "PNG file not found: $PngFile"; exit 1 }

Add-Type -AssemblyName System.Drawing | Out-Null

$src = [System.Drawing.Image]::FromFile((Resolve-Path $PngFile).Path)
$sizes = @(16, 24, 32, 48, 64, 128, 256)
[System.Collections.ArrayList]$images = @()

foreach ($s in $sizes) {
    $bmp = New-Object System.Drawing.Bitmap($s, $s)
    $g = [System.Drawing.Graphics]::FromImage($bmp)
    $g.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
    $g.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
    $g.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
    $g.DrawImage($src, 0, 0, $s, $s)
    $g.Dispose()
    $ms = New-Object System.IO.MemoryStream
    $bmp.Save($ms, [System.Drawing.Imaging.ImageFormat]::Png)
    $images.Add($ms.ToArray()) | Out-Null
    $ms.Dispose(); $bmp.Dispose()
}
$src.Dispose()

# Build ICO with all sizes
$ico = New-Object System.IO.MemoryStream
$bw = New-Object System.IO.BinaryWriter($ico)

$bw.Write([UInt16]0)  # reserved
$bw.Write([UInt16]1)  # type icon
$bw.Write([UInt16]$images.Count) # count

$offset = 6 + (16 * $images.Count)
foreach ($bytes in $images) {
    $img = [System.Drawing.Image]::FromStream([System.IO.MemoryStream]::new($bytes))
    $w = [byte]([Math]::Min($img.Width,255))
    $h = [byte]([Math]::Min($img.Height,255))
    $img.Dispose()
    $bw.Write([byte]$w)
    $bw.Write([byte]$h)
    $bw.Write([byte]0)     # colors
    $bw.Write([byte]0)     # reserved
    $bw.Write([UInt16]1)   # planes
    $bw.Write([UInt16]32)  # bit count
    $bw.Write([UInt32]$bytes.Length)
    $bw.Write([UInt32]$offset)
    $offset += $bytes.Length
}
foreach ($bytes in $images) { $bw.Write($bytes) }
$bw.Flush()
[System.IO.File]::WriteAllBytes((Join-Path $PWD $IcoFile), $ico.ToArray())
$bw.Dispose(); $ico.Dispose()

Write-Output "Created $IcoFile with sizes: $($sizes -join ', ') - Size: $(([System.IO.FileInfo]$IcoFile).Length) bytes"