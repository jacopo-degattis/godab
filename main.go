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
| |  _ / _ \ / _\` + "`" + ` |/ _\` + "`" + ` | '_ \ 
| |_| | (_) | (_| | (_| | |_) |
 \____|\___/ \__,_|\__,_|_.__/ 
`

	api.PrintColor(api.COLOR_BLUE, "%s", asciiArt)

	loggedIn, err := api.LoadCookies()

	if err != nil {
		api.PrintError("You're not logged-in, please log in before using other commands")
	}

	if !loggedIn {
		api.PrintError("You must be logged in to download from dabmusic")
	}

	cmd.Execute()

	os.Exit(0)
}
