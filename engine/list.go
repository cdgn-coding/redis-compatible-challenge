package engine

import "iter"

type Node struct {
	value interface{}
	next  *Node
}

func NewNode(value interface{}) *Node {
	return &Node{value: value}
}

type ConcurrentList struct {
	head *Node
	tail *Node
	size int
}

func NewConcurrentList() *ConcurrentList {
	return &ConcurrentList{}
}

func (cl *ConcurrentList) Len() int {
	return cl.size
}

func (cl *ConcurrentList) PushLeft(value interface{}) {
	if cl.size == 0 {
		cl.head = NewNode(value)
		cl.tail = cl.head
		cl.size++
		return
	}

	node := NewNode(value)
	node.next = cl.head
	cl.head = node
}

func (cl *ConcurrentList) PushRight(value interface{}) {
	if cl.size == 0 {
		cl.head = NewNode(value)
		cl.tail = cl.head
		cl.size++
		return
	}

	node := NewNode(value)
	cl.tail.next = node
	cl.tail = node
}

func (cl *ConcurrentList) Iterator() iter.Seq[interface{}] {
	return func(yield func(interface{}) bool) {
		for node := cl.head; node != nil; node = node.next {
			if !yield(node.value) {
				return
			}
		}

	}
}
