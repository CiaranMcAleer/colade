// main projet file, entry point of the cli
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/CiaranMcAleer/colade/internal/sitegen"
)

var version = "dev" // Version set during build with go build -ldflags "-X main.version=1.2.3"

func main() {
	rootCmd := &cobra.Command{
		Use:   "colade",
		Short: "Colade - Static site generator from Markdown",
		Long:  `Colade is a CLI tool to generate static sites from Markdown files.`,
	}

	buildCmd := &cobra.Command{
		Use:   "build [inputDir] [outputDir]",
		Short: "Build a static site from Markdown files",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			inputDir := args[0]
			outputDir := args[1]
			if err := sitegen.BuildSite(inputDir, outputDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(buildCmd)

	serveCmd := &cobra.Command{
		Use:   "serve [dir]",
		Short: "Serve a directory locally for preview",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dir := args[0]
			info, err := os.Stat(dir)
			if err != nil || !info.IsDir() {
				fmt.Fprintf(os.Stderr, "Error: '%s' is not a valid directory\n", dir)
				os.Exit(1)
			}
			err = sitegen.ServeDir(dir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
	rootCmd.AddCommand(serveCmd)

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
