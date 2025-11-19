package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"godab/config"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.senan.xyz/taglib"
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
	Title       string
	Artist      string
	Album       string
	Date        string
	Cover       string
	TrackNumber int
}

var jar, _ = cookiejar.New(nil)
var client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	},
	Jar:     jar,
	Timeout: 30 * time.Second,
}

// Support both album kind of track and single-track search
func (id *ID) UnmarshalJSON(data []byte) error {
	s := string(data)
	s = strings.Trim(s, `"`)
	val, _ := strconv.Atoi(s)
	*id = ID(val)
	return nil
}

func LoadCookies() (bool, error) {
	if _, err := os.Stat(".token"); errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	path := filepath.Join(".token")
	dat, err := os.ReadFile(path)

	if err != nil {
		return false, fmt.Errorf("unable to read .token file")
	}

	url, _ := url.Parse(config.GetEndpoint())
	client.Jar.SetCookies(url, []*http.Cookie{
		{
			Name:  "session",
			Value: string(dat),
		},
	})

	return dat != nil, nil
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
	res, err := _request(metadatas.Cover, false, []QueryParams{})

	if err != nil {
		return fmt.Errorf("can't download cover")
	}

	defer res.Body.Close()

	coverBytes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	err = taglib.WriteTags(targetFile, map[string][]string{
		taglib.Title:  {metadatas.Title},
		taglib.Artist: {metadatas.Artist},
		taglib.Album:  {metadatas.Album},
		taglib.Date:   {metadatas.Date},
	}, 0)

	if err != nil {
		return fmt.Errorf("unable to write metadata to track")
	}

	if len(coverBytes) > 0 {
		err = taglib.WriteImage(targetFile, coverBytes)
	}

	if err != nil {
		return fmt.Errorf("unable to picture meadata to track")
	}

	return nil
}

func Login(email string, password string) error {
	if email == "" || password == "" {
		return fmt.Errorf("invalid email or password")
	}

	body := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    email,
		Password: password,
	}

	out, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("cannot encode login body")
	}

	res, err := http.Post(
		fmt.Sprintf("%s/%s", config.GetEndpoint(), "api/auth/login"),
		"application/json",
		bytes.NewBuffer(out),
	)

	if res.StatusCode == 401 {
		return fmt.Errorf("invalid credentials")
	}

	if res.StatusCode == 200 {
		var token = ""
		for _, cook := range res.Cookies() {
			if cook.Name == "session" {
				token = cook.Value
			}
		}

		if token == "" {
			return fmt.Errorf("unable to get token")
		}

		path := filepath.Join(".token")
		err := os.WriteFile(path, []byte(token), 0644)

		if err != nil {
			return fmt.Errorf("unable to write .token file")
		}

		url, _ := url.Parse(config.GetEndpoint())
		client.Jar.SetCookies(url, []*http.Cookie{
			{
				Name:  "session",
				Value: token,
			},
		})

	}

	if err != nil {
		return fmt.Errorf("error while making request: %w", err)
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
		return nil, fmt.Errorf("search endpoint failed: %s", err)
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
