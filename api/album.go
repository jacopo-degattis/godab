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

	var albumLocation = fmt.Sprintf("%s/%s", rootFolder, SanitizeFilename(album.Title))

	if DirExists(albumLocation) {
		return fmt.Errorf("album directory already exists")
	}

	err := os.Mkdir(albumLocation, 0755)
	if err != nil {
		return fmt.Errorf("can't create dir %s", albumLocation)
	}

	bar := NewProgressBar(len(album.Tracks), "ALBUM", fmt.Sprintf("Downloading %s", album.Title), false)
	bar.RenderBlank()

	progressChan := make(chan int, len(album.Tracks))
	go func() {
		for range progressChan {
			bar.Add(1)
		}
	}()

	for i := range album.Tracks {
		album.Tracks[i].TrackNumber = i + 1
	}

	tracksToDownload := album.Tracks
	var failedTracks []Track
	maxRetries := 3
	maxConcurrent := 3

	for i := 0; i < maxRetries; i++ {
		if len(tracksToDownload) == 0 {
			break // All tracks downloaded successfully
		}

		if i > 0 {
			PrintColor(COLOR_YELLOW, "\nRetrying %d failed tracks (attempt %d/%d)...\n", len(tracksToDownload), i+1, maxRetries)
		}

		var wg sync.WaitGroup
		sem := make(chan struct{}, maxConcurrent)
		failedTracksChan := make(chan Track, len(tracksToDownload))

		for _, track := range tracksToDownload {
			wg.Add(1)
			sem <- struct{}{}

			go func(track Track) {
				defer wg.Done()
				defer func() { <-sem }()

				var trackName string
				if track.TrackNumber < 10 {
					trackName = fmt.Sprintf("0%d - %s", track.TrackNumber, SanitizeFilename(track.Title))
				} else {
					trackName = fmt.Sprintf("%d - %s", track.TrackNumber, SanitizeFilename(track.Title))
				}

				location := fmt.Sprintf("%s/%s.flac", albumLocation, trackName)
				err := track.downloadTrack(location, false)

				if err != nil {
					failedTracksChan <- track
				} else {
					progressChan <- 1
				}
				time.Sleep(time.Duration(rand.Intn(1500)+500) * time.Millisecond)
			}(track)
		}

		wg.Wait()
		close(failedTracksChan)

		failedTracks = nil // Reset the list for the current retry attempt
		for failedTrack := range failedTracksChan {
			failedTracks = append(failedTracks, failedTrack)
		}

		tracksToDownload = failedTracks
	}

	close(progressChan)
	bar.Finish()

	if len(failedTracks) > 0 {
		var errorMessages []string
		for _, track := range failedTracks {
			errorMessages = append(errorMessages, fmt.Sprintf("'%s' (ID: %d)", track.Title, track.Id))
		}
		return fmt.Errorf("completed with %d errors. Failed to download tracks: %s", len(failedTracks), errorMessages)
	}

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
