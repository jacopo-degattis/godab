package api

import (
	"encoding/json"
	"fmt"
	"godab/config"
	"os"

	"github.com/jedib0t/go-pretty/v6/progress"
)

type Artist struct {
	Id          ID     `json:"id"`
	Name        string `json:"name"`
	AlbumsCount int    `json:"albumsCount"`
	Albums      []Album
}

type RenderMode int

const (
	ModeDefault RenderMode = iota
	ModeTrackDownload
	ModeAlbumDownload
	ModeArtistDownload
)

type RenderContext struct {
	Pw      progress.Writer
	Mode    RenderMode
	Tracker *progress.Tracker
}

func NewArtist(artistId string) (*Artist, error) {
	type Response struct {
		Artist Artist  `json:"artist"`
		Albums []Album `json:"albums"`
	}

	res, err := _request("/api/discography", true, []QueryParams{
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

	response.Artist.Albums = response.Albums

	return &response.Artist, nil
}

func (artist *Artist) downloadArtist(format int, rc RenderContext) error {
	type Response struct {
		Artist Artist `json:"artist"`
		Album  Album  `json:"album"`
	}

	var rootFolder = fmt.Sprintf("%s/%s", config.GetDownloadLocation(), artist.Name)

	if !DirExists(rootFolder) {
		os.Mkdir(rootFolder, 0755)
	}

	// bar := progressbar.Default(int64(len(artist.Albums)))
	trackers := make([]*progress.Tracker, 0)
	for _, album := range artist.Albums {
		trk := &progress.Tracker{
			Message: album.Title,
			Total:   int64(album.TrackCount),
			Units:   progress.UnitsDefault,
		}
		trackers = append(trackers, trk)
	}

	pw := rc.Pw

	go pw.Render()
	pw.AppendTrackers(trackers)

	for idx, album := range artist.Albums {
		res, err := _request("api/album", true, []QueryParams{
			{Name: "albumId", Value: album.Id},
		})

		if err != nil {
			return fmt.Errorf("album api failed: %w", err)
		}

		var response Response
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			return fmt.Errorf("failed decoding response: %w", err)
		}

		rc.Tracker = trackers[idx]
		if err = response.Album.downloadAlbum(format, rc); err != nil {
			return fmt.Errorf("%w", err)
		}

		// bar.Add(1)
	}

	return nil
}

func (artist *Artist) Download(format int) error {
	if len(artist.Albums) == 0 {
		return fmt.Errorf("artist %d has no albums", artist.Id)
	}

	pw := InitProgress()
	rc := RenderContext{
		Pw:   pw,
		Mode: ModeArtistDownload,
	}

	PrintColor(COLOR_GREEN, "Starting download for artist %s\n", artist.Name)
	err := artist.downloadArtist(format, rc)

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
