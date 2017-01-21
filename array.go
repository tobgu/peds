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
}


// The "private" prefix is there just for Genny to match on the type name "Item"
// but we don't want to expose this type outside the package.


type privateItemNode struct {
	children interface{}
	depth int
}


var emptyItemNode privateItemNode = privateItemNode{}
const privateItemshift = 5
const privateItemNodeSize = 32



func NewItemArray(items ...Item) *ItemArray {
	if len(items) < privateItemNodeSize {
		tail := make([]Item, len(items))
		copy(tail, items)
		return &ItemArray{root: emptyItemNode, tail: tail}
	}

	panic("Not implemented yet")
}

func (a *ItemArray) Get(i int) Item {
	return a.tail[i]
}

func (a *ItemArray) Set(i int, item Item) *ItemArray {
	dst := make([]Item, len(a.tail))
	copy(dst, a.tail)
	dst[i] = item
	return &ItemArray{root: a.root, tail: dst}
}

func (a *ItemArray) Append(item Item) *ItemArray {
	dst := make([]Item, len(a.tail), len(a.tail) + 1)
	copy(dst, a.tail)
	return &ItemArray{root: a.root, tail: append(dst, item)}
}

func (a *ItemArray) Slice(start, stop int) *ItemArray {
	return &ItemArray{root: a.root, tail: a.tail[start:stop]}
}

func (a *ItemArray) Len() int {
	return len(a.tail)
}