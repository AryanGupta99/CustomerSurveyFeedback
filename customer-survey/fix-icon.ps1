# Quick and dirty icon creator that WILL work
Add-Type -AssemblyName System.Drawing

Write-Output "Loading PNG file..."
$png = [System.Drawing.Image]::FromFile((Resolve-Path "icon.png").Path)

# Create a simple 32x32 icon - most important size
$size = 32
$bmp = New-Object System.Drawing.Bitmap($size, $size)
$g = [System.Drawing.Graphics]::FromImage($bmp)
$g.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
$g.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
$g.DrawImage($png, 0, 0, $size, $size)
$g.Dispose()

# Convert to PNG bytes
$ms = New-Object System.IO.MemoryStream
$bmp.Save($ms, [System.Drawing.Imaging.ImageFormat]::Png)
$pngBytes = $ms.ToArray()
$ms.Dispose()
$bmp.Dispose()
$png.Dispose()

Write-Output "Creating ICO structure..."

# Create ICO file structure
$ico = New-Object System.IO.MemoryStream
$bw = New-Object System.IO.BinaryWriter($ico)

# ICO Header
$bw.Write([UInt16]0)        # Reserved
$bw.Write([UInt16]1)        # Type: Icon
$bw.Write([UInt16]1)        # Count: 1 image

# Icon Directory Entry
$bw.Write([byte]32)         # Width
$bw.Write([byte]32)         # Height
$bw.Write([byte]0)          # Color count
$bw.Write([byte]0)          # Reserved
$bw.Write([UInt16]1)        # Planes
$bw.Write([UInt16]32)       # Bits per pixel
$bw.Write([UInt32]$pngBytes.Length)  # Image size
$bw.Write([UInt32]22)       # Offset to image data

# Write image data
$bw.Write($pngBytes)

# Save the ICO file
$icoBytes = $ico.ToArray()
[System.IO.File]::WriteAllBytes("assets\icon.ico", $icoBytes)

$bw.Dispose()
$ico.Dispose()

Write-Output "Icon created successfully: assets\icon.ico ($($icoBytes.Length) bytes)"