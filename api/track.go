package api

import (
	"encoding/json"
	"fmt"
	"godab/config"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/progress"
)

type ProgressReader struct {
	Reader   io.Reader
	Tracker  *progress.Tracker
	Progress int64
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Progress += int64(n)
	pr.Tracker.SetValue(pr.Progress)
	return n, err
}

type Track struct {
	Id          ID     `json:"id"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"albumTitle"`
	Cover       string `json:"albumCover"`
	ReleaseDate string `json:"releaseDate"`
	Duration    int    `json:"duration"`
	TrackNumber int    `json:"-"`
}

func NewTrack(trackId string) (*Track, error) {
	id, err := strconv.Atoi(trackId)
	if err != nil {
		return nil, fmt.Errorf("invalid track id")
	}

	track := &Track{Id: ID(id)}

	metadata, err := track.GetTrackMetadata()

	if err != nil {
		fmt.Print(err)
		return nil, fmt.Errorf("track not found")
	}

	*track = metadata

	return track, nil
}

func (track *Track) TrackProgress(tk *progress.Tracker, res *http.Response, file *os.File) {
	pr := &ProgressReader{
		Reader:  res.Body,
		Tracker: tk,
	}

	_, err := io.Copy(file, pr)

	if err != nil {
		panic(err)
	}
}

func (track *Track) GetTrackMetadata() (Track, error) {
	trackId := strconv.Itoa(int(track.Id))
	res, err := Search(trackId, "track")

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

func (track *Track) GetDownloadUrl(format int) (string, error) {
	type StreamUrl struct {
		Url string `json:"url"`
	}

	res, err := _request("api/stream", true, []QueryParams{
		{Name: "trackId", Value: strconv.Itoa(int(track.Id))},
		{Name: "quality", Value: fmt.Sprint(format)},
	})
	if err != nil {
		return "", fmt.Errorf("can't get stream URL: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status: %d", res.StatusCode)
	}

	var response StreamUrl
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("can't decode stream response: %w", err)
	}

	return response.Url, nil
}

func (track *Track) downloadTrack(location string, format int, tk *progress.Tracker) error {
	streamUrl, err := track.GetDownloadUrl(format)

	if err != nil {
		return fmt.Errorf("unable to fetch stream url")
	}

	res, err := _request(streamUrl, false, []QueryParams{})
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	// With progress bar
	if tk != nil {
		out, err := os.OpenFile(location, os.O_CREATE|os.O_WRONLY, 0644)

		if err != nil {
			return fmt.Errorf("can't create file %s: %w", location, err)
		}

		// Delegate rendering the progress bar to parent which is invoking the function
		track.TrackProgress(tk, res, out)
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
		Title:       track.Title,
		Artist:      track.Artist,
		Album:       track.Album,
		Date:        track.ReleaseDate,
		Cover:       track.Cover,
		TrackNumber: track.TrackNumber,
	})

	if err != nil {
		return fmt.Errorf("cannot add metadata: %w", err)
	}

	return nil
}

func (track *Track) Download(format int) error {
	var rootFolder = fmt.Sprintf("%s/%s", config.GetDownloadLocation(), SanitizeFilename(track.Artist))

	// Create artist folder if it doesn't exist
	if !DirExists(rootFolder) {
		os.Mkdir(rootFolder, 0755)
	}

	fileFormat := "flac"

	if format == 5 {
		fileFormat = "mp3"
	}

	location := fmt.Sprintf("%s/%s.%s", rootFolder, SanitizeFilename(track.Title), fileFormat)

	if FileExists(location) {
		return fmt.Errorf("track already found at path %s", location)
	}

	PrintColor(COLOR_GREEN, "Starting download for track %s\n", track.Title)

	pw := InitProgress()
	sizes := GetTrackersTrackSizes([]Track{*track}, format)

	if len(sizes) == 0 {
		return fmt.Errorf("unable to get size of track: %s", track.Title)
	}

	pw.AppendTracker(sizes[0])
	go pw.Render()

	err := track.downloadTrack(location, format, sizes[0])

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
