package api

import (
	"encoding/json"
	"fmt"
	"godab/config"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
)

type Album struct {
	Id          string  `json:"id"`
	Title       string  `json:"title"`
	Artist      string  `json:"artist"`
	Cover       string  `json:"cover"`
	ReleaseDate string  `json:"releaseDate"`
	TrackCount  int     `json:"trackCount"`
	Tracks      []Track `json:"tracks"`
}

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

func (album *Album) downloadAlbum(format int, rc RenderContext) error {
	outputLocation := config.GetDownloadLocation()

	if !DirExists(outputLocation) {
		return fmt.Errorf("specified location for file downloads doesn't exists")
	}

	var rootFolder = fmt.Sprintf("%s/%s", outputLocation, SanitizeFilename(album.Artist))

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

	progressChan := make(chan int, len(album.Tracks))

	for i := range album.Tracks {
		album.Tracks[i].TrackNumber = i + 1
	}

	maxRetries := 3
	maxConcurrent := 3
	var failedTracks []Track
	tracksToDownload := album.Tracks

	trackers := make([]*progress.Tracker, 0)

	switch rc.Mode {
	case ModeAlbumDownload:
		pw := rc.Pw

		pw.SetNumTrackersExpected(len(tracksToDownload))
		pw.Style().Visibility.TrackerOverall = true

		trackers = GetTrackersTrackSizes(tracksToDownload, format)
		pw.AppendTrackers(trackers)

		go pw.Render()
	case ModeArtistDownload:
		rc.Pw.SetNumTrackersExpected(len(tracksToDownload))
	}

	for i := 0; i < maxRetries; i++ {
		if len(tracksToDownload) == 0 {
			break
		}

		if i > 0 {
			PrintColor(COLOR_YELLOW, "\nRetrying %d failed tracks (attempt %d/%d)...\n", len(tracksToDownload), i+1, maxRetries)
		}

		var wg sync.WaitGroup
		sem := make(chan struct{}, maxConcurrent)
		failedTracksChan := make(chan Track, len(tracksToDownload))

		for idx, track := range tracksToDownload {
			wg.Add(1)
			sem <- struct{}{}

			var trk *progress.Tracker = nil
			if rc.Mode == ModeAlbumDownload {
				trk = trackers[idx]
			}

			go func(track Track, tk *progress.Tracker) {
				defer wg.Done()
				defer func() { <-sem }()

				var trackName string
				if track.TrackNumber < 10 {
					trackName = fmt.Sprintf("0%d - %s", track.TrackNumber, SanitizeFilename(track.Title))
				} else {
					trackName = fmt.Sprintf("%d - %s", track.TrackNumber, SanitizeFilename(track.Title))
				}

				fileFormat := "flac"

				if format == 5 {
					fileFormat = "mp3"
				}

				location := fmt.Sprintf("%s/%s.%s", albumLocation, trackName, fileFormat)
				err := track.downloadTrack(location, format, tk)

				if rc.Mode == ModeArtistDownload {
					rc.Tracker.Increment(1)
				}

				if err != nil {
					failedTracksChan <- track
				} else {
					progressChan <- 1
				}
				time.Sleep(time.Duration(rand.Intn(1500)+500) * time.Millisecond)
			}(track, trk)
		}

		wg.Wait()
		close(failedTracksChan)

		failedTracks = nil
		for failedTrack := range failedTracksChan {
			failedTracks = append(failedTracks, failedTrack)
		}

		tracksToDownload = failedTracks
	}

	close(progressChan)

	if len(failedTracks) > 0 {
		var errorMessages []string
		for _, track := range failedTracks {
			errorMessages = append(errorMessages, fmt.Sprintf("'%s' (ID: %d)", track.Title, track.Id))
		}
		return fmt.Errorf("completed with %d errors. Failed to download tracks: %s", len(failedTracks), errorMessages)
	}

	return nil
}

func (album *Album) Download(format int, log bool) error {
	if log {
		PrintColor(COLOR_GREEN, "Starting download for album %s\n", album.Title)
	}

	pw := InitProgress()
	rc := RenderContext{
		Pw:   pw,
		Mode: ModeAlbumDownload,
	}

	err := album.downloadAlbum(format, rc)

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
