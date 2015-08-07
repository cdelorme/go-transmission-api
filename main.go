package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cdelorme/go-config"
	"github.com/cdelorme/go-maps"
)

const (
	sessionGet    = "{\"method\":\"session-get\"}"
	torrentGet    = "{\"method\":\"torrent-get\",\"arguments\":{\"fields\":[\"id\",\"isFinished\",\"downloadDir\",\"files\"]}}"
	torrentRemove = "{\"method\":\"torrent-remove\",\"arguments\":{\"ids\":[%s]}}"
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
	path := ""
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	conf := &settings{}
	l, err := config.Load("/etc/transmission-daemon/settings.json")
	if err != nil {
		return
	}
	maps.To(conf, l)

	url := rpc(conf)
	token := session(url)
	finished := done(url, token)
	remove(url, finished, token, path)
}

func exists(path string, num int) string {
	if num > 0 {
		path = strconv.Itoa(num) + "copy." + path
	}
	_, err := os.Stat(path)
	if err != nil {
		return path
	}
	return exists(path, num+1)
}

func copy(to, from string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0777); err != nil {
		return err
	}

	fi, e := os.Open(from)
	if e != nil {
		return e
	}

	o := exists(to, 0)

	fo, e := os.Create(o)
	if e != nil {
		return e
	}

	if _, err := io.Copy(fo, fi); err != nil {
		return err
	}

	return nil
}

func remove(route string, torrents []torrent, session string, path string) {
	list := make([]string, 0)
	for _, t := range torrents {
		list = append(list, strconv.Itoa(t.Id))
	}

	c := http.Client{}
	req, _ := http.NewRequest("POST", route, strings.NewReader(fmt.Sprintf(torrentRemove, strings.Join(list, ","))))
	req.Header.Add("X-Transmission-Session-Id", session)
	resp, err := c.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return
	}

	if path != "" {
		for _, t := range torrents {
			for _, f := range t.Files {
				if err := copy(filepath.Join(path, f.Name), filepath.Join(t.Path, f.Name)); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to copy file: %s\n", filepath.Join(t.Path, f.Name))
				}
			}
		}
	}
}

func done(route string, session string) []torrent {
	torrents := make([]torrent, 0)

	c := http.Client{}
	req, _ := http.NewRequest("POST", route, strings.NewReader(torrentGet))
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

	return torrents
}

func session(route string) string {
	resp, _ := http.Post(route, "json", strings.NewReader(sessionGet))
	return resp.Header.Get("X-Transmission-Session-Id")
}

func rpc(conf *settings) string {
	return "http://127.0.0.1:" + strconv.Itoa(conf.Port) + conf.Uri + "rpc/"
}
