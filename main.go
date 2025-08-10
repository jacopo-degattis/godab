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

	fmt.Println(asciiArt)

	if album != "" {
		dapi.DownloadAlbum(album)
	} else if track != "" {
		fmt.Println("#TODO: Feature not yet supported.")
	}
}
