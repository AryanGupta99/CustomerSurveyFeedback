from PIL import Image
ico_path = r"c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\Customer-Survey-Application\customer-survey\cmd\wails-app\build\windows\icon.ico"
output = r"c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\Customer-Survey-Application\customer-survey\tools\build_icon_preview.png"
img = Image.open(ico_path)
img.save(output)
print('Saved', output, 'size', img.size)
