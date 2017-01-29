package peds

import (
	"github.com/cheekybits/genny/generic"
)

type Item generic.Type

// TODO: Right now this is a copy on write implementation based on an underlying slice.
//       This is fairly inefficient, especially for large arrays.
//       The goal is to make it based on tries like the Clojure implementation.
type ItemArray struct {
	tail  []Item
	root  privateItemNode
	len   uint
	shift uint
}

// The "private" prefix is there just for Genny to match on the type name "Item"
// but we don't want to expose this type outside the package.

type privateItemNode interface{}

var emptyItemNode privateItemNode = []privateItemNode{}
var emptyItemTail = make([]Item, 0)
var emptyItemArray *ItemArray = &ItemArray{root: emptyItemNode, shift: privateItemshift, tail: emptyItemTail}

/*
const privateItemshift = 5
const privateItemNodeSize = 32
const privateItemBitMask = 0x1F
*/

const privateItemshift = 5
const privateItemNodeSize = 32
const privateItemBitMask = 0x1F

func NewItemArray(items ...Item) *ItemArray {
	return emptyItemArray.Append(items...)
}

func (a *ItemArray) Get(i int) Item {
	if i < 0 || uint(i) > a.len {
		panic("Index out of bounds")
	}

	return a.arrayFor(uint(i))[i&privateItemBitMask]
}

func (a *ItemArray) arrayFor(i uint) []Item {
	if i >= a.tailOffset() {
		return a.tail
	}

	node := a.root
	for level := a.shift; level > 0; level -= privateItemshift {
		node = node.([]privateItemNode)[(i>>level)&privateItemBitMask]
	}

	return node.([]Item)
}

func (a *ItemArray) tailOffset() uint {
	if a.len < privateItemNodeSize {
		return 0
	}

	return ((a.len - 1) >> privateItemshift) << privateItemshift
}

func (a *ItemArray) Set(i int, item Item) *ItemArray {
	if i < 0 || uint(i) >= a.len {
		panic("Index out of bounds")
	}

	if uint(i) >= a.tailOffset() {
		newTail := make([]Item, len(a.tail))
		copy(newTail, a.tail)
		newTail[i&privateItemBitMask] = item
		return &ItemArray{root: a.root, tail: newTail, len: a.len, shift: a.shift}
	}

	return &ItemArray{root: a.doAssoc(a.shift, a.root, uint(i), item), tail: a.tail, len: a.len, shift: a.shift}
}

func (a *ItemArray) doAssoc(level uint, node privateItemNode, i uint, item Item) privateItemNode {
	if level == 0 {
		ret := make([]Item, privateItemNodeSize)
		copy(ret, node.([]Item))
		ret[i&privateItemBitMask] = item
		return ret
	}

	ret := make([]privateItemNode, privateItemNodeSize)
	copy(ret, node.([]privateItemNode))
	subidx := (i >> level) & privateItemBitMask
	ret[subidx] = a.doAssoc(level-privateItemshift, ret[subidx], i, item)
	return ret
}

func newItemPath(shift uint, node privateItemNode) privateItemNode {
	if shift == 0 {
		return node
	}

	return newItemPath(shift-privateItemshift, privateItemNode([]privateItemNode{node}))
}

func (a *ItemArray) pushTail(level uint, parent privateItemNode, tailNode []Item) privateItemNode {
	subIdx := ((a.len - 1) >> level) & privateItemBitMask
	parentNode := parent.([]privateItemNode)
	ret := make([]privateItemNode, subIdx+1)
	copy(ret, parentNode)
	var nodeToInsert privateItemNode

	if level == privateItemshift {
		nodeToInsert = tailNode
	} else if subIdx < uint(len(parentNode)) {
		nodeToInsert = a.pushTail(level-privateItemshift, parentNode[subIdx], tailNode)
	} else {
		nodeToInsert = newItemPath(level-privateItemshift, tailNode)
	}

	ret[subIdx] = nodeToInsert
	return ret
}

func uintItemMin(a, b uint) uint {
	if a < b {
		return a
	}

	return b
}

func (a *ItemArray) Append(item... Item) *ItemArray {
	result := a
	itemLen := uint(len(item))
	for insertOffset := uint(0); insertOffset < itemLen; {
		tailLen := result.len-result.tailOffset()
		tailFree := privateItemNodeSize - tailLen
		if tailFree == 0 {
			result = result.pushLeafNode(result.tail)
			result.tail = emptyItemArray.tail
			tailFree = privateItemNodeSize
			tailLen = 0
		}

		batchLen := uintItemMin(itemLen - insertOffset, tailFree)
		newTail := make([]Item, 0, tailLen+batchLen)
		newTail = append(newTail, result.tail...)
		newTail = append(newTail, item[insertOffset:insertOffset+batchLen]...)
		result = &ItemArray{root: result.root, tail: newTail, len: result.len + batchLen, shift: result.shift}
		insertOffset += batchLen
	}

	return result
}

func (a *ItemArray) pushLeafNode(node []Item) *ItemArray {
	var newRoot privateItemNode
	newShift := a.shift

	// Root overflow?
	if (a.len >> privateItemshift) > (1 << a.shift) {
		newNode := newItemPath(a.shift, node)
		newRoot = privateItemNode([]privateItemNode{a.root, newNode})
		newShift = a.shift + privateItemshift
	} else {
		newRoot = a.pushTail(a.shift, a.root, node)
	}

	return &ItemArray{root: newRoot, tail: a.tail, len: a.len, shift: newShift}
}

func (a *ItemArray) Slice(start, stop int) *ItemArray {
	return &ItemArray{root: a.root, tail: a.tail[start:stop], len: uint(stop - start), shift: a.shift}
}

func (a *ItemArray) Len() int {
	return int(a.len)
}
