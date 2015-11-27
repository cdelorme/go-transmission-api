package transmissioner

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

var (
	getTorrentsSuccess    = []byte(`{"result":"success","arguments": {"torrents": [{"id": 1,"isFinished": false},{"id": 2,"isFinished": false},{"id": 3,"isFinished": false},{"id": 4,"isFinished": false},{"id": 5,"isFinished": false},{"id": 6,"isFinished": true},{"id": 7,"isFinished": true},{"id": 8,"isFinished": true},{"id": 9,"isFinished": true},{"id": 10,"isFinished": true}]}}`)
	moveTorrentsSuccess   = []byte(`{"result":"success"}`)
	removeTorrentsSuccess = []byte(`{"result":"success"}`)
	addTorrentSuccess     = []byte(`{"result":"success"}`)
	resumeTorrentsSuccess = []byte(`{"result":"success"}`)

	getTorrentsFail    = []byte(`{"result":"not success"}`)
	moveTorrentsFail   = []byte(`{"result":"not success"}`)
	removeTorrentsFail = []byte(`{"result":"not success"}`)
	addTorrentFail     = []byte(`{"result":"not success"}`)
	resumeTorrentsFail = []byte(`{"result":"not success"}`)

	token    = `Some Long Crazy Hash`
	metainfo = `some base64 string`
	movepath = `/new/storage/path/`
)

var fakeFSData []byte
var fakeFSError error
var fsError = errors.New("fake filesystem error")

// override transmission.go's fs with fakeFS
func init() {
	fs = &fakeFS{}
}

type fakeFS struct{}

func (self *fakeFS) ReadFile(path string) ([]byte, error) {
	return fakeFSData, fakeFSError
}

func TestPlacebo(t *testing.T) {
	t.Parallel()
	if !true {
		t.FailNow()
	}
}

func TestGetSuccess(t *testing.T) {
	t.Parallel()

	// prepare false test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("X-Transmission-Session-Id", token)

		// verify http method
		if r.Method != "POST" {
			t.Fail()
		}

		// check token
		if r.Header.Get("X-Transmission-Session-Id") != token {
			w.WriteHeader(http.StatusConflict)
			return
		}

		// verify request method
		c := &command{}
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(c)
		if c.Method != "torrent-get" {
			t.Fail()
		}

		// return get results
		w.WriteHeader(http.StatusOK)
		w.Write(getTorrentsSuccess)
	}))
	defer ts.Close()

	// parse port off ts.URL
	port, _ := strconv.Atoi(strings.Split(ts.URL, ":")[2])

	// configure transmission instance
	tr := Transmission{Port: port}

	// run Get
	l, err := tr.Get()
	if err != nil || len(l) != 10 {
		t.Logf("error (%s) or list does not have 10 records (%+v)", err, l)
		t.FailNow()
	}
}

func TestGetFail(t *testing.T) {
	t.Parallel()

	// prepare false test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("X-Transmission-Session-Id", token)

		// verify http method
		if r.Method != "POST" {
			t.Fail()
		}

		// check token
		if r.Header.Get("X-Transmission-Session-Id") != token {
			w.WriteHeader(http.StatusConflict)
			return
		}

		// verify request method
		c := &command{}
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(c)
		if c.Method != "torrent-get" {
			t.Fail()
		}

		// return get results
		w.WriteHeader(http.StatusOK)
		w.Write(getTorrentsFail)
	}))
	defer ts.Close()

	// parse port off ts.URL
	port, _ := strconv.Atoi(strings.Split(ts.URL, ":")[2])

	// configure transmission instance
	tr := Transmission{Port: port}

	// run Get
	_, err := tr.Get()
	if err == nil {
		t.Logf("expected forced error, but got: %v\n", err)
		t.FailNow()
	}
}

func TestGetFailNoServer(t *testing.T) {
	t.Parallel()

	// configure transmission instance
	tr := Transmission{}

	// run Get
	_, err := tr.Get()
	if err == nil {
		t.Logf("expected error, but got: %v\n", err)
		t.FailNow()
	}
}

func TestGetFailBadRequestAddress(t *testing.T) {
	t.Parallel()

	// configure transmission instance
	tr := Transmission{Uri: "%"}

	// run Get
	_, err := tr.Get()
	if err == nil {
		t.Logf("expected error, but got: %v\n", err)
		t.FailNow()
	}
}

func TestFinishedSuccess(t *testing.T) {
	t.Parallel()

	// prepare false test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("X-Transmission-Session-Id", token)

		// verify http method
		if r.Method != "POST" {
			t.Fail()
		}

		// check token
		if r.Header.Get("X-Transmission-Session-Id") != token {
			w.WriteHeader(http.StatusConflict)
			return
		}

		// verify request method
		c := &command{}
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(c)
		if c.Method != "torrent-get" {
			t.Fail()
		}

		// return get results
		w.WriteHeader(http.StatusOK)
		w.Write(getTorrentsSuccess)
	}))
	defer ts.Close()

	// parse port off ts.URL
	port, _ := strconv.Atoi(strings.Split(ts.URL, ":")[2])

	// configure transmission instance
	tr := Transmission{Port: port}

	// run Finished
	l, err := tr.Finished()
	if err != nil || len(l) != 5 {
		t.Logf("error (%s) or list does not have 10 records (%+v)", err, l)
		t.FailNow()
	}
}

func TestMoveSuccess(t *testing.T) {
	t.Parallel()

	// prepare false test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("X-Transmission-Session-Id", token)

		// verify http method
		if r.Method != "POST" {
			t.Fail()
		}

		// check token
		if r.Header.Get("X-Transmission-Session-Id") != token {
			w.WriteHeader(http.StatusConflict)
			return
		}

		// verify request method
		c := &command{}
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(c)
		if c.Method == "torrent-get" {
			w.WriteHeader(http.StatusOK)
			w.Write(getTorrentsSuccess)
		} else if c.Method == "torrent-set-location" {
			w.WriteHeader(http.StatusOK)
			w.Write(moveTorrentsSuccess)
		} else {
			t.Fail()
		}
	}))
	defer ts.Close()

	// parse port off ts.URL
	port, _ := strconv.Atoi(strings.Split(ts.URL, ":")[2])

	// configure transmission instance
	tr := Transmission{Port: port}

	// acquire finished list
	l, err := tr.Finished()
	if err != nil {
		t.Logf("unexpected error: %v\n", err)
		t.FailNow()
	}

	// run Move
	err = tr.Move(movepath, l)
	if err != nil {
		t.Logf("unexpected error: %v\n", err)
		t.FailNow()
	}
}

func TestMoveFail(t *testing.T) {
	t.Parallel()

	// prepare false test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("X-Transmission-Session-Id", token)

		// verify http method
		if r.Method != "POST" {
			t.Fail()
		}

		// check token
		if r.Header.Get("X-Transmission-Session-Id") != token {
			w.WriteHeader(http.StatusConflict)
			return
		}

		// verify request method
		c := &command{}
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(c)
		if c.Method == "torrent-get" {
			w.WriteHeader(http.StatusOK)
			w.Write(getTorrentsSuccess)
		} else if c.Method == "torrent-set-location" {
			w.WriteHeader(http.StatusOK)
			w.Write(moveTorrentsFail)
		} else {
			t.Fail()
		}
	}))
	defer ts.Close()

	// parse port off ts.URL
	port, _ := strconv.Atoi(strings.Split(ts.URL, ":")[2])

	// configure transmission instance
	tr := Transmission{Port: port}

	// acquire finished list
	l, err := tr.Finished()
	if err != nil {
		t.Logf("unexpected error: %v\n", err)
		t.FailNow()
	}

	// run Move
	err = tr.Move(movepath, l)
	if err == nil {
		t.Logf("expected error, but got: %v\n", err)
		t.FailNow()
	}
}

func TestMoveFailNoList(t *testing.T) {
	t.Parallel()

	// configure transmission instance
	tr := Transmission{}

	// run Move /w empty array
	err := tr.Move(movepath, []torrent{})
	if err != nil {
		t.Logf("expected error, but got: %v\n", err)
		t.FailNow()
	}
}

func TestRemoveSuccess(t *testing.T) {
	t.Parallel()

	// prepare false test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("X-Transmission-Session-Id", token)

		// verify http method
		if r.Method != "POST" {
			t.Fail()
		}

		// check token
		if r.Header.Get("X-Transmission-Session-Id") != token {
			w.WriteHeader(http.StatusConflict)
			return
		}

		// verify request method
		c := &command{}
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(c)
		if c.Method == "torrent-get" {
			w.WriteHeader(http.StatusOK)
			w.Write(getTorrentsSuccess)
		} else if c.Method == "torrent-remove" {
			w.WriteHeader(http.StatusOK)
			w.Write(removeTorrentsSuccess)
		} else {
			t.Fail()
		}
	}))
	defer ts.Close()

	// parse port off ts.URL
	port, _ := strconv.Atoi(strings.Split(ts.URL, ":")[2])

	// configure transmission instance
	tr := Transmission{Port: port}

	// acquire finished list
	l, err := tr.Finished()
	if err != nil {
		t.Logf("unexpected error: %v\n", err)
		t.FailNow()
	}

	// run Remove
	err = tr.Remove(l)
	if err != nil {
		t.Logf("unexpected error: %v\n", err)
		t.FailNow()
	}
}

func TestRemoveFail(t *testing.T) {
	t.Parallel()

	// prepare false test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("X-Transmission-Session-Id", token)

		// verify http method
		if r.Method != "POST" {
			t.Fail()
		}

		// check token
		if r.Header.Get("X-Transmission-Session-Id") != token {
			w.WriteHeader(http.StatusConflict)
			return
		}

		// verify request method
		c := &command{}
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(c)
		if c.Method == "torrent-get" {
			w.WriteHeader(http.StatusOK)
			w.Write(getTorrentsSuccess)
		} else if c.Method == "torrent-remove" {
			w.WriteHeader(http.StatusOK)
			w.Write(removeTorrentsFail)
		} else {
			t.Fail()
		}
	}))
	defer ts.Close()

	// parse port off ts.URL
	port, _ := strconv.Atoi(strings.Split(ts.URL, ":")[2])

	// configure transmission instance
	tr := Transmission{Port: port}

	// acquire finished list
	l, err := tr.Finished()
	if err != nil {
		t.Logf("unexpected error: %v\n", err)
		t.FailNow()
	}

	// run Remove
	err = tr.Remove(l)
	if err == nil {
		t.Logf("expected error, but got: %v\n", err)
		t.FailNow()
	}
}

func TestRemoveFailNoList(t *testing.T) {
	t.Parallel()

	// configure transmission instance
	tr := Transmission{}

	// run Remove /w empty array
	err := tr.Remove([]torrent{})
	if err != nil {
		t.Logf("expected error, but got: %v\n", err)
		t.FailNow()
	}
}

func TestAddSuccess(t *testing.T) {
	t.Parallel()

	// prepare false test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("X-Transmission-Session-Id", token)

		// verify http method
		if r.Method != "POST" {
			t.Fail()
		}

		// check token
		if r.Header.Get("X-Transmission-Session-Id") != token {
			w.WriteHeader(http.StatusConflict)
			return
		}

		// verify request method
		c := &command{}
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(c)
		if c.Method != "torrent-add" || c.Arguments.Metainfo != metainfo {
			t.Fail()
		}

		// return get results
		w.WriteHeader(http.StatusOK)
		w.Write(addTorrentSuccess)
	}))
	defer ts.Close()

	// parse port off ts.URL
	port, _ := strconv.Atoi(strings.Split(ts.URL, ":")[2])

	// configure transmission instance
	tr := Transmission{Port: port}

	// run Add
	err := tr.Add(metainfo)
	if err != nil {
		t.Logf("unexpected error: %v\n", err)
		t.FailNow()
	}
}

func TestAddFail(t *testing.T) {
	t.Parallel()

	// prepare false test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("X-Transmission-Session-Id", token)

		// verify http method
		if r.Method != "POST" {
			t.Fail()
		}

		// check token
		if r.Header.Get("X-Transmission-Session-Id") != token {
			w.WriteHeader(http.StatusConflict)
			return
		}

		// verify request method
		c := &command{}
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(c)
		if c.Method != "torrent-add" || c.Arguments.Metainfo != metainfo {
			t.Fail()
		}

		// return get results
		w.WriteHeader(http.StatusOK)
		w.Write(addTorrentFail)
	}))
	defer ts.Close()

	// parse port off ts.URL
	port, _ := strconv.Atoi(strings.Split(ts.URL, ":")[2])

	// configure transmission instance
	tr := Transmission{Port: port}

	// run Add
	err := tr.Add(metainfo)
	if err == nil {
		t.Logf("expected error, but got: %v\n", err)
		t.FailNow()
	}
}

func TestResumeSuccess(t *testing.T) {
	t.Parallel()

	// prepare false test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("X-Transmission-Session-Id", token)

		// verify http method
		if r.Method != "POST" {
			t.Fail()
		}

		// check token
		if r.Header.Get("X-Transmission-Session-Id") != token {
			w.WriteHeader(http.StatusConflict)
			return
		}

		// verify request method
		c := &command{}
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(c)
		if c.Method != "torrent-start-now" {
			t.Fail()
		}

		// return get results
		w.WriteHeader(http.StatusOK)
		w.Write(resumeTorrentsSuccess)
	}))
	defer ts.Close()

	// parse port off ts.URL
	port, _ := strconv.Atoi(strings.Split(ts.URL, ":")[2])

	// configure transmission instance
	tr := Transmission{Port: port}

	// run Resume
	err := tr.Resume()
	if err != nil {
		t.Logf("unexpected error: %v\n", err)
		t.FailNow()
	}
}

func TestResumeFail(t *testing.T) {
	t.Parallel()

	// prepare false test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("X-Transmission-Session-Id", token)

		// verify http method
		if r.Method != "POST" {
			t.Fail()
		}

		// check token
		if r.Header.Get("X-Transmission-Session-Id") != token {
			w.WriteHeader(http.StatusConflict)
			return
		}

		// verify request method
		c := &command{}
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(c)
		if c.Method != "torrent-start-now" {
			t.Fail()
		}

		// return get results
		w.WriteHeader(http.StatusOK)
		w.Write(resumeTorrentsFail)
	}))
	defer ts.Close()

	// parse port off ts.URL
	port, _ := strconv.Atoi(strings.Split(ts.URL, ":")[2])

	// configure transmission instance
	tr := Transmission{Port: port}

	// run Resume
	err := tr.Resume()
	if err == nil {
		t.Logf("expected error, but got: %v\n", err)
		t.FailNow()
	}
}

func TestConfigureSuccess(t *testing.T) {
	tr := Transmission{}

	// unset error & set data
	fakeFSError = nil
	fakeFSData = []byte(`{"rpc-port": 3000}`)

	// run & verify
	err := tr.Configure("some-file-path")
	if err != nil || tr.Port != 3000 {
		t.FailNow()
	}
}

func TestConfigureSuccessEmptyPath(t *testing.T) {
	tr := Transmission{}

	// unset error & set data
	fakeFSError = nil
	fakeFSData = []byte(`{"rpc-port": 3000}`)

	// run & verify
	err := tr.Configure("")
	if err != nil || tr.Port != 3000 {
		t.FailNow()
	}
}

func TestConfigureFailReadFile(t *testing.T) {
	tr := Transmission{}

	// force error
	fakeFSError = fsError
	fakeFSData = []byte{}

	// run & verify error
	err := tr.Configure("")
	if err == nil {
		t.FailNow()
	}
}
