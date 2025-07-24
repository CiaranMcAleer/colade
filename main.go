// main projet file, entry point of the cli
package main

import (
	"fmt"
	"os"
)

var version = "dev"// Version set during build with go build -ldflags "-X main.version=1.2.3"

func main() {
	//switch to handdle command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--help":
			help()
		case "--version":
			fmt.Println("Version:", version)
			os.Exit(0)
		default:
			fmt.Println("Unknown command:", os.Args[1])
			help()
		}
	} else {
		fmt.Println("No arguments provided. Use --help for usage information.")
		os.Exit(1)
	}
}

func help() {
	fmt.Println("Usage: go run main.go [options]")
	fmt.Println("Options:")
	fmt.Println("  --help\tShow this help message")
	fmt.Println("  --version\tShow the version of the CLI")
	os.Exit(0)
}
