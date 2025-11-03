package main

import (
	"customer-survey/internal/ui"
	"log"
	"runtime"
	"syscall"
)

func hideConsole() {
	if runtime.GOOS == "windows" {
		// Hide the console window on Windows
		kernel32 := syscall.NewLazyDLL("kernel32.dll")
		user32 := syscall.NewLazyDLL("user32.dll")

		getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
		showWindow := user32.NewProc("ShowWindow")

		// Get the console window handle
		hwnd, _, _ := getConsoleWindow.Call()

		// Hide the window (SW_HIDE = 0)
		if hwnd != 0 {
			showWindow.Call(hwnd, uintptr(0))
		}
	}
}

func main() {
	// Hide console window before launching UI
	hideConsole()

	// Launch native Windows desktop UI
	if err := ui.RunDesktopUI(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}
