
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



## capturing curl commands

For testing, we need to verify the behavior of transmission during successful and unsuccessful command attempts.  Also, verified that when using `omitempty` on a composite struct, it still prints an empty object instead of omitting.  This means if we want to dynamically create commands, we either need to resort to a map[string]interface{} or verify that an empty "args" property won't cause trouble with the RPC endpoint.

The RPC endpoint I tested against had a path of `http://10.0.0.2:9091/bt/rpc`

Let's start with acquiring a session:

	curl -v http://10.0.0.2:9091/bt/rpc

_This will print an error containing the session id, but if you notice verbose mode also prints the session id in the response header ``._  Thus, any request you send will give you the information you need top make a correction programmatically.  _If in the future the transmission interface I write needs to expand into a library, I should probably add a mutex for concurrency support on its properties, like the token._  We'll set the header for all subsequent requests...

Let's start with getting a list of torrents and properties we care about:

	curl -v -X POST -H "Content-Type: application/json; charset=UTF-8" -H "X-Transmission-Session-Id: q0F7blNnGVovAmXd6Q6zI5aQRTr4MEN7FvaXUFrJZaTG18gv" -d "{\"method\":\"torrent-get\",\"arguments\":{\"fields\":[\"id\",\"name\",\"isFinished\",\"downloadDir\",\"files\"]}}" http://10.0.0.2:9091/bt/rpc

Example response data:

	{
		"result":"success",
		"arguments":{
			"torrents":[
				{
				"downloadDir":"/media/transmission/downloads",
				"files":[
					{
						"bytesCompleted":487753063,
						"length":487753063,
						"name":"movie.mkv"
					}
				],
				"id":1,
				"isFinished":false,
				"name":"movie.mkv"
			},

[This might solve all of my problems](https://trac.transmissionbt.com/browser/trunk/extras/rpc-spec.txt#L408).  Previously I was removing the torrent to unlink it, then copying the data.  However, it looks like this functionality exists within transmission **and** can remain linked.  Therefore, we could perform the move leveraging transmission's RPC interface, reducing the code I have to write; assuming it ia synchronous we can then disconnect all id's afterwards.

	curl -v -X POST -H "Content-Type: application/json; charset=UTF-8" -H "X-Transmission-Session-Id: q0F7blNnGVovAmXd6Q6zI5aQRTr4MEN7FvaXUFrJZaTG18gv" -d "{\"method\":\"torrent-set-location\",\"arguments\":{\"ids\":[1],\"location\":\"/media/toshokan/transmission\",\"move\":true}}" http://10.0.0.2:9091/bt/rpc

**Success!**  We can totally leverage that over the alternative.  This is great news!  The first attempt failed, because I didn't realize that transmission needed write permissions.  Once added, a slight delay in the second attempt strongly hinted that the move is synchronous.  Either way, not having to do the copy ourselves, and the added benefit of sustaining the connection is great news.  It also means some significant changes in my design plans.  _Although, transmission doesn't have any documentation around how duplicate files are handled._

We can also reduce the request to this:

	curl -v -X POST -H "Content-Type: application/json; charset=UTF-8" -H "X-Transmission-Session-Id: 286CvsiBzHndot04TB0o62H34hfoSpxU523L8Kes1sfJFNfa" -d "{\"method\":\"torrent-get\",\"arguments\":{\"fields\":[\"id\",\"isFinished\"]}}" http://10.0.0.2:9091/bt/rpc

Which yields a much more succinct:

	{"arguments":{"torrents":[{"id":2,"isFinished":false},{"id":3,"isFinished":false},{"id":4,"isFinished":false},{"id":9,"isFinished":false},{"id":12,"isFinished":false},{"id":13,"isFinished":false},{"id":14,"isFinished":false},{"id":15,"isFinished":false},{"id":17,"isFinished":false},{"id":19,"isFinished":false},{"id":21,"isFinished":false},{"id":22,"isFinished":true},{"id":23,"isFinished":false},{"id":25,"isFinished":false},{"id":27,"isFinished":false},{"id":28,"isFinished":false},{"id":29,"isFinished":false},{"id":33,"isFinished":false},{"id":34,"isFinished":false},{"id":35,"isFinished":false},{"id":37,"isFinished":false},{"id":39,"isFinished":true},{"id":41,"isFinished":false},{"id":42,"isFinished":false},{"id":43,"isFinished":false},{"id":44,"isFinished":false},{"id":45,"isFinished":false},{"id":46,"isFinished":false},{"id":47,"isFinished":false},{"id":48,"isFinished":false},{"id":49,"isFinished":false},{"id":50,"isFinished":false},{"id":51,"isFinished":false},{"id":52,"isFinished":false},{"id":53,"isFinished":false},{"id":54,"isFinished":false},{"id":55,"isFinished":false},{"id":56,"isFinished":false},{"id":57,"isFinished":false},{"id":58,"isFinished":false},{"id":59,"isFinished":false}]},"result":"success"}

_This doesn't let us be as explicit with our logging, but since we're now leveraging transmission, it's logs are probably a more viable source._

We still need to verify what `torrent-add` looks like both with and without "valid" metadata.



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

Base setup is complete with flags to control operations as needed.  At this point we just need to cleanup and reimplemented transmission features.

_We also still need to verify what happens when I send a command with an empty args object._

New transmission method layout:

- get() []torrent
- finished() []torrent
- ids(t ...torrent) []int
- move(ids ...int) error
- delete(ids ...int) error
- load(meta []byte) error

We need structs to represent the various datatypes, and should use inline anonymous struct definitions for deep-content instead of arbitrarily named structs.

This new command layout takes advantage of transmissions own tooling better to yield the desired results.

_We might even consider making an optional delete flag, to run separate from move, or conditionally after move is successful._

**The only major change in usage is the ownership of the folder files are being moved to must be accessible by transmission.**

Also, I've confirmed that I will be attempting to ignore the `watch` folder and setting, because if we can directly load metadata we can safely ignore those features.

_It might also be really cool if I can make `torrent-add` executable, such that mimetype association would automatically load a torrent when "open" is executed, but that's a far-future feature._

If the transmission functionality grows substantially, I may consider moving to a `cmd/` folder-structure, with one or more clients.
