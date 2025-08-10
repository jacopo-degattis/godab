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

type Track struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Cover       string `json:"albumCover"`
	ReleaseDate string `json:"releaseDate"`
	// Make this a integer with seconds or a formatted string "MM:ss"
	Duration int `json:"duration"`
}

type AlbumsResults struct {
	Items []Album `json:"albums"`
}

type Album struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Artist      string  `json:"artist"`
	Cover       string  `json:"cover"`
	ReleaseDate string  `json:"releaseDate"`
	Tracks      []Track `json:"tracks"`
	// todo: add if is high-res or not?
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

type TrackMetadataInfos struct {
	Title  string
	Artist string
	Album  string
	Date   string
	Cover  []byte
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
			log.Fatalf("[!] Error while parsing url %s", fullUrl)
			return nil, err
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
		log.Fatalf("[!] Error while fetching endpoint %s with error %s", fullUrl, err)
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		log.Fatalf("[!] Request to %s failed with status code %s", fullUrl, res.Status)
		return nil, err
	}

	return res, nil
}

func (api *DabApi) Search(query *string, queryType string) (*AlbumsResults, error) {
	if query == nil || *query == "" {
		log.Fatalf("[!] Error in search(): You must provide a valid query parameter.")
	}

	if queryType != "album" && queryType != "track" {
		log.Fatalf("[!] Error in search(): you must provide a queryType of either type `track` or `album`")
	}

	res, err := api.Request("api/search", true, []QueryParams{
		{Name: "q", Value: *query},
		{Name: "type", Value: queryType},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch endpoint /api/search with status code: %d", res.StatusCode)
	}
	defer res.Body.Close()

	var albums AlbumsResults
	err = json.NewDecoder(res.Body).Decode(&albums)

	if err != nil {
		log.Fatalf("[!] Error while decoding albums into struct %s", err)
	}

	return &albums, nil
}

func (api *DabApi) DownloadTrack(trackId string, outputName string) error {
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

	trackBytes, err := io.ReadAll(res.Body)

	if err != nil {
		log.Fatalf("[!] Error while saving to local file %s", err)
	}

	os.WriteFile(outputName, trackBytes, 0644)

	return nil
}

func (api *DabApi) DownloadAlbum(albumId string) error {
	res, err := api.Request("api/album", true, []QueryParams{
		{Name: "albumId", Value: albumId},
	})

	if err != nil {
		log.Fatalf("[!] Error while downloading album %s", err)
	}
	defer res.Body.Close()

	var response AlbumResponse
	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		log.Fatalf("[!] Error while parsing result into struct: %s", err)
	}

	if !DirExists(api.outputLocation) {
		log.Fatalf("[!] Specified location for file downloads doesn't exists.")
	}

	var albumLocation = fmt.Sprintf("%s%s", api.outputLocation, ReplaceNth(response.Album.Title, " ", "", 2))

	if DirExists(albumLocation) {
		log.Fatalf("[!] Album directory already exists.")
	}

	err = os.Mkdir(albumLocation, 0755)

	if err != nil {
		log.Fatalf("[!] Error while creating directory %s: %s", albumLocation, err)
	}

	fmt.Printf("[+] Downloading album: %s by %s\n", response.Album.Title, response.Album.Artist)

	bar := progressbar.Default(int64(len(response.Album.Tracks)))

	maxConcurrent := 3 // only 3 downloads at once
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrent)

	for _, track := range response.Album.Tracks {
		wg.Add(1)
		sem <- struct{}{}

		go func(track Track) error {
			defer wg.Done()
			defer func() { <-sem }()

			outputName := fmt.Sprintf("%s/%s.flac", albumLocation, track.Title)
			api.DownloadTrack(track.ID, outputName)

			res, err = api.Request(track.Cover, false, []QueryParams{})

			if err != nil {
				return fmt.Errorf("unable to fetch full url '%s' with status code: %d", track.Cover, res.StatusCode)
			}

			defer res.Body.Close()

			coverBytes, err := io.ReadAll(res.Body)
			if err != nil {
				panic(err)
			}

			err = AddMetadata(outputName, flacgo.FlacMetadatas{
				Title:  track.Title,
				Artist: track.Artist,
				Album:  response.Album.Title,
				Date:   track.ReleaseDate,
				Cover:  coverBytes,
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
