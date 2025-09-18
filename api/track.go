package api

import (
	"encoding/json"
	"fmt"
	"godab/config"
	"io"
	"net/http"
	"os"
	"strconv"
)

type Track struct {
	Id          ID     `json:"id"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"albumTitle"`
	Cover       string `json:"albumCover"`
	ReleaseDate string `json:"releaseDate"`
	Duration    int    `json:"duration"`
}

func NewTrack(trackId string) (*Track, error) {
	id, err := strconv.Atoi(trackId)
	if err != nil {
		return nil, fmt.Errorf("invalid track id")
	}

	track := &Track{Id: ID(id)}

	metadata, err := track.GetTrackMetadata()

	if err != nil {
		return nil, fmt.Errorf("track not found")
	}

	*track = metadata

	return track, nil
}

func (track *Track) GetTrackMetadata() (Track, error) {
	trackId := strconv.Itoa(int(track.Id))
	res, err := Search(&trackId, "track")

	if err != nil {
		return Track{}, fmt.Errorf("search api failed: %w", err)
	}

	tracks := res.Tracks

	if len(tracks.Items) == 0 {
		return Track{}, fmt.Errorf("no results found for track id %s", trackId)
	}

	trackData := tracks.Items[0]

	return trackData, nil
}

func (track *Track) downloadTrack(location string, withProgress bool) error {
	type StreamUrl struct {
		Url string `json:"url"`
	}

	res, err := _request("api/stream", true, []QueryParams{
		{Name: "trackId", Value: strconv.Itoa(int(track.Id))},
		{Name: "quality", Value: "27"},
	})
	if err != nil {
		return fmt.Errorf("can't get stream URL: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %d", res.StatusCode)
	}

	var response StreamUrl
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return fmt.Errorf("can't decode stream response: %w", err)
	}

	res, err = _request(response.Url, false, []QueryParams{})
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	// With progress bar
	if withProgress {
		f, err := os.OpenFile(location, os.O_CREATE|os.O_WRONLY, 0644)

		if err != nil {
			return fmt.Errorf("can't create file %s: %w", location, err)
		}

		bar := NewProgressBar(int(res.ContentLength), "TRACK", fmt.Sprintf("Downloading track %s", track.Title), true)
		io.Copy(io.MultiWriter(f, bar), res.Body)
	} else {
		// Without progress bar
		trackBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("can't read from body: %w", err)
		}

		err = os.WriteFile(location, trackBytes, 0644)
		if err != nil {
			return fmt.Errorf("cannot save file: %w", err)
		}
	}

	err = _addMetadata(location, Metadatas{
		Title:  track.Title,
		Artist: track.Artist,
		Album:  track.Album,
		Date:   track.ReleaseDate,
		Cover:  track.Cover,
	})

	if err != nil {
		return fmt.Errorf("cannot add metadata: %w", err)
	}

	return nil
}

func (track *Track) Download() error {
	var rootFolder = fmt.Sprintf("%s/%s", config.GetDownloadLocation(), track.Artist)

	// Create artist folder if it doesn't exist
	if !DirExists(rootFolder) {
		os.Mkdir(rootFolder, 0755)
	}

	location := fmt.Sprintf("%s/%s.flac", rootFolder, track.Title)

	if FileExists(location) {
		return fmt.Errorf("track already found at path %s", location)
	}

	PrintColor(COLOR_GREEN, "Starting download for track %s\n", track.Title)
	err := track.downloadTrack(location, true)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
