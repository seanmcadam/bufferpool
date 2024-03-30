package bufferpool

//
//
//
//

import (
	"github.com/seanmcadam/counter"
	"github.com/seanmcadam/loggy"
)

type Buffer struct {
	serial counter.Count
	data   []byte
	pool   *Pool
	used   bool
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

func (b *Buffer) Trim(c int) (ret []byte) {
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
	if b.Size() < c {
		loggy.FatalfStack("trim size:%d > buffer size:%d", c, b.Size())
	}
	ret = b.data[l-c:]
	b.data = b.data[:l-c]
	return ret
}

func (b *Buffer) TrimByte() (ret byte) {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	if b.used == false {
		loggy.FatalStack("inactive buffer")
	}
	l := len(b.data)
	if 1 > l {
		loggy.FatalStack("not enough data to trim")
	}
	if b.Size() == 0 {
		loggy.FatalfStack("buffer is empty")
	}
	ret = b.data[l-1]
	b.data = b.data[:l-1]
	return ret
}

func (b *Buffer) PeekByte() (ret byte) {
	if b == nil {
		loggy.FatalStack("nil method pointer")
	}
	if b.used == false {
		loggy.FatalStack("inactive buffer")
	}
	l := len(b.data)
	if 1 > l {
		loggy.FatalStack("not enough data to trim")
	}
	if b.Size() == 0 {
		loggy.FatalfStack("buffer is empty")
	}
	ret = b.data[l-1]
	return ret
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
