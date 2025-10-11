Param(
  [string]$Out = "assets/icon.ico",
  [string]$Primary = "#0288D1",
  [string]$Secondary = "#00ACC1",
  [string]$GlyphColor = "#FFFFFF"
)

# Generate a professional-looking multi-size ICO (16..256) with a gradient circle and a star glyph
# No external tools or fonts required

Add-Type -AssemblyName System.Drawing | Out-Null
Add-Type -AssemblyName System.Drawing | Out-Null

function New-StarPoints {
  param([float]$cx, [float]$cy, [float]$outerR, [float]$innerR, [int]$points = 5)
  $pts = @()
  for ($i=0; $i -lt $points*2; $i++) {
    $angle = [Math]::PI * $i / $points
    $r = if ($i % 2 -eq 0) { $outerR } else { $innerR }
    $x = $cx + $r * [Math]::Sin($angle)
    $y = $cy - $r * [Math]::Cos($angle)
    $pts += (New-Object System.Drawing.PointF([float]$x, [float]$y))
  }
  return ,$pts
}

function Draw-IconPngBytes {
  param([int]$size, [System.Drawing.Color]$c1, [System.Drawing.Color]$c2, [System.Drawing.Color]$glyph)
  $bmp = New-Object System.Drawing.Bitmap($size, $size, [System.Drawing.Imaging.PixelFormat]::Format32bppArgb)
  $g = [System.Drawing.Graphics]::FromImage($bmp)
  $g.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
  $g.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
  $g.Clear([System.Drawing.Color]::Transparent)

  $margin = [int]([Math]::Round($size * 0.08))
  $rect = New-Object System.Drawing.Rectangle($margin, $margin, $size-2*$margin, $size-2*$margin)

  # Radial-like gradient using PathGradientBrush
  $path = New-Object System.Drawing.Drawing2D.GraphicsPath
  $path.AddEllipse($rect)
  $pbrush = New-Object System.Drawing.Drawing2D.PathGradientBrush($path)
  $pbrush.CenterColor = $c2
  $pbrush.SurroundColors = @($c1)
  $g.FillEllipse($pbrush, $rect)
  $pbrush.Dispose(); $path.Dispose()

  # Subtle inner highlight ring
  $ringPen = New-Object System.Drawing.Pen ([System.Drawing.Color]::FromArgb(80, 255, 255, 255), [float]([Math]::Max(1, $size*0.02)))
  $g.DrawEllipse($ringPen, $rect)
  $ringPen.Dispose()

  # Draw star glyph
  $cx = $size/2.0; $cy = $size/2.0
  $outerR = ($rect.Width)/2.2
  $innerR = $outerR * 0.5
  $points = New-StarPoints -cx $cx -cy $cy -outerR $outerR -innerR $innerR -points 5
  $starPath = New-Object System.Drawing.Drawing2D.GraphicsPath
  $starPath.AddPolygon($points)
  $brush = New-Object System.Drawing.SolidBrush $glyph
  $g.FillPath($brush, $starPath)
  $brush.Dispose(); $starPath.Dispose()

  $ms = New-Object System.IO.MemoryStream
  $bmp.Save($ms, [System.Drawing.Imaging.ImageFormat]::Png)
  $g.Dispose(); $bmp.Dispose()
  return ,$ms.ToArray()
}

function Write-Icon {
  param([byte[][]]$images, [string]$path)
  $dir = Split-Path -Parent $path
  if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir | Out-Null }

  $ms = New-Object System.IO.MemoryStream
  $bw = New-Object System.IO.BinaryWriter($ms)
  # ICONDIR
  $bw.Write([UInt16]0) # reserved
  $bw.Write([UInt16]1) # type icon
  $bw.Write([UInt16]$images.Count) # count

  $headerSize = 6 + 16 * $images.Count
  $offset = [UInt32]$headerSize

  # Prepare entries
  foreach ($bytes in $images) {
    # width/height bytes: 0 means 256
    $sizeImg = [System.Drawing.Image]::FromStream([System.IO.MemoryStream]::new($bytes))
    $w = [byte]([Math]::Min($sizeImg.Width,255))
    $h = [byte]([Math]::Min($sizeImg.Height,255))
    $sizeImg.Dispose()

    $bw.Write([byte]$w)
    $bw.Write([byte]$h)
    $bw.Write([byte]0)      # color count
    $bw.Write([byte]0)      # reserved
    $bw.Write([UInt16]1)    # planes
    $bw.Write([UInt16]32)   # bit count
    $bw.Write([UInt32]$bytes.Length) # bytes in res
    $bw.Write([UInt32]$offset)       # offset
    $offset = $offset + [UInt32]$bytes.Length
  }

  # Write images
  foreach ($bytes in $images) { $bw.Write($bytes) }
  $bw.Flush()
  [System.IO.File]::WriteAllBytes($path, $ms.ToArray())
  $bw.Dispose(); $ms.Dispose()
}

# Convert hex to Color
function ToColor([string]$hex) { return [System.Drawing.ColorTranslator]::FromHtml($hex) }

$primaryC = ToColor $Primary
$secondaryC = ToColor $Secondary
$glyphC = ToColor $GlyphColor

$sizes = @(16, 24, 32, 48, 64, 128, 256)
$imgs = @()
foreach ($s in $sizes) { $imgs += ,(Draw-IconPngBytes -size $s -c1 $primaryC -c2 $secondaryC -glyph $glyphC) }

Write-Icon -images $imgs -path $Out
Write-Output "Generated $Out with sizes: $($sizes -join ', ')"