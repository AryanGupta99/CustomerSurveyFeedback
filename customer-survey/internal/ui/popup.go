package ui

import (
    "embed"
    "io/fs"
    "net/http"
    "os"
)

//go:embed static
var staticFiles embed.FS

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