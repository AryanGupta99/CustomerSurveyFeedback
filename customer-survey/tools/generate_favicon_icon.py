from PIL import Image
import os

# Use the existing favicon PNG
source = r"c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\Customer-Survey-Application\customer-survey\icon files\Ace-Fav-Icon-96x96.png"
dest_ico = r"c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\Customer-Survey-Application\customer-survey\winres\assets\appicon.ico"
dest_frontend_ico = r"c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\Customer-Survey-Application\customer-survey\cmd\wails-app\frontend\favicon.png"

print("Loading favicon from", source)
img = Image.open(source).convert("RGBA")
print("Original size:", img.size)

# Generate multi-size ICO for Windows resources
sizes = [16, 24, 32, 48, 64, 128, 256]

# Ensure square
img_square = img.resize((256, 256), Image.Resampling.LANCZOS)

# Generate all size variants
icons = [img_square.resize((s, s), Image.Resampling.LANCZOS) for s in sizes]

# Save as ICO (multi-size)
icons[0].save(dest_ico, sizes=[(s, s) for s in sizes])
print("✓ Saved multi-size ICO to", dest_ico)

# Also save 96x96 as frontend favicon.png
img.save(dest_frontend_ico)
print("✓ Saved 96x96 favicon to", dest_frontend_ico)
