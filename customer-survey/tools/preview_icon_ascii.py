from PIL import Image
import sys

if len(sys.argv) < 2:
    print("Usage: preview_icon_ascii.py <image>")
    sys.exit(1)

path = sys.argv[1]
img = Image.open(path).convert("RGB").resize((32, 32))
chars = " .:-=+*#%@"
for y in range(img.height):
    row = []
    for x in range(img.width):
        r, g, b = img.getpixel((x, y))
        lum = (0.2126 * r + 0.7152 * g + 0.0722 * b) / 255
        row.append(chars[int(lum * (len(chars) - 1))])
    print("".join(row))
