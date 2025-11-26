package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "A golang dabmusic.xyz downloader",
}

func Execute() {
	rootCmd.Execute()
}
