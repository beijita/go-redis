package tcp

import (
	"context"
	"go-redis/interface/tcp"
	"go-redis/utils/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Config struct {
	Address    string        `yml:"address"`
	MaxConnect int           `yml:"max_connect"`
	Timeout    time.Duration `yml:"timeout"`
}

func ListenAndServeWithSignal(config *Config, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	listener, err := net.Listen("tcp", config.Address)
	if err != nil {
		logger.Error(err)
		return err
	}
	listenAndServe(listener, handler, closeChan)
	return nil
}

func listenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	// 监听信号 关闭观察者
	go func() {
		<-closeChan
		logger.Info("shutting down...")
		listener.Close()
		handler.Close()
	}()
	// 函数退出关闭观察者
	defer func() {
		listener.Close()
		handler.Close()
	}()
	ctx := context.Background()
	var wg sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error(err)
			break
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			handler.Handle(ctx, conn)
		}()
	}
	wg.Wait()
}
