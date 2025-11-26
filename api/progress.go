package api

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
)

func InitProgress() progress.Writer {
	pw := progress.NewWriter()
	pw.SetOutputWriter(os.Stdout)
	pw.SetTrackerLength(40)
	pw.SetMessageLength(25)
	pw.SetUpdateFrequency(time.Millisecond * 100)
	pw.SetStyle(progress.StyleDefault)
	pw.SetTrackerPosition(progress.PositionRight)
	pw.Style().Colors = progress.StyleColorsExample
	pw.Style().Options.PercentFormat = "%4.1f%%"
	pw.Style().Visibility.ETA = true
	pw.Style().Visibility.Percentage = true
	pw.Style().Visibility.Speed = true
	pw.Style().Visibility.SpeedOverall = true
	pw.Style().Visibility.Time = true
	pw.Style().Visibility.TrackerOverall = true
	pw.Style().Visibility.Value = true
	pw.Style().Visibility.Pinned = true
	pw.Style().Options.TimeInProgressPrecision = time.Second

	return pw
}

func GetTrackersTrackSizes(tracks []Track, format int) []*progress.Tracker {
	trackers := make([]*progress.Tracker, len(tracks))
	var wg sync.WaitGroup

	for i, t := range tracks {
		wg.Add(1)
		go func(i int, track Track) {
			defer wg.Done()

			url, err := track.GetDownloadUrl(format)
			if err != nil {
				return
			}

			res, err := http.Head(url)
			if err != nil {
				return
			}
			defer res.Body.Close()

			size, err := strconv.Atoi(res.Header.Get("Content-Length"))
			if err != nil {
				return
			}

			trackers[i] = &progress.Tracker{
				Message: track.Title,
				Total:   int64(size),
				Units:   progress.UnitsBytes,
			}
		}(i, t)
	}

	wg.Wait()
	return trackers
}
