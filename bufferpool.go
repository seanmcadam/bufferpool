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

const defaultPoolSize = 16
const defaultMaxPoolSize uint = 1024
const defaultLgBufSize uint = 2048
const defaultMedBufSize uint = 512
const defaultSmBufSize uint = 64
const maxBufSize uint = 2048

type PoolSize struct {
	poolmx sync.Mutex
	pool   []*Buffer
	ch     chan *Buffer
	count  uint
	max    uint
}

type Pool struct {
	count uint
	lg    *PoolSize
	med   *PoolSize
	sm    *PoolSize
}

type Buffer struct {
	serial counter.Count
	data   []byte
	pool   *Pool
	used   bool
}

var globalPool *Pool
var bufferSerial counter.Counter
var once sync.Once

// --
// --
func initGlobalPool() {
	bufferSerial = counter.New(ctx.New(), counter.BIT32)
	p := &Pool{
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
			max:   defaultLgBufSize,
		},
		sm: &PoolSize{
			pool:  make([]*Buffer, 0),
			ch:    make(chan *Buffer, defaultPoolSize),
			count: 0,
			max:   defaultLgBufSize,
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
		serial: bufferSerial.Next(),
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
	ps.poolmx.Lock()
	defer ps.poolmx.Unlock()
	ps.pool = append(ps.pool, b)
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

func (b *Buffer) ReturnToPool() {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	if !b.used {
		loggy.FatalStack("Unused buffer return")
	}
	c := b.Cap()
	if c > globalPool.med.max {
		globalPool.lg.put(b)
	} else if c > globalPool.sm.max {
		globalPool.med.put(b)
	} else {
		globalPool.sm.put(b)
	}
}

func (b *Buffer) Cap() (size uint) {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	if b.data == nil {
		loggy.Errorf("Buffer Data Nil:%v", b)
		size = 0
	} else {
		size = uint(cap(b.data))
	}
	return size
}

func (b *Buffer) Size() (size int) {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	if b.used == false {
		loggy.FatalStack("inactive buffer")
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
	if b.used == false {
		loggy.FatalStack("inactive buffer")
	}
	copy = b.pool.Get()
	copy.Append(b.data)
	return copy
}

func (b *Buffer) Append(d []byte) *Buffer {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	if b.used == false {
		loggy.FatalStack("inactive buffer")
	}
	b.data = append(b.data, d...)
	return b
}

func (b *Buffer) Trim(c int) {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	if b.used == false {
		loggy.FatalStack("inactive buffer")
	}
	l := len(b.data)
	if c > l {
		loggy.FatalStack("not enough data to trim")
	}
	b.data = b.data[:l-c]
}

func (b *Buffer) Data() (d []byte) {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	if b.used == false {
		loggy.FatalStack("inactive buffer")
	}
	return b.data
}

func (b *Buffer) Serial() (s uint64) {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	return b.serial.Uint()
}

func (b *Buffer) Used() bool {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	return b.used
}
