package main

import (
	"flag"
	"godab/api"
	"log"
)

func main() {
	// Replace URL and DOWNLOAD_PATH with a ENV variable
	dapi := api.New("https://dab.yeet.su", "<DOWNLOAD_PATH>")

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
