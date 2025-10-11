package main

import (
	"customer-survey/internal/ui"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"time"
)

func openBrowser(url string) {
	// Windows default browser
	// rundll32 url.dll,FileProtocolHandler <url>
	_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
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

func main() {
	// Wire up handlers
	http.HandleFunc("/", ui.HandleIndex)
	http.HandleFunc("/submit", ui.HandleSurveySubmission)
	http.HandleFunc("/submissions", ui.ListSubmissions)

	// Bind to the first available port in 8080-8090
	ln, port, err := getListener()
	if err != nil {
		log.Fatalf("Could not bind a port: %v", err)
	}

	// Serve in background
	srv := &http.Server{Handler: nil}
	go func() {
		log.Printf("Starting server on :%d", port)
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Give the server a moment, then open the browser
	time.Sleep(300 * time.Millisecond)
	openBrowser(fmt.Sprintf("http://localhost:%d", port))

	// Block main goroutine by serving again on the same listener would error; instead wait
	// indefinitely. The server goroutine will continue to serve requests.
	select {}
}
