package concurrency

import (
	"iter"
	"sync"
)

type MapperFunc = func(v interface{}) (interface{}, error)

type Constructor func() interface{}

type Entry struct {
	value interface{}
	lock  sync.RWMutex
}

func NewEntry(value interface{}) *Entry {
	return &Entry{
		value: value,
		lock:  sync.RWMutex{},
	}
}

func (e *Entry) Read() interface{} {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.value
}

func (e *Entry) Write(value interface{}) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.value = value
}

// Map assumes "mapper" returns the new value to set
func (e *Entry) Map(mapper MapperFunc) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	val, err := mapper(e.value)
	if err != nil {
		return err
	}
	e.value = val
	return nil
}

// Mutate assumes "mutator" takes care of race conditions for writing
func (e *Entry) Mutate(mutator MapperFunc, constructor Constructor) (interface{}, error) {
	e.lock.RLock()
	defer e.lock.RLock()

	if e.value == nil {
		e.value = constructor()
		return mutator(e.value)
	}

	val, err := mutator(e.value)
	if err != nil {
		return nil, err
	}

	return val, nil
}

type ConcurrentMap struct {
	memory  map[string]*Entry
	keyLock sync.Mutex
}

func NewConcurrentMap() *ConcurrentMap {
	return &ConcurrentMap{
		memory:  make(map[string]*Entry),
		keyLock: sync.Mutex{},
	}
}

func (c *ConcurrentMap) Set(key string, value interface{}) {
	c.keyLock.Lock()
	entry, ok := c.memory[key]

	if !ok {
		c.memory[key] = NewEntry(value)
		c.keyLock.Unlock()
		return
	}

	c.keyLock.Unlock()
	entry.Write(value)
}

func (c *ConcurrentMap) Map(key string, mapper MapperFunc) error {
	c.keyLock.Lock()
	entry, ok := c.memory[key]

	if !ok {
		defaultValue, _ := mapper(nil)
		c.memory[key] = NewEntry(defaultValue)
		c.keyLock.Unlock()
		return nil
	}

	c.keyLock.Unlock()
	return entry.Map(mapper)
}

func (c *ConcurrentMap) Mutate(key string, mutator MapperFunc, constructor Constructor) (interface{}, error) {
	c.keyLock.Lock()
	defer c.keyLock.Unlock()
	entry, ok := c.memory[key]

	if !ok {
		entry = NewEntry(nil)
		c.memory[key] = entry
	}

	return entry.Mutate(mutator, constructor)
}

func (c *ConcurrentMap) Get(key string) (interface{}, bool) {
	entry, ok := c.memory[key]
	if !ok {
		return nil, false
	}

	return entry.Read(), true
}

func (c *ConcurrentMap) Has(key string) bool {
	entry, ok := c.memory[key]
	return ok && entry.Read() != nil
}

func (c *ConcurrentMap) Delete(key string) {
	c.keyLock.Lock()
	delete(c.memory, key)
	c.keyLock.Unlock()
}

type Pair struct {
	Key   string
	Value interface{}
}

func NewPair(key string, value interface{}) Pair {
	return Pair{
		Key:   key,
		Value: value,
	}
}

func (c *ConcurrentMap) Iterable() iter.Seq[Pair] {
	return func(yield func(Pair) bool) {
		c.keyLock.Lock()
		defer c.keyLock.Unlock()
		for k, v := range c.memory {
			if !yield(NewPair(k, v.Read())) {
				return
			}
		}
	}
}
