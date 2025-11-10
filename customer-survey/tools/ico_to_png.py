from PIL import Image
import sys
ico_path = sys.argv[1]
out_path = sys.argv[2]
img = Image.open(ico_path)
img.save(out_path)
print('Saved', out_path, 'size', img.size)
