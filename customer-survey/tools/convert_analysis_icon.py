from PIL import Image
import os

src = r"c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\Customer-Survey-Application\customer-survey\SCCM-Package\analysis-icon.png"
dest = r"c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\Customer-Survey-Application\customer-survey\winres\assets\appicon.ico"

print('Loading', src)
img = Image.open(src).convert('RGBA')
print('Original size:', img.size)

sizes = [(16,16),(24,24),(32,32),(48,48),(64,64),(128,128),(256,256)]
imgs = [img.resize(s, Image.Resampling.LANCZOS) for s in sizes]

# Save as multi-size ICO
imgs[0].save(dest, sizes=sizes)
print('Saved ICO to', dest)
print('Size (bytes):', os.path.getsize(dest))
