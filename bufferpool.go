package bufferpool

//
//
//
//

import "sync"

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

func New() (p *Pool) {
	p = &Pool{
		pool: make([]*Buffer, 0),
		ch:   make(chan *Buffer, DefaultPoolSize),
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
	return p
}

func (p *Pool) Count() (c int) {
	return p.count
}

func (p *Pool) Get() (b *Buffer) {
	select {
	case b := <-p.ch:
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
		b = &Buffer{
			data: make([]byte, 0, DefaultBufSize),
			pool: p,
		}
		p.count++
		b.serial = p.count
	}
	b.used = true
	return b
}

//
// putpool()
// Return buffer back to the pool
// unless the capacity has passed the max size
//
func (p *Pool) putpool(b *Buffer) {
	if cap(b.data) > maxBufSize {
		p.count--
		if p.count < 0 {
			panic("ran out of buffers")
		}
		return
	}

	// Zero out the buffer data
	for i := range b.data {
		b.data[i] = 0
	}
	b.data = b.data[:0]
	p.poolmx.Lock()
	defer p.poolmx.Unlock()
	b.used = false
	p.pool = append(p.pool, b)
}

func (b *Buffer) ReturnToPool() {
	if !b.used {
		panic("Unused buffer return")
	}
	b.pool.putpool(b)
}

func (b *Buffer) Size() int {
	return len(b.data)
}

func (b *Buffer) Copy() (copy *Buffer) {
	copy = b.pool.getpool()
	copy.Append(b.data)
	return copy
}

func (b *Buffer) Append(d []byte) {
	b.data = append(b.data, d...)
}

func (b *Buffer) Data() (d []byte) {
	return b.data
}

func (b *Buffer) Serial() (s int) {
	return b.serial
}
