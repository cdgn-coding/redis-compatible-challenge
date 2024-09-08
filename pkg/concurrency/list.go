package concurrency

import (
	"iter"
	"reflect"
	"sync"
)

type Node struct {
	value interface{}
	next  *Node
}

func NewNode(value interface{}) *Node {
	return &Node{value: value}
}

type ConcurrentList struct {
	head    *Node
	tail    *Node
	size    int
	keyLock sync.RWMutex
}

var ConcurrentListType = reflect.TypeOf(&ConcurrentList{})

func NewConcurrentList() *ConcurrentList {
	return &ConcurrentList{}
}

func (cl *ConcurrentList) Len() int {
	return cl.size
}

func (cl *ConcurrentList) PushLeft(value interface{}) {
	cl.keyLock.Lock()
	defer cl.keyLock.Unlock()

	if cl.size == 0 {
		cl.head = NewNode(value)
		cl.tail = cl.head
		cl.size++
		return
	}

	node := NewNode(value)
	node.next = cl.head
	cl.head = node
	cl.size++
}

func (cl *ConcurrentList) PushRight(value interface{}) {
	cl.keyLock.Lock()
	defer cl.keyLock.Unlock()

	if cl.size == 0 {
		cl.head = NewNode(value)
		cl.tail = cl.head
		cl.size++
		return
	}

	node := NewNode(value)
	cl.tail.next = node
	cl.tail = node
	cl.size++
}

func (cl *ConcurrentList) Iterator() iter.Seq[interface{}] {
	return func(yield func(interface{}) bool) {
		cl.keyLock.RLock()
		defer cl.keyLock.RUnlock()
		for node := cl.head; node != nil; node = node.next {
			if !yield(node.value) {
				return
			}
		}
	}
}
