package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/jacopo-degattis/flacgo"
	"github.com/schollz/progressbar/v3"
)

type DabApi struct {
	endpoint       string
	outputLocation string
}

type AlbumTrack struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"albumTitle"`
	Cover       string `json:"albumCover"`
	ReleaseDate string `json:"releaseDate"`
	// Make this a integer with seconds or a formatted string "MM:ss"
	Duration int `json:"duration"`
}

type Track struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"albumTitle"`
	Cover       string `json:"albumCover"`
	ReleaseDate string `json:"releaseDate"`
	// Make this a integer with seconds or a formatted string "MM:ss"
	Duration int `json:"duration"`
}

type Album struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Artist      string       `json:"artist"`
	Cover       string       `json:"cover"`
	ReleaseDate string       `json:"releaseDate"`
	Tracks      []AlbumTrack `json:"tracks"`
	// todo: add if is high-res or not?
}

type AlbumsResults struct {
	Items []Album `json:"albums"`
}

type TrackResults struct {
	Items []Track `json:"tracks"`
}

type SearchResults struct {
	Tracks TrackResults
	Albums AlbumsResults
}

type AlbumResponse struct {
	Album Album `json:"album"`
}

type StreamUrl struct {
	Url string `json:"url"`
}

type QueryParams struct {
	Name  string
	Value string
}

type Metadata struct {
	Name  string
	Value string
}

type Metadatas struct {
	Title  string
	Artist string
	Album  string
	Date   string
	Cover  string
}

func New(endpoint string, outputLocation string) *DabApi {
	return &DabApi{
		endpoint:       endpoint,
		outputLocation: outputLocation,
	}
}

func (api *DabApi) Request(path string, isPathOnly bool, params []QueryParams) (resp *http.Response, err error) {
	var fullUrl string

	if isPathOnly {
		fullUrl = fmt.Sprintf("%s/%s", api.endpoint, path)

		u, err := url.Parse(fullUrl)

		if err != nil {
			return nil, fmt.Errorf("error while parsing url %s: %w", fullUrl, err)
		}

		q := u.Query()
		for _, queryParam := range params {
			q.Set(queryParam.Name, queryParam.Value)
		}
		u.RawQuery = q.Encode()

		fullUrl = u.String()
	} else {
		fullUrl = path
	}

	res, err := http.Get(fullUrl)

	if err != nil {
		return nil, fmt.Errorf("error while fetching endpoint %s with error %w", fullUrl, err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to %s failed with status code: %s", fullUrl, res.Status)
	}

	return res, nil
}

func (api *DabApi) Search(query *string, queryType string) (*SearchResults, error) {
	if query == nil || *query == "" {
		return nil, fmt.Errorf("error in search(): you must provide a valid query parameter")
	}

	if queryType != "album" && queryType != "track" {
		return nil, fmt.Errorf("error in search(): you must provide a queryType of either type `track` or `album`")
	}

	res, err := api.Request("api/search", true, []QueryParams{
		{Name: "q", Value: *query},
		{Name: "type", Value: queryType},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch endpoint /api/search with status code: %d", res.StatusCode)
	}
	defer res.Body.Close()

	var searchResponse SearchResults

	switch queryType {
	case "album":
		var albums AlbumsResults
		err = json.NewDecoder(res.Body).Decode(&albums)

		if err != nil {
			return nil, fmt.Errorf("error while decoding albums into struct %w", err)
		}

		searchResponse.Albums = albums
	case "track":
		var response TrackResults
		err = json.NewDecoder(res.Body).Decode(&response)

		if err != nil {
			return nil, fmt.Errorf("error while decoding tracks into struct %w", err)
		}

		searchResponse.Tracks = response
	}

	return &searchResponse, nil
}

func (api *DabApi) GetTrackMetadata(trackId string) (Track, error) {
	res, err := api.Search(&trackId, "track")

	if err != nil {
		return Track{}, fmt.Errorf("error whlie searching for track id %s with error: %w", trackId, err)
	}

	tracks := res.Tracks

	if len(tracks.Items) == 0 {
		return Track{}, fmt.Errorf("no results found for track id %s", trackId)
	}

	trackData := tracks.Items[0]

	return trackData, nil
}

func (api *DabApi) DownloadTrack(trackId string, outputName string, withProgress bool) error {
	res, err := api.Request("api/stream", true, []QueryParams{
		{Name: "trackId", Value: trackId},
		{Name: "quality", Value: "27"},
	})
	if err != nil {
		return fmt.Errorf("unable to fetch endpoint /api/stream with status code: %d", res.StatusCode)
	}
	defer res.Body.Close()

	var response StreamUrl
	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		log.Fatalf("[!] Error while decoding streamUrl into a struct")
	}

	res, err = api.Request(response.Url, false, []QueryParams{})
	if err != nil {
		return fmt.Errorf("unable to fetch full url: '%s' with status code: %d", response.Url, res.StatusCode)
	}
	defer res.Body.Close()

	// With progress bar
	if withProgress {
		f, _ := os.OpenFile(outputName, os.O_CREATE|os.O_WRONLY, 0644)
		bar := progressbar.DefaultBytes(
			res.ContentLength,
			fmt.Sprintf("Downloading track with id: %s", trackId),
		)
		io.Copy(io.MultiWriter(f, bar), res.Body)
		return nil
	}

	// Without progress bar
	trackBytes, err := io.ReadAll(res.Body)

	if err != nil {
		log.Fatalf("[!] Error while saving to local file %s", err)
	}

	os.WriteFile(outputName, trackBytes, 0644)

	return nil
}

func (api *DabApi) AddMetadata(targetFile string, metadatas Metadatas) error {
	reader, err := flacgo.Open(targetFile)

	if err != nil {
		return fmt.Errorf("unable to initialize flacgo: %w", err)
	}

	res, err := api.Request(metadatas.Cover, false, []QueryParams{})

	if err != nil {
		return fmt.Errorf("unable to fetch full url '%s' with status code: %d", metadatas.Cover, res.StatusCode)
	}

	defer res.Body.Close()

	coverBytes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	err = reader.BulkAddMetadata(flacgo.FlacMetadatas{
		Title:  metadatas.Title,
		Artist: metadatas.Artist,
		Album:  metadatas.Album,
		Date:   metadatas.Date,
		Cover:  coverBytes,
	})

	if err != nil {
		return fmt.Errorf("unable to add some meadata: %w", err)
	}

	err = reader.Save(nil)

	if err != nil {
		return fmt.Errorf("error while saving track with metadata: %w", err)
	}

	return nil
}

func (api *DabApi) DownloadAlbum(albumId string) error {
	res, err := api.Request("api/album", true, []QueryParams{
		{Name: "albumId", Value: albumId},
	})

	if err != nil {
		return fmt.Errorf("error while downloading album %w", err)
	}
	defer res.Body.Close()

	var response AlbumResponse
	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		return fmt.Errorf("error while parsing result into struct: %w", err)
	}

	if !DirExists(api.outputLocation) {
		return fmt.Errorf("specified location for file downloads doesn't exists")
	}

	var albumLocation = fmt.Sprintf("%s%s", api.outputLocation, ReplaceNth(response.Album.Title, " ", "", 2))

	if DirExists(albumLocation) {
		return fmt.Errorf("album directory already exists")
	}

	err = os.Mkdir(albumLocation, 0755)

	if err != nil {
		return fmt.Errorf("while creating directory %s: %w", albumLocation, err)
	}

	bar := progressbar.Default(int64(len(response.Album.Tracks)))

	maxConcurrent := 3 // only 3 downloads at once
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrent)

	for _, track := range response.Album.Tracks {
		wg.Add(1)
		sem <- struct{}{}

		go func(track AlbumTrack) error {
			defer wg.Done()
			defer func() { <-sem }()

			outputName := fmt.Sprintf("%s/%s.flac", albumLocation, track.Title)
			api.DownloadTrack(track.ID, outputName, false)

			err = api.AddMetadata(outputName, Metadatas{
				Title:  track.Title,
				Artist: track.Artist,
				Album:  response.Album.Title,
				Date:   track.ReleaseDate,
				Cover:  track.Cover,
			})

			if err != nil {
				panic(err)
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
