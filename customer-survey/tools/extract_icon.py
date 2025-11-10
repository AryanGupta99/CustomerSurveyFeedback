import ctypes
from ctypes import wintypes
import sys
from PIL import Image

if len(sys.argv) < 3:
    print("Usage: extract_icon.py <exe> <out.png>")
    sys.exit(1)

path = sys.argv[1]
out = sys.argv[2]

shell32 = ctypes.windll.shell32
user32 = ctypes.windll.user32
gdi32 = ctypes.windll.gdi32

ExtractIconExW = shell32.ExtractIconExW
ExtractIconExW.argtypes = [wintypes.LPCWSTR, ctypes.c_int, ctypes.POINTER(wintypes.HICON), ctypes.POINTER(wintypes.HICON), ctypes.c_uint]
ExtractIconExW.restype = ctypes.c_uint

DestroyIcon = user32.DestroyIcon
DestroyIcon.argtypes = [wintypes.HICON]
DestroyIcon.restype = wintypes.BOOL

large = (wintypes.HICON * 1)()
small = (wintypes.HICON * 1)()
count = ExtractIconExW(path, 0, large, small, 1)
if count <= 0:
    print("No icon found")
    sys.exit(1)

hicon = large[0] or small[0]

# Get icon info
class ICONINFO(ctypes.Structure):
    _fields_ = [
        ("fIcon", wintypes.BOOL),
        ("xHotspot", wintypes.DWORD),
        ("yHotspot", wintypes.DWORD),
        ("hbmMask", wintypes.HBITMAP),
        ("hbmColor", wintypes.HBITMAP),
    ]

class BITMAP(ctypes.Structure):
    _fields_ = [
        ("bmType", wintypes.LONG),
        ("bmWidth", wintypes.LONG),
        ("bmHeight", wintypes.LONG),
        ("bmWidthBytes", wintypes.LONG),
        ("bmPlanes", wintypes.WORD),
        ("bmBitsPixel", wintypes.WORD),
        ("bmBits", ctypes.c_void_p),
    ]

GetIconInfo = user32.GetIconInfo
GetIconInfo.argtypes = [wintypes.HICON, ctypes.POINTER(ICONINFO)]
GetIconInfo.restype = wintypes.BOOL

GetObjectW = gdi32.GetObjectW
GetObjectW.argtypes = [wintypes.HGDIOBJ, ctypes.c_int, ctypes.c_void_p]
GetObjectW.restype = ctypes.c_int

GetDIBits = gdi32.GetDIBits
GetDIBits.argtypes = [wintypes.HDC, wintypes.HBITMAP, wintypes.UINT, wintypes.UINT, ctypes.c_void_p, ctypes.c_void_p, wintypes.UINT]
GetDIBits.restype = ctypes.c_int

DeleteObject = gdi32.DeleteObject
DeleteObject.argtypes = [wintypes.HGDIOBJ]
DeleteObject.restype = wintypes.BOOL

iconinfo = ICONINFO()
if not GetIconInfo(hicon, ctypes.byref(iconinfo)):
    raise ctypes.WinError()

bmp = BITMAP()
if not GetObjectW(iconinfo.hbmColor, ctypes.sizeof(bmp), ctypes.byref(bmp)):
    raise ctypes.WinError()

width = bmp.bmWidth
height = bmp.bmHeight

# Prepare bitmap info header
class BITMAPINFOHEADER(ctypes.Structure):
    _fields_ = [
        ("biSize", wintypes.DWORD),
        ("biWidth", wintypes.LONG),
        ("biHeight", wintypes.LONG),
        ("biPlanes", wintypes.WORD),
        ("biBitCount", wintypes.WORD),
        ("biCompression", wintypes.DWORD),
        ("biSizeImage", wintypes.DWORD),
        ("biXPelsPerMeter", wintypes.LONG),
        ("biYPelsPerMeter", wintypes.LONG),
        ("biClrUsed", wintypes.DWORD),
        ("biClrImportant", wintypes.DWORD),
    ]

BI_RGB = 0
bih = BITMAPINFOHEADER()
bih.biSize = ctypes.sizeof(BITMAPINFOHEADER)
bih.biWidth = width
bih.biHeight = height
bih.biPlanes = 1
bih.biBitCount = 32
bih.biCompression = BI_RGB

buf_len = width * height * 4
buf = (ctypes.c_byte * buf_len)()

hdc = ctypes.windll.user32.GetDC(None)
try:
    res = GetDIBits(hdc, iconinfo.hbmColor, 0, height, ctypes.byref(buf), ctypes.byref(bih), 0)
    if res == 0:
        raise ctypes.WinError()
finally:
    ctypes.windll.user32.ReleaseDC(None, hdc)

# Convert BGRA bottom-up to RGBA top-down
rows = []
for row in range(height):
    start = row * width * 4
    end = start + width * 4
    rows.append(buf[start:end])
rows.reverse()

data = bytearray()
for row in rows:
    for i in range(0, len(row), 4):
        b, g, r, a = row[i:i+4]
        data.extend([r & 0xFF, g & 0xFF, b & 0xFF, a & 0xFF])

img = Image.frombytes('RGBA', (width, height), bytes(data))
img.save(out)
print(f"Saved {out} size {img.size}")

DestroyIcon(hicon)
DeleteObject(iconinfo.hbmColor)
DeleteObject(iconinfo.hbmMask)
