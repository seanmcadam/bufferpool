package bufferpool

//
//
//
//

import (
	"sync"

	"github.com/seanmcadam/loggy"
)

const DefaultPoolSize = 2
const DefaultBufSize = 2048
const maxBufSize = 2048

type Pool struct {
	poolmx sync.Mutex
	pool   []*Buffer
	ch     chan *Buffer
	count  int
}

type Buffer struct {
	serial int
	data   []byte
	pool   *Pool
	used   bool
}

var globalPool *Pool
var once sync.Once

func initGlobalPool() {
	p := &Pool{
		count: 0,
		pool:  make([]*Buffer, 0),
		ch:    make(chan *Buffer, DefaultPoolSize),
	}

	//
	// Get a buffer ready for output -> out
	//
	go func(p *Pool) {
		for {
			select {
			case p.ch <- p.getpool():
			}
		}
	}(p)

	globalPool = p

}

func New() *Pool {
	once.Do(initGlobalPool)
	return globalPool
}

func (p *Pool) Count() (c int) {
	if p == nil {
		loggy.FatalStack("nil method pointer")
	}
	return p.count
}

func (p *Pool) Get() (b *Buffer) {
	if p == nil {
		loggy.FatalStack("nil method pointer")
	}
	select {
	case b := <-p.ch:
		b.used = true
		return b
	}
}

func (p *Pool) getpool() (b *Buffer) {
	if len(p.pool) > 0 {
		p.poolmx.Lock()
		defer p.poolmx.Unlock()
		b = p.pool[len(p.pool)-1]
		p.pool = append(p.pool[:len(p.pool)-1])
	} else {
		p.count++
		b = &Buffer{
			used:   true,
			data:   make([]byte, 0, DefaultBufSize),
			pool:   p,
			serial: p.count,
		}
		p.count++
		b.serial = p.count
	}
	return b
}

// putpool()
// Return buffer back to the pool
// unless the capacity has passed the max size
func (b *Buffer) putpool() {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	if b.pool == nil {
		loggy.FatalStack("pool is nil")
	}
	if cap(b.data) > maxBufSize {
		b.pool.count--
		if b.pool.count < 0 {
			loggy.FatalStack("ran out of buffers")
		}
		return
	}

	// Zero out the buffer data
	for i := range b.data {
		b.data[i] = 0
	}
	b.data = b.data[:0]
	b.pool.poolmx.Lock()
	defer b.pool.poolmx.Unlock()
	b.used = false
	b.pool.pool = append(b.pool.pool, b)
}

func (b *Buffer) ReturnToPool() {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	if !b.used {
		loggy.FatalStack("Unused buffer return")
	}
	b.putpool()
}

func (b *Buffer) Size() (size int) {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	if b.data == nil {
		loggy.Errorf("Buffer Data Nil:%v", b)
		size = 0
	} else {
		size = len(b.data)
	}
	return size
}

func (b *Buffer) Copy() (copy *Buffer) {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	copy = b.pool.Get()
	copy.Append(b.data)
	return copy
}

func (b *Buffer) Append(d []byte) *Buffer {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	b.data = append(b.data, d...)
	return b
}

func (b *Buffer) Data() (d []byte) {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	return b.data
}

func (b *Buffer) Serial() (s int) {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	return b.serial
}
