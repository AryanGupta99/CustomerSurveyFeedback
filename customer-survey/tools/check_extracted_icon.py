from PIL import Image
import os
path = os.path.join(os.environ['TEMP'], 'customer-survey-extracted-icon.png')
img = Image.open(path)
print('size:', img.size, 'mode:', img.mode)
pixels = list(img.getdata())
avg = tuple(sum(p[i] for p in pixels)//len(pixels) for i in range(3))
print('average RGB:', avg)
center = img.getpixel((img.width//2, img.height//2))
print('center pixel RGB:', center)
img.save(os.path.join(os.environ['TEMP'], 'customer-survey-extracted-icon-preview.png'))
