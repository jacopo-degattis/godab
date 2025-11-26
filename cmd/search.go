package cmd

import (
	"godab/api"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

var queryType string

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search track, album or artist",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]

		supportedTypes := []string{"track", "artist", "album"}
		valid := slices.Contains(supportedTypes, strings.ToLower(queryType))

		if !valid {
			api.PrintError("You can search only by: track, artist and album")
		}

		results, err := api.Search(query, queryType)

		api.CheckErr(err)

		api.PrintResultsTable(results, queryType)
	},
}

func init() {
	searchCmd.Flags().StringVarP(&queryType, "type", "t", "", "Query type (track, artist, album)")
	rootCmd.AddCommand(searchCmd)
}
