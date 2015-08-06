package main

import (
	"os"

	"github.com/cdelorme/go-config"
	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-maps"
)

type settings struct{}
type torrent struct{}

func main() {
	logger := &log.Logger{}
	logger.Color()

	path := ""
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	conf := &settings{}
	l, err := config.Load("/etc/transmission-daemon/settings.json")
	if err != nil {
		logger.Error("Failed to load config: %s", err)
		return
	}
	maps.To(conf, l)

	remove(logger, list(logger, conf), path)
}

func remove(logger *log.Logger, torrents []torrent, path string) bool {
	// loop torrents
	// send remove requests to transmission
	// if path is not empty attempt to relocate the files related to the now-removed torrent
	return true
}

func list(logger *log.Logger, conf *settings) []torrent {
	// get session id?
	// get a list of torrents from the endpoint in finished state
	return nil
}
