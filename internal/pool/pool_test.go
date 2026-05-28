package pool

import (
	"sync"
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

func TestPoolPutPreservesSliceCapacity(t *testing.T) {
	p := New(func() *MyStruct {
		return &MyStruct{}
	})

	obj := p.Get()
	obj.Tags = append(obj.Tags, "x", "y", "z")
	capBefore := cap(obj.Tags)

	p.Put(obj)

	assert.Equal(t, 0, len(obj.Tags))
	assert.Equal(t, capBefore, cap(obj.Tags))
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

func TestPoolObjectReuse(t *testing.T) {
	p := New(func() *MyStruct {
		return &MyStruct{}
	})

	obj := p.Get()
	p.Put(obj)

	obj2 := p.Get()
	assert.Same(t, obj, obj2)
}

func TestPoolFactoryCalledForEachEmptyGet(t *testing.T) {
	callCount := 0
	p := New(func() *MyStruct {
		callCount++
		return &MyStruct{}
	})

	p.Get()
	p.Get()
	p.Get()

	assert.Equal(t, 3, callCount)
}

func TestPoolConcurrent(t *testing.T) {
	p := New(func() *MyStruct {
		return &MyStruct{}
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			obj := p.Get()
			obj.Name = "goroutine"
			obj.Count = 1
			obj.Tags = append(obj.Tags, "tag")
			p.Put(obj)
		}()
	}
	wg.Wait()
}
