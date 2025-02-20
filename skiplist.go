package skiplist

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

const (
	// DefaultMaxLevel default height/level of the list
	//
	// Suitable for math.Floor(math.Pow(math.E, 18)) == 65659969 elements in list
	DefaultMaxLevel int = 18
	// DefaultProbability default probability of a node could hoist to the higher level
	DefaultProbability float64 = 1 / math.E
)

// Len returns the number of elements in the list.
func (list *SkipList[T]) Len() int {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return list.length
}

// Front returns the head node of the list.
func (list *SkipList[T]) Front() *Element[T] {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return list.next[0]
}

// Set inserts a value in the list with the specified key, ordered by the key.
// If the key exists, it updates the value in the existing node.
// Returns a pointer to the new element.
// Locking is optimistic and happens only after searching.
func (list *SkipList[T]) Set(key T, value interface{}) *Element[T] {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	var element *Element[T]
	prevs := list.getPrevElementNodes(key)

	// if key == second element, than update and return the second element
	if element = prevs[0].next[0]; element != nil && element.key <= key {
		if element.key < key {
			fmt.Println(element.key)
		}

		element.value = value
		return element
	}

	// make new node and generate the random level,
	// this node will appears from level 0 to the random level.
	element = &Element[T]{
		elementNode: elementNode[T]{
			next: make([]*Element[T], list.randLevel()),
		},
		key:   key,
		value: value,
	}

	// insert new node into skiplist.
	for i := range element.next {
		element.next[i] = prevs[i].next[i]
		prevs[i].next[i] = element
	}

	list.length++
	return element
}

// Get finds an element by key. It returns element pointer if found, nil if not found.
// Locking is optimistic and happens only after searching with a fast check for deletion after locking.
func (list *SkipList[T]) Get(key T) *Element[T] {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	var prev *elementNode[T] = &list.elementNode
	var next *Element[T]

	for i := list.maxLevel - 1; i >= 0; i-- {
		next = prev.next[i]

		for next != nil && key > next.key {
			prev = &next.elementNode
			next = next.next[i]
		}
	}

	if next != nil && next.key == key {
		return next
	}

	return nil
}

// Remove deletes an element from the list.
// Returns removed element pointer if found, nil if not found.
// Locking is optimistic and happens only after searching with a fast check on adjacent nodes after locking.
func (list *SkipList[T]) Remove(key T) *Element[T] {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	prevs := list.getPrevElementNodes(key)

	// found the element, remove it
	if element := prevs[0].next[0]; element != nil && element.key <= key {
		for k, v := range element.next {
			prevs[k].next[k] = v
		}

		list.length--
		return element
	}

	return nil
}

// getPrevElementNodes is the private search mechanism that other functions use.
// Finds the previous nodes on each level relative to the current Element and
// caches them. This approach is similar to a "search finger" as described by Pugh:
// http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.17.524
//
// 从上到下，从左到右搜索跳表索引，返回每一层的命中结点。
// 下标就是层数，[0] 是最底层，[maxLevel - 1] 是最顶层。
func (list *SkipList[T]) getPrevElementNodes(key T) []*elementNode[T] {
	var prev *elementNode[T] = &list.elementNode
	var next *Element[T]

	prevs := list.prevNodesCache

	// 从最上层开始找
	for i := list.maxLevel - 1; i >= 0; i-- {
		// next 是当前层的下一个节点
		next = prev.next[i]

		// 水平遍历，直到到达尾部，或者 key > next，
		// 说明当前层就选择当前的节点（prev）
		for next != nil && key > next.key {
			prev = &next.elementNode
			next = next.next[i]
		}

		// 将每一层所选择的节点存入 prevs 中
		prevs[i] = prev
	}

	return prevs
}

// SetProbability changes the current P value of the list.
// It doesn't alter any existing data, only changes how future insert heights are calculated.
func (list *SkipList[T]) SetProbability(newProbability float64) {
	list.probability = newProbability
	list.probTable = probabilityTable(list.probability, list.maxLevel)
}

// randLevel generate the height/level of a node
//
// generate an random value, and use the probability table to find it's highest level.
func (list *SkipList[T]) randLevel() (level int) {
	// Our random number source only has Int63(), so we have to produce a float64 from it
	// Reference: https://golang.org/src/math/rand/rand.go#L150
	r := float64(list.randSource.Int63()) / (1 << 63)

	level = 1
	for level < list.maxLevel && r < list.probTable[level] {
		level++
	}
	return
}

// probabilityTable calculates in advance the probability of a new node having a given level.
// probability is in [0, 1], MaxLevel is (0, 64]
// Returns a table of floating point probabilities that each level should be included during an insert.
func probabilityTable(probability float64, MaxLevel int) (table []float64) {
	for i := 1; i <= MaxLevel; i++ {
		prob := math.Pow(probability, float64(i-1))
		table = append(table, prob)
	}
	return table
}

// NewWithMaxLevel creates a new skip list with MaxLevel set to the provided number.
// maxLevel has to be int(math.Ceil(math.Log(N))) for DefaultProbability (where N is an upper bound on the
// number of elements in a skip list). See http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.17.524
// Returns a pointer to the new list.
func NewWithMaxLevel[T Sortable](maxLevel int) *SkipList[T] {
	if maxLevel < 1 || maxLevel > 64 {
		panic("maxLevel for a SkipList must be a positive integer <= 64")
	}

	return &SkipList[T]{
		elementNode:    elementNode[T]{next: make([]*Element[T], maxLevel)},
		prevNodesCache: make([]*elementNode[T], maxLevel),
		maxLevel:       maxLevel,
		randSource:     rand.New(rand.NewSource(time.Now().UnixNano())),
		probability:    DefaultProbability,
		probTable:      probabilityTable(DefaultProbability, maxLevel),
	}
}

// New creates a new skip list with default parameters. Returns a pointer to the new list.
func New[T Sortable]() *SkipList[T] {
	return NewWithMaxLevel[T](DefaultMaxLevel)
}
