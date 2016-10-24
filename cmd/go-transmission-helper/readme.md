
# [go-transmission-helper](https://github.com/cdelorme/go-transmission-api/tree/master/cmd/go-transmission-helper)

This is the command line utility for the [`go-transmission-api`](https://github.com/cdelorme/go-transmission-api) library.  It creates a CLI interface to solve problems with adding torrents to transmission, and moving files that are in a `finished` state (_eg. have reached their seed ratio_).


## requirements

**This utility was written and tested on `linux` and has not been tried on any other operating system.**

To use this utility:

- you must have read permissions on transmissions `settings.json` file
- transmission must have write permissions on (or full ownership of) the folder the files are moving to

I did not emulate `watch` functionality due to time and utility, here are crontab entries instead:

	*/2 * * * * go-transmission-helper -a $HOME
	@daily go-transmission-helper -d -m /backup/folder

_The example assumes `go-transmission-helper` is in your default `PATH`. otherwise you may prefix it with `. ~/.bash_profile; `._


## why

To scratch my own itch; there are two functions transmission does not properly handle:

- dealing with seeded torrents
- automatically loading downloaded torrents

_I download a lot of files, so the lack of these features was really troublesome,_ and the core of this library is based around solving those problems, with potential room for new features.

**It does have a state-change handler,** but that handler only identifies "download completed" status not "seeding completed" status.  The RPC interface describes the latter as "isFinished", downloaded != "isFinished".  As a result you can't rely on this script trigger to correctly deal with completed torrents.

Yes, transmission does have a "watch" folder, but there are a few very strange behaviors surrounding it.  If you run transmission as a daemon, it won't have permissions on your user folders (eg. ~/Downloads), which prevents it from loading and removing loaded torrents.  It also appears to ignore moved files, whether it is permission or timestamp related that's a problem.  Another _awesome feature_ is that it does not automatically resume loaded torrents and often they will sit idle, often for an extended duration.

_There are a myriad of other issues I have that aren't part of what this solves, but those two are the largest low-hanging fruits that I could resolve with my own utility._


## usage

To install this utility:

	go install github.com/cdelorme/go-transmission-api/...

_The `...` will expand the `cmd/` folder installing the library and the cli utility._

The utility comes with help output from `-h`, `--help`, and `help`.

The two basic behaviors include loading torrent files:

	go-transmission-helper -a .

_This is not recursive, and will default to `$HOME/Downloads/` if it cannot find the directory supplied._

The other will relocate files safely, even to another disk, using transmissions own API to handle the relocation, preventing you from bumping into file-in-use errors or any of the thousand other edge cases associated:

	go-transmission-helper -d -m /new/path

_Since this modifies the hard drive it's best to run it less often, such as nightly._
