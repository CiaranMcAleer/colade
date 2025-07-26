package sitegen

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
)

// TODO Allow user to set threshold via CLI
func CheckGzipSize(path string, threshold int, out chan<- string) {
	go func() {
		data, err := os.ReadFile(path)
		if err != nil {
			return
		}
		var gzBuf bytes.Buffer
		gz := gzip.NewWriter(&gzBuf)
		_, gzErr := gz.Write(data)
		gz.Close()
		if gzErr != nil {
			return
		}
		sizeKB := float64(gzBuf.Len()) / 1024
		threshKB := float64(threshold) / 1024
		msg := fmt.Sprintf("[Size] %s: compressed size is %.1fKB\n", path, sizeKB)
		if gzBuf.Len() > threshold {
			msg += fmt.Sprintf("[WARN] %s: compressed size is %.1fKB (> %.1fKB)\n", path, sizeKB, threshKB)
		}
		out <- msg
	}()
}
