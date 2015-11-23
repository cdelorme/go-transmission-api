
# go transmission helper

This is a helper library abbreviated as `transmissoner`.  The intended function is to provide a go stack to execute calls against the RPC interface.

A supporting CLI will provide real functionality leveraging this interface.

**This implementation is not feature-complete.**


## why

To scratch my own itch; there are two functions transmission is missing:

- dealing with seeded torrents
- automatically loading downloaded torrents

_I download a lot of files, so the lack of these features was really troublesome,_ and the core of this library is based around solving those problems, with potential room for new features.

**It does have a state-change handler,** but that handler only identifies "download completed" status not "seeding completed" status.  The RPC interface describes the latter as "isFinished", downloaded != "isFinished".  As a result you can't rely on this script trigger to accurately deal with completed torrents.

Yes, transmission does have a "watch" folder, but there are a few very strange behaviors surrounding it.  First, if you run transmission as a daemon, it won't have permission to deal with your user folders (eg. ~/Downloads), which prevents it from loading and removing loaded torrents.  Second, the watch folder ignores files that are placed via `mv`, it requires them to be `cp`'d instead (on linux).

_There are a myriad of other issues I have that aren't part of what this solves, but those two are the largest low-hanging fruit available._


## usage

To install this utility:

	go install github.com/cdelorme/go-transmission/helper/...

_The `...` will expand the `cmd/` folder installing the library and the cli utility._

View the [cli readme](cmd/go-transmission-helper/readme.md) for execution details.


# references

- [rpc-spec document](https://trac.transmissionbt.com/browser/trunk/extras/rpc-spec.txt)
