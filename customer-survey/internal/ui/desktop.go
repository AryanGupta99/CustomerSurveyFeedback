package ui

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"
)

//go:embed static
var staticFiles embed.FS

func getScreenResolution() (int, int) {
	if runtime.GOOS == "windows" {
		user32 := syscall.NewLazyDLL("user32.dll")
		getSystemMetrics := user32.NewProc("GetSystemMetrics")
		width, _, _ := getSystemMetrics.Call(0)  // SM_CXSCREEN
		height, _, _ := getSystemMetrics.Call(1) // SM_CYSCREEN
		return int(width), int(height)
	}
	// Fallback for other OSes
	return 1920, 1080
}

// RunDesktopUI launches the survey as a desktop application
// It starts a local server and opens it in app mode (frameless window)
func RunDesktopUI() error {
	// Wire up handlers
	http.HandleFunc("/", HandleIndex)
	http.HandleFunc("/submit", HandleSurveySubmission)

	// Bind to the first available port
	ln, port, err := getListener()
	if err != nil {
		return fmt.Errorf("could not bind a port: %w", err)
	}

	// Serve in background
	srv := &http.Server{Handler: nil}
	go func() {
		log.Printf("Starting server on :%d", port)
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Give the server a moment
	time.Sleep(300 * time.Millisecond)

	// Open in app mode (frameless browser window that looks like a desktop app)
	url := fmt.Sprintf("http://localhost:%d", port)
	openAsApp(url)

	// Block indefinitely
	select {}
}

func getListener() (net.Listener, int, error) {
	for p := 8080; p <= 8090; p++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
		if err == nil {
			return ln, p, nil
		}
	}
	return nil, 0, fmt.Errorf("no free port in 8080-8090")
}

// openAsApp opens the URL in application mode (frameless window) if possible
func openAsApp(url string) {
	// Set window size to fit the survey form perfectly
	formWidth := 420
	formHeight := 700

	// Get screen resolution
	screenW, screenH := getScreenResolution()

	// Center position
	posX := (screenW - formWidth) / 2
	posY := (screenH - formHeight) / 2

	windowSize := fmt.Sprintf("--window-size=%d,%d", formWidth, formHeight)
	windowPos := fmt.Sprintf("--window-position=%d,%d", posX, posY)

	edgePaths := []string{
		"C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe",
		"C:\\Program Files\\Microsoft\\Edge\\Application\\msedge.exe",
	}
	for _, edgePath := range edgePaths {
		if _, err := os.Stat(edgePath); err == nil {
			cmd := exec.Command(edgePath,
				"--app="+url,
				windowSize,
				windowPos,
				"--disable-features=TranslateUI,VizDisplayCompositor,EdgeShortcutsSync,MicrosoftEdgeIntroShowNotification",
				"--no-first-run",
				"--no-default-browser-check",
				"--disable-gpu",
				"--hide-scrollbars",
				"--disable-background-timer-throttling",
				"--disable-renderer-backgrounding",
				"--disable-backgrounding-occluded-windows",
				"--disable-sync",
				"--disable-popup-blocking",
				"--user-data-dir="+os.Getenv("TEMP")+"\\ace-survey-browser-"+fmt.Sprint(os.Getpid()),
			)
			if err := cmd.Start(); err == nil {
				log.Printf("Launched in app mode via Edge (separate profile)")
				return
			}
		}
	}
	chromePaths := []string{
		os.Getenv("LOCALAPPDATA") + "\\Google\\Chrome\\Application\\chrome.exe",
		"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
		"C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",
	}
	for _, chromePath := range chromePaths {
		if _, err := os.Stat(chromePath); err == nil {
			cmd := exec.Command(chromePath,
				"--app="+url,
				windowSize,
				windowPos,
				"--disable-gpu",
				"--hide-scrollbars",
				"--disable-features=VizDisplayCompositor,ChromeHeadless",
				"--disable-background-timer-throttling",
				"--disable-renderer-backgrounding",
				"--disable-backgrounding-occluded-windows",
				"--no-first-run",
				"--no-default-browser-check",
				"--disable-sync",
				"--disable-popup-blocking",
				"--user-data-dir="+os.Getenv("TEMP")+"\\ace-survey-browser-"+fmt.Sprint(os.Getpid()),
			)
			if err := cmd.Start(); err == nil {
				log.Printf("Launched in app mode via Chrome (separate profile)")
				return
			}
		}
	}

	// Fallback: regular browser
	log.Printf("App mode not available, using default browser")
	_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

// HandleIndex serves the main popup/index page
func HandleIndex(w http.ResponseWriter, r *http.Request) {
	// serve embedded static files
	sub, _ := fs.Sub(staticFiles, "static")
	http.FileServer(http.FS(sub)).ServeHTTP(w, r)
}

// ListSubmissions serves the contents of submissions.log as a simple HTML page
func ListSubmissions(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("submissions.log")
	if err != nil {
		http.Error(w, "no submissions found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<html><body><h2>Saved Submissions</h2><pre>"))
	w.Write(data)
	w.Write([]byte("</pre></body></html>"))
}
