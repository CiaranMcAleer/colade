package sitegen

import (
	"fmt"
	"net/http"
)

func ServeDir(dir string, port int) error {
	handler := http.FileServer(http.Dir(dir))
	fmt.Printf("Serving '%s' at http://localhost:%d\n", dir, port)
	fmt.Println("Press Ctrl+C to stop.")
	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, handler)
}
