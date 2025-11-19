package main

import (
	"flag"
	"fmt"
	"godab/api"
	"godab/config"
	"log"
	"os"
	"strings"
)

var formatMap = map[string]int{
	"mp3":  5,
	"flac": 27,
}

func PrintError(msg string) {
	api.PrintColor(api.COLOR_RED, "%s", msg)
}

func main() {
	if !api.DirExists(config.GetDownloadLocation()) {
		log.Fatalf("You must provide a valid DOWNLOAD_LOCATION folder")
	}

	asciiArt := `
  ____           _       _     
 / ___| ___   __| | __ _| |__  
| |  _ / _ \ / _\` + "`" + ` |/ _\` + "`" + ` | '_ \ 
| |_| | (_) | (_| | (_| | |_) |
 \____|\___/ \__,_|\__,_|_.__/ 
`

	var (
		email    string
		password string
		album    string
		track    string
		artist   string
		format   string
	)

	flag.StringVar(&email, "email", "", "Dabmusic email")
	flag.StringVar(&password, "password", "", "Dabmusic password")
	flag.StringVar(&album, "album", "", "Album URL to download")
	flag.StringVar(&track, "track", "", "Track URL to download")
	flag.StringVar(&artist, "artist", "", "Artist URL to download")
	flag.StringVar(&format, "format", "", "Track download format")
	flag.Parse()

	// fmt.Println(asciiArt)
	api.PrintColor(api.COLOR_BLUE, "%s", asciiArt)

	if album == "" && track == "" && artist == "" && email == "" && password == "" {
		flag.Usage()
	}

	if format != "" && (format != "flac" && format != "mp3") {
		PrintError("Invalid audio format, you must choose between MP3 or FLAC")
		os.Exit(1)
	}

	if (email != "" && password == "") || (email == "" && password != "") {
		PrintError("In order to login you must provide both -email and -password.")
		flag.Usage()
	}

	if (album != "" && track != "") || (artist != "" && track != "") || (album != "" && artist != "") {
		PrintError("You can download only one between `album` and `track` at a time.")
		flag.Usage()
	}

	loggedIn, err := api.LoadCookies()

	if err != nil {
		api.PrintColor(api.COLOR_YELLOW, "You're not logged-in, please log in using: -email and -password")
		os.Exit(1)
	}

	if email != "" && password != "" {
		err := api.Login(email, password)

		if err != nil {
			log.Fatalf("Error: %s", err)
		}

		api.PrintColor(api.COLOR_GREEN, "[+] Logged in succesfully!")
		return
	}

	if !loggedIn {
		api.PrintColor(api.COLOR_YELLOW, "You must be logged in to download from dabmusic")
		os.Exit(1)
	}

	downloadFormat := formatMap[strings.ToLower(format)]

	// TODO: can I improve this?
	if downloadFormat == 0 {
		downloadFormat = 27
	}

	if album != "" {
		album, err := api.NewAlbum(album)

		if err != nil {
			PrintError(fmt.Sprintf("Error: %s", err))
		}

		if err := album.Download(downloadFormat, true); err != nil {
			PrintError(fmt.Sprintf("Cannot download album %s: %s", album.Title, err))
		}
	} else if track != "" {
		track, err := api.NewTrack(track)

		if err != nil {
			PrintError(fmt.Sprintf("Error: %s", err))
		}

		if err := track.Download(downloadFormat); err != nil {
			PrintError(fmt.Sprintf("Cannot download track %s: %s", track.Title, err))
		}
	} else if artist != "" {
		artist, err := api.NewArtist(artist)

		if err != nil {
			PrintError(fmt.Sprintf("Error: %s", err))
		}

		if err := artist.Download(downloadFormat); err != nil {
			PrintError(fmt.Sprintf("Cannot download artist %s: %s", artist.Name, err))
		}
	}

	os.Exit(0)
}
