package config

import "os"

var Env = map[string]string{
	"DAB_ENDPOINT":      os.Getenv("DAB_ENDPOINT"),
	"DOWNLOAD_LOCATION": os.Getenv("DOWNLOAD_LOCATION"),
	"VERSION":           "1.0.0",
}

func GetEndpoint() string {
	endpoint := Env["DAB_ENDPOINT"]
	if endpoint != "" {
		return endpoint
	}
	return "https://dabmusic.xyz"
}

func GetDownloadLocation() string {
	downloadLocation := Env["DOWNLOAD_LOCATION"]
	if downloadLocation != "" {
		return downloadLocation
	}
	return "."
}

func GetVersion() string {
	return Env["VERSION"]
}
