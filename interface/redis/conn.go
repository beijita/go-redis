package redis

type Connection interface {
	Write([]byte) error
	SetPassword(string)
	GetPassword() string

	Subscribe(channel string)
	Unsubscribe(channel string)
	SubCount() int
	GetChannels() []string

	InMultiState() bool
	SetMultiState(bool)
	GetQueueCmdLine() [][][]byte
	EnqueueCmd([][]byte)
	ClearQueuedCmd()
	GetWatching() map[string]uint32
	AddTxError(err error)
	GetTxErrors() []error

	GetDBIndex() int
	SelectDB(int)
	GetRole() int32
	SetRole(int32)
}
