package main

import (
	"encoding/base64"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-transmission-api"
	"github.com/cdelorme/gonf"
)

type logger interface {
	Debug(string, ...interface{})
	Error(string, ...interface{})
}

type config struct {
	Add    string `json:"add,omitempty"`
	Remove bool   `json:"remove,omitempty"`
	Move   string `json:"move,omitempty"`
	File   string `json:"configFile,omitempty"`
}

func main() {

	// load cli options
	c := new(config)
	g := gonf.Gonf{Description: "A utility to help wield the power of transmission through cli & automation", Configuration: c}
	g.Add("configFile", "transmission config file path", "", "-c", "--config")
	g.Add("add", "add torrent(s) from supplied path or home folder", "", "-a", "--add")
	g.Add("move", "move torrents in finished state to this folder", "", "-m", "--move")
	g.Add("remove", "remove torrents in finished state from transmission", "", "-r", "--remove")
	g.Example("-a")
	g.Example("-a /tmp/special-torrent-stash/")
	g.Example("-r -m /backup/drive/")
	g.Load()

	// prepare & configure logger
	l := &log.Logger{}

	// prepare transmission instance, and apply settings
	t := &transmission.Transmission{}
	if err := t.Configure(c.File); err != nil {
		l.Error("Failed to read transmission configuration: %s", err.Error())
		return
	}
	l.Debug("transmission configuration: %+v", t)

	// conditionally add from downloads, then force-resume of paused/new downloads (eg. Resume Now)
	if c.Add != "" {
		add(t, l, c.Add)
		t.Resume()
	}

	// conditionally move & remove
	if c.Move != "" {
		move(t, l, c.Move, c.Remove)
	}
}

func fileInList(needle os.FileInfo, haystack []os.FileInfo) bool {
	for _, s := range haystack {
		if needle.Name() == s.Name() {
			return true
		}
	}
	return false
}

func copy(from, to string) error {
	fi, e := os.Open(from)
	if e != nil {
		return e
	}

	fo, e := os.Create(to)
	if e != nil {
		return e
	}

	if _, err := io.Copy(fo, fi); err != nil {
		return err
	}

	return nil
}

func load64(f string) (string, error) {
	d, err := ioutil.ReadFile(f)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(d), nil
}

func addFile(t *transmission.Transmission, f string) error {

	// attempt to read file contents
	meta, err := load64(f)
	if err != nil {
		return err
	}

	// attempt to add metadata
	if err := t.Add(meta); err != nil {
		return err
	}

	return nil
}

func add(t *transmission.Transmission, l logger, p string) {
	d, err := os.Stat(p)
	if err != nil {
		l.Debug("no or unreadable add-file supplied, switching to downloads")
		downloads := path.Join(os.Getenv("HOME"), "Downloads")
		if h, e := os.Stat(downloads); e != nil || !h.IsDir() {
			l.Error("unable to read downloads folder...")
			return
		}
		add(t, l, downloads)
		return
	}

	// attempt to load a single file
	if !d.IsDir() {
		l.Debug("adding torrent file %s", p)
		if err := addFile(t, p); err != nil {
			l.Error("failed to add %s (%s)", p, err.Error)
			return
		}
		os.Remove(p)
		return
	}

	// get list of files
	files, err := ioutil.ReadDir(p)
	if err != nil {
		l.Error("failed to read files in %s...", p)
		return
	}

	// iterate files, add if .torrent, remove if successful
	l.Debug("adding torrents from %s", p)
	for _, f := range files {
		l.Debug("adding file %s", f.Name())
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".torrent") {
			d := path.Join(p, f.Name())
			if err := addFile(t, d); err != nil {
				l.Error("unable to load %s (%v)", d, err)
			} else {
				os.Remove(d)
			}
		}
	}
}

func move(t *transmission.Transmission, l logger, m string, r bool) {
	if fi, err := os.Stat(m); err == nil && !fi.IsDir() {
		l.Error("file exists at supplied path...")
		return
	}

	l.Debug("searching for finished torrents...")
	list, err := t.Finished()
	if err != nil {
		l.Error("failed to get a list of completed torrents: %s", err.Error())
		return
	}
	l.Debug("list: %+v\n", list)

	l.Debug("moving finished torrent downloads to %s", m)
	err = t.Move(m, list)
	if err != nil {
		l.Error("failed to get a list of completed torrents: %s", err.Error())
		return
	}
	l.Debug("finished moving list...")

	// remove matching list if allowed
	// @note: alternative implementation is to return a list of torrents
	//   without that, we can't ensure the list of isFinished will match
	//   which may result in removing files before they have been relocated
	if r == false {
		return
	}
	l.Debug("removing complete torrents from transmission...")
	err = t.Remove(list)
	if err != nil {
		l.Error("failed to get a list of completed torrents: %s", err.Error())
		return
	}
}
