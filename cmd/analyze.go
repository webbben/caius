/*
Copyright © 2025 Ben Webb ben.webb340@gmail.com
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/webbben/caius/internal/metrics"
	"github.com/webbben/caius/internal/project"
	"github.com/webbben/caius/internal/utils"
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "Usage: a directory or file path is required.")
			os.Exit(1)
		}
		path, err := filepath.Abs(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error resolving path:", err)
			os.Exit(1)
		}

		fileinfo, err := os.Stat(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error getting file information:", err)
			os.Exit(1)
		}

		if fileinfo.IsDir() {
			output, err := project.AnalyzeDirectory(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, "\nfailed to analyze directory;", err)
				os.Exit(1)
			}
			fmt.Println(output)
			elapsed := metrics.SpeedRecord("AnalyzeDirectory").GetAverageDuration().Round(time.Second)
			utils.Terminal.Lowkey(fmt.Sprintf("(%s elapsed)", elapsed))
		} else {
			filename := filepath.Base(path)
			fmt.Printf("Analyzing %s ...\n", filename)
			response, err := project.AnalyzeFileBasic(path, filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to analyze file: %q", err)
			}
			fmt.Println("file type:", response.Type)
			fmt.Println("description:", response.Description)
		}
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// analyzeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// analyzeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
