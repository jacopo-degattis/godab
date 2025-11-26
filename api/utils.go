package api

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/table"
)

type Color int

var colorMapping = map[Color]string{
	COLOR_RESET:  "\033[0m",
	COLOR_RED:    "\033[31m",
	COLOR_GREEN:  "\033[32m",
	COLOR_YELLOW: "\033[33m",
	COLOR_BLUE:   "\033[34m",
	COLOR_PURPLE: "\033[35m",
	COLOR_CYAN:   "\033[36m",
	COLOR_GRAY:   "\033[37m",
	COLOR_WHITE:  "\033[97m",
}

const (
	COLOR_RESET Color = iota
	COLOR_RED
	COLOR_GREEN
	COLOR_YELLOW
	COLOR_BLUE
	COLOR_PURPLE
	COLOR_CYAN
	COLOR_GRAY
	COLOR_WHITE
)

var FormatMap = map[string]int{
	"mp3":  5,
	"flac": 27,
}

func PrintError(msg string) {
	PrintColor(COLOR_RED, "%s", msg)
	os.Exit(1)
}

func CheckErr(err error) {
	if err != nil {
		PrintError(err.Error())
	}
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && info.IsDir()
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func PrintColor(color Color, format string, args ...any) {
	statement := fmt.Sprintf(format, args...)
	println(colorMapping[color] + statement + colorMapping[COLOR_RESET])
}

func SanitizeFilename(filename string) string {
	badCharacters := []string{"\\", "/", "<", ">", "?", "*", "|", "\"", ":"}
	sanitized := filename

	for _, char := range badCharacters {
		sanitized = strings.ReplaceAll(sanitized, char, "_")
	}

	sanitized = strings.Trim(sanitized, " .")
	if len(sanitized) > 255 {
		sanitized = sanitized[:255]
	}

	return sanitized
}

func PrintResultsTable(results *SearchResults, resultType string) {
	var (
		colIndex       = "#"
		colId          = "Track ID"
		colTitle       = "Title"
		colArtist      = "Artist"
		colReleaseDate = "Release date"
	)

	tw := table.NewWriter()

	switch resultType {
	case "track":
		tw.AppendHeader(table.Row{colIndex, colId, colTitle, colArtist, colReleaseDate})
		for idx, track := range results.Tracks.Items {
			tw.AppendRow(table.Row{idx, track.Id, track.Title, track.Artist, track.ReleaseDate})
		}
	case "album":
		tw.AppendHeader(table.Row{colIndex, "Album ID", colTitle, colArtist, colReleaseDate})
		for idx, album := range results.Albums.Items {
			tw.AppendRow(table.Row{idx, album.Id, album.Title, album.Artist, album.ReleaseDate})
		}
	case "artist":
		tw.AppendHeader(table.Row{colIndex, "Artist ID", "Name"})
		for idx, artist := range results.Artists.Items {
			tw.AppendRow(table.Row{idx, artist.Id, artist.Name})
		}
	}

	fmt.Println(tw.Render())
}
