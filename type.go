package skiplist

import (
	"math/rand"
	"sync"
)

// Number is a number type
type Number interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float32 | float64
}

// Sortable Data types that can be compared by >, <, ==
type Sortable interface {
	Number | string
}

type elementNode[T Sortable] struct {
	// key forward pointers of one node
	//
	// length means this node's height.
	// if node has forward pointers in level 0~3,
	// then length is 4.
	next []*Element[T]
}

type Element[T Sortable] struct {
	elementNode[T]
	key   T
	value interface{}
}

// Key allows retrieval of the key for a given Element
func (e *Element[T]) Key() T {
	return e.key
}

// Value allows retrieval of the value for a given Element
func (e *Element[T]) Value() interface{} {
	return e.value
}

// Next returns the following Element or nil if we're at the end of the list.
// Only operates on the bottom level of the skip list (a fully linked list).
func (e *Element[T]) Next() *Element[T] {
	return e.next[0]
}

type SkipList[T Sortable] struct {
	// elementNode forward pointers
	elementNode[T]
	// maxLevel 最大高度
	maxLevel   int
	Length     int
	randSource rand.Source
	// probability 节点上升的概率
	probability float64
	// probTable 预先算好上升到每一层的概率
	probTable []float64
	mutex     sync.RWMutex
	// prevNodesCache 在跳表中查询某一个 key 的时候，
	// 需要从上到下，从左到右，逐层遍历。
	// prevNodesCache 的长度等同于高的，存储的是查找过程中，每一层的节点。
	//
	// 从最后一个元素开始，依次向前。
	prevNodesCache []*elementNode[T]
}
