package investing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var clientIps = []int{
	1, 3, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 28, 29, 30,
	31, 32, 38, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 59, 60, 61, 62, 63,
	64, 67, 73, 79, 80, 81, 83, 84, 86, 88, 91, 92, 96, 99, 102, 103, 105, 108, 109, 110, 117, 118, 120,
	121, 122, 123, 124, 125, 126, 127, 129, 130, 131, 132, 133, 134, 135, 136, 137, 138, 139, 140, 141,
	142, 143, 144, 145, 146, 147, 148, 149, 150, 151, 152, 153, 154, 155, 156, 157, 158, 159, 160, 161,
	162, 163, 164, 165, 166, 167, 168, 169, 170, 171, 173, 174, 175, 176, 177, 178, 179, 180, 181, 182,
	183, 184, 185, 186, 187, 188, 189, 190, 191, 192,
}

type Client struct {
	pairIds      []int
	connNum      int
	sendInterval time.Duration
	ips          []int //可用ip的最后一个字节
	channel      *channel
	Chan         chan struct{}
	dialer       *websocket.Dialer
	head         http.Header
	errors       []error
	errMux       sync.Mutex
}

func NewClient(pairIds []int, connNum int, sendInterval time.Duration) *Client {
	c := &Client{
		pairIds:      pairIds,
		sendInterval: sendInterval,
		ips:          clientIps,
		Chan:         make(chan struct{}, 1),
		dialer:       websocket.DefaultDialer,
		head:         http.Header{},
	}
	c.head.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	c.head.Set("Origin", "https://cn.investing.com")
	c.dialer.HandshakeTimeout = time.Second * 10
	c.channel = newChannel(100, sendInterval, c.Chan)

	l := len(pairIds)
	c.connNum = l / connNum
	if l%connNum != 0 {
		c.connNum += 1
	}
	return c
}

func (c *Client) Receive() []RealData {
	msgs := c.channel.get()
	data := make([]RealData, len(msgs))
	i := 0
	for _, msg := range msgs {
		if err := json.Unmarshal(msg[bytes.Index(msg, []byte("::"))+2:], &data[i]); err != nil {
			c.addError(fmt.Errorf("%s:%s", err.Error(), string(msg)))
		} else {
			i++
		}
	}
	return data[:i]
}

func (c *Client) Start() error {
	l := len(c.pairIds)
	n := c.connNum
	var firstErr error
	nPerConn := (l / n)
	if l%n > 0 {
		nPerConn++
	}
	for i := 0; i < n; i++ {
		start := i * nPerConn
		end := start + nPerConn
		if end > l {
			end = l
		}
		e := newConn(i, c, c.pairIds[start:end], 4*time.Second).start()
		if firstErr == nil {
			firstErr = e
		} else {
			c.addError(e)
		}
	}
	return firstErr
}

func (c *Client) addError(err error) {
	c.errMux.Lock()
	c.errors = append(c.errors, err)
	c.errMux.Unlock()
}

func (c *Client) GetErrors() (errors []error) {
	c.errMux.Lock()
	errors = make([]error, len(c.errors))
	copy(errors, c.errors)
	c.errors = c.errors[:0]
	c.errMux.Unlock()
	return
}

type RealData struct {
	PairId      json.Number `json:"pid"`
	LastDir     string      `json:"last_dir"`
	LastNumeric float64     `json:"last_numeric"`
	Last        json.Number
	Bid         json.Number
	Ask         json.Number
	High        json.Number
	Low         json.Number
	Pc          json.Number
	Pcp         json.Number
	PcCol       string `json:"pc_col"`
	Time        string
	Timestamp   int
}
