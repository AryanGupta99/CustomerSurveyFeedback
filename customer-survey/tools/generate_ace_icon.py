from PIL import Image
import os

source = os.path.join(os.path.dirname(__file__), "..", "icon files", "Ace-Cloud-Icon-Dark.png")
source = os.path.abspath(source)

dest_png = os.path.join(os.path.dirname(__file__), "AceCloudIcon_512.png")
dest_ico = os.path.join(os.path.dirname(__file__), "..", "winres", "assets", "appicon.ico")

dest_png2 = os.path.join(os.path.dirname(__file__), "appicon_preview_from_ace.png")

print("Loading", source)
img = Image.open(source).convert("RGBA")
print("Original size:", img.size)

sizes = [16, 24, 32, 48, 64, 128, 256]

# Ensure square by fitting within min dimension
min_side = min(img.size)
img_cropped = img.crop((0, 0, min_side, min_side))
img_resized = img_cropped.resize((512, 512), Image.Resampling.LANCZOS)
img_resized.save(dest_png)
print("Saved 512x512 reference PNG to", dest_png)

icons = [img_resized.resize((s, s), Image.Resampling.LANCZOS) for s in sizes]
icons[0].save(dest_ico, sizes=[(s, s) for s in sizes])
print("Saved ICO to", dest_ico)

icons[2].save(dest_png2)
print("Saved preview PNG to", dest_png2)
