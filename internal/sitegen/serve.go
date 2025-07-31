package sitegen

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// loggingHandler wraps an http.Handler to log requests
type loggingHandler struct {
	handler http.Handler
}

func (lh *loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Wrap the ResponseWriter to capture status code
	wrapper := &responseWrapper{ResponseWriter: w, statusCode: 200}

	lh.handler.ServeHTTP(wrapper, r)

	duration := time.Since(start)
	log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapper.statusCode, duration)
}

// responseWrapper captures the status code
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// customFileServer handles custom 404 and index.html serving
type customFileServer struct {
	root http.Dir
	dir  string
}

func (cfs *customFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clean the path
	path := r.URL.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Handle root path - serve index.html if it exists
	if path == "/" {
		indexPath := filepath.Join(string(cfs.root), "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
			return
		}
	}

	// Try to serve the requested file
	fullPath := filepath.Join(string(cfs.root), path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// File doesn't exist, try to serve custom 404.html
		custom404Path := filepath.Join(string(cfs.root), "404.html")
		if _, err := os.Stat(custom404Path); err == nil {
			w.WriteHeader(http.StatusNotFound)
			http.ServeFile(w, r, custom404Path)
			return
		}

		// Serve hardcoded 404 page
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>404 - Page Not Found</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; margin-top: 100px; }
        h1 { color: #333; }
        p { color: #666; }
    </style>
</head>
<body>
    <h1>404 - Page Not Found</h1>
    <p>The requested page could not be found.</p>
    <p><a href="/">Return to home</a></p>
</body>
</html>`)
		return
	}

	// Serve the file normally
	http.ServeFile(w, r, fullPath)
}

// checkPortAvailable checks if a port is available
func checkPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

func ServeDir(dir string, port int) error {
	// Check if port is available
	if !checkPortAvailable(port) {
		fmt.Printf("Port %d is already in use. Try a different port.\n", port)
		// Suggest alternative ports
		fmt.Println("Trying alternative ports...")
		for i := port + 1; i <= port+10; i++ {
			if checkPortAvailable(i) {
				fmt.Printf("Port %d is available. Use --port %d\n", i, i)
				break
			}
		}
		return fmt.Errorf("port %d is not available", port)
	}

	// Create custom file server
	customHandler := &customFileServer{
		root: http.Dir(dir),
		dir:  dir,
	}

	// Wrap with logging
	loggingWrapper := &loggingHandler{handler: customHandler}

	fmt.Printf("Serving '%s' at http://localhost:%d\n", dir, port)
	fmt.Println("Press Ctrl+C to stop.")

	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, loggingWrapper)
}
