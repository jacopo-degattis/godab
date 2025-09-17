package main

import (
	"flag"
	"fmt"
	"godab/api"
	"log"
	"os"
)

func main() {
	serverEndpoint := os.Getenv("DAB_ENDPOINT")
	downloadLocation := os.Getenv("DOWNLOAD_LOCATION")

	if !api.DirExists(downloadLocation) {
		log.Fatalf("You must provide a valid DOWNLOAD_LOCATION folder")
	}

	asciiArt := `
  ____           _       _     
 / ___| ___   __| | __ _| |__  
| |  _ / _ \ / _\` + "`" + ` |/ _\` + "`" + ` | '_ \ 
| |_| | (_) | (_| | (_| | |_) |
 \____|\___/ \__,_|\__,_|_.__/ 
`

	if serverEndpoint == "" {
		panic("You must provide a valid `DAB_ENDPOINT` env variable")
	}

	if downloadLocation == "" {
		panic("You must provide a valid `DOWNLOAD_LOCATION` env variable")
	}

	dapi := api.New(serverEndpoint, downloadLocation)

	var (
		album  string
		track  string
		artist string
	)

	flag.StringVar(&album, "album", "", "Album URL to download")
	flag.StringVar(&track, "track", "", "Track URL to download")
	flag.StringVar(&artist, "artist", "", "Artist URL to download")
	flag.Parse()

	if album == "" && track == "" && artist == "" {
		flag.Usage()
	}

	if (album != "" && track != "") || (artist != "" && track != "") || (album != "" && artist != "") {
		log.Fatalf("You can download only one between `album` and `track` at a time.")
		flag.Usage()
	}

	fmt.Println(asciiArt)

	if album != "" {
		err := dapi.DownloadAlbum(album)

		if err != nil {
			log.Fatalf("Cannot download album %s: %s", album, err)
		}
	} else if track != "" {
		err := dapi.DownloadTrack(track)

		if err != nil {
			log.Fatalf("Cannot download track %s: %s", track, err)
		}
	} else if artist != "" {
		err := dapi.DownloadArtist(artist)

		if err != nil {
			log.Fatalf("Cannot download artist %s: %s", track, err)
		}
	}
}
