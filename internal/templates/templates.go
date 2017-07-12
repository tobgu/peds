package templates

// NOTE: This file is auto generated, don't edit manually!
const CommonTemplate string = `
// TODO: Need a way to specify imports required by different pieces of the code
import (
	"fmt"
	"encoding/binary"
	"hash/crc32"
	"math"
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

//////////////////////////
//// Hash functions //////
//////////////////////////

func hash(x []byte) uint32 {
	return crc32.ChecksumIEEE(x)
}

func interfaceHash(x interface{}) uint32 {
	return hash([]byte(fmt.Sprintf("%v", x)))
}

func byteHash(x byte) uint32 {
	return hash([]byte{x})
}

func uint8Hash(x uint8) uint32 {
	return byteHash(byte(x))
}

func int8Hash(x int8) uint32 {
	return uint8Hash(uint8(x))
}

func uint16Hash(x uint16) uint32 {
	bX := make([]byte, 2)
	binary.LittleEndian.PutUint16(bX, x)
	return hash(bX)
}

func int16Hash(x int16) uint32 {
	return uint16Hash(uint16(x))
}

func uint32Hash(x uint32) uint32 {
	bX := make([]byte, 4)
	binary.LittleEndian.PutUint32(bX, x)
	return hash(bX)
}

func int32Hash(x int32) uint32 {
	return uint32Hash(uint32(x))
}

func uint64Hash(x uint64) uint32 {
	bX := make([]byte, 8)
	binary.LittleEndian.PutUint64(bX, x)
	return hash(bX)
}

func int64Hash(x int64) uint32 {
	return uint64Hash(uint64(x))
}

func intHash(x int) uint32 {
	return int64Hash(int64(x))
}

func uintHash(x uint) uint32 {
	return uint64Hash(uint64(x))
}

func boolHash(x bool) uint32 {
	if x {
		return 1
	}

	return 0
}

func runeHash(x rune) uint32 {
	return int32Hash(int32(x))
}

func stringHash(x string) uint32 {
	return hash([]byte(x))
}

func float64Hash(x float64) uint32 {
	return uint64Hash(math.Float64bits(x))
}

func float32Hash(x float32) uint32 {
	return uint32Hash(math.Float32bits(x))
}

`
const PrivateMapTemplate string = `
///////////
/// Map ///
///////////

/////////////////////
/// Backing vector ///
/////////////////////

type {{.MapItemTypeName}}BucketVector struct {
	tail  []{{.MapItemTypeName}}Bucket
	root  commonNode
	len   uint
	shift uint
}

type {{.MapItemTypeName}} struct {
	Key   {{.MapKeyTypeName}}
	Value {{.MapValueTypeName}}
}

type {{.MapItemTypeName}}Bucket []{{.MapItemTypeName}}

var empty{{.MapItemTypeName}}BucketVectorTail = make([]{{.MapItemTypeName}}Bucket, 0)
var empty{{.MapItemTypeName}}BucketVector *{{.MapItemTypeName}}BucketVector = &{{.MapItemTypeName}}BucketVector{root: emptyCommonNode, shift: shiftSize, tail: empty{{.MapItemTypeName}}BucketVectorTail}

func (v *{{.MapItemTypeName}}BucketVector) Get(i int) {{.MapItemTypeName}}Bucket {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	return v.sliceFor(uint(i))[i&shiftBitMask]
}

func (v *{{.MapItemTypeName}}BucketVector) sliceFor(i uint) []{{.MapItemTypeName}}Bucket {
	if i >= v.tailOffset() {
		return v.tail
	}

	node := v.root
	for level := v.shift; level > 0; level -= shiftSize {
		node = node.([]commonNode)[(i>>level)&shiftBitMask]
	}

	return node.([]{{.MapItemTypeName}}Bucket)
}

func (v *{{.MapItemTypeName}}BucketVector) tailOffset() uint {
	if v.len < nodeSize {
		return 0
	}

	return ((v.len - 1) >> shiftSize) << shiftSize
}

func (v *{{.MapItemTypeName}}BucketVector) Set(i int, item {{.MapItemTypeName}}Bucket) *{{.MapItemTypeName}}BucketVector {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	if uint(i) >= v.tailOffset() {
		newTail := make([]{{.MapItemTypeName}}Bucket, len(v.tail))
		copy(newTail, v.tail)
		newTail[i&shiftBitMask] = item
		return &{{.MapItemTypeName}}BucketVector{root: v.root, tail: newTail, len: v.len, shift: v.shift}
	}

	return &{{.MapItemTypeName}}BucketVector{root: v.doAssoc(v.shift, v.root, uint(i), item), tail: v.tail, len: v.len, shift: v.shift}
}

func (v *{{.MapItemTypeName}}BucketVector) doAssoc(level uint, node commonNode, i uint, item {{.MapItemTypeName}}Bucket) commonNode {
	if level == 0 {
		ret := make([]{{.MapItemTypeName}}Bucket, nodeSize)
		copy(ret, node.([]{{.MapItemTypeName}}Bucket))
		ret[i&shiftBitMask] = item
		return ret
	}

	ret := make([]commonNode, nodeSize)
	copy(ret, node.([]commonNode))
	subidx := (i >> level) & shiftBitMask
	ret[subidx] = v.doAssoc(level-shiftSize, ret[subidx], i, item)
	return ret
}

func (v *{{.MapItemTypeName}}BucketVector) pushTail(level uint, parent commonNode, tailNode []{{.MapItemTypeName}}Bucket) commonNode {
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

func (v *{{.MapItemTypeName}}BucketVector) Append(item ...{{.MapItemTypeName}}Bucket) *{{.MapItemTypeName}}BucketVector {
	result := v
	itemLen := uint(len(item))
	for insertOffset := uint(0); insertOffset < itemLen; {
		tailLen := result.len - result.tailOffset()
		tailFree := nodeSize - tailLen
		if tailFree == 0 {
			result = result.pushLeafNode(result.tail)
			result.tail = empty{{.MapItemTypeName}}BucketVector.tail
			tailFree = nodeSize
			tailLen = 0
		}

		batchLen := uintMin(itemLen-insertOffset, tailFree)
		newTail := make([]{{.MapItemTypeName}}Bucket, 0, tailLen+batchLen)
		newTail = append(newTail, result.tail...)
		newTail = append(newTail, item[insertOffset:insertOffset+batchLen]...)
		result = &{{.MapItemTypeName}}BucketVector{root: result.root, tail: newTail, len: result.len + batchLen, shift: result.shift}
		insertOffset += batchLen
	}

	return result
}

func (v *{{.MapItemTypeName}}BucketVector) pushLeafNode(node []{{.MapItemTypeName}}Bucket) *{{.MapItemTypeName}}BucketVector {
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

	return &{{.MapItemTypeName}}BucketVector{root: newRoot, tail: v.tail, len: v.len, shift: newShift}
}

func (v *{{.MapItemTypeName}}BucketVector) Len() int {
	return int(v.len)
}

func (v *{{.MapItemTypeName}}BucketVector) Iter() *{{.MapItemTypeName}}BucketVectorIterator {
	return new{{.MapItemTypeName}}BucketVectorIterator(v, 0, v.Len())
}

//////////////////
//// Iterator ////
//////////////////

type {{.MapItemTypeName}}BucketVectorIterator struct {
	vector      *{{.MapItemTypeName}}BucketVector
	currentNode []{{.MapItemTypeName}}Bucket
	stop, pos   int
}

func new{{.MapItemTypeName}}BucketVectorIterator(vector *{{.MapItemTypeName}}BucketVector, start, stop int) *{{.MapItemTypeName}}BucketVectorIterator {
	it := {{.MapItemTypeName}}BucketVectorIterator{vector: vector, pos: start, stop: stop}
	it.currentNode = vector.sliceFor(uint(it.pos))
	return &it
}

func (it *{{.MapItemTypeName}}BucketVectorIterator) Next() (value {{.MapItemTypeName}}Bucket, ok bool) {
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

type {{.MapTypeName}} struct {
	backingVector *{{.MapItemTypeName}}BucketVector
	len           int
}

/////////////////////////
/// Private functions ///
/////////////////////////

func (m *{{.MapTypeName}}) pos(key {{.MapKeyTypeName}}) int {
	return int(uint64({{.MapKeyHashFunc}}(key)) % uint64(m.backingVector.Len()))
}

func (m *{{.MapTypeName}}) load(key {{.MapKeyTypeName}}) (value {{.MapValueTypeName}}, ok bool) {
	bucket := m.backingVector.Get(m.pos(key))
	if bucket != nil {
		for _, item := range bucket {
			if item.Key == key {
				return item.Value, true
			}
		}
	}

	var zeroValue {{.MapValueTypeName}}
	return zeroValue, false
}

func (m *{{.MapTypeName}}) store(key {{.MapKeyTypeName}}, value {{.MapValueTypeName}}) *{{.MapTypeName}} {
	pos := m.pos(key)
	bucket := m.backingVector.Get(pos)
	if bucket != nil {
		for ix, item := range bucket {
			if item.Key == key {
				// Overwrite existing item
				newBucket := make({{.MapItemTypeName}}Bucket, len(bucket))
				copy(newBucket, bucket)
				newBucket[ix] = {{.MapItemTypeName}}{Key: key, Value: value}
				return &{{.MapTypeName}}{backingVector: m.backingVector.Set(pos, newBucket), len: m.len}
			}
		}

		// Add new item to bucket
		newBucket := make({{.MapItemTypeName}}Bucket, len(bucket), len(bucket)+1)
		copy(newBucket, bucket)
		newBucket = append(newBucket, {{.MapItemTypeName}}{Key: key, Value: value})
		return &{{.MapTypeName}}{backingVector: m.backingVector.Set(pos, newBucket), len: m.len + 1}
	}

	item := {{.MapItemTypeName}}{Key: key, Value: value}
	newBucket := {{.MapItemTypeName}}Bucket{item}
	return &{{.MapTypeName}}{backingVector: m.backingVector.Set(pos, newBucket), len: m.len + 1}
}

func (m *{{.MapTypeName}}) delete(key {{.MapKeyTypeName}}) *{{.MapTypeName}} {
	pos := m.pos(key)
	bucket := m.backingVector.Get(pos)
	if bucket != nil {
		newBucket := make({{.MapItemTypeName}}Bucket, 0)
		for _, item := range bucket {
			if item.Key != key {
				newBucket = append(newBucket, item)
			}
		}

		removedItemCount := len(bucket) - len(newBucket)
		if removedItemCount == 0 {
			return m
		}

		if len(newBucket) == 0 {
			newBucket = nil
		}

		return &{{.MapTypeName}}{backingVector: m.backingVector.Set(pos, newBucket), len: m.len - removedItemCount}
	}

	return m
}

func (m *{{.MapTypeName}}) prange(f func(key {{.MapKeyTypeName}}, value {{.MapValueTypeName}}) bool) {
	it := m.backingVector.Iter()
	for bucket, ok := it.Next(); ok; bucket, ok = it.Next() {
		for _, item := range bucket {
			if !f(item.Key, item.Value) {
				return
			}
		}
	}
}

`
const PublicMapTemplate string = `
////////////////////////
/// Public functions ///
////////////////////////

func New{{.MapTypeName}}(items ...{{.MapItemTypeName}}) *{{.MapTypeName}} {
	// TODO: Vary size depending on input size
	vectorSize := nodeSize
	input := make([]{{.MapItemTypeName}}Bucket, vectorSize)
	length := 0
	for _, item := range items {
		ix := int(uint64({{.MapKeyHashFunc}}(item.Key)) % uint64(vectorSize))
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
				input[ix] = append(bucket, {{.MapItemTypeName}}{Key: item.Key, Value: item.Value})
				length++
			}
		} else {
			input[ix] = {{.MapItemTypeName}}Bucket{item}
			length++
		}
	}

	return &{{.MapTypeName}}{backingVector: empty{{.MapItemTypeName}}BucketVector.Append(input...), len: length}
}

func (m *{{.MapTypeName}}) Len() int {
	return int(m.len)
}

func (m *{{.MapTypeName}}) Load(key {{.MapKeyTypeName}}) (value {{.MapValueTypeName}}, ok bool) {
	return m.load(key)
}

func (m *{{.MapTypeName}}) Store(key {{.MapKeyTypeName}}, value {{.MapValueTypeName}}) *{{.MapTypeName}} {
	return m.store(key, value)
}

func (m *{{.MapTypeName}}) Delete(key {{.MapKeyTypeName}}) *{{.MapTypeName}} {
	return m.delete(key)
}

func (m *{{.MapTypeName}}) Range(f func(key {{.MapKeyTypeName}}, value {{.MapValueTypeName}}) bool) {
	m.prange(f)
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
const otherStuff string = `
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
`
