package peds

import (
	"github.com/cheekybits/genny/generic"
)

type Item generic.Type

// TODO: Right now this is a copy on write implementation based on an underlying slice.
//       This is fairly inefficient, especially for large arrays.
//       The goal is to make it based on tries like the Clojure implementation.
type ItemArray struct {
	tail []Item
	root privateItemNode
	len uint
	shift uint
}


// The "private" prefix is there just for Genny to match on the type name "Item"
// but we don't want to expose this type outside the package.


type privateItemNode interface{}

var emptyItemNode privateItemNode = nil
const privateItemshift = 5
const privateItemNodeSize = 32

func NewItemArray(items ...Item) *ItemArray {
	if len(items) < privateItemNodeSize {
		tail := make([]Item, len(items))
		copy(tail, items)
		return &ItemArray{root: emptyItemNode, tail: tail, len: uint(len(tail))}
	}

	panic("Not implemented yet")
}

func (a *ItemArray) Get(i int) Item {
	return a.tail[i]
}

func (a *ItemArray) tailOffset() uint {
	if a.len < 32 {
		return 0
	}

	return ((a.len - 1) >> privateItemshift) << privateItemshift
}

func (a *ItemArray) Set(i int, item Item) *ItemArray {
	dst := make([]Item, len(a.tail))
	copy(dst, a.tail)
	dst[i] = item
	return &ItemArray{root: a.root, tail: dst}
}

func newItemPath(shift uint, node privateItemNode) privateItemNode {
	if shift == 0 {
		return node
	}

	return newItemPath(shift - 5, privateItemNode([]privateItemNode{node}))
}

func (a *ItemArray) pushTail(level uint, parent privateItemNode, tailNode []Item) privateItemNode {
	subIdx := ((a.len - 1) >> level) & 0x01F
	parentNode := parent.([]privateItemNode)
	ret := make([]privateItemNode, subIdx+1)
	copy(ret, parentNode)
	var nodeToInsert privateItemNode

	if level == privateItemshift {
		nodeToInsert = tailNode
	} else {
		if subIdx < uint(len(parentNode)) {
			nodeToInsert = newItemPath(level-privateItemshift, tailNode)
		} else {
			nodeToInsert = a.pushTail(level-privateItemshift, parentNode[subIdx], tailNode)
		}
	}

	ret[subIdx] = nodeToInsert
	return ret
}

// TODO: Should take variadic arguments
func (a *ItemArray) Append(item Item) *ItemArray {
	if a.len - a.tailOffset() < 32 {
		newTail := make([]Item, len(a.tail) + 1)
		copy(newTail, a.tail)
		newTail[len(a.tail)] = item
		return &ItemArray{root: a.root, tail: newTail, len: a.len + 1, shift: a.shift}
	}

	// Tail full, push into tree
	var newRoot privateItemNode
	newShift := a.shift

	// Root overflow?
	if (a.len >> privateItemshift) > (1 << a.shift) {
		newNode := newItemPath(a.shift, a.tail)
		newRoot = privateItemNode([]privateItemNode{a.root, newNode})
		newShift = a.shift + privateItemshift
	} else {
		newRoot = a.pushTail(a.shift, a.root, a.tail)
	}

	return &ItemArray{root: newRoot, tail: []Item{item}, len: a.len + 1, shift: newShift}
}

func (a *ItemArray) Slice(start, stop int) *ItemArray {
	return &ItemArray{root: a.root, tail: a.tail[start:stop], len: uint(stop - start), shift: a.shift}
}

func (a *ItemArray) Len() int {
	return int(a.len)
}