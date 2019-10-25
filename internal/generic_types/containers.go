// Package generic_types contains underlying Go implementation of types that
// will be used in text templates to generate specific implementations.
package generic_types

//template:CommonTemplate

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"math"
	"unsafe"
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

//go:noescape
//go:linkname nilinterhash runtime.nilinterhash
func nilinterhash(p unsafe.Pointer, h uintptr) uintptr

func interfaceHash(x interface{}) uint32 {
	return uint32(nilinterhash(unsafe.Pointer(&x), 0))
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

//template:VectorTemplate

//////////////
/// Vector ///
//////////////

// A GenericVectorType is an ordered persistent/immutable collection of items corresponding roughly
// to the use cases for a slice.
type GenericVectorType struct {
	tail  []GenericType
	root  commonNode
	len   uint
	shift uint
}

var emptyGenericVectorTypeTail = make([]GenericType, 0)
var emptyGenericVectorType *GenericVectorType = &GenericVectorType{root: emptyCommonNode, shift: shiftSize, tail: emptyGenericVectorTypeTail}

// NewGenericVectorType returns a new GenericVectorType containing the items provided in items.
func NewGenericVectorType(items ...GenericType) *GenericVectorType {
	return emptyGenericVectorType.Append(items...)
}

// Get returns the element at position i.
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

// Set returns a new vector with the element at position i set to item.
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

// Append returns a new vector with item(s) appended to it.
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

// Slice returns a GenericVectorTypeSlice that refers to all elements [start,stop) in v.
func (v *GenericVectorType) Slice(start, stop int) *GenericVectorTypeSlice {
	assertSliceOk(start, stop, v.Len())
	return &GenericVectorTypeSlice{vector: v, start: start, stop: stop}
}

// Len returns the length of v.
func (v *GenericVectorType) Len() int {
	return int(v.len)
}

// Range calls f repeatedly passing it each element in v in order as argument until either
// all elements have been visited or f returns false.
func (v *GenericVectorType) Range(f func(GenericType) bool) {
	var currentNode []GenericType
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
func (v *GenericVectorType) ToNativeSlice() []GenericType {
	result := make([]GenericType, 0, v.len)
	for i := uint(0); i < v.len; i += nodeSize {
		result = append(result, v.sliceFor(i)...)
	}

	return result
}

//template:SliceTemplate

////////////////
//// Slice /////
////////////////

// GenericVectorTypeSlice is a slice type backed by a GenericVectorType.
type GenericVectorTypeSlice struct {
	vector      *GenericVectorType
	start, stop int
}

// NewGenericVectorTypeSlice returns a new NewGenericVectorTypeSlice containing the items provided in items.
func NewGenericVectorTypeSlice(items ...GenericType) *GenericVectorTypeSlice {
	return &GenericVectorTypeSlice{vector: emptyGenericVectorType.Append(items...), start: 0, stop: len(items)}
}

// Len returns the length of s.
func (s *GenericVectorTypeSlice) Len() int {
	return s.stop - s.start
}

// Get returns the element at position i.
func (s *GenericVectorTypeSlice) Get(i int) GenericType {
	if i < 0 || s.start+i >= s.stop {
		panic("Index out of bounds")
	}

	return s.vector.Get(s.start + i)
}

// Set returns a new slice with the element at position i set to item.
func (s *GenericVectorTypeSlice) Set(i int, item GenericType) *GenericVectorTypeSlice {
	if i < 0 || s.start+i >= s.stop {
		panic("Index out of bounds")
	}

	return s.vector.Set(s.start+i, item).Slice(s.start, s.stop)
}

// Append returns a new slice with item(s) appended to it.
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

// Slice returns a GenericVectorTypeSlice that refers to all elements [start,stop) in s.
func (s *GenericVectorTypeSlice) Slice(start, stop int) *GenericVectorTypeSlice {
	assertSliceOk(start, stop, s.stop-s.start)
	return &GenericVectorTypeSlice{vector: s.vector, start: s.start + start, stop: s.start + stop}
}

// Range calls f repeatedly passing it each element in s in order as argument until either
// all elements have been visited or f returns false.
func (s *GenericVectorTypeSlice) Range(f func(GenericType) bool) {
	var currentNode []GenericType
	for i := uint(s.start); i < uint(s.stop); i++ {
		if i&shiftBitMask == 0 || i == uint(s.start) {
			currentNode = s.vector.sliceFor(uint(i))
		}

		if !f(currentNode[i&shiftBitMask]) {
			return
		}
	}
}

//template:PrivateMapTemplate

///////////
/// Map ///
///////////

//////////////////////
/// Backing vector ///
//////////////////////

type privateGenericMapItemBucketVector struct {
	tail  []privateGenericMapItemBucket
	root  commonNode
	len   uint
	shift uint
}

type GenericMapItem struct {
	Key   GenericMapKeyType
	Value GenericMapValueType
}

type privateGenericMapItemBucket []GenericMapItem

var emptyGenericMapItemBucketVectorTail = make([]privateGenericMapItemBucket, 0)
var emptyGenericMapItemBucketVector *privateGenericMapItemBucketVector = &privateGenericMapItemBucketVector{root: emptyCommonNode, shift: shiftSize, tail: emptyGenericMapItemBucketVectorTail}

func (v *privateGenericMapItemBucketVector) Get(i int) privateGenericMapItemBucket {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	return v.sliceFor(uint(i))[i&shiftBitMask]
}

func (v *privateGenericMapItemBucketVector) sliceFor(i uint) []privateGenericMapItemBucket {
	if i >= v.tailOffset() {
		return v.tail
	}

	node := v.root
	for level := v.shift; level > 0; level -= shiftSize {
		node = node.([]commonNode)[(i>>level)&shiftBitMask]
	}

	return node.([]privateGenericMapItemBucket)
}

func (v *privateGenericMapItemBucketVector) tailOffset() uint {
	if v.len < nodeSize {
		return 0
	}

	return ((v.len - 1) >> shiftSize) << shiftSize
}

func (v *privateGenericMapItemBucketVector) Set(i int, item privateGenericMapItemBucket) *privateGenericMapItemBucketVector {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	if uint(i) >= v.tailOffset() {
		newTail := make([]privateGenericMapItemBucket, len(v.tail))
		copy(newTail, v.tail)
		newTail[i&shiftBitMask] = item
		return &privateGenericMapItemBucketVector{root: v.root, tail: newTail, len: v.len, shift: v.shift}
	}

	return &privateGenericMapItemBucketVector{root: v.doAssoc(v.shift, v.root, uint(i), item), tail: v.tail, len: v.len, shift: v.shift}
}

func (v *privateGenericMapItemBucketVector) doAssoc(level uint, node commonNode, i uint, item privateGenericMapItemBucket) commonNode {
	if level == 0 {
		ret := make([]privateGenericMapItemBucket, nodeSize)
		copy(ret, node.([]privateGenericMapItemBucket))
		ret[i&shiftBitMask] = item
		return ret
	}

	ret := make([]commonNode, nodeSize)
	copy(ret, node.([]commonNode))
	subidx := (i >> level) & shiftBitMask
	ret[subidx] = v.doAssoc(level-shiftSize, ret[subidx], i, item)
	return ret
}

func (v *privateGenericMapItemBucketVector) pushTail(level uint, parent commonNode, tailNode []privateGenericMapItemBucket) commonNode {
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

func (v *privateGenericMapItemBucketVector) Append(item ...privateGenericMapItemBucket) *privateGenericMapItemBucketVector {
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
		newTail := make([]privateGenericMapItemBucket, 0, tailLen+batchLen)
		newTail = append(newTail, result.tail...)
		newTail = append(newTail, item[insertOffset:insertOffset+batchLen]...)
		result = &privateGenericMapItemBucketVector{root: result.root, tail: newTail, len: result.len + batchLen, shift: result.shift}
		insertOffset += batchLen
	}

	return result
}

func (v *privateGenericMapItemBucketVector) pushLeafNode(node []privateGenericMapItemBucket) *privateGenericMapItemBucketVector {
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

	return &privateGenericMapItemBucketVector{root: newRoot, tail: v.tail, len: v.len, shift: newShift}
}

func (v *privateGenericMapItemBucketVector) Len() int {
	return int(v.len)
}

func (v *privateGenericMapItemBucketVector) Range(f func(privateGenericMapItemBucket) bool) {
	var currentNode []privateGenericMapItemBucket
	for i := uint(0); i < v.len; i++ {
		if i&shiftBitMask == 0 {
			currentNode = v.sliceFor(uint(i))
		}

		if !f(currentNode[i&shiftBitMask]) {
			return
		}
	}
}

// GenericMapType is a persistent key - value map
type GenericMapType struct {
	backingVector *privateGenericMapItemBucketVector
	len           int
}

func (m *GenericMapType) pos(key GenericMapKeyType) int {
	return int(uint64(genericHash(key)) % uint64(m.backingVector.Len()))
}

// Helper type used during map creation and reallocation
type privateGenericMapItemBuckets struct {
	buckets []privateGenericMapItemBucket
	length  int
}

func newPrivateGenericMapItemBuckets(itemCount int) *privateGenericMapItemBuckets {
	size := int(float64(itemCount)/initialMapLoadFactor) + 1
	buckets := make([]privateGenericMapItemBucket, size)
	return &privateGenericMapItemBuckets{buckets: buckets}
}

func (b *privateGenericMapItemBuckets) AddItem(item GenericMapItem) {
	ix := int(uint64(genericHash(item.Key)) % uint64(len(b.buckets)))
	bucket := b.buckets[ix]
	if bucket != nil {
		// Hash collision, merge with existing bucket
		for keyIx, bItem := range bucket {
			if item.Key == bItem.Key {
				bucket[keyIx] = item
				return
			}
		}

		b.buckets[ix] = append(bucket, GenericMapItem{Key: item.Key, Value: item.Value})
		b.length++
	} else {
		bucket := make(privateGenericMapItemBucket, 0, int(math.Max(initialMapLoadFactor, 1.0)))
		b.buckets[ix] = append(bucket, item)
		b.length++
	}
}

func (b *privateGenericMapItemBuckets) AddItemsFromMap(m *GenericMapType) {
	m.backingVector.Range(func(bucket privateGenericMapItemBucket) bool {
		for _, item := range bucket {
			b.AddItem(item)
		}
		return true
	})
}

func newGenericMapType(items []GenericMapItem) *GenericMapType {
	buckets := newPrivateGenericMapItemBuckets(len(items))
	for _, item := range items {
		buckets.AddItem(item)
	}

	return &GenericMapType{backingVector: emptyGenericMapItemBucketVector.Append(buckets.buckets...), len: buckets.length}
}

// Len returns the number of items in m.
func (m *GenericMapType) Len() int {
	return int(m.len)
}

// Load returns value identified by key. ok is set to true if key exists in the map, false otherwise.
func (m *GenericMapType) Load(key GenericMapKeyType) (value GenericMapValueType, ok bool) {
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

// Store returns a new GenericMapType containing value identified by key.
func (m *GenericMapType) Store(key GenericMapKeyType, value GenericMapValueType) *GenericMapType {
	// Grow backing vector if load factor is too high
	if m.Len() >= m.backingVector.Len()*int(upperMapLoadFactor) {
		buckets := newPrivateGenericMapItemBuckets(m.Len() + 1)
		buckets.AddItemsFromMap(m)
		buckets.AddItem(GenericMapItem{Key: key, Value: value})
		return &GenericMapType{backingVector: emptyGenericMapItemBucketVector.Append(buckets.buckets...), len: buckets.length}
	}

	pos := m.pos(key)
	bucket := m.backingVector.Get(pos)
	if bucket != nil {
		for ix, item := range bucket {
			if item.Key == key {
				// Overwrite existing item
				newBucket := make(privateGenericMapItemBucket, len(bucket))
				copy(newBucket, bucket)
				newBucket[ix] = GenericMapItem{Key: key, Value: value}
				return &GenericMapType{backingVector: m.backingVector.Set(pos, newBucket), len: m.len}
			}
		}

		// Add new item to bucket
		newBucket := make(privateGenericMapItemBucket, len(bucket), len(bucket)+1)
		copy(newBucket, bucket)
		newBucket = append(newBucket, GenericMapItem{Key: key, Value: value})
		return &GenericMapType{backingVector: m.backingVector.Set(pos, newBucket), len: m.len + 1}
	}

	item := GenericMapItem{Key: key, Value: value}
	newBucket := privateGenericMapItemBucket{item}
	return &GenericMapType{backingVector: m.backingVector.Set(pos, newBucket), len: m.len + 1}
}

// Delete returns a new GenericMapType without the element identified by key.
func (m *GenericMapType) Delete(key GenericMapKeyType) *GenericMapType {
	pos := m.pos(key)
	bucket := m.backingVector.Get(pos)
	if bucket != nil {
		newBucket := make(privateGenericMapItemBucket, 0)
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

		newMap := &GenericMapType{backingVector: m.backingVector.Set(pos, newBucket), len: m.len - removedItemCount}
		if newMap.backingVector.Len() > 1 && newMap.Len() < newMap.backingVector.Len()*int(lowerMapLoadFactor) {
			// Shrink backing vector if needed to avoid occupying excessive space
			buckets := newPrivateGenericMapItemBuckets(newMap.Len())
			buckets.AddItemsFromMap(newMap)
			return &GenericMapType{backingVector: emptyGenericMapItemBucketVector.Append(buckets.buckets...), len: buckets.length}
		}

		return newMap
	}

	return m
}

// Range calls f repeatedly passing it each key and value as argument until either
// all elements have been visited or f returns false.
func (m *GenericMapType) Range(f func(GenericMapKeyType, GenericMapValueType) bool) {
	m.backingVector.Range(func(bucket privateGenericMapItemBucket) bool {
		for _, item := range bucket {
			if !f(item.Key, item.Value) {
				return false
			}
		}
		return true
	})
}

// ToNativeMap returns a native Go map containing all elements of m.
func (m *GenericMapType) ToNativeMap() map[GenericMapKeyType]GenericMapValueType {
	result := make(map[GenericMapKeyType]GenericMapValueType)
	m.Range(func(key GenericMapKeyType, value GenericMapValueType) bool {
		result[key] = value
		return true
	})

	return result
}

//template:PublicMapTemplate

////////////////////
/// Constructors ///
////////////////////

// NewGenericMapType returns a new GenericMapType containing all items in items.
func NewGenericMapType(items ...GenericMapItem) *GenericMapType {
	return newGenericMapType(items)
}

// NewGenericMapTypeFromNativeMap returns a new GenericMapType containing all items in m.
func NewGenericMapTypeFromNativeMap(m map[GenericMapKeyType]GenericMapValueType) *GenericMapType {
	buckets := newPrivateGenericMapItemBuckets(len(m))
	for key, value := range m {
		buckets.AddItem(GenericMapItem{Key: key, Value: value})
	}

	return &GenericMapType{backingVector: emptyGenericMapItemBucketVector.Append(buckets.buckets...), len: buckets.length}
}

//template:SetTemplate

// GenericSetType is a persistent set
type GenericSetType struct {
	backingMap *GenericMapType
}

// NewGenericSetType returns a new GenericSetType containing items.
func NewGenericSetType(items ...GenericMapKeyType) *GenericSetType {
	mapItems := make([]GenericMapItem, 0, len(items))
	var mapValue GenericMapValueType
	for _, x := range items {
		mapItems = append(mapItems, GenericMapItem{Key: x, Value: mapValue})
	}

	return &GenericSetType{backingMap: newGenericMapType(mapItems)}
}

// Add returns a new GenericSetType containing item.
func (s *GenericSetType) Add(item GenericMapKeyType) *GenericSetType {
	var mapValue GenericMapValueType
	return &GenericSetType{backingMap: s.backingMap.Store(item, mapValue)}
}

// Delete returns a new GenericSetType without item.
func (s *GenericSetType) Delete(item GenericMapKeyType) *GenericSetType {
	newMap := s.backingMap.Delete(item)
	if newMap == s.backingMap {
		return s
	}

	return &GenericSetType{backingMap: newMap}
}

// Contains returns true if item is present in s, false otherwise.
func (s *GenericSetType) Contains(item GenericMapKeyType) bool {
	_, ok := s.backingMap.Load(item)
	return ok
}

// Range calls f repeatedly passing it each element in s as argument until either
// all elements have been visited or f returns false.
func (s *GenericSetType) Range(f func(GenericMapKeyType) bool) {
	s.backingMap.Range(func(k GenericMapKeyType, _ GenericMapValueType) bool {
		return f(k)
	})
}

// IsSubset returns true if all elements in s are present in other, false otherwise.
func (s *GenericSetType) IsSubset(other *GenericSetType) bool {
	if other.Len() < s.Len() {
		return false
	}

	isSubset := true
	s.Range(func(item GenericMapKeyType) bool {
		if !other.Contains(item) {
			isSubset = false
		}

		return isSubset
	})

	return isSubset
}

// IsSuperset returns true if all elements in other are present in s, false otherwise.
func (s *GenericSetType) IsSuperset(other *GenericSetType) bool {
	return other.IsSubset(s)
}

// Union returns a new GenericSetType containing all elements present
// in either s or other.
func (s *GenericSetType) Union(other *GenericSetType) *GenericSetType {
	result := s

	// Simplest possible solution right now. Would probable be more efficient
	// to concatenate two slices of elements from the two sets and create a
	// new set from that slice for many cases.
	other.Range(func(item GenericMapKeyType) bool {
		result = result.Add(item)
		return true
	})

	return result
}

// Equals returns true if s and other contains the same elements, false otherwise.
func (s *GenericSetType) Equals(other *GenericSetType) bool {
	return s.Len() == other.Len() && s.IsSubset(other)
}

func (s *GenericSetType) difference(other *GenericSetType) []GenericMapKeyType {
	items := make([]GenericMapKeyType, 0)
	s.Range(func(item GenericMapKeyType) bool {
		if !other.Contains(item) {
			items = append(items, item)
		}

		return true
	})

	return items
}

// Difference returns a new GenericSetType containing all elements present
// in s but not in other.
func (s *GenericSetType) Difference(other *GenericSetType) *GenericSetType {
	return NewGenericSetType(s.difference(other)...)
}

// SymmetricDifference returns a new GenericSetType containing all elements present
// in either s or other but not both.
func (s *GenericSetType) SymmetricDifference(other *GenericSetType) *GenericSetType {
	items := s.difference(other)
	items = append(items, other.difference(s)...)
	return NewGenericSetType(items...)
}

// Intersection returns a new GenericSetType containing all elements present in both
// s and other.
func (s *GenericSetType) Intersection(other *GenericSetType) *GenericSetType {
	items := make([]GenericMapKeyType, 0)
	s.Range(func(item GenericMapKeyType) bool {
		if other.Contains(item) {
			items = append(items, item)
		}

		return true
	})

	return NewGenericSetType(items...)
}

// Len returns the number of elements in s.
func (s *GenericSetType) Len() int {
	return s.backingMap.Len()
}

// ToNativeSlice returns a native Go slice containing all elements of s.
func (s *GenericSetType) ToNativeSlice() []GenericMapKeyType {
	items := make([]GenericMapKeyType, 0, s.Len())
	s.Range(func(item GenericMapKeyType) bool {
		items = append(items, item)
		return true
	})

	return items
}

//template:commentsNotWantedInGeneratedCode

// peds -maps "FooMap<int, string>;BarMap<int16, int32>"
//      -sets "FooSet<mypackage.MyType>"
//      -vectors "FooVec<io.Bar>"
//      -imports "io;github.com/my/mypackage"
//      -package mycontainers
//      -file mycontainers_gen.go
