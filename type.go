package skiplist

import (
	"math/rand"
	"sync"
)

type elementNode struct {
	// key forward pointers of one node
	//
	// length means this node's height.
	// if node has forward pointers in level 0~3,
	// then length is 4.
	next []*Element
}

type Element struct {
	elementNode
	key   float64
	value interface{}
}

// Key allows retrieval of the key for a given Element
func (e *Element) Key() float64 {
	return e.key
}

// Value allows retrieval of the value for a given Element
func (e *Element) Value() interface{} {
	return e.value
}

// Next returns the following Element or nil if we're at the end of the list.
// Only operates on the bottom level of the skip list (a fully linked list).
func (element *Element) Next() *Element {
	return element.next[0]
}

type SkipList struct {
	// elementNode forward pointers
	elementNode
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
	prevNodesCache []*elementNode
}
