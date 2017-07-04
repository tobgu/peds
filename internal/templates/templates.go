package templates

// NOTE: This file is auto generated, don't edit manually!
const publicMapTemplate string = `
////////////////////////
/// Public functions ///
////////////////////////

type GenericMapType struct {
	backingVector {{.MapBucketTypeName}}Vector
	len           int
}

func (m *GenericMapType) Len() int {
	return int(m.len)
}

func (m *GenericMapType) Load(key {{.MapKeyTypeName}}) (value {{.MapValueTypeName}}, ok bool) {
	var temp {{.MapValueTypeName}}
	return temp, false
}

func (m *GenericMapType) Store(key {{.MapKeyTypeName}}, value {{.MapValueTypeName}}) *GenericMapType {
	return &GenericMapType{}
}

func (m *GenericMapType) Delete(key {{.MapKeyTypeName}}) *GenericMapType {
	return &GenericMapType{}
}

func (m *GenericMapType) Range(f func(key {{.MapKeyTypeName}}, value {{.MapValueTypeName}}) bool) {
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
// or types in other packages need to be inspected to create v custom hash function for them (only public fields
// will be accessible I guess). Types in the same package also need to be inspected. This is only true for keys.
// Values can be whatever type.

// As v first step only support built in types (and any redeclarations of them). Composite/custom types would
// get their hash through:
//   key := fmt.Sprintf("%#v", value)
// Potentially printing v warning that this is supported but likely not very fast.
`
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

func (v *{{.VectorTypeName}}) Get(i int) {{.TypeName}} {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	return v.sliceFor(uint(i))[i&shiftBitMask]
}

func (v *{{.VectorTypeName}}) sliceFor(i uint) []{{.TypeName}} {
	if i >= v.tailOffset() {
		return v.tail
	}

	node := v.root
	for level := v.shift; level > 0; level -= shiftSize {
		node = node.([]commonNode)[(i>>level)&shiftBitMask]
	}

	return node.([]{{.TypeName}})
}

func (v *{{.VectorTypeName}}) tailOffset() uint {
	if v.len < nodeSize {
		return 0
	}

	return ((v.len - 1) >> shiftSize) << shiftSize
}

func (v *{{.VectorTypeName}}) Set(i int, item {{.TypeName}}) *{{.VectorTypeName}} {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	if uint(i) >= v.tailOffset() {
		newTail := make([]{{.TypeName}}, len(v.tail))
		copy(newTail, v.tail)
		newTail[i&shiftBitMask] = item
		return &{{.VectorTypeName}}{root: v.root, tail: newTail, len: v.len, shift: v.shift}
	}

	return &{{.VectorTypeName}}{root: v.doAssoc(v.shift, v.root, uint(i), item), tail: v.tail, len: v.len, shift: v.shift}
}

func (v *{{.VectorTypeName}}) doAssoc(level uint, node commonNode, i uint, item {{.TypeName}}) commonNode {
	if level == 0 {
		ret := make([]{{.TypeName}}, nodeSize)
		copy(ret, node.([]{{.TypeName}}))
		ret[i&shiftBitMask] = item
		return ret
	}

	ret := make([]commonNode, nodeSize)
	copy(ret, node.([]commonNode))
	subidx := (i >> level) & shiftBitMask
	ret[subidx] = v.doAssoc(level-shiftSize, ret[subidx], i, item)
	return ret
}

func (v *{{.VectorTypeName}}) pushTail(level uint, parent commonNode, tailNode []{{.TypeName}}) commonNode {
	subIdx := ((v.len - 1) >> level) & shiftBitMask
	parentNode := parent.([]commonNode)
	ret := make([]commonNode, subIdx+1)
	copy(ret, parentNode)
	var nodeToInsert commonNode

	if level == shiftSize {
		nodeToInsert = tailNode
	} else if subIdx < uint(len(parentNode)) {
		nodeToInsert = v.pushTail(level-shiftSize, parentNode[subIdx], tailNode)
	} else {
		nodeToInsert = newPath(level-shiftSize, tailNode)
	}

	ret[subIdx] = nodeToInsert
	return ret
}

func (v *{{.VectorTypeName}}) Append(item ...{{.TypeName}}) *{{.VectorTypeName}} {
	result := v
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

func (v *{{.VectorTypeName}}) pushLeafNode(node []{{.TypeName}}) *{{.VectorTypeName}} {
	var newRoot commonNode
	newShift := v.shift

	// Root overflow?
	if (v.len >> shiftSize) > (1 << v.shift) {
		newNode := newPath(v.shift, node)
		newRoot = commonNode([]commonNode{v.root, newNode})
		newShift = v.shift + shiftSize
	} else {
		newRoot = v.pushTail(v.shift, v.root, node)
	}

	return &{{.VectorTypeName}}{root: newRoot, tail: v.tail, len: v.len, shift: newShift}
}

func (v *{{.VectorTypeName}}) Slice(start, stop int) *{{.VectorTypeName}}Slice {
	assertSliceOk(start, stop, v.Len())
	return &{{.VectorTypeName}}Slice{vector: v, start: start, stop: stop}
}

func (v *{{.VectorTypeName}}) Len() int {
	return int(v.len)
}

func (v *{{.VectorTypeName}}) Iter() *{{.VectorTypeName}}Iterator {
	return new{{.VectorTypeName}}Iterator(v, 0, v.Len())
}

//////////////////
//// Iterator ////
//////////////////

type {{.VectorTypeName}}Iterator struct {
	vector      *{{.VectorTypeName}}
	currentNode []{{.TypeName}}
	stop, pos   int
}

func new{{.VectorTypeName}}Iterator(vector *{{.VectorTypeName}}, start, stop int) *{{.VectorTypeName}}Iterator {
	it := {{.VectorTypeName}}Iterator{vector: vector, pos: start, stop: stop}
	it.currentNode = vector.sliceFor(uint(it.pos))
	return &it
}

func (it *{{.VectorTypeName}}Iterator) Next() (value {{.TypeName}}, ok bool) {
	if it.pos >= it.stop {
		return value, false
	}

	if it.pos&shiftBitMask == 0 {
		it.currentNode = it.vector.sliceFor(uint(it.pos))
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
const privateMapTemplate string = `
///////////
/// Map ///
///////////

/////////////////////
/// Backing vector ///
/////////////////////

type {{.MapBucketTypeName}}Vector struct {
	tail  []{{.MapBucketTypeName}}
	root  commonNode
	len   uint
	shift uint
}

type {{.MapItemTypeName}} struct {
	Key   {{.MapKeyTypeName}}
	Value {{.MapValueTypeName}}
}

type {{.MapBucketTypeName}} []{{.MapItemTypeName}}

var empty{{.MapBucketTypeName}}Tail = make([]{{.MapBucketTypeName}}, 0)
var empty{{.MapBucketTypeName}}Vector *{{.MapBucketTypeName}}Vector = &{{.MapBucketTypeName}}Vector{root: emptyCommonNode, shift: shiftSize, tail: empty{{.MapBucketTypeName}}Tail}

func (v *{{.MapBucketTypeName}}Vector) Get(i int) {{.MapBucketTypeName}} {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	return v.sliceFor(uint(i))[i&shiftBitMask]
}

func (v *{{.MapBucketTypeName}}Vector) sliceFor(i uint) []{{.MapBucketTypeName}} {
	if i >= v.tailOffset() {
		return v.tail
	}

	node := v.root
	for level := v.shift; level > 0; level -= shiftSize {
		node = node.([]commonNode)[(i>>level)&shiftBitMask]
	}

	return node.([]{{.MapBucketTypeName}})
}

func (v *{{.MapBucketTypeName}}Vector) tailOffset() uint {
	if v.len < nodeSize {
		return 0
	}

	return ((v.len - 1) >> shiftSize) << shiftSize
}

func (v *{{.MapBucketTypeName}}Vector) Set(i int, item {{.MapBucketTypeName}}) *{{.MapBucketTypeName}}Vector {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	if uint(i) >= v.tailOffset() {
		newTail := make([]{{.MapBucketTypeName}}, len(v.tail))
		copy(newTail, v.tail)
		newTail[i&shiftBitMask] = item
		return &{{.MapBucketTypeName}}Vector{root: v.root, tail: newTail, len: v.len, shift: v.shift}
	}

	return &{{.MapBucketTypeName}}Vector{root: v.doAssoc(v.shift, v.root, uint(i), item), tail: v.tail, len: v.len, shift: v.shift}
}

func (v *{{.MapBucketTypeName}}Vector) doAssoc(level uint, node commonNode, i uint, item {{.MapBucketTypeName}}) commonNode {
	if level == 0 {
		ret := make([]{{.MapBucketTypeName}}, nodeSize)
		copy(ret, node.([]{{.MapBucketTypeName}}))
		ret[i&shiftBitMask] = item
		return ret
	}

	ret := make([]commonNode, nodeSize)
	copy(ret, node.([]commonNode))
	subidx := (i >> level) & shiftBitMask
	ret[subidx] = v.doAssoc(level-shiftSize, ret[subidx], i, item)
	return ret
}

func (v *{{.MapBucketTypeName}}Vector) pushTail(level uint, parent commonNode, tailNode []{{.MapBucketTypeName}}) commonNode {
	subIdx := ((v.len - 1) >> level) & shiftBitMask
	parentNode := parent.([]commonNode)
	ret := make([]commonNode, subIdx+1)
	copy(ret, parentNode)
	var nodeToInsert commonNode

	if level == shiftSize {
		nodeToInsert = tailNode
	} else if subIdx < uint(len(parentNode)) {
		nodeToInsert = v.pushTail(level-shiftSize, parentNode[subIdx], tailNode)
	} else {
		nodeToInsert = newPath(level-shiftSize, tailNode)
	}

	ret[subIdx] = nodeToInsert
	return ret
}

func (v *{{.MapBucketTypeName}}Vector) Append(item ...{{.MapBucketTypeName}}) *{{.MapBucketTypeName}}Vector {
	result := v
	itemLen := uint(len(item))
	for insertOffset := uint(0); insertOffset < itemLen; {
		tailLen := result.len - result.tailOffset()
		tailFree := nodeSize - tailLen
		if tailFree == 0 {
			result = result.pushLeafNode(result.tail)
			result.tail = empty{{.MapBucketTypeName}}Vector.tail
			tailFree = nodeSize
			tailLen = 0
		}

		batchLen := uintMin(itemLen-insertOffset, tailFree)
		newTail := make([]{{.MapBucketTypeName}}, 0, tailLen+batchLen)
		newTail = append(newTail, result.tail...)
		newTail = append(newTail, item[insertOffset:insertOffset+batchLen]...)
		result = &{{.MapBucketTypeName}}Vector{root: result.root, tail: newTail, len: result.len + batchLen, shift: result.shift}
		insertOffset += batchLen
	}

	return result
}

func (v *{{.MapBucketTypeName}}Vector) pushLeafNode(node []{{.MapBucketTypeName}}) *{{.MapBucketTypeName}}Vector {
	var newRoot commonNode
	newShift := v.shift

	// Root overflow?
	if (v.len >> shiftSize) > (1 << v.shift) {
		newNode := newPath(v.shift, node)
		newRoot = commonNode([]commonNode{v.root, newNode})
		newShift = v.shift + shiftSize
	} else {
		newRoot = v.pushTail(v.shift, v.root, node)
	}

	return &{{.MapBucketTypeName}}Vector{root: newRoot, tail: v.tail, len: v.len, shift: newShift}
}

func (v *{{.MapBucketTypeName}}Vector) Len() int {
	return int(v.len)
}

func (v *{{.MapBucketTypeName}}Vector) Iter() *{{.MapBucketTypeName}}VectorIterator {
	return new{{.MapBucketTypeName}}VectorIterator(v, 0, v.Len())
}

//////////////////
//// Iterator ////
//////////////////

type {{.MapBucketTypeName}}VectorIterator struct {
	vector      *{{.MapBucketTypeName}}Vector
	currentNode []{{.MapBucketTypeName}}
	stop, pos   int
}

func new{{.MapBucketTypeName}}VectorIterator(vector *{{.MapBucketTypeName}}Vector, start, stop int) *{{.MapBucketTypeName}}VectorIterator {
	it := {{.MapBucketTypeName}}VectorIterator{vector: vector, pos: start, stop: stop}
	it.currentNode = vector.sliceFor(uint(it.pos))
	return &it
}

func (it *{{.MapBucketTypeName}}VectorIterator) Next() (value {{.MapBucketTypeName}}, ok bool) {
	if it.pos >= it.stop {
		return value, false
	}

	if it.pos&shiftBitMask == 0 {
		it.currentNode = it.vector.sliceFor(uint(it.pos))
	}

	value = it.currentNode[it.pos&shiftBitMask]
	it.pos++
	return value, true
}

`
const SliceTemplate string = `
////////////////
//// Slice /////
////////////////

type {{.VectorTypeName}}Slice struct {
	vector      *{{.VectorTypeName}}
	start, stop int
}

func New{{.VectorTypeName}}Slice(items ...{{.TypeName}}) *{{.VectorTypeName}}Slice {
	return &{{.VectorTypeName}}Slice{vector: empty{{.VectorTypeName}}.Append(items...), start: 0, stop: len(items)}
}

func (s *{{.VectorTypeName}}Slice) Len() int {
	return s.stop - s.start
}

func (s *{{.VectorTypeName}}Slice) Get(i int) {{.TypeName}} {
	if i < 0 || s.start+i >= s.stop {
		panic("Index out of bounds")
	}

	return s.vector.Get(s.start + i)
}

func (s *{{.VectorTypeName}}Slice) Set(i int, item {{.TypeName}}) *{{.VectorTypeName}}Slice {
	if i < 0 || s.start+i >= s.stop {
		panic("Index out of bounds")
	}

	return s.vector.Set(s.start+i, item).Slice(s.start, s.stop)
}

func (s *{{.VectorTypeName}}Slice) Append(items ...{{.TypeName}}) *{{.VectorTypeName}}Slice {
	newSlice := {{.VectorTypeName}}Slice{vector: s.vector, start: s.start, stop: s.stop + len(items)}

	// If this is v slice that has an upper bound that is lower than the backing
	// vector then set the values in the backing vector to achieve some structural
	// sharing.
	itemPos := 0
	for ; s.stop+itemPos < s.vector.Len() && itemPos < len(items); itemPos++ {
		newSlice.vector = newSlice.vector.Set(s.stop+itemPos, items[itemPos])
	}

	// For the rest just append it to the underlying vector
	newSlice.vector = newSlice.vector.Append(items[itemPos:]...)
	return &newSlice
}

func (s *{{.VectorTypeName}}Slice) Slice(start, stop int) *{{.VectorTypeName}}Slice {
	assertSliceOk(start, stop, s.stop-s.start)
	return &{{.VectorTypeName}}Slice{vector: s.vector, start: s.start + start, stop: s.start + stop}
}

func (s *{{.VectorTypeName}}Slice) Iter() *{{.VectorTypeName}}Iterator {
	return new{{.VectorTypeName}}Iterator(s.vector, s.start, s.stop)
}

`
