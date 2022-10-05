package tcp

import (
	"bufio"
	"go-redis/transport/tcp/echo"
	"go-redis/utils/logger"
	"math/rand"
	"net"
	"strconv"
	"testing"
)

func TestListenAndServe(t *testing.T) {
	listener, _ := net.Listen("tcp", ":0")
	closeChan := make(chan struct{})
	go listenAndServe(listener, echo.NewEchoHandler(), closeChan)
	addr := listener.Addr().String()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}
	for i := 0; i < 10; i++ {
		val := strconv.Itoa(rand.Int())
		_, err := conn.Write([]byte(val + "\n"))
		if err != nil {
			return
		}
		reader := bufio.NewReader(conn)
		line, _, err := reader.ReadLine()
		if err != nil {
			return
		}
		if string(line) != val {
			logger.Error(" error response ")
			return
		}
		logger.Info(string(line), val)
	}
	conn.Close()
	for i := 0; i < 5; i++ {
		net.Dial("tcp", addr)
	}
	//time.Sleep(5 * time.Second)
	closeChan <- struct{}{}
}
