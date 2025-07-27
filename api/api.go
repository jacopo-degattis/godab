package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
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

func New(endpoint string, outputLocation string) *DabApi {
	return &DabApi{
		endpoint:       endpoint,
		outputLocation: outputLocation,
	}
}

func (api *DabApi) DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && info.IsDir()
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

	log.Printf("[+] Fetching url %s", fullUrl)

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

func (api *DabApi) Search(query *string, queryType string) AlbumsResults {
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
	defer res.Body.Close()

	var albums AlbumsResults
	err = json.NewDecoder(res.Body).Decode(&albums)

	if err != nil {
		log.Fatalf("[!] Error while decoding albums into struct %s", err)
	}

	return albums
}

func (api *DabApi) DownloadTrack(trackId string, outputName string) {
	log.Printf("[+] Downloading track with id %s and name %s", trackId, outputName)

	res, err := api.Request("api/stream", true, []QueryParams{
		{Name: "trackId", Value: trackId},
		{Name: "quality", Value: "27"},
	})
	defer res.Body.Close()

	// Get stream url from response

	var response StreamUrl
	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		log.Fatalf("[!] Error while decoding streamUrl into a struct")
	}

	res, err = api.Request(response.Url, false, []QueryParams{})
	defer res.Body.Close()

	trackBytes, err := io.ReadAll(res.Body)

	if err != nil {
		log.Fatalf("[!] Error while saving to local file %s", err)
	}

	os.WriteFile(outputName, trackBytes, 0644)

	log.Print("[+] Done")
}

func (api *DabApi) DownloadAlbum(albumId string) {
	res, err := api.Request("api/album", true, []QueryParams{
		{Name: "albumId", Value: albumId},
	})

	if err != nil {
		log.Fatalf("[!] Error while downloading album %s", err)
	}
	defer res.Body.Close()

	log.Printf("[+] Starting download of album with id %s", albumId)

	var response AlbumResponse
	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		log.Fatalf("[!] Error while parsing result into struct: %s", err)
	}

	if !api.DirExists(fmt.Sprintf("%s", api.outputLocation)) {
		log.Fatalf("[!] Specified location for file downloads doesn't exists.")
	}

	var albumLocation = fmt.Sprintf("%s/%s", api.outputLocation, response.Album.Title)

	if api.DirExists(albumLocation) {
		log.Fatalf("[!] Album directory already exists.")
	}

	err = os.Mkdir(albumLocation, 0755)

	if err != nil {
		log.Fatalf("[!] Error while creating directory %s: %s", albumLocation, err)
	}

	for _, track := range response.Album.Tracks {
		api.DownloadTrack(track.ID, fmt.Sprintf("%s/%s.flac", albumLocation, track.Title))

		// TODO: eventually add a random sleep time to bypass rate limiting
		// Maybe also use a more appropriate User-Agent to trick server
	}
}
