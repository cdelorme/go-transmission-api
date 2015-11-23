
# go-transmission-helper

This is the command line utility for the `go-transmission-helper` library.  It creates a CLI interface to solve problems with adding torrents to transmission, and


## requirements

**This utility is only intended for `linux` users**, and probably won't work elsewhere.

To use this utility:

- you must have read permissions on transmissions `settings.json` file
- transmission must have write permissions on the folder the files are moving to

_I did not emulate `watch` functionality due to time and utility._


## usage

To install this utility:

	go install github.com/cdelorme/go-transmission/helper/...

_The `...` will expand the `cmd/` folder installing the library and the cli utility._

There are three uses currently:

- add torrents
- move files & delete torrents

The first should be run as your own user in a cronjob, eg:

	*/2 * * * * go-transmission-helper -a

_By default it looks at `~/Downloads`, but you can set a path if you want to explicitly load a file or .torrent files from another folder._

The second function is great if you have a storage drive that you want to move files to after seeding is completed.  Pass it the path, and optionally whether to remove the torrent from transmission after relocating the files:

	go-transmission-helper -d -m /new/path

_If run on a cronjob, it's probably best to do this nightly to reduce disk io._

**For all available options, run `go-transmission-helper -h`.**

