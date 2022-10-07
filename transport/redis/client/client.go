package client

import (
	"errors"
	"go-redis/interface/redis"
	"go-redis/transport/redis/parser"
	"go-redis/transport/redis/protocol"
	"go-redis/utils/logger"
	"go-redis/utils/sync/wait"
	"net"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	created = iota
	running
	closed
)

const (
	chanSize = 256
	maxWait  = 3 * time.Second
)

type Client struct {
	conn        net.Conn
	pendingReqs chan *request
	waitingReqs chan *request
	ticker      *time.Ticker
	addr        string
	status      int32
	working     *sync.WaitGroup
}

type request struct {
	id        uint64
	args      [][]byte
	reply     redis.Reply
	heartbeat bool
	waiting   *wait.Wait
	err       error
}

func NewClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Client{
		addr:        addr,
		conn:        conn,
		pendingReqs: make(chan *request, chanSize),
		waitingReqs: make(chan *request, chanSize),
		working:     &sync.WaitGroup{},
	}, nil
}

func (c *Client) Start() {
	c.ticker = time.NewTicker(10 * time.Second)
	go c.handleWrite()
	go c.handleRead()
	go c.heartbeat()
	atomic.StoreInt32(&c.status, running)
}

func (c *Client) handleWrite() {
	for req := range c.pendingReqs {
		c.doRequest(req)
	}
}

func (c *Client) handleRead() {
	ch := parser.ParseStream(c.conn)
	for payload := range ch {
		if payload.Err != nil {
			status := atomic.LoadInt32(&c.status)
			if status == closed {
				return
			}
			c.reconnect()
			return
		}
		c.finishRequest(payload.Data)
	}
}

func (c *Client) heartbeat() {
	for range c.ticker.C {
		c.doHeartbeat()
	}
}

func (c *Client) doRequest(req *request) {
	if req == nil || len(req.args) == 0 {
		return
	}
	reply := protocol.NewMultiBulkReply(req.args)
	bytes := reply.ToBytes()
	var err error
	for i := 0; i < 3; i++ {
		_, err = c.conn.Write(bytes)
		if err == nil || (!strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline exceeded")) {
			break
		}
	}
	if err == nil {
		c.waitingReqs <- req
	} else {
		req.err = err
		req.waiting.Done()
	}
}

func (c *Client) reconnect() {
	logger.Info("reconnect with " + c.addr)
	c.conn.Close()

	var conn net.Conn
	for i := 0; i < 3; i++ {
		var err error
		conn, err = net.Dial("tcp", c.addr)
		if err != nil {
			logger.Error("reconnect error: " + err.Error())
			time.Sleep(time.Second)
			continue
		} else {
			break
		}
	}
	if conn == nil {
		c.Close()
		return
	}
	c.conn = conn
	close(c.waitingReqs)
	for req := range c.waitingReqs {
		req.err = errors.New("")
		req.waiting.Done()
	}
	c.waitingReqs = make(chan *request, chanSize)
	go c.handleRead()
}

func (c *Client) finishRequest(reply redis.Reply) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			logger.Error(err)
		}
	}()

	req := <-c.waitingReqs
	if req == nil {
		return
	}
	req.reply = reply
	if req.waiting != nil {
		req.waiting.Done()
	}
}

func (c *Client) Close() {
	atomic.StoreInt32(&c.status, closed)
	c.ticker.Stop()
	close(c.pendingReqs)
}

func (c *Client) doHeartbeat() {
	req := &request{
		args:      [][]byte{[]byte("PING")},
		heartbeat: true,
		waiting:   &wait.Wait{},
	}
	req.waiting.Add(1)
	c.working.Add(1)
	defer c.working.Done()
	c.pendingReqs <- req
	req.waiting.WaitWithTimeout(maxWait)
}
