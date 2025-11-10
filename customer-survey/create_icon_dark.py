from PIL import Image
import os

# Use the dark icon - better for taskbar
ace_png = r"c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\Customer-Survey-Application\customer-survey\icon files\Ace-Cloud-Icon-Dark.png"
ico_path = r"c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\Customer-Survey-Application\customer-survey\winres\assets\icon.ico"

print(f"Loading PNG: {ace_png}")
img = Image.open(ace_png).convert('RGBA')
print(f"Original size: {img.size}")

# Create all common icon sizes with high quality
sizes = [(16, 16), (24, 24), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)]
images = []

for size in sizes:
    # Use high-quality resampling (LANCZOS)
    resized = img.resize(size, Image.Resampling.LANCZOS)
    images.append(resized)
    print(f"Created {size[0]}x{size[1]} icon")

# Save as ICO with all sizes
images[0].save(ico_path, sizes=sizes)
print(f"\n✓ Saved multi-size ICO to: {ico_path}")
print(f"✓ File size: {os.path.getsize(ico_path)} bytes")
