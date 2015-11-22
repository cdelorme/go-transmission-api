package main

import (
	"encoding/base64"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/cdelorme/go-config"
	"github.com/cdelorme/go-log"
	"github.com/cdelorme/go-maps"
	"github.com/cdelorme/go-option"
)

type logger interface {
	Debug(string, ...interface{})
	Error(string, ...interface{})
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

func main() {

	// load cli options
	cli := &option.App{Description: "A utility to help wield the power of transmission through cli & automation"}
	cli.Flag("verbose", "debug mode", "-v", "--verbose")
	cli.Flag("silent", "silent mode", "-s", "--silent")
	cli.Flag("configFile", "transmission config file path", "-c", "--config")
	cli.Flag("load", "load torrents from user home folder", "-l", "--load")
	cli.Flag("move", "move torrents in finished state to this folder", "-m", "--move")
	cli.Flag("remove", "remove torrents in finished state from transmission", "-r", "--remove")
	flags := cli.Parse()

	// prepare & configure logger
	l := &log.Logger{Severity: log.Error}
	if b, _ := maps.Bool(flags, false, "silent"); b {
		l.Silent = b
	}
	if b, _ := maps.Bool(flags, false, "verbose"); b {
		l.Severity = log.Debug
	}

	// load config file via sane-default or supplied
	f, _ := maps.String(flags, transmissionConfigPath, "configFile")
	conf, err := config.Load(f)
	if err != nil {
		l.Error("Failed to read transmission configuration (%s): %s", transmissionConfigPath, err.Error())
		return
	}

	// prepare transmission instance, and apply settings
	t := &Transmission{}
	maps.To(t, conf)
	l.Debug("transmission configuration: %+v", t)

	// conditionally load from downloads
	if b, _ := maps.Bool(flags, false, "load"); b {
		load(t, l)
	}

	// conditionally move & remove
	if s, _ := maps.String(flags, "", "move"); len(s) > 0 {
		d, _ := maps.Bool(flags, false, "remove")
		move(t, l, s, d)
	}
}

func load(t *Transmission, l logger) {
	home := os.Getenv("HOME")
	if len(home) == 0 {
		l.Error("unable to determine home folder for executing user...")
		return
	}
	downloads := path.Join(home, "Downloads")
	fi, err := os.Stat(downloads)
	if err != nil || !fi.IsDir() {
		l.Error("expected ~/Downloads, but was unable to access...")
		return
	}
	l.Debug("loading torrent files from %s", downloads)

	// prepare directory for torrent storage relative to transmission downloads
	// to support multi-user transmission instances with a single shared torrent-store
	torrentStore := path.Join(path.Dir(t.Downloads), ".torrents")
	os.MkdirAll(torrentStore, 0777)
	torrents, err := ioutil.ReadDir(torrentStore)
	if err != nil {
		l.Error("failed to access torret storage directory...")
		return
	}

	files, err := ioutil.ReadDir(downloads)
	if err != nil {
		l.Error("failed to read files in downloads folder...")
	}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".torrent") {
			if fileInList(f, torrents) {
				d, err := ioutil.ReadFile(path.Join(downloads, f.Name()))
				if err != nil {
					l.Error("failed to read %s", f.Name())
					break
				}
				meta := base64.StdEncoding.EncodeToString(d)
				err = t.Add(meta)
				if err != nil {
					l.Error("failed to load torrent %s (%s)", f.Name(), err.Error())
					break
				}
				copy(path.Join(downloads, f.Name()), path.Join(torrentStore, f.Name()))
			} else {
				l.Debug("%s was already loaded previously...", f.Name())
			}
			os.Remove(path.Join(downloads, f.Name()))
		}
	}
}

func move(t *Transmission, l logger, m string, d bool) {
	if fi, err := os.Stat(m); err == nil && !fi.IsDir() {
		l.Error("file exists at supplied path...")
		return
	}
	l.Debug("moving finished torrent downloads to %s", m)

	list, err := t.Finished()
	if err != nil {
		l.Error("failed to get a list of completed torrents: %s", err.Error())
		return
	}
	err = t.Move(m, list...)
	if err != nil {
		l.Error("failed to get a list of completed torrents: %s", err.Error())
		return
	}

	// conditionally handle removal, if a separate function
	// is called we need to communicate the shared list or else
	// new items may enter isFinished state before move is done
	if d == false {
		return
	}
	l.Debug("removing complete torrents from transmission...")
	err = t.Remove(list...)
	if err != nil {
		l.Error("failed to get a list of completed torrents: %s", err.Error())
		return
	}
}
