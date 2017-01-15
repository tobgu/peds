package peds

import (
	"github.com/cheekybits/genny/generic"
)

type Item generic.Type

// TODO: Right now this is a copy on write implementation based on an underlying slice.
//       This is fairly inefficient, especially for large arrays.
//       The goal is to make it based on tries like the Clojure implementation.
type ItemArray struct {
	array []Item
	/*
	tail [privateItemNodeSize]Item
	root privateItemNode
	*/
}


// The "private" prefix is there just for Genny to match on the type name "Item"
// but we don't want to expose this type outside the package.

/*
type privateItemNode struct {
	children interface{}
	depth int
}
*/

/*
const privateItemshift = 5
const privateItemNodeSize = 32
*/


func NewItemArray(items ...Item) *ItemArray {
	dst := make([]Item, len(items))
	copy(dst, items)
	return &ItemArray{array: dst}
}

func (a *ItemArray) Get(i int) Item {
	return a.array[i]
}

func (a *ItemArray) Set(i int, item Item) *ItemArray {
	dst := make([]Item, len(a.array))
	copy(dst, a.array)
	dst[i] = item
	return &ItemArray{array: dst}
}

func (a *ItemArray) Append(item Item) *ItemArray {
	dst := make([]Item, len(a.array), len(a.array) + 1)
	copy(dst, a.array)
	return &ItemArray{array: append(dst, item)}
}

func (a *ItemArray) Slice(start, stop int) *ItemArray {
	return &ItemArray{array: a.array[start:stop]}
}

func (a *ItemArray) Len() int {
	return len(a.array)
}