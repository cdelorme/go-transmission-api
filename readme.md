
# [go transmission api](https://github.com/cdelorme/go-transmission-api)

This is an api wrapper written in go for the express purpose of making calls to the RPC API exposed by the transmission daemon.

A [cli implementation](cmd/go-transmission-helper) is being created to leverage the functionality as a `helper`.

**This implementation is not yet feature-complete.**


## why

To provide a way to talk to transmission from my preferred language.

Also to create a CLI utility that can help shore up some of the functionality I wish existed by default or worked better within transmission.


## usage

To get the library:

	go get github.com/cdelorme/go-transmission-api

Import the library:

	import "github.com/cdelorme/go-transmission-api"

Create an instance and load the configuration:

	trans := transmission.Transmission{}
	trans.Configure("/optional/custom/path/to/settings.json")

_See the code for available function signatures and implementation._


# references

- [rpc-spec document](https://trac.transmissionbt.com/browser/trunk/extras/rpc-spec.txt)
