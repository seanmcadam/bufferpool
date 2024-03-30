package bufferpool

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestCompileCheck(t *testing.T) {
}

func TestNew(t *testing.T) {

	p := New()

	b := p.Get()
	if b.Size() != 0 {
		t.Errorf("Buf.Size() is not 0: %d", b.Size())
	}

	b.Append([]byte("x"))
	if b.Size() != 1 {
		t.Errorf("Buf.Size() is not 1: %d", b.Size())
	}

	for i := 0; i < int(defaultLgBufSize); i++ {
		b.Append([]byte("x"))
	}

	if b.Size() != int(defaultLgBufSize+1) {
		t.Errorf("Buf.Size() is not 1: %d", b.Size())
	}

	b.ReturnToPool()

}

func TestBufferNew(t *testing.T) {

	const hithere = "Hi There........................................................................................................................................................."
	p := New()
	b := p.GetSm()

	b.Append([]byte(hithere))
	if b.Size() != len(hithere) {
		t.Errorf("Buffer Size mismatch")
	}

	b.ReturnToPool()
	p.GetLg().ReturnToPool()
	p.GetLg().ReturnToPool()
	p.GetLg().ReturnToPool()
	p.GetLg().ReturnToPool()
	p.GetLg().ReturnToPool()
	p.GetLg().ReturnToPool()
	p.GetMed().ReturnToPool()
	p.GetMed().ReturnToPool()
	p.GetMed().ReturnToPool()
	p.GetMed().ReturnToPool()
	p.GetMed().ReturnToPool()
	p.GetMed().ReturnToPool()
	p.GetSm().ReturnToPool()
	p.GetSm().ReturnToPool()
	p.GetSm().ReturnToPool()
	p.GetSm().ReturnToPool()
	p.GetSm().ReturnToPool()
	p.GetSm().ReturnToPool()

	b = p.GetSm()
	b.Append([]byte(hithere))
	b.Append([]byte(hithere))
	b.Append([]byte(hithere))

	c := b.Copy()
	defer c.ReturnToPool()
	d := b.Copy()
	defer d.ReturnToPool()
	b.ReturnToPool()

	if c.Size() != d.Size() {
		t.Errorf("Copy sizes dont match")
	}
}

func TestBufferLoad(t *testing.T) {
	const hithere = "FillerData"

	var wg sync.WaitGroup
	depth := 1024

	p := New()

	wg.Add(1)
	go func(p *Pool) {
		var arr = make([]*Buffer, 0, depth)

		defer func() {
			for _, b := range arr {
				b.ReturnToPool()
			}
			wg.Done()
		}()

		for i := 0; i < depth; i++ {
			b := p.Get()
			RandBufferFill(b)
			arr = append(arr, b)
		}

	}(p)

	wg.Add(1)
	go func(p *Pool) {
		var arr = make([]*Buffer, 0, depth)

		defer func() {
			for _, b := range arr {
				b.ReturnToPool()
			}
			wg.Done()
		}()

		for i := 0; i < depth; i++ {
			b := p.Get()
			RandBufferFill(b)
			arr = append(arr, b)
			if i > 0 && 0 == i%2 {
				randIndex := rand.Intn(len(arr) - 1)
				newArr := make([]*Buffer, len(arr)-1)
				copy(newArr[:randIndex], arr[:randIndex])
				copy(newArr[randIndex:], arr[randIndex+1:])
				arr = newArr
			}
		}

	}(p)

	wg.Wait()
	t.Logf("Pool Count %d", p.Count())
}

func RandBufferFill(b *Buffer) {
	size := rand.Intn(4098 - 1)
	for i := 0; i < size; i++ {
		b.Append(getRandomPrintableChar())
	}
}

func getRandomPrintableChar() []byte {
	b := make([]byte, 1, 1)
	b[0] = byte(rand.Intn(94) + 32)
	return b
}

func TestBufferQty(t *testing.T) {
	pool := New()

	var bufsm []*Buffer = make([]*Buffer, 0, defaultPoolSize)
	var bufmed []*Buffer = make([]*Buffer, 0, defaultPoolSize)
	var buflg []*Buffer = make([]*Buffer, 0, defaultPoolSize)

	for i := 0; i < defaultPoolSize; i++ {
		bufsm = append(bufsm, pool.GetSm())
		bufmed = append(bufmed, pool.GetMed())
		buflg = append(buflg, pool.GetLg())
	}

	for _, b := range bufsm {
		b.ReturnToPool()
	}
	for _, b := range bufmed {
		b.ReturnToPool()
	}
	for _, b := range buflg {
		b.ReturnToPool()
	}

}

func TestBufferMethods(t *testing.T) {
	pool := New()

	b := pool.Get()

	b.Append([]byte("testing"))
	b1 := b.PeekByte()
	if b1 != byte('g') {
		t.Errorf("PeekByte() != g")
	}

	b2 := b.TrimByte()
	if b2 != byte('g') {
		t.Errorf("TrimByte() != g")
	}
	if b.Size() != 6 {
		t.Errorf("Trim() size != 6")
	}

	b3 := b.Trim(4)
	if string(b3) != "stin" {
		t.Errorf("Trim(4) != stin")
	}
	if b.Size() != 2 {
		t.Errorf("Trim() size != 2")
	}

	b4 := b.Data()
	if string(b4) != "te" {
		t.Errorf("Data != 'te'")
	}

	_ = b.Serial()

	if !b.Used() {
		t.Errorf("Buffer is not marked used")
	}

	b.ReturnToPool()

}
