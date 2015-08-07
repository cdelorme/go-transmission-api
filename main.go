package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/cdelorme/go-config"
	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-maps"
)

const (
	GetSession     = "{\"method\":\"session-get\"}"
	GetTorrents    = "{\"method\":\"torrent-get\",\"arguments\":{\"fields\":[\"id\",\"isFinished\",\"downloadDir\",\"files\"]}}"
	RemoveTorrents = ""
)

type settings struct {
	Port int    `json:"rpc-port"`
	Uri  string `json:"rpc-url"`
}

type file struct {
	Name string `json:"name"`
}

type torrent struct {
	Id       int    `json:"id"`
	Files    []file `json:""files""`
	Path     string `json:"downloadDir"`
	Finished bool   `json:"isFinished"`
}

type internal struct {
	Torrents []torrent `json:"torrents"`
}

type response struct {
	Args   internal `json:"arguments"`
	Result string   `json:"result"`
}

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

	url := rpc(conf)
	token := session(url)
	finished := done(url, token)
	remove(url, finished, token, path)
}

func remove(route string, torrents *[]torrent, session string, path string) {

	// create removal list

	// remove torrents

	// if not successful, exit before causing damage

	if path != "" {
		// move files to path
	}
}

func done(route string, session string) *[]torrent {
	torrents := make([]torrent, 0)

	c := http.Client{}
	req, _ := http.NewRequest("POST", route, strings.NewReader(GetTorrents))
	req.Header.Add("X-Transmission-Session-Id", session)
	resp, _ := c.Do(req)

	res := &response{}
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(res)

	for _, t := range res.Args.Torrents {
		if t.Finished {
			torrents = append(torrents, t)
		}
	}

	return &torrents
}

func session(route string) string {
	resp, _ := http.Post(route, "json", strings.NewReader(GetSession))
	return resp.Header.Get("X-Transmission-Session-Id")
}

func rpc(conf *settings) string {
	return "http://127.0.0.1:" + strconv.Itoa(conf.Port) + conf.Uri + "rpc/"
}
