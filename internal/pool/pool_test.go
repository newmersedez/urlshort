package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MyStruct struct {
	Name  string
	Count int
	Tags  []string
}

func (m *MyStruct) Reset() {
	m.Name = ""
	m.Count = 0
	m.Tags = m.Tags[:0]
}

func TestPoolGetReturnsNewObject(t *testing.T) {
	p := New(func() *MyStruct {
		return &MyStruct{}
	})

	obj := p.Get()
	assert.NotNil(t, obj)
}

func TestPoolPutResetsObject(t *testing.T) {
	p := New(func() *MyStruct {
		return &MyStruct{}
	})

	obj := p.Get()
	obj.Name = "hello"
	obj.Count = 42
	obj.Tags = append(obj.Tags, "a", "b")

	p.Put(obj)

	assert.Equal(t, "", obj.Name)
	assert.Equal(t, 0, obj.Count)
	assert.Equal(t, 0, len(obj.Tags))
}

func TestPoolGetAfterPut(t *testing.T) {
	p := New(func() *MyStruct {
		return &MyStruct{}
	})

	obj := p.Get()
	obj.Name = "test"
	p.Put(obj)

	obj2 := p.Get()
	assert.Equal(t, "", obj2.Name)
}
