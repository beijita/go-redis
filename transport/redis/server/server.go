package server

import (
	"context"
	"go-redis/interface/database"
	"go-redis/transport/redis/connection"
	"go-redis/transport/redis/parser"
	"go-redis/transport/redis/protocol"
	"go-redis/utils/logger"
	"go-redis/utils/sync/atomic"
	"io"
	"net"
	"strings"
	"sync"
)

type Handler struct {
	activeConn sync.Map
	db         database.DB
	closing    atomic.Boolean
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h Handler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		conn.Close()
		return
	}
	client := connection.NewConnection(conn)
	h.activeConn.Store(client, 1)

	ch := parser.ParseStream(conn)
	for payload := range ch {
		if payload.Err != nil {
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			errReply := protocol.NewErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				h.closeClient(client)
				logger.Info("")
				return
			}
			continue
		}
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}
		reply, ok := payload.Data.(*protocol.MultiBulkReply)
		if !ok {
			logger.Error(" require multi bulk protocol ")
			continue
		}
		result := h.db.Exec(client, reply.Args)
		if result != nil {
			client.Write(result.ToBytes())
		}
	}
}

func (h *Handler) Close() error {
	logger.Info()
	h.closing.Set(true)
	h.activeConn.Range(func(key, value any) bool {
		key.(*connection.Connection).Close()
		return true
	})
	h.db.Close()
	return nil
}

func (h *Handler) closeClient(client *connection.Connection) {

}
