// main projet file, entry point of the cli
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/CiaranMcAleer/colade/internal/sitegen"
)

var version = "dev" // Version set during build with go build -ldflags "-X main.version=1.2.3"
var coladeAscii = `
 ██████╗ ██████╗ ██╗      █████╗ ██████╗ ███████╗
██╔════╝██╔═══██╗██║     ██╔══██╗██╔══██╗██╔════╝
██║     ██║   ██║██║     ███████║██║  ██║█████╗  
██║     ██║   ██║██║     ██╔══██║██║  ██║██╔══╝  
╚██████╗╚██████╔╝███████╗██║  ██║██████╔╝███████╗
 ╚═════╝ ╚═════╝ ╚══════╝╚═╝  ╚═╝╚═════╝ ╚══════╝
`

func main() {
	fmt.Print(coladeAscii)
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
			threshold, _ := cmd.Flags().GetInt("size-threshold")
			noIncremental, _ := cmd.Flags().GetBool("no-incremental")
			rssURL, _ := cmd.Flags().GetString("rss")
			rssMaxItems, _ := cmd.Flags().GetInt("rss-max-items")
			keepOrphaned, _ := cmd.Flags().GetBool("keep-orphaned")
			templateOpt, _ := cmd.Flags().GetString("template")
			headerFile, _ := cmd.Flags().GetString("header-file")
			footerFile, _ := cmd.Flags().GetString("footer-file")
			noHeader, _ := cmd.Flags().GetBool("no-header")
			noFooter, _ := cmd.Flags().GetBool("no-footer")
			if err := sitegen.BuildSite(
				inputDir, outputDir, threshold*1024, noIncremental, rssURL, rssMaxItems, keepOrphaned, templateOpt,
				headerFile, footerFile, noHeader, noFooter,
			); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
	buildCmd.Flags().IntP("size-threshold", "s", 14, "Size threshold in KB for gzip compression warnings")
	buildCmd.Flags().Bool("no-incremental", false, "Disable incremental build and force full rebuild")
	buildCmd.Flags().StringP("rss", "r", "", "Generate RSS feed with specified base URL (e.g., https://example.com)")
	buildCmd.Flags().Int("rss-max-items", 20, "Maximum number of items to include in RSS feed (default 20)")
	buildCmd.Flags().Bool("keep-orphaned", false, "Keep orphaned files in output directory instead of deleting them")
	buildCmd.Flags().String("template", "default", "Template to use for HTML output (name of bundled template or path to custom template)")
	buildCmd.Flags().String("header-file", "", "Markdown file to use as header (default: header.md in inputDir)")
	buildCmd.Flags().String("footer-file", "", "Markdown file to use as footer (default: footer.md in inputDir)")
	buildCmd.Flags().Bool("no-header", false, "Disable header injection")
	buildCmd.Flags().Bool("no-footer", false, "Disable footer injection")

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
			port, _ := cmd.Flags().GetInt("port")
			err = sitegen.ServeDir(dir, port)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
	serveCmd.Flags().IntP("port", "p", 8080, "Port to serve on (default 8080)")
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
