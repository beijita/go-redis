package echo

import (
	"go-redis/utils/sync/wait"
	"net"
	"time"
)

type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait
}

func (c EchoClient) Close() error {
	c.Waiting.WaitWithTimeout(10 * time.Second)
	return c.Conn.Close()
}
