package investing

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/gorilla/websocket"
)

var (
	heartBeatMessag     = []byte(`["{\"_event\":\"heartbeat\",\"data\":\"h\"}"]`)
	startReceiveMessage = []byte(`["{\"_event\":\"UID\",\"UID\":0}"]`)
	subscriptMessage    = `["{\"_event\":\"subscribe\",\"tzID\":28,\"message\":\"pid-%d:\"}"]`
)

type conn struct {
	id                int
	conn              *websocket.Conn
	pids              []int
	client            *Client
	heartBeatInterval time.Duration
}

func newConn(id int, client *Client, pids []int, heartBeatInterval time.Duration) *conn {
	return &conn{
		id:                id,
		pids:              pids,
		client:            client,
		heartBeatInterval: heartBeatInterval,
	}
}
func (c *conn) start() error {
	if err := c.connect(); err != nil {
		return err
	}
	go c.handleMsg()
	return nil
}

func (c *conn) connect() error {
	cn, _, err := c.client.dialer.Dial(genUrl(c.client.ips[r.Intn(len(c.client.ips))]), c.client.head)
	if err != nil {
		return err
	}
	c.conn = cn
	if _, _, err := cn.ReadMessage(); err != nil {
		return err
	}

	for _, id := range c.client.pairIds {
		if err = cn.WriteMessage(websocket.TextMessage, getSubscribeMessage(id)); err != nil {
			return err
		}
	}

	if err = cn.WriteMessage(websocket.TextMessage, startReceiveMessage); err != nil {
		return err
	}
	go c.heartBeat(cn)
	return nil
}

func (c *conn) handleMsg() {
	defer func() {
		if e := recover(); e != nil {
			c.client.addError(fmt.Errorf("stack :%s,err:%v", string(debug.Stack()), e))
			time.Sleep(time.Minute)
			c.handleMsg()
		}
	}()

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			c.client.addError(err)
			c.conn.Close()
			c.tryConnect()
		} else {
			if m, err := isDataMsg(msg); err != nil {
				c.client.addError(err)
			} else if len(m) != 0 {
				c.client.channel.add(m)
			}
		}
	}
}
func (c *conn) tryConnect() {
	for err := c.connect(); err != nil; err = c.connect() {
		c.client.addError(err)
	}
}

func (c *conn) heartBeat(cn *websocket.Conn) {
	for {
		if err := c.writeHeartBeat(cn); err != nil {
			c.client.addError(err)
			cn.Close()
			break
		}
		time.Sleep(c.heartBeatInterval)
	}
}

func (c *conn) writeHeartBeat(cn *websocket.Conn) error {
	return cn.WriteMessage(websocket.TextMessage, heartBeatMessag)
}

func getSubscribeMessage(pid int) []byte {
	message := fmt.Sprintf(subscriptMessage, pid)
	return []byte(message)
}

func genUrl(i int) string {
	return fmt.Sprintf("wss://stream%d.forexpros.com/echo/%03d/%s/websocket", i, r.Intn(1000), randomString(8))
}
func isDataMsg(msg []byte) ([]byte, error) {
	if msg[0] == 'a' {
		msg = msg[1:]
		var msgPackage []string
		err := json.Unmarshal(msg, &msgPackage)
		if err != nil {
			return nil, err
		}
		for _, element := range msgPackage {
			var m = make(map[string]string)
			if err := json.Unmarshal([]byte(element), &m); err != nil {
				return nil, err
			}
			for key, val := range m {
				if key == "message" {
					return []byte(val), nil
				}
				if key == "_event" {
					return nil, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("unknown message package:%s", string(msg))
}
