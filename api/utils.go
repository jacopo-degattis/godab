package api

import (
	"os"
	"strings"
)

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && info.IsDir()
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
