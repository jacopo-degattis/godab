package cmd

import (
	"godab/api"
	"strings"

	"github.com/spf13/cobra"
)

var downloadFormat string

func getFormat() int {
	format := api.FormatMap[strings.ToLower(downloadFormat)]

	if format == 0 {
		format = 27
	}

	return format
}

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Download a track",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		format := getFormat()

		track, err := api.NewTrack(id)
		api.CheckErr(err)
		err = track.Download(format)
		api.CheckErr(err)
	},
}

var albumCmd = &cobra.Command{
	Use:   "album",
	Short: "Download a album",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		format := getFormat()

		album, err := api.NewAlbum(id)
		api.CheckErr(err)

		err = album.Download(format, true)
		api.CheckErr(err)
	},
}

var artistCmd = &cobra.Command{
	Use:   "arist",
	Short: "Download a artist",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		format := getFormat()

		artist, err := api.NewArtist(id)
		api.CheckErr(err)

		err = artist.Download(format)
		api.CheckErr(err)
	},
}

func init() {
	trackCmd.Flags().StringVarP(&downloadFormat, "format", "f", "", "Download format")
	albumCmd.Flags().StringVarP(&downloadFormat, "format", "f", "", "Download format")
	artistCmd.Flags().StringVarP(&downloadFormat, "format", "f", "", "Download format")
	rootCmd.AddCommand(trackCmd)
	rootCmd.AddCommand(albumCmd)
	rootCmd.AddCommand(artistCmd)
}
