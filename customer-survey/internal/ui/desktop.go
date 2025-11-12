package ui

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os/exec"
	"time"
)

//go:embed static
var staticFiles embed.FS

func RunDesktopUI() error {
	// Start lightweight HTTP server
	ln, port, err := getListener()
	if err != nil {
		return err
	}
	defer ln.Close()

	// Setup routes
	mux := http.NewServeMux()
	sub, _ := fs.Sub(staticFiles, "static")
	mux.Handle("/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("/submit", HandleSurveySubmission) // Match client-side script

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Start server in background
	go server.Serve(ln)

	url := fmt.Sprintf("http://localhost:%d", port)

	// Open ONLY in default browser (no Edge/Chrome spawning)
	// This uses the user's already-running browser tab (minimal memory)
	if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start(); err != nil {
		log.Printf("Failed to open browser: %v", err)
		return err
	}

	// Keep server alive for 5 minutes max
	time.Sleep(5 * time.Minute)
	return nil
}

func getListener() (net.Listener, int, error) {
	for p := 8080; p <= 8090; p++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", p))
		if err == nil {
			return ln, p, nil
		}
	}
	return nil, 0, fmt.Errorf("no free port in 8080-8090")
}
