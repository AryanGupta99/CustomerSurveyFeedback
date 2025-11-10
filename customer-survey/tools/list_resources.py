import ctypes
from ctypes import wintypes
import sys

if len(sys.argv) < 2:
    raise SystemExit("Usage: list_resources.py <exe>")

path = sys.argv[1]

kernel32 = ctypes.windll.kernel32

LOAD_LIBRARY_AS_DATAFILE = 0x00000002
LOAD_LIBRARY_AS_IMAGE_RESOURCE = 0x00000020

def makeintresource(i: int):
    return ctypes.cast(ctypes.c_void_p(i), wintypes.LPWSTR)

hModule = kernel32.LoadLibraryExW(path, None, LOAD_LIBRARY_AS_DATAFILE | LOAD_LIBRARY_AS_IMAGE_RESOURCE)
if not hModule:
    raise OSError("LoadLibraryEx failed")

names = []

ENUMRESNAMEPROCW = ctypes.WINFUNCTYPE(wintypes.BOOL, wintypes.HMODULE, wintypes.LPCWSTR, wintypes.LPCWSTR, wintypes.LPARAM)

def enum_proc(hModule, lpszType, lpszName, lParam):
    ptr_value = ctypes.cast(lpszName, ctypes.c_void_p).value
    if ptr_value >> 16 == 0:
        names.append(ptr_value)
    else:
        names.append(ctypes.wstring_at(lpszName))
    return True

cb = ENUMRESNAMEPROCW(enum_proc)
if not kernel32.EnumResourceNamesW(hModule, makeintresource(14), cb, 0):
    err = ctypes.GetLastError()
    kernel32.FreeLibrary(hModule)
    raise OSError(f"EnumResourceNames failed: {err}")

kernel32.FreeLibrary(hModule)

print("RT_GROUP_ICON resources:", names)
