package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type FakeFile struct {
	name string
	size int64
	mode os.FileMode
	mod  time.Time
	dir  bool
}

func (self FakeFile) Name() string       { return self.name }
func (self FakeFile) Size() int64        { return self.size }
func (self FakeFile) Mode() os.FileMode  { return self.mode }
func (self FakeFile) ModTime() time.Time { return self.mod }
func (self FakeFile) IsDir() bool        { return self.dir }
func (self FakeFile) Sys() interface{}   { return nil }

var (
	readfileBytes []byte
	readfileError error
	statError     error
	readdirError  error
	removeError   error

	fakeFile                = FakeFile{}
	readdirFiles            = []os.FileInfo{&FakeFile{name: "test.torrent"}}
	mockError               = errors.New("mock error...")
	jsonTransmissionSuccess = []byte(`{"result":"success"}`)
	jsonTransmissionList    = []byte(`{"result":"success","arguments": {"torrents": [{"id": 1,"isFinished": false},{"id": 2,"isFinished": false},{"id": 3,"isFinished": false},{"id": 4,"isFinished": false},{"id": 5,"isFinished": false},{"id": 6,"isFinished": true},{"id": 7,"isFinished": true},{"id": 8,"isFinished": true},{"id": 9,"isFinished": true},{"id": 10,"isFinished": true}]}}`)
	token                   = `Some Long Crazy Hash`
)

func init() {
	readfile = func(_ string) ([]byte, error) { return readfileBytes, readfileError }
	stat = func(_ string) (os.FileInfo, error) { return &fakeFile, statError }
	remove = func(_ string) error { return removeError }
	readdir = func(_ string) ([]os.FileInfo, error) { return readdirFiles, readdirError }
}

func TestHelperPublicInit(t *testing.T) {
	os.Clearenv()
	os.Args = []string{}
	h := &helper{}

	// try with env
	os.Setenv("TRANSMISSION_CONFIG", "/tmp/config")
	os.Setenv("TRANSMISSION_ADD", "/tmp/Downloads")
	os.Setenv("TRANSMISSION_MOVE", "/tmp/moved")
	os.Setenv("TRANSMISSION_REMOVE", "true")
	h.Init()
	if h.File != "/tmp/config" || h.Move != "/tmp/moved" || h.Add != "/tmp/Downloads" || !h.Remove {
		t.FailNow()
	}

	// try with cli overrides
	os.Args = []string{"-c", "~/config", "-a", "~/Downloads", "-m", "~/moved", "-r", "false"}
	h.Init()
	if h.File != "~/config" || h.Move != "~/moved" || h.Add != "~/Downloads" || h.Remove {
		t.FailNow()
	}
}

func TestHelperPublicRun(t *testing.T) {
	statError = nil
	h := &helper{}

	// successfully do nothing
	if e := h.Run(); e != 0 {
		t.FailNow()
	}

	// fail move
	h.Move = "/tmp"
	fakeFile.dir = false
	if e := h.Run(); e != 1 {
		t.FailNow()
	}

	// fail add
	h.Add = "/tmp"
	statError = mockError
	if e := h.Run(); e != 1 {
		t.FailNow()
	}

	// fail transmission configure
	h.File = "/doesnotexist"
	if e := h.Run(); e != 1 {
		t.FailNow()
	}
}

func TestHelperPrivateMove(t *testing.T) {
	readfileBytes = []byte("")
	readfileError = nil
	statError = nil
	getStatus := http.StatusOK
	moveStatus := http.StatusOK
	removeStatus := http.StatusOK
	h := &helper{}
	h.Remove = true

	// setup mock transmission endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("X-Transmission-Session-Id", token)

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var m map[string]interface{}
		decoder.Decode(&m)

		if v, ok := m["method"]; !ok {
			t.Logf("unexpected command %s\n", m)
			t.Fail()
		} else if v == "torrent-get" {
			w.WriteHeader(getStatus)
			w.Write(jsonTransmissionList)
			return
		} else if v == "torrent-set-location" {
			w.WriteHeader(moveStatus)
			w.Write(jsonTransmissionSuccess)
			return
		}

		w.WriteHeader(removeStatus)
		w.Write(jsonTransmissionSuccess)
	}))
	defer ts.Close()
	h.Transmission.Port, _ = strconv.Atoi(strings.Split(ts.URL, ":")[2])

	// test empty move
	if err := h.move(); err != nil {
		t.FailNow()
	}
	h.Move = "/tmp"

	// test successful response
	fakeFile.dir = true
	if err := h.move(); err != nil {
		t.FailNow()
	}

	// test bad remove
	removeStatus = http.StatusInternalServerError
	if err := h.move(); err == nil {
		t.FailNow()
	}

	// test bad move response
	moveStatus = http.StatusInternalServerError
	if err := h.move(); err == nil {
		t.FailNow()
	}

	// test bad finished response
	getStatus = http.StatusInternalServerError
	if err := h.move(); err == nil {
		t.FailNow()
	}

	// test bad stat
	fakeFile.dir = false
	if err := h.move(); err == nil {
		t.FailNow()
	}
}

func TestHelperPrivateAdd(t *testing.T) {
	readfileBytes = []byte("")
	readfileError = nil
	statError = nil
	status := http.StatusOK
	h := &helper{}

	// setup mock transmission endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("X-Transmission-Session-Id", token)
		w.WriteHeader(status)
		w.Write(jsonTransmissionSuccess)
	}))
	defer ts.Close()
	h.Transmission.Port, _ = strconv.Atoi(strings.Split(ts.URL, ":")[2])

	// test with empty add
	if err := h.add(); err != nil {
		t.FailNow()
	}
	h.Add = "/tmp"

	// test success with dir
	fakeFile.dir = true
	if err := h.add(); err != nil {
		t.FailNow()
	}

	// // test failure with dir at addfile
	status = http.StatusInternalServerError
	if err := h.add(); err == nil {
		t.FailNow()
	}

	// test failure with dir at readdir
	readdirError = mockError
	if err := h.add(); err == nil {
		t.FailNow()
	}

	// test success at addFile with file
	fakeFile.dir = false
	status = http.StatusOK
	if err := h.add(); err != nil {
		t.FailNow()
	}

	// test failure at addFile with file
	status = http.StatusInternalServerError
	if err := h.add(); err == nil {
		t.FailNow()
	}

	// test with failed stat
	statError = mockError
	if err := h.add(); err == nil {
		t.FailNow()
	}
}

func TestHelperPrivateAddFile(t *testing.T) {
	readfileBytes = []byte("")
	readfileError = nil
	status := http.StatusOK
	h := &helper{}

	// setup mock transmission endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("X-Transmission-Session-Id", token)
		w.WriteHeader(status)
		w.Write(jsonTransmissionSuccess)
	}))
	defer ts.Close()
	h.Transmission.Port, _ = strconv.Atoi(strings.Split(ts.URL, ":")[2])

	// test running add file success
	if err := h.addFile(""); err != nil {
		t.FailNow()
	}

	// test running add file failure at transmission
	status = http.StatusInternalServerError
	if err := h.addFile(""); err == nil {
		t.FailNow()
	}

	// test running add file failure at encoding
	readfileError = mockError
	if err := h.addFile(""); err == nil {
		t.FailNow()
	}
}

func TestHelperPrivateLoad64(t *testing.T) {
	readfileBytes = []byte("test")
	readfileError = nil
	h := &helper{}

	// test readfile
	if b, err := h.load64(""); err != nil || b != "dGVzdA==" {
		t.FailNow()
	}

	// test readfile with mock error
	readfileError = mockError
	if _, err := h.load64(""); err == nil {
		t.FailNow()
	}
}
