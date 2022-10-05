package echo

import (
	"bufio"
	"context"
	"go-redis/utils/logger"
	"go-redis/utils/sync/atomic"
	"io"
	"net"
	"sync"
)

type EchoHandler struct {
	activeConn sync.Map
	closing    atomic.Boolean
}

func NewEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		conn.Close()
		return
	}

	client := &EchoClient{Conn: conn}
	h.activeConn.Store(client, struct{}{})
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				h.activeConn.Delete(client)
			} else {
				logger.Error(err)
			}
			return
		}
		client.Waiting.Add(1)
		conn.Write([]byte(msg))
		client.Waiting.Done()
	}
}

func (h *EchoHandler) Close() error {
	h.closing.Set(true)
	h.activeConn.Range(func(key, value any) bool {
		err := key.(*EchoClient).Close()
		return err == nil
	})
	return nil
}
