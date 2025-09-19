package config

import "os"

var Env = map[string]string{
	"DAB_ENDPOINT":      os.Getenv("DAB_ENDPOINT"),
	"DOWNLOAD_LOCATION": os.Getenv("DOWNLOAD_LOCATION"),
}

func GetEndpoint() string {
	endpoint := Env["DAB_ENDPOINT"]
	if endpoint != "" {
		return endpoint
	}
	return "https://dab.yeet.su"
}

func GetDownloadLocation() string {
	downloadLocation := Env["DOWNLOAD_LOCATION"]
	if downloadLocation != "" {
		return downloadLocation
	}
	return "."
}
