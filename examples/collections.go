package examples

import (
	"encoding/binary"
	"fmt"
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

const upperMapLoadFactor float64 = 8.0
const lowerMapLoadFactor float64 = 2.0
const initialMapLoadFactor float64 = (upperMapLoadFactor + lowerMapLoadFactor) / 2

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

//////////////
/// Vector ///
//////////////

// A IntVector is an ordered persistent/immutable collection of items corresponding roughly
// to the use cases for a slice.
type IntVector struct {
	tail  []int
	root  commonNode
	len   uint
	shift uint
}

var emptyIntVectorTail = make([]int, 0)
var emptyIntVector *IntVector = &IntVector{root: emptyCommonNode, shift: shiftSize, tail: emptyIntVectorTail}

// NewIntVector returns a new IntVector containing the items provided in items.
func NewIntVector(items ...int) *IntVector {
	return emptyIntVector.Append(items...)
}

// Get returns the element at position i.
func (v *IntVector) Get(i int) int {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	return v.sliceFor(uint(i))[i&shiftBitMask]
}

func (v *IntVector) sliceFor(i uint) []int {
	if i >= v.tailOffset() {
		return v.tail
	}

	node := v.root
	for level := v.shift; level > 0; level -= shiftSize {
		node = node.([]commonNode)[(i>>level)&shiftBitMask]
	}

	return node.([]int)
}

func (v *IntVector) tailOffset() uint {
	if v.len < nodeSize {
		return 0
	}

	return ((v.len - 1) >> shiftSize) << shiftSize
}

// Set returns a new vector with the element at position i set to item.
func (v *IntVector) Set(i int, item int) *IntVector {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	if uint(i) >= v.tailOffset() {
		newTail := make([]int, len(v.tail))
		copy(newTail, v.tail)
		newTail[i&shiftBitMask] = item
		return &IntVector{root: v.root, tail: newTail, len: v.len, shift: v.shift}
	}

	return &IntVector{root: v.doAssoc(v.shift, v.root, uint(i), item), tail: v.tail, len: v.len, shift: v.shift}
}

func (v *IntVector) doAssoc(level uint, node commonNode, i uint, item int) commonNode {
	if level == 0 {
		ret := make([]int, nodeSize)
		copy(ret, node.([]int))
		ret[i&shiftBitMask] = item
		return ret
	}

	ret := make([]commonNode, nodeSize)
	copy(ret, node.([]commonNode))
	subidx := (i >> level) & shiftBitMask
	ret[subidx] = v.doAssoc(level-shiftSize, ret[subidx], i, item)
	return ret
}

func (v *IntVector) pushTail(level uint, parent commonNode, tailNode []int) commonNode {
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

// Append returns a new vector with item(s) appended to it.
func (v *IntVector) Append(item ...int) *IntVector {
	result := v
	itemLen := uint(len(item))
	for insertOffset := uint(0); insertOffset < itemLen; {
		tailLen := result.len - result.tailOffset()
		tailFree := nodeSize - tailLen
		if tailFree == 0 {
			result = result.pushLeafNode(result.tail)
			result.tail = emptyIntVector.tail
			tailFree = nodeSize
			tailLen = 0
		}

		batchLen := uintMin(itemLen-insertOffset, tailFree)
		newTail := make([]int, 0, tailLen+batchLen)
		newTail = append(newTail, result.tail...)
		newTail = append(newTail, item[insertOffset:insertOffset+batchLen]...)
		result = &IntVector{root: result.root, tail: newTail, len: result.len + batchLen, shift: result.shift}
		insertOffset += batchLen
	}

	return result
}

func (v *IntVector) pushLeafNode(node []int) *IntVector {
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

	return &IntVector{root: newRoot, tail: v.tail, len: v.len, shift: newShift}
}

// Slice returns a IntVectorSlice that refers to all elements [start,stop) in v.
func (v *IntVector) Slice(start, stop int) *IntVectorSlice {
	assertSliceOk(start, stop, v.Len())
	return &IntVectorSlice{vector: v, start: start, stop: stop}
}

// Len returns the length of v.
func (v *IntVector) Len() int {
	return int(v.len)
}

// Range calls f repeatedly passing it each element in v in order as argument until either
// all elements have been visited or f returns false.
func (v *IntVector) Range(f func(int) bool) {
	var currentNode []int
	for i := uint(0); i < v.len; i++ {
		if i&shiftBitMask == 0 {
			currentNode = v.sliceFor(i)
		}

		if !f(currentNode[i&shiftBitMask]) {
			return
		}
	}
}

// ToNativeSlice returns a Go slice containing all elements of v
func (v *IntVector) ToNativeSlice() []int {
	result := make([]int, 0, v.len)
	for i := uint(0); i < v.len; i += nodeSize {
		result = append(result, v.sliceFor(i)...)
	}

	return result
}

////////////////
//// Slice /////
////////////////

// IntVectorSlice is a slice type backed by a IntVector.
type IntVectorSlice struct {
	vector      *IntVector
	start, stop int
}

// NewIntVectorSlice returns a new NewIntVectorSlice containing the items provided in items.
func NewIntVectorSlice(items ...int) *IntVectorSlice {
	return &IntVectorSlice{vector: emptyIntVector.Append(items...), start: 0, stop: len(items)}
}

// Len returns the length of s.
func (s *IntVectorSlice) Len() int {
	return s.stop - s.start
}

// Get returns the element at position i.
func (s *IntVectorSlice) Get(i int) int {
	if i < 0 || s.start+i >= s.stop {
		panic("Index out of bounds")
	}

	return s.vector.Get(s.start + i)
}

// Set returns a new slice with the element at position i set to item.
func (s *IntVectorSlice) Set(i int, item int) *IntVectorSlice {
	if i < 0 || s.start+i >= s.stop {
		panic("Index out of bounds")
	}

	return s.vector.Set(s.start+i, item).Slice(s.start, s.stop)
}

// Append returns a new slice with item(s) appended to it.
func (s *IntVectorSlice) Append(items ...int) *IntVectorSlice {
	newSlice := IntVectorSlice{vector: s.vector, start: s.start, stop: s.stop + len(items)}

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

// Slice returns a IntVectorSlice that refers to all elements [start,stop) in s.
func (s *IntVectorSlice) Slice(start, stop int) *IntVectorSlice {
	assertSliceOk(start, stop, s.stop-s.start)
	return &IntVectorSlice{vector: s.vector, start: s.start + start, stop: s.start + stop}
}

// Range calls f repeatedly passing it each element in s in order as argument until either
// all elements have been visited or f returns false.
func (s *IntVectorSlice) Range(f func(int) bool) {
	var currentNode []int
	for i := uint(s.start); i < uint(s.stop); i++ {
		if i&shiftBitMask == 0 || i == uint(s.start) {
			currentNode = s.vector.sliceFor(uint(i))
		}

		if !f(currentNode[i&shiftBitMask]) {
			return
		}
	}
}

///////////
/// Map ///
///////////

//////////////////////
/// Backing vector ///
//////////////////////

type privatePersonBySsnItemBucketVector struct {
	tail  []privatePersonBySsnItemBucket
	root  commonNode
	len   uint
	shift uint
}

type PersonBySsnItem struct {
	Key   string
	Value Person
}

type privatePersonBySsnItemBucket []PersonBySsnItem

var emptyPersonBySsnItemBucketVectorTail = make([]privatePersonBySsnItemBucket, 0)
var emptyPersonBySsnItemBucketVector *privatePersonBySsnItemBucketVector = &privatePersonBySsnItemBucketVector{root: emptyCommonNode, shift: shiftSize, tail: emptyPersonBySsnItemBucketVectorTail}

func (v *privatePersonBySsnItemBucketVector) Get(i int) privatePersonBySsnItemBucket {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	return v.sliceFor(uint(i))[i&shiftBitMask]
}

func (v *privatePersonBySsnItemBucketVector) sliceFor(i uint) []privatePersonBySsnItemBucket {
	if i >= v.tailOffset() {
		return v.tail
	}

	node := v.root
	for level := v.shift; level > 0; level -= shiftSize {
		node = node.([]commonNode)[(i>>level)&shiftBitMask]
	}

	return node.([]privatePersonBySsnItemBucket)
}

func (v *privatePersonBySsnItemBucketVector) tailOffset() uint {
	if v.len < nodeSize {
		return 0
	}

	return ((v.len - 1) >> shiftSize) << shiftSize
}

func (v *privatePersonBySsnItemBucketVector) Set(i int, item privatePersonBySsnItemBucket) *privatePersonBySsnItemBucketVector {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	if uint(i) >= v.tailOffset() {
		newTail := make([]privatePersonBySsnItemBucket, len(v.tail))
		copy(newTail, v.tail)
		newTail[i&shiftBitMask] = item
		return &privatePersonBySsnItemBucketVector{root: v.root, tail: newTail, len: v.len, shift: v.shift}
	}

	return &privatePersonBySsnItemBucketVector{root: v.doAssoc(v.shift, v.root, uint(i), item), tail: v.tail, len: v.len, shift: v.shift}
}

func (v *privatePersonBySsnItemBucketVector) doAssoc(level uint, node commonNode, i uint, item privatePersonBySsnItemBucket) commonNode {
	if level == 0 {
		ret := make([]privatePersonBySsnItemBucket, nodeSize)
		copy(ret, node.([]privatePersonBySsnItemBucket))
		ret[i&shiftBitMask] = item
		return ret
	}

	ret := make([]commonNode, nodeSize)
	copy(ret, node.([]commonNode))
	subidx := (i >> level) & shiftBitMask
	ret[subidx] = v.doAssoc(level-shiftSize, ret[subidx], i, item)
	return ret
}

func (v *privatePersonBySsnItemBucketVector) pushTail(level uint, parent commonNode, tailNode []privatePersonBySsnItemBucket) commonNode {
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

func (v *privatePersonBySsnItemBucketVector) Append(item ...privatePersonBySsnItemBucket) *privatePersonBySsnItemBucketVector {
	result := v
	itemLen := uint(len(item))
	for insertOffset := uint(0); insertOffset < itemLen; {
		tailLen := result.len - result.tailOffset()
		tailFree := nodeSize - tailLen
		if tailFree == 0 {
			result = result.pushLeafNode(result.tail)
			result.tail = emptyPersonBySsnItemBucketVector.tail
			tailFree = nodeSize
			tailLen = 0
		}

		batchLen := uintMin(itemLen-insertOffset, tailFree)
		newTail := make([]privatePersonBySsnItemBucket, 0, tailLen+batchLen)
		newTail = append(newTail, result.tail...)
		newTail = append(newTail, item[insertOffset:insertOffset+batchLen]...)
		result = &privatePersonBySsnItemBucketVector{root: result.root, tail: newTail, len: result.len + batchLen, shift: result.shift}
		insertOffset += batchLen
	}

	return result
}

func (v *privatePersonBySsnItemBucketVector) pushLeafNode(node []privatePersonBySsnItemBucket) *privatePersonBySsnItemBucketVector {
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

	return &privatePersonBySsnItemBucketVector{root: newRoot, tail: v.tail, len: v.len, shift: newShift}
}

func (v *privatePersonBySsnItemBucketVector) Len() int {
	return int(v.len)
}

func (v *privatePersonBySsnItemBucketVector) Range(f func(privatePersonBySsnItemBucket) bool) {
	var currentNode []privatePersonBySsnItemBucket
	for i := uint(0); i < v.len; i++ {
		if i&shiftBitMask == 0 {
			currentNode = v.sliceFor(uint(i))
		}

		if !f(currentNode[i&shiftBitMask]) {
			return
		}
	}
}

// PersonBySsn is a persistent key - value map
type PersonBySsn struct {
	backingVector *privatePersonBySsnItemBucketVector
	len           int
}

func (m *PersonBySsn) pos(key string) int {
	return int(uint64(stringHash(key)) % uint64(m.backingVector.Len()))
}

// Helper type used during map creation and reallocation
type privatePersonBySsnItemBuckets struct {
	buckets []privatePersonBySsnItemBucket
	length  int
}

func newPrivatePersonBySsnItemBuckets(itemCount int) *privatePersonBySsnItemBuckets {
	size := int(float64(itemCount)/initialMapLoadFactor) + 1
	buckets := make([]privatePersonBySsnItemBucket, size)
	return &privatePersonBySsnItemBuckets{buckets: buckets}
}

func (b *privatePersonBySsnItemBuckets) AddItem(item PersonBySsnItem) {
	ix := int(uint64(stringHash(item.Key)) % uint64(len(b.buckets)))
	bucket := b.buckets[ix]
	if bucket != nil {
		// Hash collision, merge with existing bucket
		for keyIx, bItem := range bucket {
			if item.Key == bItem.Key {
				bucket[keyIx] = item
				return
			}
		}

		b.buckets[ix] = append(bucket, PersonBySsnItem{Key: item.Key, Value: item.Value})
		b.length++
	} else {
		bucket := make(privatePersonBySsnItemBucket, 0, int(math.Max(initialMapLoadFactor, 1.0)))
		b.buckets[ix] = append(bucket, item)
		b.length++
	}
}

func (b *privatePersonBySsnItemBuckets) AddItemsFromMap(m *PersonBySsn) {
	m.backingVector.Range(func(bucket privatePersonBySsnItemBucket) bool {
		for _, item := range bucket {
			b.AddItem(item)
		}
		return true
	})
}

func newPersonBySsn(items []PersonBySsnItem) *PersonBySsn {
	buckets := newPrivatePersonBySsnItemBuckets(len(items))
	for _, item := range items {
		buckets.AddItem(item)
	}

	return &PersonBySsn{backingVector: emptyPersonBySsnItemBucketVector.Append(buckets.buckets...), len: buckets.length}
}

// Len returns the number of items in m.
func (m *PersonBySsn) Len() int {
	return int(m.len)
}

// Load returns value identified by key. ok is set to true if key exists in the map, false otherwise.
func (m *PersonBySsn) Load(key string) (value Person, ok bool) {
	bucket := m.backingVector.Get(m.pos(key))
	if bucket != nil {
		for _, item := range bucket {
			if item.Key == key {
				return item.Value, true
			}
		}
	}

	var zeroValue Person
	return zeroValue, false
}

// Store returns a new PersonBySsn containing value identified by key.
func (m *PersonBySsn) Store(key string, value Person) *PersonBySsn {
	// Grow backing vector if load factor is too high
	if m.Len() >= m.backingVector.Len()*int(upperMapLoadFactor) {
		buckets := newPrivatePersonBySsnItemBuckets(m.Len() + 1)
		buckets.AddItemsFromMap(m)
		buckets.AddItem(PersonBySsnItem{Key: key, Value: value})
		return &PersonBySsn{backingVector: emptyPersonBySsnItemBucketVector.Append(buckets.buckets...), len: buckets.length}
	}

	pos := m.pos(key)
	bucket := m.backingVector.Get(pos)
	if bucket != nil {
		for ix, item := range bucket {
			if item.Key == key {
				// Overwrite existing item
				newBucket := make(privatePersonBySsnItemBucket, len(bucket))
				copy(newBucket, bucket)
				newBucket[ix] = PersonBySsnItem{Key: key, Value: value}
				return &PersonBySsn{backingVector: m.backingVector.Set(pos, newBucket), len: m.len}
			}
		}

		// Add new item to bucket
		newBucket := make(privatePersonBySsnItemBucket, len(bucket), len(bucket)+1)
		copy(newBucket, bucket)
		newBucket = append(newBucket, PersonBySsnItem{Key: key, Value: value})
		return &PersonBySsn{backingVector: m.backingVector.Set(pos, newBucket), len: m.len + 1}
	}

	item := PersonBySsnItem{Key: key, Value: value}
	newBucket := privatePersonBySsnItemBucket{item}
	return &PersonBySsn{backingVector: m.backingVector.Set(pos, newBucket), len: m.len + 1}
}

// Delete returns a new PersonBySsn without the element identified by key.
func (m *PersonBySsn) Delete(key string) *PersonBySsn {
	pos := m.pos(key)
	bucket := m.backingVector.Get(pos)
	if bucket != nil {
		newBucket := make(privatePersonBySsnItemBucket, 0)
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

		newMap := &PersonBySsn{backingVector: m.backingVector.Set(pos, newBucket), len: m.len - removedItemCount}
		if newMap.backingVector.Len() > 1 && newMap.Len() < newMap.backingVector.Len()*int(lowerMapLoadFactor) {
			// Shrink backing vector if needed to avoid occupying excessive space
			buckets := newPrivatePersonBySsnItemBuckets(newMap.Len())
			buckets.AddItemsFromMap(newMap)
			return &PersonBySsn{backingVector: emptyPersonBySsnItemBucketVector.Append(buckets.buckets...), len: buckets.length}
		}

		return newMap
	}

	return m
}

// Range calls f repeatedly passing it each key and value as argument until either
// all elements have been visited or f returns false.
func (m *PersonBySsn) Range(f func(string, Person) bool) {
	m.backingVector.Range(func(bucket privatePersonBySsnItemBucket) bool {
		for _, item := range bucket {
			if !f(item.Key, item.Value) {
				return false
			}
		}
		return true
	})
}

// ToNativeMap returns a native Go map containing all elements of m.
func (m *PersonBySsn) ToNativeMap() map[string]Person {
	result := make(map[string]Person)
	m.Range(func(key string, value Person) bool {
		result[key] = value
		return true
	})

	return result
}

////////////////////
/// Constructors ///
////////////////////

// NewPersonBySsn returns a new PersonBySsn containing all items in items.
func NewPersonBySsn(items ...PersonBySsnItem) *PersonBySsn {
	return newPersonBySsn(items)
}

// NewPersonBySsnFromNativeMap returns a new PersonBySsn containing all items in m.
func NewPersonBySsnFromNativeMap(m map[string]Person) *PersonBySsn {
	buckets := newPrivatePersonBySsnItemBuckets(len(m))
	for key, value := range m {
		buckets.AddItem(PersonBySsnItem{Key: key, Value: value})
	}

	return &PersonBySsn{backingVector: emptyPersonBySsnItemBucketVector.Append(buckets.buckets...), len: buckets.length}
}

///////////
/// Map ///
///////////

//////////////////////
/// Backing vector ///
//////////////////////

type privateprivatePersonsMapItemBucketVector struct {
	tail  []privateprivatePersonsMapItemBucket
	root  commonNode
	len   uint
	shift uint
}

type privatePersonsMapItem struct {
	Key   Person
	Value struct{}
}

type privateprivatePersonsMapItemBucket []privatePersonsMapItem

var emptyprivatePersonsMapItemBucketVectorTail = make([]privateprivatePersonsMapItemBucket, 0)
var emptyprivatePersonsMapItemBucketVector *privateprivatePersonsMapItemBucketVector = &privateprivatePersonsMapItemBucketVector{root: emptyCommonNode, shift: shiftSize, tail: emptyprivatePersonsMapItemBucketVectorTail}

func (v *privateprivatePersonsMapItemBucketVector) Get(i int) privateprivatePersonsMapItemBucket {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	return v.sliceFor(uint(i))[i&shiftBitMask]
}

func (v *privateprivatePersonsMapItemBucketVector) sliceFor(i uint) []privateprivatePersonsMapItemBucket {
	if i >= v.tailOffset() {
		return v.tail
	}

	node := v.root
	for level := v.shift; level > 0; level -= shiftSize {
		node = node.([]commonNode)[(i>>level)&shiftBitMask]
	}

	return node.([]privateprivatePersonsMapItemBucket)
}

func (v *privateprivatePersonsMapItemBucketVector) tailOffset() uint {
	if v.len < nodeSize {
		return 0
	}

	return ((v.len - 1) >> shiftSize) << shiftSize
}

func (v *privateprivatePersonsMapItemBucketVector) Set(i int, item privateprivatePersonsMapItemBucket) *privateprivatePersonsMapItemBucketVector {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	if uint(i) >= v.tailOffset() {
		newTail := make([]privateprivatePersonsMapItemBucket, len(v.tail))
		copy(newTail, v.tail)
		newTail[i&shiftBitMask] = item
		return &privateprivatePersonsMapItemBucketVector{root: v.root, tail: newTail, len: v.len, shift: v.shift}
	}

	return &privateprivatePersonsMapItemBucketVector{root: v.doAssoc(v.shift, v.root, uint(i), item), tail: v.tail, len: v.len, shift: v.shift}
}

func (v *privateprivatePersonsMapItemBucketVector) doAssoc(level uint, node commonNode, i uint, item privateprivatePersonsMapItemBucket) commonNode {
	if level == 0 {
		ret := make([]privateprivatePersonsMapItemBucket, nodeSize)
		copy(ret, node.([]privateprivatePersonsMapItemBucket))
		ret[i&shiftBitMask] = item
		return ret
	}

	ret := make([]commonNode, nodeSize)
	copy(ret, node.([]commonNode))
	subidx := (i >> level) & shiftBitMask
	ret[subidx] = v.doAssoc(level-shiftSize, ret[subidx], i, item)
	return ret
}

func (v *privateprivatePersonsMapItemBucketVector) pushTail(level uint, parent commonNode, tailNode []privateprivatePersonsMapItemBucket) commonNode {
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

func (v *privateprivatePersonsMapItemBucketVector) Append(item ...privateprivatePersonsMapItemBucket) *privateprivatePersonsMapItemBucketVector {
	result := v
	itemLen := uint(len(item))
	for insertOffset := uint(0); insertOffset < itemLen; {
		tailLen := result.len - result.tailOffset()
		tailFree := nodeSize - tailLen
		if tailFree == 0 {
			result = result.pushLeafNode(result.tail)
			result.tail = emptyprivatePersonsMapItemBucketVector.tail
			tailFree = nodeSize
			tailLen = 0
		}

		batchLen := uintMin(itemLen-insertOffset, tailFree)
		newTail := make([]privateprivatePersonsMapItemBucket, 0, tailLen+batchLen)
		newTail = append(newTail, result.tail...)
		newTail = append(newTail, item[insertOffset:insertOffset+batchLen]...)
		result = &privateprivatePersonsMapItemBucketVector{root: result.root, tail: newTail, len: result.len + batchLen, shift: result.shift}
		insertOffset += batchLen
	}

	return result
}

func (v *privateprivatePersonsMapItemBucketVector) pushLeafNode(node []privateprivatePersonsMapItemBucket) *privateprivatePersonsMapItemBucketVector {
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

	return &privateprivatePersonsMapItemBucketVector{root: newRoot, tail: v.tail, len: v.len, shift: newShift}
}

func (v *privateprivatePersonsMapItemBucketVector) Len() int {
	return int(v.len)
}

func (v *privateprivatePersonsMapItemBucketVector) Range(f func(privateprivatePersonsMapItemBucket) bool) {
	var currentNode []privateprivatePersonsMapItemBucket
	for i := uint(0); i < v.len; i++ {
		if i&shiftBitMask == 0 {
			currentNode = v.sliceFor(uint(i))
		}

		if !f(currentNode[i&shiftBitMask]) {
			return
		}
	}
}

// privatePersonsMap is a persistent key - value map
type privatePersonsMap struct {
	backingVector *privateprivatePersonsMapItemBucketVector
	len           int
}

func (m *privatePersonsMap) pos(key Person) int {
	return int(uint64(interfaceHash(key)) % uint64(m.backingVector.Len()))
}

// Helper type used during map creation and reallocation
type privateprivatePersonsMapItemBuckets struct {
	buckets []privateprivatePersonsMapItemBucket
	length  int
}

func newPrivateprivatePersonsMapItemBuckets(itemCount int) *privateprivatePersonsMapItemBuckets {
	size := int(float64(itemCount)/initialMapLoadFactor) + 1
	buckets := make([]privateprivatePersonsMapItemBucket, size)
	return &privateprivatePersonsMapItemBuckets{buckets: buckets}
}

func (b *privateprivatePersonsMapItemBuckets) AddItem(item privatePersonsMapItem) {
	ix := int(uint64(interfaceHash(item.Key)) % uint64(len(b.buckets)))
	bucket := b.buckets[ix]
	if bucket != nil {
		// Hash collision, merge with existing bucket
		for keyIx, bItem := range bucket {
			if item.Key == bItem.Key {
				bucket[keyIx] = item
				return
			}
		}

		b.buckets[ix] = append(bucket, privatePersonsMapItem{Key: item.Key, Value: item.Value})
		b.length++
	} else {
		bucket := make(privateprivatePersonsMapItemBucket, 0, int(math.Max(initialMapLoadFactor, 1.0)))
		b.buckets[ix] = append(bucket, item)
		b.length++
	}
}

func (b *privateprivatePersonsMapItemBuckets) AddItemsFromMap(m *privatePersonsMap) {
	m.backingVector.Range(func(bucket privateprivatePersonsMapItemBucket) bool {
		for _, item := range bucket {
			b.AddItem(item)
		}
		return true
	})
}

func newprivatePersonsMap(items []privatePersonsMapItem) *privatePersonsMap {
	buckets := newPrivateprivatePersonsMapItemBuckets(len(items))
	for _, item := range items {
		buckets.AddItem(item)
	}

	return &privatePersonsMap{backingVector: emptyprivatePersonsMapItemBucketVector.Append(buckets.buckets...), len: buckets.length}
}

// Len returns the number of items in m.
func (m *privatePersonsMap) Len() int {
	return int(m.len)
}

// Load returns value identified by key. ok is set to true if key exists in the map, false otherwise.
func (m *privatePersonsMap) Load(key Person) (value struct{}, ok bool) {
	bucket := m.backingVector.Get(m.pos(key))
	if bucket != nil {
		for _, item := range bucket {
			if item.Key == key {
				return item.Value, true
			}
		}
	}

	var zeroValue struct{}
	return zeroValue, false
}

// Store returns a new privatePersonsMap containing value identified by key.
func (m *privatePersonsMap) Store(key Person, value struct{}) *privatePersonsMap {
	// Grow backing vector if load factor is too high
	if m.Len() >= m.backingVector.Len()*int(upperMapLoadFactor) {
		buckets := newPrivateprivatePersonsMapItemBuckets(m.Len() + 1)
		buckets.AddItemsFromMap(m)
		buckets.AddItem(privatePersonsMapItem{Key: key, Value: value})
		return &privatePersonsMap{backingVector: emptyprivatePersonsMapItemBucketVector.Append(buckets.buckets...), len: buckets.length}
	}

	pos := m.pos(key)
	bucket := m.backingVector.Get(pos)
	if bucket != nil {
		for ix, item := range bucket {
			if item.Key == key {
				// Overwrite existing item
				newBucket := make(privateprivatePersonsMapItemBucket, len(bucket))
				copy(newBucket, bucket)
				newBucket[ix] = privatePersonsMapItem{Key: key, Value: value}
				return &privatePersonsMap{backingVector: m.backingVector.Set(pos, newBucket), len: m.len}
			}
		}

		// Add new item to bucket
		newBucket := make(privateprivatePersonsMapItemBucket, len(bucket), len(bucket)+1)
		copy(newBucket, bucket)
		newBucket = append(newBucket, privatePersonsMapItem{Key: key, Value: value})
		return &privatePersonsMap{backingVector: m.backingVector.Set(pos, newBucket), len: m.len + 1}
	}

	item := privatePersonsMapItem{Key: key, Value: value}
	newBucket := privateprivatePersonsMapItemBucket{item}
	return &privatePersonsMap{backingVector: m.backingVector.Set(pos, newBucket), len: m.len + 1}
}

// Delete returns a new privatePersonsMap without the element identified by key.
func (m *privatePersonsMap) Delete(key Person) *privatePersonsMap {
	pos := m.pos(key)
	bucket := m.backingVector.Get(pos)
	if bucket != nil {
		newBucket := make(privateprivatePersonsMapItemBucket, 0)
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

		newMap := &privatePersonsMap{backingVector: m.backingVector.Set(pos, newBucket), len: m.len - removedItemCount}
		if newMap.backingVector.Len() > 1 && newMap.Len() < newMap.backingVector.Len()*int(lowerMapLoadFactor) {
			// Shrink backing vector if needed to avoid occupying excessive space
			buckets := newPrivateprivatePersonsMapItemBuckets(newMap.Len())
			buckets.AddItemsFromMap(newMap)
			return &privatePersonsMap{backingVector: emptyprivatePersonsMapItemBucketVector.Append(buckets.buckets...), len: buckets.length}
		}

		return newMap
	}

	return m
}

// Range calls f repeatedly passing it each key and value as argument until either
// all elements have been visited or f returns false.
func (m *privatePersonsMap) Range(f func(Person, struct{}) bool) {
	m.backingVector.Range(func(bucket privateprivatePersonsMapItemBucket) bool {
		for _, item := range bucket {
			if !f(item.Key, item.Value) {
				return false
			}
		}
		return true
	})
}

// ToNativeMap returns a native Go map containing all elements of m.
func (m *privatePersonsMap) ToNativeMap() map[Person]struct{} {
	result := make(map[Person]struct{})
	m.Range(func(key Person, value struct{}) bool {
		result[key] = value
		return true
	})

	return result
}

// Persons is a persistent set
type Persons struct {
	backingMap *privatePersonsMap
}

// NewPersons returns a new Persons containing items.
func NewPersons(items ...Person) *Persons {
	mapItems := make([]privatePersonsMapItem, 0, len(items))
	var mapValue struct{}
	for _, x := range items {
		mapItems = append(mapItems, privatePersonsMapItem{Key: x, Value: mapValue})
	}

	return &Persons{backingMap: newprivatePersonsMap(mapItems)}
}

// Add returns a new Persons containing item.
func (s *Persons) Add(item Person) *Persons {
	var mapValue struct{}
	return &Persons{backingMap: s.backingMap.Store(item, mapValue)}
}

// Delete returns a new Persons without item.
func (s *Persons) Delete(item Person) *Persons {
	newMap := s.backingMap.Delete(item)
	if newMap == s.backingMap {
		return s
	}

	return &Persons{backingMap: newMap}
}

// Contains returns true if item is present in s, false otherwise.
func (s *Persons) Contains(item Person) bool {
	_, ok := s.backingMap.Load(item)
	return ok
}

// Range calls f repeatedly passing it each element in s as argument until either
// all elements have been visited or f returns false.
func (s *Persons) Range(f func(Person) bool) {
	s.backingMap.Range(func(k Person, _ struct{}) bool {
		return f(k)
	})
}

// IsSubset returns true if all elements in s are present in other, false otherwise.
func (s *Persons) IsSubset(other *Persons) bool {
	if other.Len() < s.Len() {
		return false
	}

	isSubset := true
	s.Range(func(item Person) bool {
		if !other.Contains(item) {
			isSubset = false
		}

		return isSubset
	})

	return isSubset
}

// IsSuperset returns true if all elements in other are present in s, false otherwise.
func (s *Persons) IsSuperset(other *Persons) bool {
	return other.IsSubset(s)
}

// Union returns a new Persons containing all elements present
// in either s or other.
func (s *Persons) Union(other *Persons) *Persons {
	result := s

	// Simplest possible solution right now. Would probable be more efficient
	// to concatenate two slices of elements from the two sets and create a
	// new set from that slice for many cases.
	other.Range(func(item Person) bool {
		result = result.Add(item)
		return true
	})

	return result
}

// Equals returns true if s and other contains the same elements, false otherwise.
func (s *Persons) Equals(other *Persons) bool {
	return s.Len() == other.Len() && s.IsSubset(other)
}

func (s *Persons) difference(other *Persons) []Person {
	items := make([]Person, 0)
	s.Range(func(item Person) bool {
		if !other.Contains(item) {
			items = append(items, item)
		}

		return true
	})

	return items
}

// Difference returns a new Persons containing all elements present
// in s but not in other.
func (s *Persons) Difference(other *Persons) *Persons {
	return NewPersons(s.difference(other)...)
}

// SymmetricDifference returns a new Persons containing all elements present
// in either s or other but not both.
func (s *Persons) SymmetricDifference(other *Persons) *Persons {
	items := s.difference(other)
	items = append(items, other.difference(s)...)
	return NewPersons(items...)
}

// Intersection returns a new Persons containing all elements present in both
// s and other.
func (s *Persons) Intersection(other *Persons) *Persons {
	items := make([]Person, 0)
	s.Range(func(item Person) bool {
		if other.Contains(item) {
			items = append(items, item)
		}

		return true
	})

	return NewPersons(items...)
}

// Len returns the number of elements in s.
func (s *Persons) Len() int {
	return s.backingMap.Len()
}

// ToNativeSlice returns a native Go slice containing all elements of s.
func (s *Persons) ToNativeSlice() []Person {
	items := make([]Person, 0, s.Len())
	s.Range(func(item Person) bool {
		items = append(items, item)
		return true
	})

	return items
}
