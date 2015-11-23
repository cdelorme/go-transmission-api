
# tasks

This is a plan tracking document for a history of changes.  Summary:

- changelog
- current status
- curl commands against RPC to test behavior


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

I was looking at torrent structure and ran across [this](https://en.wikipedia.org/wiki/Torrent_file#File_structure).  It appears that a torrent **does** contain the files it intends to download.  There is also a, rather complex, hashing system.  _I still need to verify what transmission does when two downloads have the same "files" locally._  If it renames them, that's fine, but I'll need to deal with that accordingly.  The conflicts include torrent name, files, and of course the hash (sha1).  Two of three would require my system be able to read the torrent file, at which point I may as well be rewriting transmission in go.

**I have decided to abandon deduplication at load.**  It's not that I don't want this functionality, but I could not find any solid source on what transmission does in the event of name conflicts, and more importantly I could not find a way to embed a storage path for old torrents, or to continue tracking removed torrents to prevent duplicate downloads.  Transmission literally forgets what it downloaded if you remove the torrent, which while correct is not the behavior I would want from a home torrent server.

Switched to library file system, due to lack of naming options calling the library `transmissioner` (short for transmission helper?).

Add `Configure()` method, accepts config file path, loads json directly, returns an error.

Created a new readme for external execution.

updated readme to reflect new state and reference new readme & install method.  Document interface of transmission instance.

switch `load` flag to `add` flag, and accept as a boolean but pass as a string.  We check if the string exists in the file system, if not we default to `$HOME/Downloads`.  If it does exist, and is a .torrent file, we add it, if it's a different directory, we scan that instead of `$HOME/Downloads`.  Created a modular addFile sub-method to handle each file in a loop.

verified base64 encoding raw torrent file contents and passing works as expected.  _There may be an edge case, such as a .torrent file already being base64 encoded, but I haven't confirmed behavior yet._

After considering my options, I've decided to abandon any attempt to deduplicate at load.  I am going to pursue a golang client implementation of my own that addresses all these quirks as a longer-term objective.

Add a sleep to the loop so that the first command waits 2 seconds prior to attempting the next command.

Removed cmd layer testing, integration tests would require a lot of work with either too many parts (running instance of transmission, fake file system, valid fake torrent files, etc).  Assuming the rest of the code follows the RPC specification, the only issues with the client are file-system related, and with appropriate debugging that should be an obvious read.  We also won't test the Configure() method, that requires file system intgeration testing as well.  If time allows I may try [this solution](https://talks.golang.org/2012/10things.slide#8).

Verified that on fail, 200 StatusCode is still sent, but result is not "success", added check for valid result and logic to retry when result != "success".


# status

To finish up test coverage, I need to create file system abstraction layers mirroring [this example](https://talks.golang.org/2012/10things.slide#8), but targeting the `ioutil` ReadFile() method.  It is not a priority so long as the current implementation works.


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

Which yields the much more succinct result list:

	{"arguments":{"torrents":[{"id":2,"isFinished":false},{"id":3,"isFinished":false},{"id":4,"isFinished":false},{"id":9,"isFinished":false},{"id":12,"isFinished":false},{"id":13,"isFinished":false},{"id":14,"isFinished":false},{"id":15,"isFinished":false},{"id":17,"isFinished":false},{"id":19,"isFinished":false},{"id":21,"isFinished":false},{"id":22,"isFinished":true},{"id":23,"isFinished":false},{"id":25,"isFinished":false},{"id":27,"isFinished":false},{"id":28,"isFinished":false},{"id":29,"isFinished":false},{"id":33,"isFinished":false},{"id":34,"isFinished":false},{"id":35,"isFinished":false},{"id":37,"isFinished":false},{"id":39,"isFinished":true},{"id":41,"isFinished":false},{"id":42,"isFinished":false},{"id":43,"isFinished":false},{"id":44,"isFinished":false},{"id":45,"isFinished":false},{"id":46,"isFinished":false},{"id":47,"isFinished":false},{"id":48,"isFinished":false},{"id":49,"isFinished":false},{"id":50,"isFinished":false},{"id":51,"isFinished":false},{"id":52,"isFinished":false},{"id":53,"isFinished":false},{"id":54,"isFinished":false},{"id":55,"isFinished":false},{"id":56,"isFinished":false},{"id":57,"isFinished":false},{"id":58,"isFinished":false},{"id":59,"isFinished":false}]},"result":"success"}

_Since transmission is now responsible, we don't need names for logging, just the id and state._

I also was able to verify that I can force start all torrents with this command:

	curl -v -X POST -H "Content-Type: application/json; charset=UTF-8" -H "X-Transmission-Session-Id: 286CvsiBzHndot04TB0o62H34hfoSpxU523L8Kes1sfJFNfa" -d "{\"method\":\"torrent-start-now\",\"arguments\":{}}" http://10.0.0.2:9091/bt/rpc

_As part of this test I purposefully included an empty arguments object, since that is what go translates my entity to, which confirms that I don't have to worry about empty arguments causing disruption (yet)._

Finally, I verified the following works passing the base64 encoded file contents of a .torrent file; granted I've omitted the metainfo because it was massive:

	curl -v -X POST -H "Content-Type: application/json; charset=UTF-8" -H "X-Transmission-Session-Id: y3yI1IvpmnJgtzPhAmgivxny8zSenmcQYfX6dfUKOnY1OHcE" -d "{\"method\":\"torrent-add\",\"arguments\":{\"metainfo\": \"\"}}" http://10.0.0.2:9091/bt/rpc

The response looked like this (again sub'd name and hash for fake data):

	{"arguments":{"torrent-added":{"hashString":"sha1-of-file","id":82,"name":"filename"}},"result":"success"}

Adding the same torrent twice returned:

	{"arguments":{"torrent-duplicate":{"hashString":"sha1-of-file","id":83,"name":"filename"}},"result":"success"}

Adding an invalid torrent metainfo returned:

	{"arguments":{},"result":"invalid or corrupt torrent file"}

_All three had StatusCode 200, so failure is not reflected by the http status code._
