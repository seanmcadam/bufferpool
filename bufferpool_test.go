package bufferpool

import (
	"testing"
)

func TestCompileCheck(t *testing.T) {
}

func TestNew(t *testing.T) {

	p := New()

	//
	// Count should be 2, one in the pipe, one ready to go
	//
	if p.Count() != 2 {
		t.Errorf("Count() is not zero")
	}

	b := p.Get()
	b.ReturnToPool()

	if p.Count() != 3 {
		t.Errorf("Count() is not zero")
	}
}

func TestBufferNew(t *testing.T) {

	const hithere = "Hi There"

	p := New()

	b := p.Get()

	b.Append([]byte(hithere))
	if b.Size() != len(hithere){
		t.Errorf("Buffer Size mismatch")
	}

	b.ReturnToPool()
	p.Get().ReturnToPool()
	p.Get().ReturnToPool()
	p.Get().ReturnToPool()
	p.Get().ReturnToPool()
	p.Get().ReturnToPool()
	p.Get().ReturnToPool()

	b = p.Get()
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