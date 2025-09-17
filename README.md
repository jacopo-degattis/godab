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

### First export needed env variables

```sh
export DAB_ENDPOINT=<DAB_ENDPOINT> 
export DOWNLOAD_LOCATION=<LOCATION>
```

Like in the following example:
```sh
export DAB_ENDPOINT=https://dab.yeet.su
export DOWNLOAD_LOCATION=.
```

Then you can download

```sh
# ALBUM
go run main.go -album <ALBUM_ID>

#TRACK
go run main.go -track <TRACK_ID>

#ARTIST
go run main.go -artist <ARTIST_ID>
