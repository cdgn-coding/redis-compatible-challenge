package engine

import "sync"

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

type ConcurrentMap struct {
	memory map[string]*Entry
}

func NewConcurrentMap() ConcurrentMap {
	return ConcurrentMap{
		memory: make(map[string]*Entry),
	}
}

func (c *ConcurrentMap) Set(key string, value interface{}) {
	entry, ok := c.memory[key]
	if !ok {
		c.memory[key] = NewEntry(value)
		return
	}

	entry.Write(value)
}

func (c *ConcurrentMap) Get(key string) (interface{}, bool) {
	entry, ok := c.memory[key]
	if !ok {
		return nil, false
	}

	return entry.Read(), true
}

func (c *ConcurrentMap) Delete(key string) {
	c.Set(key, nil)
}
