package main

import (
	"godab/api"
	"godab/cmd"
	"godab/config"
	"os"
)

func main() {
	if !api.DirExists(config.GetDownloadLocation()) {
		api.PrintError("You must provide a valid DOWNLOAD_LOCATION folder")
	}

	asciiArt := `
  ____           _       _
 / ___| ___   __| | __ _| |__
| |  _ / _ \ / _` + "`" + ` |/ _` + "`" + ` | '_ \
| |_| | (_) | (_| | (_| | |_) |
 \____|\___/ \__,_|\__,_|_.__/
`

	api.PrintColor(api.COLOR_BLUE, "%s", asciiArt)
	api.PrintColor(api.COLOR_BLUE, "v%s", config.GetVersion())

	isLoginCommand := false
	if len(os.Args) > 1 && os.Args[1] == "login" {
		isLoginCommand = true
	}

	loggedIn, err := api.LoadCookies()

	if !isLoginCommand {
		if err != nil {
			api.PrintError("You're not logged-in. Run 'login' command first.")
		}

		if !loggedIn {
			api.PrintError("You must be logged in to download from dabmusic")
		}
	}

	cmd.Execute()
}