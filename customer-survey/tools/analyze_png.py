from PIL import Image
import sys
path = sys.argv[1]
img = Image.open(path)
print('Analyzing', path)
print('size:', img.size, 'mode:', img.mode)
pixels = list(img.getdata())
avg = tuple(sum(p[i] for p in pixels)//len(pixels) for i in range(3))
print('average RGB:', avg)
center = img.getpixel((img.width//2, img.height//2))
print('center pixel:', center)
