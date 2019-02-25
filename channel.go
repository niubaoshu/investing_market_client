package investing

import (
	"sync"
	"time"
)

type channel struct {
	mx        sync.Mutex
	cache     [][]byte
	cache2    [][]byte
	fullNum   int
	FullChan  chan struct{}
	isNoticed bool
}

func newChannel(fullNum int, interval time.Duration, ch chan struct{}) *channel {
	c := &channel{
		fullNum:  fullNum,
		cache:    make([][]byte, 0, fullNum*2),
		cache2:   make([][]byte, 0, fullNum*2),
		FullChan: ch,
	}
	go func() {
		for {
			time.Sleep(interval)
			if len(c.cache) > 0 {
				c.FullChan <- struct{}{}
			}
		}
	}()
	return c
}

func (c *channel) add(msg []byte) {
	var needNotice bool
	c.mx.Lock()
	c.cache = append(c.cache, msg)
	needNotice = len(c.cache) >= c.fullNum && !c.isNoticed
	if needNotice {
		c.isNoticed = true
	}
	c.mx.Unlock()
	if needNotice {
		c.FullChan <- struct{}{}
	}
}

func (c *channel) get() (ret [][]byte) {
	c.mx.Lock()
	c.cache, c.cache2 = c.cache2, c.cache
	c.cache = c.cache[:0]
	c.isNoticed = false
	c.mx.Unlock()
	return c.cache2
}

func (c *channel) len() int {
	c.mx.Lock()
	l := len(c.cache)
	c.mx.Unlock()
	return l
}
