package bufferpool

//
//
//
//

import (
	"sync"

	"github.com/seanmcadam/counter"
	"github.com/seanmcadam/ctx"
	"github.com/seanmcadam/loggy"
)

type PoolSize struct {
	poolmx sync.Mutex
	pool   []*Buffer
	ch     chan *Buffer
	count  uint
	max    uint
}

type Pool struct {
	count        uint
	bufferSerial counter.Counter
	lg           *PoolSize
	med          *PoolSize
	sm           *PoolSize
}

var globalPool *Pool
var once sync.Once

// --
// --
func initGlobalPool() {
	p := &Pool{
		bufferSerial: counter.New(ctx.New(), counter.BIT32),
		lg: &PoolSize{
			pool:  make([]*Buffer, 0),
			ch:    make(chan *Buffer, defaultPoolSize),
			count: 0,
			max:   defaultLgBufSize,
		},
		med: &PoolSize{
			pool:  make([]*Buffer, 0),
			ch:    make(chan *Buffer, defaultPoolSize),
			count: 0,
			max:   defaultMedBufSize,
		},
		sm: &PoolSize{
			pool:  make([]*Buffer, 0),
			ch:    make(chan *Buffer, defaultPoolSize),
			count: 0,
			max:   defaultSmBufSize,
		},
	}
	p.lg.goChan()
	p.med.goChan()
	p.sm.goChan()

	globalPool = p

}

// --
//
// --
func New() *Pool {
	once.Do(initGlobalPool)
	return globalPool
}

// --
//
// --
func (p *Pool) Get() (b *Buffer) {
	return p.GetLg()
}

// --
//
// --
func (p *Pool) GetLg() (b *Buffer) {
	return p.lg.getch()
}

// --
//
// --
func (p *Pool) GetMed() (b *Buffer) {
	return p.med.getch()
}

// --
//
// --
func (p *Pool) GetSm() (b *Buffer) {
	return p.sm.getch()
}

// --
//
// --
func (p *Pool) Count() uint {
	return p.count
}

// --
//
// --
func (p *Pool) newBuffer(size uint) (b *Buffer) {
	p.count++
	b = &Buffer{
		serial: p.bufferSerial.Next(),
		used:   true,
		data:   make([]byte, 0, size),
		pool:   p,
	}
	return b
}

// --
// Pool Size
// --
func (ps *PoolSize) Count() uint {
	return ps.count
}

// --
//
// --
func (ps *PoolSize) getch() (b *Buffer) {
	if ps == nil {
		loggy.FatalStack("nil method pointer")
	}
	select {
	case b := <-ps.ch:
		b.used = true
		return b
	}
}

// --
//
// --
func (ps *PoolSize) get() (b *Buffer) {
	if len(ps.pool) > 0 {
		ps.poolmx.Lock()
		defer ps.poolmx.Unlock()
		b = ps.pool[len(ps.pool)-1]
		ps.pool = append(ps.pool[:len(ps.pool)-1])
	} else {
		b = globalPool.newBuffer(ps.max)
		ps.count++
	}
	return b
}

// --
//
// --
func (ps *PoolSize) put(b *Buffer) {
	if ps.count < defaultPoolSize {
		b.used = false
		b.data = b.data[:0]
		ps.poolmx.Lock()
		defer ps.poolmx.Unlock()
		ps.pool = append(ps.pool, b)
	} else {
		loggy.Debug("Dropping buffer")
	}
}

func (ps *PoolSize) goChan() {
	go func(ps *PoolSize) {
		for {
			select {
			case ps.ch <- ps.get():
			}
		}
	}(ps)
}
