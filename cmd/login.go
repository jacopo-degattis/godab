package cmd

import (
	"godab/api"
	"os"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login using credentials",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		email := args[0]
		password := args[1]
		err := api.Login(email, password)

		if err != nil {
			api.PrintColor(api.COLOR_RED, "%s", err)
			os.Exit(1)
		}

		api.PrintColor(api.COLOR_GREEN, "Login successfull")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
