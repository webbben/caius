/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/webbben/caius/internal/utils"
	"github.com/webbben/caius/internal/websearch"
)

// websearchCmd represents the websearch command
var websearchCmd = &cobra.Command{
	Use:   "websearch",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		websites, err := websearch.WebSearch("Trump Epstein files")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("found %v websites:\n", len(websites))
		for _, website := range websites {
			utils.Terminal.Lowkey(fmt.Sprintf("%s\n%s / %s", website.URL, website.WebsiteName, website.Title))
		}

		output, err := websearch.SummarizeListOfWebsites(websites)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(output)
	},
}

func init() {
	rootCmd.AddCommand(websearchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// websearchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// websearchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
