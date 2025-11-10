import ctypes
from ctypes import wintypes
import struct
import sys

def load_ico(path):
    with open(path, 'rb') as f:
        data = f.read()
    reserved, ico_type, count = struct.unpack_from('<HHH', data, 0)
    if reserved != 0 or ico_type != 1:
        raise ValueError('Not a standard icon file')
    entries = []
    offset = 6
    for idx in range(count):
        bWidth, bHeight, bColorCount, bReserved, wPlanes, wBitCount, dwBytesInRes, dwImageOffset = struct.unpack_from('<BBBBHHII', data, offset)
        offset += 16
        img_data = data[dwImageOffset:dwImageOffset + dwBytesInRes]
        entries.append({
            'width': bWidth or 256,
            'height': bHeight or 256,
            'colorcount': bColorCount,
            'reserved': bReserved,
            'planes': wPlanes,
            'bitcount': wBitCount,
            'bytes': dwBytesInRes,
            'offset': dwImageOffset,
            'data': img_data,
        })
    return data, entries

RT_ICON = 3
RT_GROUP_ICON = 14

kernel32 = ctypes.windll.kernel32
kernel32.BeginUpdateResourceW.argtypes = [wintypes.LPCWSTR, wintypes.BOOL]
kernel32.BeginUpdateResourceW.restype = wintypes.HANDLE
kernel32.UpdateResourceW.argtypes = [wintypes.HANDLE, wintypes.LPCWSTR, wintypes.LPCWSTR, wintypes.WORD, ctypes.c_void_p, wintypes.DWORD]
kernel32.UpdateResourceW.restype = wintypes.BOOL
kernel32.EndUpdateResourceW.argtypes = [wintypes.HANDLE, wintypes.BOOL]
kernel32.EndUpdateResourceW.restype = wintypes.BOOL

def makeintresource(i):
    return ctypes.cast(ctypes.c_void_p(i), wintypes.LPWSTR)

def replace_icon(exe_path, ico_path):
    _, images = load_ico(ico_path)
    handle = kernel32.BeginUpdateResourceW(exe_path, True)
    if not handle:
        raise OSError('BeginUpdateResource failed')
    try:
        buffers = []
        # Add each image
        for idx, entry in enumerate(images, start=1):
            buf = ctypes.create_string_buffer(entry['data'])
            buffers.append(buf)
            ptr = ctypes.c_void_p(ctypes.addressof(buf))
            res = kernel32.UpdateResourceW(handle, makeintresource(RT_ICON), makeintresource(idx), 0, ptr, len(entry['data']))
            if not res:
                raise OSError(f'UpdateResource icon {idx} failed')
        # Build group icon
        grp_header = struct.pack('<HHH', 0, 1, len(images))
        grp_entries = []
        for idx, entry in enumerate(images, start=1):
            grp_entries.append(struct.pack('<BBBBHHIH',
                                           entry['width'] if entry['width'] != 256 else 0,
                                           entry['height'] if entry['height'] != 256 else 0,
                                           entry['colorcount'],
                                           entry['reserved'],
                                           entry['planes'] or 1,
                                           entry['bitcount'] or 32,
                                           entry['bytes'],
                                           idx))
        grp_data = grp_header + b''.join(grp_entries)
        grp_buf = ctypes.create_string_buffer(grp_data)
        buffers.append(grp_buf)
        res = kernel32.UpdateResourceW(handle, makeintresource(RT_GROUP_ICON), makeintresource(1), 0, ctypes.c_void_p(ctypes.addressof(grp_buf)), len(grp_data))
        if not res:
            raise OSError('UpdateResource group icon failed')
    finally:
        if not kernel32.EndUpdateResourceW(handle, False):
            raise OSError('EndUpdateResource failed')

if __name__ == '__main__':
    exe = sys.argv[1]
    ico = sys.argv[2]
    replace_icon(exe, ico)
    print('Replaced icon for', exe)
