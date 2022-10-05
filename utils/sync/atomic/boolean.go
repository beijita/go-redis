package atomic

import "sync/atomic"

type Boolean uint64

func (b *Boolean) Get() bool {
	return atomic.LoadUint64((*uint64)(b)) != 0
}

func (b *Boolean) Set(value bool) {
	if value {
		atomic.StoreUint64((*uint64)(b), 1)
	} else {
		atomic.StoreUint64((*uint64)(b), 0)
	}
}
