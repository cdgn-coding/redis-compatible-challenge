package engine

import "sync"

type MapperFunc = func(v interface{}) (interface{}, error)

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
	c.Set(key, nil)
}
