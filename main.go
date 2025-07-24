// main projet file, entry point of the cli
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev" // Version set during build with go build -ldflags "-X main.version=1.2.3"

func main() {
	rootCmd := &cobra.Command{
		Use:   "colade",
		Short: "Colade - Static site generator from Markdown",
		Long:  `Colade is a CLI tool to generate static sites from Markdown files.`,
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show the version of Colade",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Colade version:", version)
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
