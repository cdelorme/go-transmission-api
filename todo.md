
# tasks

This will help keep a history of research and tentative activity on the project.

I attempted a quick fix, and introduced a bug with out-of-array-bounds.

Realizing the difficulty surrounding this project, specifically it's dependency on transmission's RPC and config file, I figured it might make more sense to create tests that emulate the behavior of those resources so I can easily validate behavior of my own code and expedite creating a solid bugfix.

I also began to see a large number of "copies", and have decided to adjust how I deal with duplicates.


## goals

- define execution behavior to be placed into the readme later
- introduce tests to emulate the http RPC server responses and validating code behavior
- update how to deal with duplicate downloads
- future fuctionality, specifically load-torrents & level6 integration


### execution behavior

Three important areas:

1. when to execute
2. execution permissions
3. atomic behavior

Ideally the execution should take place on a schedule instead of after each download completes.  This is because the code does not accept input to determine which torrent to remove, and duplicate execution can lead to undesired behavior (eg. two attempts to copy, a crash, and a corrected file).  It also reduces load on the storage drives.

The executor must:

- have read permissions on transmissions `settings.json` file
- have write permissions on transmission's downloads directory
- have write permissions on the folder the files are moving to

Currently my script copies the files one at a time recursively, creating matching folders as needed.  While `os.Rename` is a great atomic (one-uninterrupted-step) alternative, it only works when moving a file across the same disk, and fails to be atomic when going from one disk to another.  _This ruins most scenarios where one may want to move files._


### testing

To reduce the error-prone behavior of my code I want to add tests that emulate the resources my code acts upon to help me verify correct (or known) behavior.

To achieve this I need to:

- revisit the RPC interface and determine curl commands to test responses
- add those commands to my `todo.md` file for historical tracking
- create a test file to validate behavior
- create an httptest server that emulates the RPC

Ideally the tests will not need to manipulate real data, and we can still verify behavior.  Ideally abstracting the method that copies records so we can replace it at runtime during tests.

_It may also be time to break the code into multiple files for modularity._

This should allow me to address not only the current "out-of-bounds" array bug, but also any future bugs that may come along.


### duplicates

I need a long-term plan to deal with duplicates that probably won't be part of this update.

I actually want to take a total of four steps:

- add load-torrents logic to this command
- make load-torrents smart, able to track already-loaded files
- integrate level6 for file comparison
- update logic to delete when duplicate, or copy at parent directory

_Because many of the files I download are compressed, once decompressed the last two steps fail to deal with duplicates adequately._


### load torrents

This is a feature I wish transmission had, which is a historical record of all downloaded items.

When attempting to load torrents automatically, such as a "watch" list, there is no user interaction, and therefore it should have a default behavior such as "ignore and delete".

This would also mean that when a torrent finishes the torrent file should not be deleted.

In fact it shouldn't be suffixed but moved from the watched folder as well.

**To address these short-comings, I need to make an intelligent loader.**

It will:

- create a folder to copy every torrent file I've ever downloaded into
- compare new `.torrent` files against that folder, deleting files that are "named" duplicates
- the rest will be copied both into that folder, as well as to the transmission watch folder

This behavior yields all the benefits I desire.  A complete copy of every torrent file means I have a way to recover files that get corrupted without re-locating the torrent on the internet (sometimes a hassle) and I prevent duplicates at the source (me).


### level6

I still have to update my `level6` project before this becomes a possibility, but in the event that a duplicate file is downloaded, I want the behavior of my software to leverage level6 and compare.

This, _much smarter_, approach will allow me to delete duplicates instead of copying them.


# status

I'm going to tackle several of these items at once via a design change:

- define non-`main.go` files for modular code
- create `_test.go` files for testing
- leverage more cli flags
	- -o --operation
	- -c --config (transmission's settings.json)
	- -w --working-dir (the folder we're copying from/to)
- figure out how to load torrents through RPC, or add to Transmission entity:
	- `watch` setting (enabled or disabled)
	- watch folder (for torrents)

_I don't yet have a reason to create a `cmd/` folder for this project._

Needs the ability to capture the user that is executing's home folder.  Preferably that info would be gathered via init() into global vars and not during main() execution.

**I can use details from transmission's [rpc-spec](https://trac.transmissionbt.com/browser/trunk/extras/rpc-spec.txt).**

So, CSRF protection uses a `409` response with a valid CSRF and expects the client to deal with the refresh.  _This is very similar to oauth, so it should be easy to implement._

I want to add a new `const` for loading torrents manually via RPC in the event that the watch feature is disabled.  If that works I won't even need to leverage the watch behavior of transmission, which in my opinion is a huge benefit.  _There is a `torrent-start-now` method that can be executed after loading new torrents as well, and I may leverage `time.Sleep()` to delay between loading all new files and starting._  **There does indeed appear to be a `torrent-add` that accepts either the filename path or url, or metainfo (base64 encoded contents of .torrent) which is probably better since then transmission never has to care about the associated file.**
