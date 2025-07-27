package main

import (
	"flag"
	"godab/api"
	"log"
	"os"
)

func main() {
	serverEndpoint := os.Getenv("DAB_ENDPOINT")
	downloadLocation := os.Getenv("DOWNLOAD_LOCATION")

	if serverEndpoint == "" {
		panic("You must provide a valid `DAB_ENDPOINT` env variable")
	}

	if downloadLocation == "" {
		panic("You must provide a valid `DOWNLOAD_LOCATION` env variable")
	}

	// Replace URL and DOWNLOAD_PATH with a ENV variable
	dapi := api.New(serverEndpoint, downloadLocation)

	var (
		album string
		track string
	)

	flag.StringVar(&album, "album", "", "Album URL to download")
	flag.StringVar(&track, "track", "", "Track URL to download")
	flag.Parse()

	if album == "" && track == "" {
		flag.Usage()
	}

	if album != "" && track != "" {
		log.Fatalf("You can download only one between `album` and `track` at a time.")
		flag.Usage()
	}

	if album != "" {
		dapi.DownloadAlbum(album)
	} else if track != "" {
		// TODO: first I need to fetch track infos in order to get filename and then use it
		// dapi.DownloadTrack(*track)
	}
}
