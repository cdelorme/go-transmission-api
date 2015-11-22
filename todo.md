
# tasks

This is a plan tracking document for a history of changes.  Summary:

- changelog
- current status
- curl commands against RPC to test behavior
- testing
- pending readme updates


# changelog

Many changes were not committed as it's a work-in-progress, but because this is alpha software commits are going strait to master until we have reached a suitable level of stability.

We had two major bugs in previous iterations:

- out-of-array-bounds issue with the latest code meant to address:
- failure to cleanup folders after moving files

_With the operation failing to execute well, or failing to execute, I needed to fix it, and testing against transmission is difficult when I have to deal with live data, so testing seemed like the best option to move forward quickly with._

Since the last code commit there are many major changes.

- added significantly more clarity to cli operations and main() structure
- data types are more explicit but also succinct
- file structure is broken down for cleaner abstraction
- switched from private `transmission` to public `Transmission`
- separated move from remove and added a third operation to load torrents
- I added a mutex for concurrent safety of Token acquisition

Future iterations will very likely follow the same library structure, where `cmd/` houses the cli client.  This is more reusable and testable; separating the parts that require integration from the parts that do not.

Another major change is deciding which parts of transmission to leverage.

For example, the load-torrents command ignores the watch directory and setting, instead it directly loads files from a folder and copies them to a transmission-relative directory.

That decision was made because transmission lacks a suitable means of keeping a history of torrents, and often when disk issues occur corrupt files exist and the process of re-locating the same torrent files is a pain.  Further while it won't load two of the same torrent files actively, it does nothing to keep a history of old torrent files from previously completed and removed items.

This becomes more of a problem when downloading compressed files, where after extracting it increases the likelyhood of never realizing duplicates.

At the same time, we're going from a process that copies files directly, to leveraging the `torrent-set-location` feature, which benefits us in two ways.  First it reduces what our code is responsible for (formerly the data), and at the same time prevents time-gaps between torrent removal and file relocation which may formerly have caused copy or delete errors.


# status

Currently my focus is on:

- test sending raw base64 encoded metadata to `torrent-add`
- switching to `cmd/` for the client and going full-library system
- adding a `Load()` or `Configure()` command to transmission and directly loading the config file
	- this would eliminate conflicting concerns with my `go-config` libraries XDG pathing
- alternative interpretation of `load` operation as a string, which may allow alternative to ~/Downloads
	- useful override, especially for testing
- adding tests to identify any design errors and bugs in transmission library
	- ideally using an httptest server to mock behaviors, both success & fail cases

There is currently no safety around concurrent execution of commands that interact with the file system.  For example, we cannot lock the list

Additionally, there is concerns regarding three forms of duplication:

- torrent names
- torrent file names
- torrent sha256 hash

None of these alone resolve the problem we face.  Two differently named torrents may have the same sha256 hash. Similarly two same-named torrents may have a different sha256 hash.

While the most ideal solution would be to parse duplicates by file names in the torrent, I do not believe that sort of information can be acquired from the raw .torrent files.

I need to do some investigating, because that will fundamentally change how I deal with loading torrent files going forward.


## curl commands

My endpoint looks like this: `http://10.0.0.2:9091/bt/rpc`

If I curl it with verbose mode, I get an html response with the code, but verbose shows the `X-Transmission-Session-Id` header is set:

	curl -v http://10.0.0.2:9091/bt/rpc

_It appears that every command can returna 409 with that header, and that means we can simply extract it as a response._

Historically I was running commands to get a list with lost of details, like this:

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

I grab the `downloadDir` to ensure I had the right parent path, since torrent download paths can be changed (I had never done it before so it seems like an odd feature, but not unreasonable).  I need the state of `isFinished` to determine that not only is it seeding, but the seeding hit our ratio goal.  I was grabbing the name for error reporting via logs.  The id's are needed for all other commands, such as removing them pre-move.  Finally, the entire list of files is necessary because I cannot simply execute a recursive copy from go.

During my review session on the rpc specification I ran across [this alternative solution](https://trac.transmissionbt.com/browser/trunk/extras/rpc-spec.txt#L408), which after a short test proved to work exellently.  It offloads moving the data to transmission's responsibility, reducing code and risk on our end.  Further, it prevents clashes with moving data pre-deletion, by keeping the association with transmission.  We can even choose to leave them attached if we want to.

There are noticeable delay in the return from this command, which hinted to me that it is synchronous, making it safe to run and continue:

	curl -v -X POST -H "Content-Type: application/json; charset=UTF-8" -H "X-Transmission-Session-Id: q0F7blNnGVovAmXd6Q6zI5aQRTr4MEN7FvaXUFrJZaTG18gv" -d "{\"method\":\"torrent-set-location\",\"arguments\":{\"ids\":[1],\"location\":\"/new/path\",\"move\":true}}" http://10.0.0.2:9091/bt/rpc

_Confirmation would be nice, but the documentation does not, so I may have to run a sizable number of more detailed tests._

The other huge benefit is it reduces the operation and response to this:

	curl -v -X POST -H "Content-Type: application/json; charset=UTF-8" -H "X-Transmission-Session-Id: 286CvsiBzHndot04TB0o62H34hfoSpxU523L8Kes1sfJFNfa" -d "{\"method\":\"torrent-get\",\"arguments\":{\"fields\":[\"id\",\"isFinished\"]}}" http://10.0.0.2:9091/bt/rpc

	{"arguments":{"torrents":[{"id":2,"isFinished":false},{"id":3,"isFinished":false},{"id":4,"isFinished":false},{"id":9,"isFinished":false},{"id":12,"isFinished":false},{"id":13,"isFinished":false},{"id":14,"isFinished":false},{"id":15,"isFinished":false},{"id":17,"isFinished":false},{"id":19,"isFinished":false},{"id":21,"isFinished":false},{"id":22,"isFinished":true},{"id":23,"isFinished":false},{"id":25,"isFinished":false},{"id":27,"isFinished":false},{"id":28,"isFinished":false},{"id":29,"isFinished":false},{"id":33,"isFinished":false},{"id":34,"isFinished":false},{"id":35,"isFinished":false},{"id":37,"isFinished":false},{"id":39,"isFinished":true},{"id":41,"isFinished":false},{"id":42,"isFinished":false},{"id":43,"isFinished":false},{"id":44,"isFinished":false},{"id":45,"isFinished":false},{"id":46,"isFinished":false},{"id":47,"isFinished":false},{"id":48,"isFinished":false},{"id":49,"isFinished":false},{"id":50,"isFinished":false},{"id":51,"isFinished":false},{"id":52,"isFinished":false},{"id":53,"isFinished":false},{"id":54,"isFinished":false},{"id":55,"isFinished":false},{"id":56,"isFinished":false},{"id":57,"isFinished":false},{"id":58,"isFinished":false},{"id":59,"isFinished":false}]},"result":"success"}

_Since transmission is now responsible, we don't need names for logging, just the id and state._

I also was able to verify that I can force start all torrents with this command:

	curl -v -X POST -H "Content-Type: application/json; charset=UTF-8" -H "X-Transmission-Session-Id: 286CvsiBzHndot04TB0o62H34hfoSpxU523L8Kes1sfJFNfa" -d "{\"method\":\"torrent-start-now\",\"arguments\":{}}" http://10.0.0.2:9091/bt/rpc

_As part of this test I purposefully included an empty arguments object, since that is what go translates my entity to, which confirms that I don't have to worry about empty arguments causing disruption (yet)._


**The last thing that needs testing is the process of base64 encoding and supplying metadata via `torrent-add`.**


## testing

Once the library has been separated from the core code, we can directly manipulate the behavior of our code and run tests against an httptest server instance.

This will let us verify all possible behaviors, and correct any mistakes or bugs in the process, without having to touch a live running transmission instance.

_The same cannot be said for the client code,_ but integration needs to happen sometime, so we can verify actually loading torrents, and reading contents from the file system.  _Although how the load operation will work in the future is subject to a significant overhaul._



## readme

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

