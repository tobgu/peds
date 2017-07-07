package templates_template

//template:CommonTemplate

// TODO: Need a way to specify imports required by different pieces of the code
import (
	"fmt"
	"hash/fnv"
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

func hash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func genericHashFunc(x interface{}) uint64 {
	return hash(fmt.Sprintf("%v", x))
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

//template:PrivateMapTemplate

///////////
/// Map ///
///////////

/////////////////////
/// Backing vector ///
/////////////////////

type GenericMapItemBucketVector struct {
	tail  []GenericMapItemBucket
	root  commonNode
	len   uint
	shift uint
}

type GenericMapItem struct {
	Key   GenericMapKeyType
	Value GenericMapValueType
}

type GenericMapItemBucket []GenericMapItem

var emptyGenericMapItemBucketVectorTail = make([]GenericMapItemBucket, 0)
var emptyGenericMapItemBucketVector *GenericMapItemBucketVector = &GenericMapItemBucketVector{root: emptyCommonNode, shift: shiftSize, tail: emptyGenericMapItemBucketVectorTail}

func (v *GenericMapItemBucketVector) Get(i int) GenericMapItemBucket {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	return v.sliceFor(uint(i))[i&shiftBitMask]
}

func (v *GenericMapItemBucketVector) sliceFor(i uint) []GenericMapItemBucket {
	if i >= v.tailOffset() {
		return v.tail
	}

	node := v.root
	for level := v.shift; level > 0; level -= shiftSize {
		node = node.([]commonNode)[(i>>level)&shiftBitMask]
	}

	return node.([]GenericMapItemBucket)
}

func (v *GenericMapItemBucketVector) tailOffset() uint {
	if v.len < nodeSize {
		return 0
	}

	return ((v.len - 1) >> shiftSize) << shiftSize
}

func (v *GenericMapItemBucketVector) Set(i int, item GenericMapItemBucket) *GenericMapItemBucketVector {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	if uint(i) >= v.tailOffset() {
		newTail := make([]GenericMapItemBucket, len(v.tail))
		copy(newTail, v.tail)
		newTail[i&shiftBitMask] = item
		return &GenericMapItemBucketVector{root: v.root, tail: newTail, len: v.len, shift: v.shift}
	}

	return &GenericMapItemBucketVector{root: v.doAssoc(v.shift, v.root, uint(i), item), tail: v.tail, len: v.len, shift: v.shift}
}

func (v *GenericMapItemBucketVector) doAssoc(level uint, node commonNode, i uint, item GenericMapItemBucket) commonNode {
	if level == 0 {
		ret := make([]GenericMapItemBucket, nodeSize)
		copy(ret, node.([]GenericMapItemBucket))
		ret[i&shiftBitMask] = item
		return ret
	}

	ret := make([]commonNode, nodeSize)
	copy(ret, node.([]commonNode))
	subidx := (i >> level) & shiftBitMask
	ret[subidx] = v.doAssoc(level-shiftSize, ret[subidx], i, item)
	return ret
}

func (v *GenericMapItemBucketVector) pushTail(level uint, parent commonNode, tailNode []GenericMapItemBucket) commonNode {
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

func (v *GenericMapItemBucketVector) Append(item ...GenericMapItemBucket) *GenericMapItemBucketVector {
	result := v
	itemLen := uint(len(item))
	for insertOffset := uint(0); insertOffset < itemLen; {
		tailLen := result.len - result.tailOffset()
		tailFree := nodeSize - tailLen
		if tailFree == 0 {
			result = result.pushLeafNode(result.tail)
			result.tail = emptyGenericMapItemBucketVector.tail
			tailFree = nodeSize
			tailLen = 0
		}

		batchLen := uintMin(itemLen-insertOffset, tailFree)
		newTail := make([]GenericMapItemBucket, 0, tailLen+batchLen)
		newTail = append(newTail, result.tail...)
		newTail = append(newTail, item[insertOffset:insertOffset+batchLen]...)
		result = &GenericMapItemBucketVector{root: result.root, tail: newTail, len: result.len + batchLen, shift: result.shift}
		insertOffset += batchLen
	}

	return result
}

func (v *GenericMapItemBucketVector) pushLeafNode(node []GenericMapItemBucket) *GenericMapItemBucketVector {
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

	return &GenericMapItemBucketVector{root: newRoot, tail: v.tail, len: v.len, shift: newShift}
}

func (v *GenericMapItemBucketVector) Len() int {
	return int(v.len)
}

func (v *GenericMapItemBucketVector) Iter() *GenericMapItemBucketVectorIterator {
	return newGenericMapItemBucketVectorIterator(v, 0, v.Len())
}

//////////////////
//// Iterator ////
//////////////////

type GenericMapItemBucketVectorIterator struct {
	vector      *GenericMapItemBucketVector
	currentNode []GenericMapItemBucket
	stop, pos   int
}

func newGenericMapItemBucketVectorIterator(vector *GenericMapItemBucketVector, start, stop int) *GenericMapItemBucketVectorIterator {
	it := GenericMapItemBucketVectorIterator{vector: vector, pos: start, stop: stop}
	it.currentNode = vector.sliceFor(uint(it.pos))
	return &it
}

func (it *GenericMapItemBucketVectorIterator) Next() (value GenericMapItemBucket, ok bool) {
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

type GenericMapType struct {
	backingVector *GenericMapItemBucketVector
	len           int
}

/////////////////////////
/// Private functions ///
/////////////////////////

func (m *GenericMapType) pos(key GenericMapKeyType) int {
	return int(genericHashFunc(key) % uint64(m.backingVector.Len()))
}

func (m *GenericMapType) load(key GenericMapKeyType) (value GenericMapValueType, ok bool) {
	bucket := m.backingVector.Get(m.pos(key))
	if bucket != nil {
		for _, item := range bucket {
			if item.Key == key {
				return item.Value, true
			}
		}
	}

	var zeroValue GenericMapValueType
	return zeroValue, false
}

func (m *GenericMapType) store(key GenericMapKeyType, value GenericMapValueType) *GenericMapType {
	pos := m.pos(key)
	bucket := m.backingVector.Get(pos)
	if bucket != nil {
		for ix, item := range bucket {
			if item.Key == key {
				// Overwrite existing item
				newBucket := make(GenericMapItemBucket, len(bucket))
				copy(newBucket, bucket)
				newBucket[ix] = GenericMapItem{Key: key, Value: value}
				return &GenericMapType{backingVector: m.backingVector.Set(pos, newBucket), len: m.len}
			}
		}

		// Add new item to bucket
		newBucket := make(GenericMapItemBucket, len(bucket), len(bucket)+1)
		copy(newBucket, bucket)
		newBucket = append(newBucket, GenericMapItem{Key: key, Value: value})
		return &GenericMapType{backingVector: m.backingVector.Set(pos, newBucket), len: m.len + 1}
	}

	item := GenericMapItem{Key: key, Value: value}
	newBucket := GenericMapItemBucket{item}
	return &GenericMapType{backingVector: m.backingVector.Set(pos, newBucket), len: m.len + 1}
}

//template:PublicMapTemplate

////////////////////////
/// Public functions ///
////////////////////////

func NewGenericMapType(items ...GenericMapItem) *GenericMapType {
	// TODO: Vary size depending on input size
	vectorSize := nodeSize
	input := make([]GenericMapItemBucket, vectorSize)
	length := 0
	for _, item := range items {
		ix := int(genericHashFunc(item) % uint64(vectorSize))
		bucket := input[ix]
		if bucket != nil {
			// Hash collision, merge with existing bucket
			found := false
			for keyIx, bItem := range bucket {
				if item.Key == bItem.Key {
					bucket[keyIx] = item
					found = true
					break
				}
			}

			if !found {
				input[ix] = append(bucket, GenericMapItem{Key: item.Key, Value: item.Value})
				length++
			}
		} else {
			input[ix] = GenericMapItemBucket{item}
			length++
		}
	}

	return &GenericMapType{backingVector: emptyGenericMapItemBucketVector.Append(input...), len: length}
}

func (m *GenericMapType) Len() int {
	return int(m.len)
}

func (m *GenericMapType) Load(key GenericMapKeyType) (value GenericMapValueType, ok bool) {
	return m.load(key)
}

func (m *GenericMapType) Store(key GenericMapKeyType, value GenericMapValueType) *GenericMapType {
	return m.store(key, value)
}

func (m *GenericMapType) Delete(key GenericMapKeyType) *GenericMapType {
	return &GenericMapType{}
}

func (m *GenericMapType) Range(f func(key GenericMapKeyType, value GenericMapValueType) bool) {
}

//template:otherStuff

//////////////////////
/// Hash functions ///
//////////////////////

/*
Types for which special hash functions could be implemented:

uint8       the set of all unsigned  8-bit integers (0 to 255)
uint16      the set of all unsigned 16-bit integers (0 to 65535)
uint32      the set of all unsigned 32-bit integers (0 to 4294967295)
uint64      the set of all unsigned 64-bit integers (0 to 18446744073709551615)

int8        the set of all signed  8-bit integers (-128 to 127)
int16       the set of all signed 16-bit integers (-32768 to 32767)
int32       the set of all signed 32-bit integers (-2147483648 to 2147483647)
int64       the set of all signed 64-bit integers (-9223372036854775808 to 9223372036854775807)

float32     the set of all IEEE-754 32-bit floating-point numbers
float64     the set of all IEEE-754 64-bit floating-point numbers

complex64   the set of all complex numbers with float32 real and imaginary parts
complex128  the set of all complex numbers with float64 real and imaginary parts

byte        alias for uint8
rune        alias for int32

bool

string

Everything else will use the generic hash based on string representation for now.

General idea:

If the generic hash based on string representation currently used proves to be a bottleneck
implement custom hashes for the above.

Use fnv hash function to start with. If still bottle neck test something like murmur3.

Use binary.LittleEndian.PutUint32 and friends to convert integers to bytes.

Use math.Float64/32bits to convert float to uints.

Equivalent types:
We may want to denote equivalent types (eg. redefinitions of any of the above types) to make
it possible to use the faster hash functions for those as well. Perhaps there is some kind
of introspection possible that would make this possible without user input?

*/

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
