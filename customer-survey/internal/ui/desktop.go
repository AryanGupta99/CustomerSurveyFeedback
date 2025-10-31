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
	"time"
)

//go:embed static
var staticFiles embed.FS

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
	// Try Edge in app mode first (looks like a native app - no address bar)
	edgePaths := []string{
		"C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe",
		"C:\\Program Files\\Microsoft\\Edge\\Application\\msedge.exe",
	}

	for _, edgePath := range edgePaths {
		if _, err := os.Stat(edgePath); err == nil {
			// Launch Edge in app mode (frameless window) - compact size
			cmd := exec.Command(edgePath,
				"--app="+url,
				"--window-size=420,600",
				"--disable-features=TranslateUI",
				"--no-first-run",
				"--force-device-scale-factor=0.9",
			)
			if err := cmd.Start(); err == nil {
				log.Printf("Launched in app mode via Edge")
				return
			}
		}
	}

	// Try Chrome in app mode
	chromePaths := []string{
		os.Getenv("LOCALAPPDATA") + "\\Google\\Chrome\\Application\\chrome.exe",
		"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
		"C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",
	}

	for _, chromePath := range chromePaths {
		if _, err := os.Stat(chromePath); err == nil {
			cmd := exec.Command(chromePath,
				"--app="+url,
				"--window-size=420,600",
				"--force-device-scale-factor=0.9",
			)
			if err := cmd.Start(); err == nil {
				log.Printf("Launched in app mode via Chrome")
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
