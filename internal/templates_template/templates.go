package templates_template

//template:CommonTemplate

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

//template:VectorTemplate

//////////////
/// Vector ///
//////////////

type GenericVectorType struct {
	tail  []GenericType
	root  commonNode
	len   uint
	shift uint
}

var emptyGenericTypeTail = make([]GenericType, 0)
var emptyGenericVectorType *GenericVectorType = &GenericVectorType{root: emptyCommonNode, shift: shiftSize, tail: emptyGenericTypeTail}

func NewGenericVectorType(items ...GenericType) *GenericVectorType {
	return emptyGenericVectorType.Append(items...)
}

func (v *GenericVectorType) Get(i int) GenericType {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	return v.sliceFor(uint(i))[i&shiftBitMask]
}

func (v *GenericVectorType) sliceFor(i uint) []GenericType {
	if i >= v.tailOffset() {
		return v.tail
	}

	node := v.root
	for level := v.shift; level > 0; level -= shiftSize {
		node = node.([]commonNode)[(i>>level)&shiftBitMask]
	}

	return node.([]GenericType)
}

func (v *GenericVectorType) tailOffset() uint {
	if v.len < nodeSize {
		return 0
	}

	return ((v.len - 1) >> shiftSize) << shiftSize
}

func (v *GenericVectorType) Set(i int, item GenericType) *GenericVectorType {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	if uint(i) >= v.tailOffset() {
		newTail := make([]GenericType, len(v.tail))
		copy(newTail, v.tail)
		newTail[i&shiftBitMask] = item
		return &GenericVectorType{root: v.root, tail: newTail, len: v.len, shift: v.shift}
	}

	return &GenericVectorType{root: v.doAssoc(v.shift, v.root, uint(i), item), tail: v.tail, len: v.len, shift: v.shift}
}

func (v *GenericVectorType) doAssoc(level uint, node commonNode, i uint, item GenericType) commonNode {
	if level == 0 {
		ret := make([]GenericType, nodeSize)
		copy(ret, node.([]GenericType))
		ret[i&shiftBitMask] = item
		return ret
	}

	ret := make([]commonNode, nodeSize)
	copy(ret, node.([]commonNode))
	subidx := (i >> level) & shiftBitMask
	ret[subidx] = v.doAssoc(level-shiftSize, ret[subidx], i, item)
	return ret
}

func (v *GenericVectorType) pushTail(level uint, parent commonNode, tailNode []GenericType) commonNode {
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

func (v *GenericVectorType) Append(item ...GenericType) *GenericVectorType {
	result := v
	itemLen := uint(len(item))
	for insertOffset := uint(0); insertOffset < itemLen; {
		tailLen := result.len - result.tailOffset()
		tailFree := nodeSize - tailLen
		if tailFree == 0 {
			result = result.pushLeafNode(result.tail)
			result.tail = emptyGenericVectorType.tail
			tailFree = nodeSize
			tailLen = 0
		}

		batchLen := uintMin(itemLen-insertOffset, tailFree)
		newTail := make([]GenericType, 0, tailLen+batchLen)
		newTail = append(newTail, result.tail...)
		newTail = append(newTail, item[insertOffset:insertOffset+batchLen]...)
		result = &GenericVectorType{root: result.root, tail: newTail, len: result.len + batchLen, shift: result.shift}
		insertOffset += batchLen
	}

	return result
}

func (v *GenericVectorType) pushLeafNode(node []GenericType) *GenericVectorType {
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

	return &GenericVectorType{root: newRoot, tail: v.tail, len: v.len, shift: newShift}
}

func (v *GenericVectorType) Slice(start, stop int) *GenericVectorTypeSlice {
	assertSliceOk(start, stop, v.Len())
	return &GenericVectorTypeSlice{vector: v, start: start, stop: stop}
}

func (v *GenericVectorType) Len() int {
	return int(v.len)
}

func (v *GenericVectorType) Iter() *GenericVectorTypeIterator {
	return newGenericVectorTypeIterator(v, 0, v.Len())
}

//////////////////
//// Iterator ////
//////////////////

type GenericVectorTypeIterator struct {
	vector      *GenericVectorType
	currentNode []GenericType
	stop, pos   int
}

func newGenericVectorTypeIterator(vector *GenericVectorType, start, stop int) *GenericVectorTypeIterator {
	it := GenericVectorTypeIterator{vector: vector, pos: start, stop: stop}
	it.currentNode = vector.sliceFor(uint(it.pos))
	return &it
}

func (it *GenericVectorTypeIterator) Next() (value GenericType, ok bool) {
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

//template:SliceTemplate

////////////////
//// Slice /////
////////////////

type GenericVectorTypeSlice struct {
	vector      *GenericVectorType
	start, stop int
}

func NewGenericVectorTypeSlice(items ...GenericType) *GenericVectorTypeSlice {
	return &GenericVectorTypeSlice{vector: emptyGenericVectorType.Append(items...), start: 0, stop: len(items)}
}

func (s *GenericVectorTypeSlice) Len() int {
	return s.stop - s.start
}

func (s *GenericVectorTypeSlice) Get(i int) GenericType {
	if i < 0 || s.start+i >= s.stop {
		panic("Index out of bounds")
	}

	return s.vector.Get(s.start + i)
}

func (s *GenericVectorTypeSlice) Set(i int, item GenericType) *GenericVectorTypeSlice {
	if i < 0 || s.start+i >= s.stop {
		panic("Index out of bounds")
	}

	return s.vector.Set(s.start+i, item).Slice(s.start, s.stop)
}

func (s *GenericVectorTypeSlice) Append(items ...GenericType) *GenericVectorTypeSlice {
	newSlice := GenericVectorTypeSlice{vector: s.vector, start: s.start, stop: s.stop + len(items)}

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

func (s *GenericVectorTypeSlice) Slice(start, stop int) *GenericVectorTypeSlice {
	assertSliceOk(start, stop, s.stop-s.start)
	return &GenericVectorTypeSlice{vector: s.vector, start: s.start + start, stop: s.start + stop}
}

func (s *GenericVectorTypeSlice) Iter() *GenericVectorTypeIterator {
	return newGenericVectorTypeIterator(s.vector, s.start, s.stop)
}

//template:privateMapTemplate

///////////
/// Map ///
///////////

/////////////////////
/// Backing vector ///
/////////////////////

type GenericBucketVector struct {
	tail  []GenericBucket
	root  commonNode
	len   uint
	shift uint
}

type GenericMapItem struct {
	Key   GenericMapKeyType
	Value GenericMapValueType
}

type GenericBucket []GenericMapItem

var emptyGenericBucketTail = make([]GenericBucket, 0)
var emptyGenericBucketVector *GenericBucketVector = &GenericBucketVector{root: emptyCommonNode, shift: shiftSize, tail: emptyGenericBucketTail}

func (v *GenericBucketVector) Get(i int) GenericBucket {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	return v.sliceFor(uint(i))[i&shiftBitMask]
}

func (v *GenericBucketVector) sliceFor(i uint) []GenericBucket {
	if i >= v.tailOffset() {
		return v.tail
	}

	node := v.root
	for level := v.shift; level > 0; level -= shiftSize {
		node = node.([]commonNode)[(i>>level)&shiftBitMask]
	}

	return node.([]GenericBucket)
}

func (v *GenericBucketVector) tailOffset() uint {
	if v.len < nodeSize {
		return 0
	}

	return ((v.len - 1) >> shiftSize) << shiftSize
}

func (v *GenericBucketVector) Set(i int, item GenericBucket) *GenericBucketVector {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	if uint(i) >= v.tailOffset() {
		newTail := make([]GenericBucket, len(v.tail))
		copy(newTail, v.tail)
		newTail[i&shiftBitMask] = item
		return &GenericBucketVector{root: v.root, tail: newTail, len: v.len, shift: v.shift}
	}

	return &GenericBucketVector{root: v.doAssoc(v.shift, v.root, uint(i), item), tail: v.tail, len: v.len, shift: v.shift}
}

func (v *GenericBucketVector) doAssoc(level uint, node commonNode, i uint, item GenericBucket) commonNode {
	if level == 0 {
		ret := make([]GenericBucket, nodeSize)
		copy(ret, node.([]GenericBucket))
		ret[i&shiftBitMask] = item
		return ret
	}

	ret := make([]commonNode, nodeSize)
	copy(ret, node.([]commonNode))
	subidx := (i >> level) & shiftBitMask
	ret[subidx] = v.doAssoc(level-shiftSize, ret[subidx], i, item)
	return ret
}

func (v *GenericBucketVector) pushTail(level uint, parent commonNode, tailNode []GenericBucket) commonNode {
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

func (v *GenericBucketVector) Append(item ...GenericBucket) *GenericBucketVector {
	result := v
	itemLen := uint(len(item))
	for insertOffset := uint(0); insertOffset < itemLen; {
		tailLen := result.len - result.tailOffset()
		tailFree := nodeSize - tailLen
		if tailFree == 0 {
			result = result.pushLeafNode(result.tail)
			result.tail = emptyGenericBucketVector.tail
			tailFree = nodeSize
			tailLen = 0
		}

		batchLen := uintMin(itemLen-insertOffset, tailFree)
		newTail := make([]GenericBucket, 0, tailLen+batchLen)
		newTail = append(newTail, result.tail...)
		newTail = append(newTail, item[insertOffset:insertOffset+batchLen]...)
		result = &GenericBucketVector{root: result.root, tail: newTail, len: result.len + batchLen, shift: result.shift}
		insertOffset += batchLen
	}

	return result
}

func (v *GenericBucketVector) pushLeafNode(node []GenericBucket) *GenericBucketVector {
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

	return &GenericBucketVector{root: newRoot, tail: v.tail, len: v.len, shift: newShift}
}

func (v *GenericBucketVector) Len() int {
	return int(v.len)
}

func (v *GenericBucketVector) Iter() *GenericBucketVectorIterator {
	return newGenericBucketVectorIterator(v, 0, v.Len())
}

//////////////////
//// Iterator ////
//////////////////

type GenericBucketVectorIterator struct {
	vector      *GenericBucketVector
	currentNode []GenericBucket
	stop, pos   int
}

func newGenericBucketVectorIterator(vector *GenericBucketVector, start, stop int) *GenericBucketVectorIterator {
	it := GenericBucketVectorIterator{vector: vector, pos: start, stop: stop}
	it.currentNode = vector.sliceFor(uint(it.pos))
	return &it
}

func (it *GenericBucketVectorIterator) Next() (value GenericBucket, ok bool) {
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

//template:publicMapTemplate

////////////////////////
/// Public functions ///
////////////////////////

type GenericMapType struct {
	backingVector GenericBucketVector
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
// or types in other packages need to be inspected to create v custom hash function for them (only public fields
// will be accessible I guess). Types in the same package also need to be inspected. This is only true for keys.
// Values can be whatever type.

// As v first step only support built in types (and any redeclarations of them). Composite/custom types would
// get their hash through:
//   key := fmt.Sprintf("%#v", value)
// Potentially printing v warning that this is supported but likely not very fast.
