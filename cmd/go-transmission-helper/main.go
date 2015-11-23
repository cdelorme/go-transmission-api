package main

import (
	"encoding/base64"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-maps"
	"github.com/cdelorme/go-option"

	"github.com/cdelorme/go-transmission-helper"
)

type logger interface {
	Debug(string, ...interface{})
	Error(string, ...interface{})
}

func main() {

	// load cli options
	cli := &option.App{Description: "A utility to help wield the power of transmission through cli & automation"}
	cli.Flag("verbose", "debug mode", "-v", "--verbose")
	cli.Flag("silent", "silent mode", "-s", "--silent")
	cli.Flag("configFile", "transmission config file path", "-c", "--config")
	cli.Flag("add", "add torrent(s) from supplied path or home folder", "-a", "--add")
	cli.Flag("move", "move torrents in finished state to this folder", "-m", "--move")
	cli.Flag("remove", "remove torrents in finished state from transmission", "-r", "--remove")
	cli.Example("-v -a")
	cli.Example("-a /tmp/special-torrent-stash/")
	cli.Example("-s -r -m /backup/drive/")
	flags := cli.Parse()

	// prepare & configure logger
	l := &log.Logger{Severity: log.Error}
	if b, _ := maps.Bool(flags, false, "silent"); b {
		l.Silent = b
	}
	if b, _ := maps.Bool(flags, false, "verbose"); b {
		l.Severity = log.Debug
	}

	// prepare transmission instance, and apply settings
	f, _ := maps.String(flags, "", "configFile")
	t := &transmissioner.Transmission{}
	if err := t.Configure(f); err != nil {
		l.Error("Failed to read transmission configuration (%s): %s", f, err.Error())
		return
	}
	l.Debug("transmission configuration: %+v", t)

	// conditionally add from downloads, then force-resume of paused/new downloads (eg. Resume Now)
	if b, _ := maps.Bool(flags, false, "add"); b {
		s, _ := maps.String(flags, "", "add")
		add(t, l, s)
		t.Resume()
	}

	// conditionally move & remove
	if s, _ := maps.String(flags, "", "move"); len(s) > 0 {
		r, _ := maps.Bool(flags, false, "remove")
		move(t, l, s, r)
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

func addFile(t *transmissioner.Transmission, f string) error {

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

func add(t *transmissioner.Transmission, l logger, p string) {
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

func move(t *transmissioner.Transmission, l logger, m string, r bool) {
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
