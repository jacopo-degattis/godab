package main

import (
	"flag"
	"godab/api"
	"godab/config"
	"log"
)

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
	)

	flag.StringVar(&email, "email", "", "Dabmusic email")
	flag.StringVar(&password, "password", "", "Dabmusic password")
	flag.StringVar(&album, "album", "", "Album URL to download")
	flag.StringVar(&track, "track", "", "Track URL to download")
	flag.StringVar(&artist, "artist", "", "Artist URL to download")
	flag.Parse()

	if album == "" && track == "" && artist == "" && email == "" && password == "" {
		flag.Usage()
	}

	if (email != "" && password == "") || (email == "" && password != "") {
		log.Fatalf("In order to login you must provide both -email and -password.")
		flag.Usage()
	}

	if (album != "" && track != "") || (artist != "" && track != "") || (album != "" && artist != "") {
		log.Fatalf("You can download only one between `album` and `track` at a time.")
		flag.Usage()
	}

	// fmt.Println(asciiArt)
	api.PrintColor(api.COLOR_BLUE, "%s", asciiArt)

	loggedIn, err := api.LoadCookies()

	if err != nil {
		api.PrintColor(api.COLOR_YELLOW, "You're not logged-in, please log in using: -email and -password")
		return
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
		return
	}

	if album != "" {
		album, err := api.NewAlbum(album)

		if err != nil {
			log.Fatalf("Error: %s", err)
		}

		if err := album.Download(true); err != nil {
			log.Fatalf("Cannot download album %s: %s", album.Title, err)
		}
	} else if track != "" {
		track, err := api.NewTrack(track)

		if err != nil {
			log.Fatalf("Error: %s", err)
		}

		if err := track.Download(); err != nil {
			log.Fatalf("Cannot download track %s: %s", track.Title, err)
		}
	} else if artist != "" {
		artist, err := api.NewArtist(artist)

		if err != nil {
			log.Fatalf("Error: %s", err)
		}

		if err := artist.Download(); err != nil {
			log.Fatalf("Cannot download artist %s: %s", artist.Name, err)
		}
	}
}
