package config

import "os"

var Env = map[string]string{
	"DAB_ENDPOINT":      os.Getenv("DAB_ENDPOINT"),
	"DOWNLOAD_LOCATION": os.Getenv("DOWNLOAD_LOCATION"),
}

func GetEndpoint() string {
	return Env["DAB_ENDPOINT"]
}

func GetDownloadLocation() string {
	return Env["DOWNLOAD_LOCATION"]
}
