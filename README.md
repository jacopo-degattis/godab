# Godab

## Description

A [dab.yeet.su](https://dab.yeet.su) CLI downloader written in go.

## Build

In order to create a binary from the given source you can use

```sh
$ go build
$ ./godab
```

Otherwise you can always run it directly with

```sh
$ go run main.go
```

N.B: Remember that you always have to define two env variables
- DOWNLOAD_LOCATION: to specify the location where you want your files to be downloaded
- DAB_ENDPOINT: url of the `dab.yeet` domain you want to hit

## Usage

You can download any album or track using the following commands

### Album download

```sh
export DAB_ENDPOINT=<ENDPOINT>
export DOWNLOAD_LOCATION=<LOCATION>
go run main.go -album <ALBUM_ID>
```

### Track download

```sh
export DAB_ENDPOINT=<ENDPOINT>
export DOWNLOAD_LOCATION=<LOCATION>
go run main.go -track <TRACK_ID>
```
