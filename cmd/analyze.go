/*
Copyright © 2025 Ben Webb ben.webb340@gmail.com
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/webbben/caius/internal/project"
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
		response, err := project.AnalyzeFileBasic("/Users/benwebb/dev/personal/ai-tools/ai-agent/internal/project/tests/javascript01.txt", "index.js")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to analyze file: %q", err)
		}
		fmt.Println(response.Type)
		fmt.Println(response.Description)
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
