package sitegen

import (
	"fmt"
	"net/http"
)

// ServeDir starts a local web server to serve static files from the given directory. This allows users to preview their site locally.
func ServeDir(dir string) error {
	handler := http.FileServer(http.Dir(dir))
	fmt.Printf("Serving '%s' at http://localhost:8080\n", dir)
	fmt.Println("Press Ctrl+C to stop.")
	return http.ListenAndServe(":8080", handler)
}
