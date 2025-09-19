package api

import (
	"encoding/json"
	"fmt"
	"godab/config"
	"math/rand"
	"os"
	"sync"
	"time"
)

type Album struct {
	Id          string  `json:"id"`
	Title       string  `json:"title"`
	Artist      string  `json:"artist"`
	Cover       string  `json:"cover"`
	ReleaseDate string  `json:"releaseDate"`
	Tracks      []Track `json:"tracks"`
}

// outputLocation and DAB_ENDPOINT should be loaded from a config or from a .env file to fix the api reference problem

func NewAlbum(albumId string) (*Album, error) {
	type Response struct {
		Album Album `json:"album"`
	}

	res, err := _request("api/album", true, []QueryParams{
		{Name: "albumId", Value: albumId},
	})

	if err != nil {
		return nil, fmt.Errorf("cannot fetch album (reason: %w)", err)
	}
	defer res.Body.Close()

	var response Response
	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		return nil, fmt.Errorf("cannot decode result: %w", err)
	}

	if response.Album.Id == "0" {
		return nil, fmt.Errorf("album not found")
	}

	return &response.Album, nil
}

func (album *Album) downloadAlbum() error {
	outputLocation := config.GetDownloadLocation()

	if !DirExists(outputLocation) {
		return fmt.Errorf("specified location for file downloads doesn't exists")
	}

	var rootFolder = fmt.Sprintf("%s/%s", outputLocation, SanitizeFilename(album.Artist))

	// Create artist folder if it doesn't exist
	if !DirExists(rootFolder) {
		os.Mkdir(rootFolder, 0755)
	}

	var albumLocation = fmt.Sprintf("%s/%s", rootFolder, ReplaceNth(SanitizeFilename(album.Title), " ", "", 2))

	if DirExists(albumLocation) {
		return fmt.Errorf("album directory already exists")
	}

	err := os.Mkdir(albumLocation, 0755)

	if err != nil {
		return fmt.Errorf("can't create dir %s", albumLocation)
	}

	bar := NewProgressBar(len(album.Tracks), "ALBUM", fmt.Sprintf("Downloading %s", album.Title), false)

	maxConcurrent := 3
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrent)

	for _, track := range album.Tracks {
		wg.Add(1)
		sem <- struct{}{}

		go func(track Track) error {
			defer wg.Done()
			defer func() { <-sem }()

			location := fmt.Sprintf("%s/%s.flac", albumLocation, SanitizeFilename(track.Title))
			err = track.downloadTrack(location, false)

			if err != nil {
				return fmt.Errorf("can't download track %d", track.Id)
			}

			bar.Add(1)
			time.Sleep(time.Millisecond)
			time.Sleep(time.Duration(rand.Intn(1500)+500) * time.Millisecond)

			return nil
		}(track)
	}

	wg.Wait()
	return nil
}

func (album *Album) Download(log bool) error {
	if log {
		PrintColor(COLOR_GREEN, "Starting download for album %s\n", album.Title)
	}
	err := album.downloadAlbum()

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
