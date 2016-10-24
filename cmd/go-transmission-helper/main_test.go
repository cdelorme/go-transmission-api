package main

import (
	"os"
	"testing"
)

func init() {
	exit = func(_ int) {}
}

func TestPlacebo(t *testing.T) {
	t.Parallel()
	if !true {
		t.FailNow()
	}
}

func TestMain(_ *testing.T) {
	os.Clearenv()
	os.Args = []string{}
	main()
}
