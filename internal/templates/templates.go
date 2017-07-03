package templates

// NOTE: This file is auto generated, don't edit manually!
const VectorTemplate string = `
//////////////
/// Vector ///
//////////////

type {{.VectorTypeName}} struct {
	tail  []{{.TypeName}}
	root  commonNode
	len   uint
	shift uint
}

var empty{{.TypeName}}Tail = make([]{{.TypeName}}, 0)
var empty{{.VectorTypeName}} *{{.VectorTypeName}} = &{{.VectorTypeName}}{root: emptyCommonNode, shift: shiftSize, tail: empty{{.TypeName}}Tail}

func New{{.VectorTypeName}}(items ...{{.TypeName}}) *{{.VectorTypeName}} {
	return empty{{.VectorTypeName}}.Append(items...)
}

func (a *{{.VectorTypeName}}) Get(i int) {{.TypeName}} {
	if i < 0 || uint(i) >= a.len {
		panic("Index out of bounds")
	}

	return a.arrayFor(uint(i))[i&shiftBitMask]
}

func (a *{{.VectorTypeName}}) arrayFor(i uint) []{{.TypeName}} {
	if i >= a.tailOffset() {
		return a.tail
	}

	node := a.root
	for level := a.shift; level > 0; level -= shiftSize {
		node = node.([]commonNode)[(i>>level)&shiftBitMask]
	}

	return node.([]{{.TypeName}})
}

func (a *{{.VectorTypeName}}) tailOffset() uint {
	if a.len < nodeSize {
		return 0
	}

	return ((a.len - 1) >> shiftSize) << shiftSize
}

func (a *{{.VectorTypeName}}) Set(i int, item {{.TypeName}}) *{{.VectorTypeName}} {
	if i < 0 || uint(i) >= a.len {
		panic("Index out of bounds")
	}

	if uint(i) >= a.tailOffset() {
		newTail := make([]{{.TypeName}}, len(a.tail))
		copy(newTail, a.tail)
		newTail[i&shiftBitMask] = item
		return &{{.VectorTypeName}}{root: a.root, tail: newTail, len: a.len, shift: a.shift}
	}

	return &{{.VectorTypeName}}{root: a.doAssoc(a.shift, a.root, uint(i), item), tail: a.tail, len: a.len, shift: a.shift}
}

func (a *{{.VectorTypeName}}) doAssoc(level uint, node commonNode, i uint, item {{.TypeName}}) commonNode {
	if level == 0 {
		ret := make([]{{.TypeName}}, nodeSize)
		copy(ret, node.([]{{.TypeName}}))
		ret[i&shiftBitMask] = item
		return ret
	}

	ret := make([]commonNode, nodeSize)
	copy(ret, node.([]commonNode))
	subidx := (i >> level) & shiftBitMask
	ret[subidx] = a.doAssoc(level-shiftSize, ret[subidx], i, item)
	return ret
}

func (a *{{.VectorTypeName}}) pushTail(level uint, parent commonNode, tailNode []{{.TypeName}}) commonNode {
	subIdx := ((a.len - 1) >> level) & shiftBitMask
	parentNode := parent.([]commonNode)
	ret := make([]commonNode, subIdx+1)
	copy(ret, parentNode)
	var nodeToInsert commonNode

	if level == shiftSize {
		nodeToInsert = tailNode
	} else if subIdx < uint(len(parentNode)) {
		nodeToInsert = a.pushTail(level-shiftSize, parentNode[subIdx], tailNode)
	} else {
		nodeToInsert = newPath(level-shiftSize, tailNode)
	}

	ret[subIdx] = nodeToInsert
	return ret
}

func (a *{{.VectorTypeName}}) Append(item ...{{.TypeName}}) *{{.VectorTypeName}} {
	result := a
	itemLen := uint(len(item))
	for insertOffset := uint(0); insertOffset < itemLen; {
		tailLen := result.len - result.tailOffset()
		tailFree := nodeSize - tailLen
		if tailFree == 0 {
			result = result.pushLeafNode(result.tail)
			result.tail = empty{{.VectorTypeName}}.tail
			tailFree = nodeSize
			tailLen = 0
		}

		batchLen := uintMin(itemLen-insertOffset, tailFree)
		newTail := make([]{{.TypeName}}, 0, tailLen+batchLen)
		newTail = append(newTail, result.tail...)
		newTail = append(newTail, item[insertOffset:insertOffset+batchLen]...)
		result = &{{.VectorTypeName}}{root: result.root, tail: newTail, len: result.len + batchLen, shift: result.shift}
		insertOffset += batchLen
	}

	return result
}

func (a *{{.VectorTypeName}}) pushLeafNode(node []{{.TypeName}}) *{{.VectorTypeName}} {
	var newRoot commonNode
	newShift := a.shift

	// Root overflow?
	if (a.len >> shiftSize) > (1 << a.shift) {
		newNode := newPath(a.shift, node)
		newRoot = commonNode([]commonNode{a.root, newNode})
		newShift = a.shift + shiftSize
	} else {
		newRoot = a.pushTail(a.shift, a.root, node)
	}

	return &{{.VectorTypeName}}{root: newRoot, tail: a.tail, len: a.len, shift: newShift}
}

func (a *{{.VectorTypeName}}) Slice(start, stop int) *{{.VectorTypeName}}Slice {
	assertSliceOk(start, stop, a.Len())
	return &{{.VectorTypeName}}Slice{array: a, start: start, stop: stop}
}

func (a *{{.VectorTypeName}}) Len() int {
	return int(a.len)
}

func (a *{{.VectorTypeName}}) Iter() *{{.VectorTypeName}}Iterator {
	return new{{.VectorTypeName}}Iterator(a, 0, a.Len())
}

//////////////////
//// Iterator ////
//////////////////

type {{.VectorTypeName}}Iterator struct {
	array       *{{.VectorTypeName}}
	currentNode []{{.TypeName}}
	stop, pos   int
}

func new{{.VectorTypeName}}Iterator(array *{{.VectorTypeName}}, start, stop int) *{{.VectorTypeName}}Iterator {
	it := {{.VectorTypeName}}Iterator{array: array, pos: start, stop: stop}
	it.currentNode = array.arrayFor(uint(it.pos))
	return &it
}

func (it *{{.VectorTypeName}}Iterator) Next() (value {{.TypeName}}, ok bool) {
	if it.pos >= it.stop {
		return value, false
	}

	if it.pos&shiftBitMask == 0 {
		it.currentNode = it.array.arrayFor(uint(it.pos))
	}

	value = it.currentNode[it.pos&shiftBitMask]
	it.pos++
	return value, true
}

`
const CommonTemplate string = `
import (
	"fmt"
)

const shiftSize = 5
const nodeSize = 32
const shiftBitMask = 0x1F

type commonNode interface{}

var emptyCommonNode commonNode = []commonNode{}

func uintMin(a, b uint) uint {
	if a < b {
		return a
	}

	return b
}

func newPath(shift uint, node commonNode) commonNode {
	if shift == 0 {
		return node
	}

	return newPath(shift-shiftSize, commonNode([]commonNode{node}))
}

func assertSliceOk(start, stop, len int) {
	if start < 0 {
		panic(fmt.Sprintf("Invalid slice index %d (index must be non-negative)", start))
	}

	if start > stop {
		panic(fmt.Sprintf("Invalid slice index: %d > %d", start, stop))
	}

	if stop > len {
		panic(fmt.Sprintf("Slice bounds out of range, start=%d, stop=%d, len=%d", start, stop, len))
	}
}

`
const MapTemplate string = `
///////////
/// Map ///
///////////

////////////////////////
/// Public functions ///
////////////////////////

type GenericMapType struct {
	backingVector {{.VectorTypeName}}
	len           int
}

func (m *GenericMapType) Len() int {
	return int(m.len)
}

func (m *GenericMapType) Load(key GenericMapKeyType) (value GenericMapValueType, ok bool) {
	var temp GenericMapValueType
	return temp, false
}

func (m *GenericMapType) Store(key GenericMapKeyType, value GenericMapValueType) *GenericMapType {
	return &GenericMapType{}
}

func (m *GenericMapType) Delete(key GenericMapKeyType) *GenericMapType {
	return &GenericMapType{}
}

func (m *GenericMapType) Range(f func(key GenericMapKeyType, value GenericMapValueType) bool) {
}

// Check during generation that key is comparable, this can be done using reflection
// Generate hashing code during generation based on key type (should be possible to get this info from AST or similar)

// peds -maps "FooMap<int, string>;BarMap<int16, int32>"
//      -sets "FooSet<mypackage.MyType>"
//      -vectors "FooVec<io.Bar>"
//      -imports "io;github.com/my/mypackage"
//      -package mycontainers
//      -file mycontainers_gen.go

// Built in types can more or less be used as is, custom hash function needed depending on type. Third party types
// or types in other packages need to be inspected to create a custom hash function for them (only public fields
// will be accessible I guess). Types in the same package also need to be inspected. This is only true for keys.
// Values can be whatever type.

// As a first step only support built in types (and any redeclarations of them). Composite/custom types would
// get their hash through:
//   key := fmt.Sprintf("%#v", value)
// Potentially printing a warning that this is supported but likely not very fast.
`
const SliceTemplate string = `
////////////////
//// Slice /////
////////////////

type {{.VectorTypeName}}Slice struct {
	array       *{{.VectorTypeName}}
	start, stop int
}

func New{{.VectorTypeName}}Slice(items ...{{.TypeName}}) *{{.VectorTypeName}}Slice {
	return &{{.VectorTypeName}}Slice{array: empty{{.VectorTypeName}}.Append(items...), start: 0, stop: len(items)}
}

func (s *{{.VectorTypeName}}Slice) Len() int {
	return s.stop - s.start
}

func (s *{{.VectorTypeName}}Slice) Get(i int) {{.TypeName}} {
	if i < 0 || s.start+i >= s.stop {
		panic("Index out of bounds")
	}

	return s.array.Get(s.start + i)
}

func (s *{{.VectorTypeName}}Slice) Set(i int, item {{.TypeName}}) *{{.VectorTypeName}}Slice {
	if i < 0 || s.start+i >= s.stop {
		panic("Index out of bounds")
	}

	return s.array.Set(s.start+i, item).Slice(s.start, s.stop)
}

func (s *{{.VectorTypeName}}Slice) Append(items ...{{.TypeName}}) *{{.VectorTypeName}}Slice {
	newSlice := {{.VectorTypeName}}Slice{array: s.array, start: s.start, stop: s.stop + len(items)}

	// If this is a slice that has an upper bound that is lower than the backing
	// array then set the values in the backing array to achieve some structural
	// sharing.
	itemPos := 0
	for ; s.stop+itemPos < s.array.Len() && itemPos < len(items); itemPos++ {
		newSlice.array = newSlice.array.Set(s.stop+itemPos, items[itemPos])
	}

	// For the rest just append it to the underlying array
	newSlice.array = newSlice.array.Append(items[itemPos:]...)
	return &newSlice
}

func (s *{{.VectorTypeName}}Slice) Slice(start, stop int) *{{.VectorTypeName}}Slice {
	assertSliceOk(start, stop, s.stop-s.start)
	return &{{.VectorTypeName}}Slice{array: s.array, start: s.start + start, stop: s.start + stop}
}

func (s *{{.VectorTypeName}}Slice) Iter() *{{.VectorTypeName}}Iterator {
	return new{{.VectorTypeName}}Iterator(s.array, s.start, s.stop)
}

`
