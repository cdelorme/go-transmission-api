
# go transmission helper

This is a cli utility written to support the linux transmission daemon.  It loads the configuration file, grabs the RPC address, and can run requests against it.


## requests

Currently the only operation being planned is a way to handle completed torrents.  The definition of "complete" being that the download has finished, and the seeding ratio has been met.

When executed it should:

- request torrents that are labeled as "finished"
- remove each of them
- optionally move the files of each into another path

For me this is helpful in two ways.  For one my main storage disk is not the same as the download disk, and I do not have infinite storage space on the download disk.  For two, it lets me clearly know when downloads are able to be shifted around.

_This is a feature I expected to exist within the client, much like utorrent used to have._


## future

While I can think of at half a dozen more features I'd like to see, I'd rather write my own client to handle them if I was going to go that far.

So this library will likely only ever provide one function.
