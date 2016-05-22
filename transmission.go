package transmission

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"sync"
	"time"
)

const transmissionConfigPath = "/etc/transmission-daemon/settings.json"

var readFile = ioutil.ReadFile

type filesystem interface {
	ReadFile(string) ([]byte, error)
}

type Transmission struct {
	sync.RWMutex
	Token     string `json:"-"`
	Downloads string `json:"download-dir"`
	Port      int    `json:"rpc-port"`
	Uri       string `json:"rpc-url"`
}

type torrent struct {
	Id       int  `json:"id,omitempty"`
	Finished bool `json:"isFinished,omitempty"`
}

type arguments struct {
	Torrents []torrent `json:"torrents,omitempty"`
	Ids      []int     `json:"ids,omitempty"`
	Fields   []string  `json:"fields,omitempty"`
	Location string    `json:"location,omitempty"`
	Metainfo string    `json:"metainfo,omitempty"`
	Move     bool      `json:"move,omitempty"`
}

type command struct {
	Method    string    `json:"method"`
	Result    string    `json:"result,omitempty"`
	Arguments arguments `json:"arguments,omitempty"`
}

var errorRetryFailed = errors.New("failed to get a valid response from transmission")

// consolidated method for sending http requests to transmission
// computes the endpoint, loops with 3 retries and grabbing tokens
// @link: https://trac.transmissionbt.com/browser/trunk/extras/rpc-spec.txt#L61
func (self *Transmission) send(cmd *command) ([]torrent, error) {

	// compute RPC address
	self.RLock()
	route := "http://127.0.0.1:" + path.Join(strconv.Itoa(self.Port), self.Uri, "rpc/")
	self.RUnlock()

	// error for later
	var results []torrent

	// json marshal cmd for request
	d, _ := json.Marshal(cmd)

	// prepare client to send requests
	c := http.Client{}

	// three-attempts per operation
	for i := 0; i < 3; i++ {

		// prepare request
		r, err := http.NewRequest("POST", route, bytes.NewReader(d))
		if err != nil {
			return results, err
		}

		// apply token header
		self.RLock()
		r.Header.Set("X-Transmission-Session-Id", self.Token)
		self.RUnlock()

		// deal with the aftermath
		resp, err := c.Do(r)
		if err != nil || resp == nil {
			time.Sleep(time.Second * 2)
			continue
		} else if resp.StatusCode == http.StatusConflict {
			self.Lock()
			self.Token = resp.Header.Get("X-Transmission-Session-Id")
			self.Unlock()
			continue
		} else if resp.StatusCode == http.StatusOK {
			rc := &command{}
			decoder := json.NewDecoder(resp.Body)
			decoder.Decode(rc)
			if rc.Result != "success" {
				time.Sleep(time.Second * 2)
				continue
			}
			results = rc.Arguments.Torrents
			return results, nil
		}
	}
	return results, errorRetryFailed
}

func (self *Transmission) ids(torrents ...torrent) []int {
	var ids []int
	for _, t := range torrents {
		ids = append(ids, t.Id)
	}
	return ids
}

func (self *Transmission) Configure(path string) error {
	if len(path) == 0 {
		path = transmissionConfigPath
	}

	// read file
	d, e := readFile(path)
	if e != nil {
		return e
	}

	// unmarshal onto self
	return json.Unmarshal(d, self)
}

// @link: https://trac.transmissionbt.com/browser/trunk/extras/rpc-spec.txt#L129
func (self *Transmission) Get() ([]torrent, error) {
	cmd := &command{Method: "torrent-get", Arguments: arguments{Fields: []string{"id", "isFinished"}}}
	return self.send(cmd)
}

func (self *Transmission) Finished() ([]torrent, error) {
	torrents, err := self.Get()
	var results []torrent
	for _, t := range torrents {
		if t.Finished == true {
			results = append(results, t)
		}
	}
	return results, err
}

// @link: https://trac.transmissionbt.com/browser/trunk/extras/rpc-spec.txt#L408
// @note: returns success status even if files are not moved due to permissions
//   be careful if misconfigured data may not be relocated, only unlinked of -r
func (self *Transmission) Move(path string, torrents []torrent) error {
	if len(torrents) == 0 {
		return nil
	}
	cmd := &command{Method: "torrent-set-location", Arguments: arguments{Ids: self.ids(torrents...), Location: path, Move: true}}
	_, err := self.send(cmd)
	return err
}

// @link: https://trac.transmissionbt.com/browser/trunk/extras/rpc-spec.txt#L394
func (self *Transmission) Remove(torrents []torrent) error {
	if len(torrents) == 0 {
		return nil
	}
	cmd := &command{Method: "torrent-remove", Arguments: arguments{Ids: self.ids(torrents...)}}
	_, err := self.send(cmd)
	return err
}

// @link: https://trac.transmissionbt.com/browser/trunk/extras/rpc-spec.txt#L358
func (self *Transmission) Add(meta string) error {
	cmd := &command{Method: "torrent-add", Arguments: arguments{Metainfo: meta}}
	_, err := self.send(cmd)
	return err
}

// @link: https://trac.transmissionbt.com/browser/trunk/extras/rpc-spec.txt#L76
func (self *Transmission) Resume() error {
	cmd := &command{Method: "torrent-start-now"}
	_, err := self.send(cmd)
	return err
}
