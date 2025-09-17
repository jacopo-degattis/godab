package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jacopo-degattis/flacgo"
	"github.com/schollz/progressbar/v3"
)

type ID int

type DabApi struct {
	endpoint       string
	outputLocation string
}

type Track struct {
	Id          ID     `json:"id"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"albumTitle"`
	Cover       string `json:"albumCover"`
	ReleaseDate string `json:"releaseDate"`
	Duration    int    `json:"duration"`
}

type Album struct {
	Id          string  `json:"id"`
	Title       string  `json:"title"`
	Artist      string  `json:"artist"`
	Cover       string  `json:"cover"`
	ReleaseDate string  `json:"releaseDate"`
	Tracks      []Track `json:"tracks"`
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

type QueryParams struct {
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

// Support both album kind of track and single-track search
func (id *ID) UnmarshalJSON(data []byte) error {
	s := string(data)
	s = strings.Trim(s, `"`)
	val, _ := strconv.Atoi(s)
	*id = ID(val)
	return nil
}

func New(endpoint string, outputLocation string) *DabApi {
	return &DabApi{
		endpoint:       endpoint,
		outputLocation: outputLocation,
	}
}

func (api *DabApi) _request(path string, isPathOnly bool, params []QueryParams) (resp *http.Response, err error) {
	var fullUrl string

	if isPathOnly {
		fullUrl = fmt.Sprintf("%s/%s", api.endpoint, path)

		u, err := url.Parse(fullUrl)

		if err != nil {
			return nil, fmt.Errorf("cannot parse url %s", fullUrl)
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
		return nil, fmt.Errorf("can't fetch endpoint %s: %w", fullUrl, err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to %s failed with status code: %s", fullUrl, res.Status)
	}

	return res, nil
}

func (api *DabApi) _addMetadata(targetFile string, metadatas Metadatas) error {
	reader, err := flacgo.Open(targetFile)

	if err != nil {
		return fmt.Errorf("unable to initialize flacgo: %w", err)
	}

	res, err := api._request(metadatas.Cover, false, []QueryParams{})

	if err != nil {
		return fmt.Errorf("can't download cover")
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
		return fmt.Errorf("can't save track with metadata: %w", err)
	}

	return nil
}

func (api *DabApi) Search(query *string, queryType string) (*SearchResults, error) {
	if query == nil || *query == "" {
		return nil, fmt.Errorf("you must provide a valid query parameter")
	}

	if queryType != "album" && queryType != "track" {
		return nil, fmt.Errorf("you must provide a queryType of either type `track` or `album`")
	}

	res, err := api._request("api/search", true, []QueryParams{
		{Name: "q", Value: *query},
		{Name: "type", Value: queryType},
	})
	if err != nil {
		return nil, fmt.Errorf("search endpoint failed with status code: %d", res.StatusCode)
	}
	defer res.Body.Close()

	var searchResponse SearchResults

	switch queryType {
	case "album":
		var albums AlbumsResults
		err = json.NewDecoder(res.Body).Decode(&albums)

		if err != nil {
			return nil, fmt.Errorf("cannot decode response: %w", err)
		}

		searchResponse.Albums = albums
	case "track":
		var response TrackResults
		err = json.NewDecoder(res.Body).Decode(&response)

		if err != nil {
			return nil, fmt.Errorf("cannot decode response: %w", err)
		}

		searchResponse.Tracks = response
	}

	return &searchResponse, nil
}

func (api *DabApi) GetTrackMetadata(trackId string) (Track, error) {
	res, err := api.Search(&trackId, "track")

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

func (api *DabApi) _downloadTrack(track Track, location string, withProgress bool) error {
	type StreamUrl struct {
		Url string `json:"url"`
	}

	res, err := api._request("api/stream", true, []QueryParams{
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

	res, err = api._request(response.Url, false, []QueryParams{})
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

		bar := progressbar.DefaultBytes(
			res.ContentLength,
			fmt.Sprintf("Downloading track with id: %s", strconv.Itoa(int(track.Id))),
		)
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

	err = api._addMetadata(location, Metadatas{
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

func (api *DabApi) DownloadTrack(trackId string) error {
	track, err := api.GetTrackMetadata(trackId)
	if err != nil {
		return fmt.Errorf("track not found")
	}

	var rootFolder = fmt.Sprintf("%s/%s", api.outputLocation, track.Artist)

	// Create artist folder if it doesn't exist
	if !DirExists(rootFolder) {
		os.Mkdir(rootFolder, 0755)
	}

	location := fmt.Sprintf("%s/%s.flac", rootFolder, track.Title)

	if FileExists(location) {
		return fmt.Errorf("track already found at path %s", location)
	}

	err = api._downloadTrack(track, location, true)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (api *DabApi) _downloadAlbum(album Album) error {

	if !DirExists(api.outputLocation) {
		return fmt.Errorf("specified location for file downloads doesn't exists")
	}

	var rootFolder = fmt.Sprintf("%s/%s", api.outputLocation, album.Artist)

	// Create artist folder if it doesn't exist
	if !DirExists(rootFolder) {
		os.Mkdir(rootFolder, 0755)
	}

	var albumLocation = fmt.Sprintf("%s/%s", rootFolder, ReplaceNth(album.Title, " ", "", 2))

	if DirExists(albumLocation) {
		return fmt.Errorf("error: album directory already exists")
	}

	err := os.Mkdir(albumLocation, 0755)

	if err != nil {
		return fmt.Errorf("can't create dir %s", albumLocation)
	}

	bar := progressbar.Default(int64(len(album.Tracks)), fmt.Sprintf("Downloading %s", album.Title))

	maxConcurrent := 3
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrent)

	for _, track := range album.Tracks {
		wg.Add(1)
		sem <- struct{}{}

		go func(track Track) error {
			defer wg.Done()
			defer func() { <-sem }()

			location := fmt.Sprintf("%s/%s.flac", albumLocation, track.Title)
			err = api._downloadTrack(track, location, false)

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

func (api *DabApi) DownloadAlbum(albumId string) error {
	type Response struct {
		Album Album `json:"album"`
	}

	res, err := api._request("api/album", true, []QueryParams{
		{Name: "albumId", Value: albumId},
	})

	if err != nil {
		return fmt.Errorf("error while downloading album %w", err)
	}
	defer res.Body.Close()

	var response Response
	err = json.NewDecoder(res.Body).Decode(&response)

	if response.Album.Id == "0" {
		return fmt.Errorf("album not found")
	}

	if err != nil {
		return fmt.Errorf("failed decoding result: %w", err)
	}

	err = api._downloadAlbum(response.Album)

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (api *DabApi) GetArtistDiscography(artistId string) ([]Album, error) {
	type Response struct {
		Albums []Album `json:"albums"`
	}

	res, err := api._request("/api/discography", true, []QueryParams{
		{Name: "artistId", Value: artistId},
	})

	if err != nil {
		return nil, fmt.Errorf("discography api failed: %w", err)
	}
	defer res.Body.Close()

	var response Response
	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		return nil, fmt.Errorf("failed decoding into struct: %w", err)
	}

	return response.Albums, nil
}

func (api *DabApi) _downloadArtist(albums []Album) error {
	type Response struct {
		Album Album `json:"album"`
	}

	artistName := albums[0].Artist

	var rootFolder = fmt.Sprintf("%s/%s", api.outputLocation, artistName)

	if !DirExists(rootFolder) {
		os.Mkdir(rootFolder, 0755)
	}

	bar := progressbar.Default(int64(len(albums)))

	for _, album := range albums {
		res, err := api._request("api/album", true, []QueryParams{
			{Name: "albumId", Value: album.Id},
		})

		if err != nil {
			return fmt.Errorf("album api failed: %w", err)
		}

		var response Response
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			return fmt.Errorf("failed decoding response: %w", err)
		}

		if err = api._downloadAlbum(response.Album); err != nil {
			return fmt.Errorf("%w", err)
		}

		bar.Add(1)
	}

	return nil
}

func (api *DabApi) DownloadArtist(artistId string) error {
	albums, err := api.GetArtistDiscography(artistId)

	if err != nil {
		return fmt.Errorf("artist discography not found")
	}

	if len(albums) == 0 {
		return fmt.Errorf("artist %s has no albums", artistId)
	}

	err = api._downloadArtist(albums)

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
