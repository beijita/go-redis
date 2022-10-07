package connection

import (
	"go-redis/utils/sync/wait"
	"net"
	"sync"
	"time"
)

type Connection struct {
	conn         net.Conn
	waitingReply wait.Wait
	mu           sync.Mutex
	subMap       map[string]bool
	password     string
	multiState   bool
	queue        [][][]byte
	watching     map[string]uint32
	txErrors     []error
	selectedDB   int
	role         int32
}

func (c *Connection) AddTxError(err error) {
	c.txErrors = append(c.txErrors, err)
}

func (c *Connection) GetTxErrors() []error {
	return c.txErrors
}

func (c *Connection) GetRole() int32 {
	return c.role
}

func (c *Connection) SetRole(role int32) {
	c.role = role
}

func (c *Connection) Write(bytes []byte) error {
	if len(bytes) <= 0 {
		return nil
	}
	c.waitingReply.Add(1)
	defer func() { c.waitingReply.Done() }()
	_, err := c.conn.Write(bytes)
	return err
}

func (c *Connection) SetPassword(password string) {
	c.password = password
}

func (c *Connection) GetPassword() string {
	return c.password
}

func (c *Connection) Subscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.subMap == nil {
		c.subMap = make(map[string]bool)
	}
	c.subMap[channel] = true
}

func (c *Connection) Unsubscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.subMap) == 0 {
		return
	}
	delete(c.subMap, channel)
}

func (c *Connection) SubCount() int {
	return len(c.subMap)
}

func (c *Connection) GetChannels() []string {
	if c.subMap == nil {
		return make([]string, 0)
	}
	lenSub := len(c.subMap)
	i := 0
	channelList := make([]string, lenSub)
	for channel := range c.subMap {
		channelList[i] = channel
		i++
	}
	return channelList
}

func (c *Connection) InMultiState() bool {
	return c.multiState
}

func (c *Connection) SetMultiState(state bool) {
	if !state {
		c.watching = nil
		c.queue = nil
	}
	c.multiState = state
}

func (c *Connection) GetQueueCmdLine() [][][]byte {
	return c.queue
}

func (c *Connection) EnqueueCmd(cmdLine [][]byte) {
	c.queue = append(c.queue, cmdLine)
}

func (c *Connection) ClearQueuedCmd() {
	c.queue = nil
}

func (c *Connection) GetWatching() map[string]uint32 {
	if c.watching == nil {
		c.watching = make(map[string]uint32)
	}
	return c.watching
}

func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

func (c *Connection) SelectDB(dbNum int) {
	c.selectedDB = dbNum
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{conn: conn}
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) Close() error {
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	return c.conn.Close()
}
