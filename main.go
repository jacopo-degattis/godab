package main

import (
	"flag"
	"fmt"
	"godab/api"
	"log"
	"os"
	"strconv"
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
			panic(err)
		}
	} else if track != "" {
		track, err := dapi.GetTrackMetadata(track)

		if err != nil {
			panic(err)
		}

		dapi.DownloadTrack(
			strconv.Itoa(track.ID),
			fmt.Sprintf("%s/%s.flac", downloadLocation, track.Title),
			true,
		)

		dapi.AddMetadata(fmt.Sprintf("%s/%s.flac", downloadLocation, track.Title), api.Metadatas{
			Title:  track.Title,
			Artist: track.Artist,
			Album:  track.Album,
			Date:   track.ReleaseDate,
			Cover:  track.Cover,
		})
	}

	// TODO: add support for entire artist download
}
