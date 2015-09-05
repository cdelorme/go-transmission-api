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
	"time"

	"github.com/cdelorme/go-config"
	"github.com/cdelorme/go-maps"
)

const (
	sessionGet    = "{\"method\":\"session-get\"}"
	torrentGet    = "{\"method\":\"torrent-get\",\"arguments\":{\"fields\":[\"id\",\"name\",\"isFinished\",\"downloadDir\",\"files\"]}}"
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
	Name     string `json:"name"`
	Files    []file `json:"files"`
	Path     string `json:"downloadDir"`
	Finished bool   `json:"isFinished"`
}

func (self *torrent) Move(path string) error {
	containers := make(map[string]struct{}, 0)
	for _, f := range self.Files {
		if e := copy(filepath.Join(path, f.Name), filepath.Join(self.Path, f.Name)); e != nil {
			return e
		}
		containers[strings.Split(f.Name, string(os.PathSeparator))[0]] = struct{}{}
	}

	for k, _ := range containers {
		if e := os.RemoveAll(k); e != nil && !os.IsNotExist(e) {
			return e
		}
	}

	return nil
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
	remove(url, token, path, finished)
}

func tsugi(path string, num int) string {
	if num > 0 {
		ext := filepath.Ext(path)
		path = path[0:len(path)-len(ext)] + "(copy " + strconv.Itoa(num) + ")" + ext
	}
	_, err := os.Stat(path)
	if err != nil {
		return path
	}
	return tsugi(path, num+1)
}

// @note: os.Rename does not work across drives, even if it did it would no-longer be atomic
func copy(to, from string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0777); err != nil {
		return err
	}

	fi, e := os.Open(from)
	if e != nil {
		return e
	}

	o := tsugi(to, 0)

	fo, e := os.Create(o)
	if e != nil {
		return e
	}

	if _, err := io.Copy(fo, fi); err != nil {
		return err
	}

	return nil
}

func remove(route, session, path string, torrents []torrent) {
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
		for r := 0; r < 3; r++ {
			if len(torrents) == 0 {
				break
			} else {
				time.Sleep(time.Second * 15)
			}
			for i, t := range torrents {
				if t.Move(path) == nil {
					torrents = append(torrents[:i], torrents[i+1:]...)
				}
			}
		}
		if len(torrents) > 0 {
			fmt.Fprintln(os.Stderr, "Failed to relocate files for these torrents:")
			for _, t := range torrents {
				fmt.Fprintln(os.Stderr, t.Name)
			}
		}
	}
}

func done(route, session string) []torrent {
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
