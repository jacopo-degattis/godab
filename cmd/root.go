package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "dabmusic.xyz downloader",
}

func Execute() {
	rootCmd.Execute()
}
