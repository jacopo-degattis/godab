package config

import (
	"os"
	"time"
)

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

func GetIdleConnTimeout() time.Duration {
	if val := os.Getenv("IDLE_CONN_TIMEOUT"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return 120 * time.Second
}

func GetTLSHandshakeTimeout() time.Duration {
	if val := os.Getenv("TLS_HANDSHAKE_TIMEOUT"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return 120 * time.Second
}

func GetExpectContinueTimeout() time.Duration {
	if val := os.Getenv("EXPECT_CONTINUE_TIMEOUT"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return 120 * time.Second
}

func GetTimeout() time.Duration {
	if val := os.Getenv("TIMEOUT"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return 120 * time.Second
}
