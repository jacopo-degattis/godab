package api

import (
	"fmt"
	"os"
	"strings"

	"github.com/jacopo-degattis/flacgo"
)

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && info.IsDir()
}

func AddMetadata(targetFile string, metadatas flacgo.FlacMetadatas) error {
	reader, err := flacgo.Open(targetFile)

	if err != nil {
		return fmt.Errorf("unable to initialize flacgo: %w", err)
	}

	err = reader.BulkAddMetadata(metadatas)

	if err != nil {
		return fmt.Errorf("unable to add some meadata: %w", err)
	}

	err = reader.Save(nil)

	if err != nil {
		panic(err)
	}

	return nil
}

func ReplaceNth(s, old, new string, n int) string {
	i := 0
	start := 0

	for count := 0; count < n; count++ {
		idx := strings.Index(s[start:], old)
		if idx == -1 {
			return s // less than n occurrences
		}
		i = start + idx
		start = i + len(old)
	}

	return s[:i] + new + s[i+len(old):]
}
