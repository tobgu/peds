package templates_template

//template:CommonTemplate

// TODO: Need a way to specify imports required by different pieces of the code
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

// TODO: Perhaps make this private?
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
	return int(uint64(genericHash(key)) % uint64(m.backingVector.Len()))
}

// Helper type used during map creation and reallocation
type privateGenericMapItemBuckets struct {
	buckets []GenericMapItemBucket
	length  int
}

func newPrivateGenericMapItemBuckets(itemCount int) *privateGenericMapItemBuckets {
	size := int(float64(itemCount)/initialMapLoadFactor) + 1
	buckets := make([]GenericMapItemBucket, size)
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
		bucket := make(GenericMapItemBucket, 0, int(math.Max(initialMapLoadFactor, 1.0)))
		b.buckets[ix] = append(bucket, item)
		b.length++
	}
}

func (b *privateGenericMapItemBuckets) AddItemsFromMap(m *GenericMapType) {
	it := m.backingVector.Iter()
	for bucket, ok := it.Next(); ok; bucket, ok = it.Next() {
		for _, item := range bucket {
			b.AddItem(item)
		}
	}
}

func newGenericMapType(items []GenericMapItem) *GenericMapType {
	buckets := newPrivateGenericMapItemBuckets(len(items))
	for _, item := range items {
		buckets.AddItem(item)
	}

	return &GenericMapType{backingVector: emptyGenericMapItemBucketVector.Append(buckets.buckets...), len: buckets.length}
}

func (m *GenericMapType) Len() int {
	return int(m.len)
}

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

func (m *GenericMapType) Delete(key GenericMapKeyType) *GenericMapType {
	pos := m.pos(key)
	bucket := m.backingVector.Get(pos)
	if bucket != nil {
		newBucket := make(GenericMapItemBucket, 0)
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

func (m *GenericMapType) Range(f func(key GenericMapKeyType, value GenericMapValueType) bool) {
	it := m.backingVector.Iter()
	for bucket, ok := it.Next(); ok; bucket, ok = it.Next() {
		for _, item := range bucket {
			if !f(item.Key, item.Value) {
				return
			}
		}
	}
}

func (m *GenericMapType) ToNativeMap() map[GenericMapKeyType]GenericMapValueType {
	result := make(map[GenericMapKeyType]GenericMapValueType)
	m.Range(func(key GenericMapKeyType, value GenericMapValueType) bool {
		result[key] = value
		return true
	})

	return result
}

//template:PublicMapTemplate

////////////////////////
/// Public functions ///
////////////////////////

func NewGenericMapType(items ...GenericMapItem) *GenericMapType {
	return newGenericMapType(items)
}

func NewGenericMapTypeFromNativeMap(m map[GenericMapKeyType]GenericMapValueType) *GenericMapType {
	buckets := newPrivateGenericMapItemBuckets(len(m))
	for key, value := range m {
		buckets.AddItem(GenericMapItem{Key: key, Value: value})
	}

	return &GenericMapType{backingVector: emptyGenericMapItemBucketVector.Append(buckets.buckets...), len: buckets.length}
}

//template:SetTemplate

type GenericSetType struct {
	backingMap *GenericMapType
}

func NewGenericSetType(items ...GenericMapKeyType) *GenericSetType {
	mapItems := make([]GenericMapItem, 0, len(items))
	var mapValue GenericMapValueType
	for _, x := range items {
		mapItems = append(mapItems, GenericMapItem{Key: x, Value: mapValue})
	}

	return &GenericSetType{backingMap: newGenericMapType(mapItems)}
}

// TODO: Variadic?
func (s *GenericSetType) Add(item GenericMapKeyType) *GenericSetType {
	var mapValue GenericMapValueType
	return &GenericSetType{backingMap: s.backingMap.Store(item, mapValue)}
}

// TODO: Variadic?
func (s *GenericSetType) Delete(item GenericMapKeyType) *GenericSetType {
	newMap := s.backingMap.Delete(item)
	if newMap == s.backingMap {
		return s
	}

	return &GenericSetType{backingMap: newMap}
}

func (s *GenericSetType) Contains(item GenericMapKeyType) bool {
	_, ok := s.backingMap.Load(item)
	return ok
}

func (s *GenericSetType) Range(f func(item GenericMapKeyType) bool) {
	s.backingMap.Range(func(k GenericMapKeyType, _ GenericMapValueType) bool {
		return f(k)
	})
}

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

func (s *GenericSetType) IsSuperset(other *GenericSetType) bool {
	return other.IsSubset(s)
}

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

func (s *GenericSetType) Difference(other *GenericSetType) *GenericSetType {
	return NewGenericSetType(s.difference(other)...)
}

func (s *GenericSetType) SymmetricDifference(other *GenericSetType) *GenericSetType {
	items := s.difference(other)
	items = append(items, other.difference(s)...)
	return NewGenericSetType(items...)
}

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

func (s *GenericSetType) Len() int {
	return s.backingMap.Len()
}

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
