package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"godab/config"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jacopo-degattis/flacgo"
)

type ID int

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

func _request(path string, isPathOnly bool, params []QueryParams) (resp *http.Response, err error) {
	var fullUrl string

	if isPathOnly {
		fullUrl = fmt.Sprintf("%s/%s", config.GetEndpoint(), path)

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

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := &http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest(http.MethodGet, fullUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("can't create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")

	res, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("can't fetch endpoint %s: %w", fullUrl, err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to %s failed with status code: %s", fullUrl, res.Status)
	}

	return res, nil
}

func _addMetadata(targetFile string, metadatas Metadatas) error {
	reader, err := flacgo.Open(targetFile)

	if err != nil {
		return fmt.Errorf("unable to initialize flacgo: %w", err)
	}

	res, err := _request(metadatas.Cover, false, []QueryParams{})

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

func Search(query *string, queryType string) (*SearchResults, error) {
	if query == nil || *query == "" {
		return nil, fmt.Errorf("you must provide a valid query parameter")
	}

	if queryType != "album" && queryType != "track" {
		return nil, fmt.Errorf("you must provide a queryType of either type `track` or `album`")
	}

	res, err := _request("api/search", true, []QueryParams{
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
