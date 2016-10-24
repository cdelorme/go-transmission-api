package main

import "os"

var exit = os.Exit

func main() {
	h := new(helper)
	h.Init()
	exit(h.Run())
}
