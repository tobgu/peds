package templates

// NOTE: This file is auto generated, don't edit manually!
const CommonTemplate string = `
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

`
const PrivateMapTemplate string = `
///////////
/// Map ///
///////////

//////////////////////
/// Backing vector ///
//////////////////////

type private{{.MapItemTypeName}}BucketVector struct {
	tail  []private{{.MapItemTypeName}}Bucket
	root  commonNode
	len   uint
	shift uint
}

type {{.MapItemTypeName}} struct {
	Key   {{.MapKeyTypeName}}
	Value {{.MapValueTypeName}}
}

type private{{.MapItemTypeName}}Bucket []{{.MapItemTypeName}}

var empty{{.MapItemTypeName}}BucketVectorTail = make([]private{{.MapItemTypeName}}Bucket, 0)
var empty{{.MapItemTypeName}}BucketVector *private{{.MapItemTypeName}}BucketVector = &private{{.MapItemTypeName}}BucketVector{root: emptyCommonNode, shift: shiftSize, tail: empty{{.MapItemTypeName}}BucketVectorTail}

func (v *private{{.MapItemTypeName}}BucketVector) Get(i int) private{{.MapItemTypeName}}Bucket {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	return v.sliceFor(uint(i))[i&shiftBitMask]
}

func (v *private{{.MapItemTypeName}}BucketVector) sliceFor(i uint) []private{{.MapItemTypeName}}Bucket {
	if i >= v.tailOffset() {
		return v.tail
	}

	node := v.root
	for level := v.shift; level > 0; level -= shiftSize {
		node = node.([]commonNode)[(i>>level)&shiftBitMask]
	}

	return node.([]private{{.MapItemTypeName}}Bucket)
}

func (v *private{{.MapItemTypeName}}BucketVector) tailOffset() uint {
	if v.len < nodeSize {
		return 0
	}

	return ((v.len - 1) >> shiftSize) << shiftSize
}

func (v *private{{.MapItemTypeName}}BucketVector) Set(i int, item private{{.MapItemTypeName}}Bucket) *private{{.MapItemTypeName}}BucketVector {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	if uint(i) >= v.tailOffset() {
		newTail := make([]private{{.MapItemTypeName}}Bucket, len(v.tail))
		copy(newTail, v.tail)
		newTail[i&shiftBitMask] = item
		return &private{{.MapItemTypeName}}BucketVector{root: v.root, tail: newTail, len: v.len, shift: v.shift}
	}

	return &private{{.MapItemTypeName}}BucketVector{root: v.doAssoc(v.shift, v.root, uint(i), item), tail: v.tail, len: v.len, shift: v.shift}
}

func (v *private{{.MapItemTypeName}}BucketVector) doAssoc(level uint, node commonNode, i uint, item private{{.MapItemTypeName}}Bucket) commonNode {
	if level == 0 {
		ret := make([]private{{.MapItemTypeName}}Bucket, nodeSize)
		copy(ret, node.([]private{{.MapItemTypeName}}Bucket))
		ret[i&shiftBitMask] = item
		return ret
	}

	ret := make([]commonNode, nodeSize)
	copy(ret, node.([]commonNode))
	subidx := (i >> level) & shiftBitMask
	ret[subidx] = v.doAssoc(level-shiftSize, ret[subidx], i, item)
	return ret
}

func (v *private{{.MapItemTypeName}}BucketVector) pushTail(level uint, parent commonNode, tailNode []private{{.MapItemTypeName}}Bucket) commonNode {
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

func (v *private{{.MapItemTypeName}}BucketVector) Append(item ...private{{.MapItemTypeName}}Bucket) *private{{.MapItemTypeName}}BucketVector {
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
		newTail := make([]private{{.MapItemTypeName}}Bucket, 0, tailLen+batchLen)
		newTail = append(newTail, result.tail...)
		newTail = append(newTail, item[insertOffset:insertOffset+batchLen]...)
		result = &private{{.MapItemTypeName}}BucketVector{root: result.root, tail: newTail, len: result.len + batchLen, shift: result.shift}
		insertOffset += batchLen
	}

	return result
}

func (v *private{{.MapItemTypeName}}BucketVector) pushLeafNode(node []private{{.MapItemTypeName}}Bucket) *private{{.MapItemTypeName}}BucketVector {
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

	return &private{{.MapItemTypeName}}BucketVector{root: newRoot, tail: v.tail, len: v.len, shift: newShift}
}

func (v *private{{.MapItemTypeName}}BucketVector) Len() int {
	return int(v.len)
}

func (v *private{{.MapItemTypeName}}BucketVector) Range(f func(private{{.MapItemTypeName}}Bucket) bool) {
	var currentNode []private{{.MapItemTypeName}}Bucket
	for i := uint(0); i < v.len; i++ {
		if i&shiftBitMask == 0 {
			currentNode = v.sliceFor(uint(i))
		}

		if !f(currentNode[i&shiftBitMask]) {
			return
		}
	}
}

// {{.MapTypeName}} is a persistent key - value map
type {{.MapTypeName}} struct {
	backingVector *private{{.MapItemTypeName}}BucketVector
	len           int
}

func (m *{{.MapTypeName}}) pos(key {{.MapKeyTypeName}}) int {
	return int(uint64({{.MapKeyHashFunc}}(key)) % uint64(m.backingVector.Len()))
}

// Helper type used during map creation and reallocation
type private{{.MapItemTypeName}}Buckets struct {
	buckets []private{{.MapItemTypeName}}Bucket
	length  int
}

func newPrivate{{.MapItemTypeName}}Buckets(itemCount int) *private{{.MapItemTypeName}}Buckets {
	size := int(float64(itemCount)/initialMapLoadFactor) + 1
	buckets := make([]private{{.MapItemTypeName}}Bucket, size)
	return &private{{.MapItemTypeName}}Buckets{buckets: buckets}
}

func (b *private{{.MapItemTypeName}}Buckets) AddItem(item {{.MapItemTypeName}}) {
	ix := int(uint64({{.MapKeyHashFunc}}(item.Key)) % uint64(len(b.buckets)))
	bucket := b.buckets[ix]
	if bucket != nil {
		// Hash collision, merge with existing bucket
		for keyIx, bItem := range bucket {
			if item.Key == bItem.Key {
				bucket[keyIx] = item
				return
			}
		}

		b.buckets[ix] = append(bucket, {{.MapItemTypeName}}{Key: item.Key, Value: item.Value})
		b.length++
	} else {
		bucket := make(private{{.MapItemTypeName}}Bucket, 0, int(math.Max(initialMapLoadFactor, 1.0)))
		b.buckets[ix] = append(bucket, item)
		b.length++
	}
}

func (b *private{{.MapItemTypeName}}Buckets) AddItemsFromMap(m *{{.MapTypeName}}) {
	m.backingVector.Range(func(bucket private{{.MapItemTypeName}}Bucket) bool {
		for _, item := range bucket {
			b.AddItem(item)
		}
		return true
	})
}

func new{{.MapTypeName}}(items []{{.MapItemTypeName}}) *{{.MapTypeName}} {
	buckets := newPrivate{{.MapItemTypeName}}Buckets(len(items))
	for _, item := range items {
		buckets.AddItem(item)
	}

	return &{{.MapTypeName}}{backingVector: empty{{.MapItemTypeName}}BucketVector.Append(buckets.buckets...), len: buckets.length}
}

// Len returns the number of items in m.
func (m *{{.MapTypeName}}) Len() int {
	return int(m.len)
}

// Load returns value identified by key. ok is set to true if key exists in the map, false otherwise.
func (m *{{.MapTypeName}}) Load(key {{.MapKeyTypeName}}) (value {{.MapValueTypeName}}, ok bool) {
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

// Store returns a new {{.MapTypeName}} containing value identified by key.
func (m *{{.MapTypeName}}) Store(key {{.MapKeyTypeName}}, value {{.MapValueTypeName}}) *{{.MapTypeName}} {
	// Grow backing vector if load factor is too high
	if m.Len() >= m.backingVector.Len()*int(upperMapLoadFactor) {
		buckets := newPrivate{{.MapItemTypeName}}Buckets(m.Len() + 1)
		buckets.AddItemsFromMap(m)
		buckets.AddItem({{.MapItemTypeName}}{Key: key, Value: value})
		return &{{.MapTypeName}}{backingVector: empty{{.MapItemTypeName}}BucketVector.Append(buckets.buckets...), len: buckets.length}
	}

	pos := m.pos(key)
	bucket := m.backingVector.Get(pos)
	if bucket != nil {
		for ix, item := range bucket {
			if item.Key == key {
				// Overwrite existing item
				newBucket := make(private{{.MapItemTypeName}}Bucket, len(bucket))
				copy(newBucket, bucket)
				newBucket[ix] = {{.MapItemTypeName}}{Key: key, Value: value}
				return &{{.MapTypeName}}{backingVector: m.backingVector.Set(pos, newBucket), len: m.len}
			}
		}

		// Add new item to bucket
		newBucket := make(private{{.MapItemTypeName}}Bucket, len(bucket), len(bucket)+1)
		copy(newBucket, bucket)
		newBucket = append(newBucket, {{.MapItemTypeName}}{Key: key, Value: value})
		return &{{.MapTypeName}}{backingVector: m.backingVector.Set(pos, newBucket), len: m.len + 1}
	}

	item := {{.MapItemTypeName}}{Key: key, Value: value}
	newBucket := private{{.MapItemTypeName}}Bucket{item}
	return &{{.MapTypeName}}{backingVector: m.backingVector.Set(pos, newBucket), len: m.len + 1}
}

// Delete returns a new {{.MapTypeName}} without the element identified by key.
func (m *{{.MapTypeName}}) Delete(key {{.MapKeyTypeName}}) *{{.MapTypeName}} {
	pos := m.pos(key)
	bucket := m.backingVector.Get(pos)
	if bucket != nil {
		newBucket := make(private{{.MapItemTypeName}}Bucket, 0)
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

		newMap := &{{.MapTypeName}}{backingVector: m.backingVector.Set(pos, newBucket), len: m.len - removedItemCount}
		if newMap.backingVector.Len() > 1 && newMap.Len() < newMap.backingVector.Len()*int(lowerMapLoadFactor) {
			// Shrink backing vector if needed to avoid occupying excessive space
			buckets := newPrivate{{.MapItemTypeName}}Buckets(newMap.Len())
			buckets.AddItemsFromMap(newMap)
			return &{{.MapTypeName}}{backingVector: empty{{.MapItemTypeName}}BucketVector.Append(buckets.buckets...), len: buckets.length}
		}

		return newMap
	}

	return m
}

// Range calls f repeatedly passing it each key and value as argument until either
// all elements have been visited or f returns false.
func (m *{{.MapTypeName}}) Range(f func({{.MapKeyTypeName}}, {{.MapValueTypeName}}) bool) {
	m.backingVector.Range(func(bucket private{{.MapItemTypeName}}Bucket) bool {
		for _, item := range bucket {
			if !f(item.Key, item.Value) {
				return false
			}
		}
		return true
	})
}

// ToNativeMap returns a native Go map containing all elements of m.
func (m *{{.MapTypeName}}) ToNativeMap() map[{{.MapKeyTypeName}}]{{.MapValueTypeName}} {
	result := make(map[{{.MapKeyTypeName}}]{{.MapValueTypeName}})
	m.Range(func(key {{.MapKeyTypeName}}, value {{.MapValueTypeName}}) bool {
		result[key] = value
		return true
	})

	return result
}

`
const PublicMapTemplate string = `
////////////////////
/// Constructors ///
////////////////////

// New{{.MapTypeName}} returns a new {{.MapTypeName}} containing all items in items.
func New{{.MapTypeName}}(items ...{{.MapItemTypeName}}) *{{.MapTypeName}} {
	return new{{.MapTypeName}}(items)
}

// New{{.MapTypeName}}FromNativeMap returns a new {{.MapTypeName}} containing all items in m.
func New{{.MapTypeName}}FromNativeMap(m map[{{.MapKeyTypeName}}]{{.MapValueTypeName}}) *{{.MapTypeName}} {
	buckets := newPrivate{{.MapItemTypeName}}Buckets(len(m))
	for key, value := range m {
		buckets.AddItem({{.MapItemTypeName}}{Key: key, Value: value})
	}

	return &{{.MapTypeName}}{backingVector: empty{{.MapItemTypeName}}BucketVector.Append(buckets.buckets...), len: buckets.length}
}

`
const SetTemplate string = `
// {{.SetTypeName}} is a persistent set
type {{.SetTypeName}} struct {
	backingMap *{{.MapTypeName}}
}

// New{{.SetTypeName}} returns a new {{.SetTypeName}} containing items.
func New{{.SetTypeName}}(items ...{{.MapKeyTypeName}}) *{{.SetTypeName}} {
	mapItems := make([]{{.MapItemTypeName}}, 0, len(items))
	var mapValue {{.MapValueTypeName}}
	for _, x := range items {
		mapItems = append(mapItems, {{.MapItemTypeName}}{Key: x, Value: mapValue})
	}

	return &{{.SetTypeName}}{backingMap: new{{.MapTypeName}}(mapItems)}
}

// Add returns a new {{.SetTypeName}} containing item.
func (s *{{.SetTypeName}}) Add(item {{.MapKeyTypeName}}) *{{.SetTypeName}} {
	var mapValue {{.MapValueTypeName}}
	return &{{.SetTypeName}}{backingMap: s.backingMap.Store(item, mapValue)}
}

// Delete returns a new {{.SetTypeName}} without item.
func (s *{{.SetTypeName}}) Delete(item {{.MapKeyTypeName}}) *{{.SetTypeName}} {
	newMap := s.backingMap.Delete(item)
	if newMap == s.backingMap {
		return s
	}

	return &{{.SetTypeName}}{backingMap: newMap}
}

// Contains returns true if item is present in s, false otherwise.
func (s *{{.SetTypeName}}) Contains(item {{.MapKeyTypeName}}) bool {
	_, ok := s.backingMap.Load(item)
	return ok
}

// Range calls f repeatedly passing it each element in s as argument until either
// all elements have been visited or f returns false.
func (s *{{.SetTypeName}}) Range(f func({{.MapKeyTypeName}}) bool) {
	s.backingMap.Range(func(k {{.MapKeyTypeName}}, _ {{.MapValueTypeName}}) bool {
		return f(k)
	})
}

// IsSubset returns true if all elements in s are present in other, false otherwise.
func (s *{{.SetTypeName}}) IsSubset(other *{{.SetTypeName}}) bool {
	if other.Len() < s.Len() {
		return false
	}

	isSubset := true
	s.Range(func(item {{.MapKeyTypeName}}) bool {
		if !other.Contains(item) {
			isSubset = false
		}

		return isSubset
	})

	return isSubset
}

// IsSuperset returns true if all elements in other are present in s, false otherwise.
func (s *{{.SetTypeName}}) IsSuperset(other *{{.SetTypeName}}) bool {
	return other.IsSubset(s)
}

// Union returns a new {{.SetTypeName}} containing all elements present
// in either s or other.
func (s *{{.SetTypeName}}) Union(other *{{.SetTypeName}}) *{{.SetTypeName}} {
	result := s

	// Simplest possible solution right now. Would probable be more efficient
	// to concatenate two slices of elements from the two sets and create a
	// new set from that slice for many cases.
	other.Range(func(item {{.MapKeyTypeName}}) bool {
		result = result.Add(item)
		return true
	})

	return result
}

// Equals returns true if s and other contains the same elements, false otherwise.
func (s *{{.SetTypeName}}) Equals(other *{{.SetTypeName}}) bool {
	return s.Len() == other.Len() && s.IsSubset(other)
}

func (s *{{.SetTypeName}}) difference(other *{{.SetTypeName}}) []{{.MapKeyTypeName}} {
	items := make([]{{.MapKeyTypeName}}, 0)
	s.Range(func(item {{.MapKeyTypeName}}) bool {
		if !other.Contains(item) {
			items = append(items, item)
		}

		return true
	})

	return items
}

// Difference returns a new {{.SetTypeName}} containing all elements present
// in s but not in other.
func (s *{{.SetTypeName}}) Difference(other *{{.SetTypeName}}) *{{.SetTypeName}} {
	return New{{.SetTypeName}}(s.difference(other)...)
}

// SymmetricDifference returns a new {{.SetTypeName}} containing all elements present
// in either s or other but not both.
func (s *{{.SetTypeName}}) SymmetricDifference(other *{{.SetTypeName}}) *{{.SetTypeName}} {
	items := s.difference(other)
	items = append(items, other.difference(s)...)
	return New{{.SetTypeName}}(items...)
}

// Intersection returns a new {{.SetTypeName}} containing all elements present in both
// s and other.
func (s *{{.SetTypeName}}) Intersection(other *{{.SetTypeName}}) *{{.SetTypeName}} {
	items := make([]{{.MapKeyTypeName}}, 0)
	s.Range(func(item {{.MapKeyTypeName}}) bool {
		if other.Contains(item) {
			items = append(items, item)
		}

		return true
	})

	return New{{.SetTypeName}}(items...)
}

// Len returns the number of elements in s.
func (s *{{.SetTypeName}}) Len() int {
	return s.backingMap.Len()
}

// ToNativeSlice returns a native Go slice containing all elements of s.
func (s *{{.SetTypeName}}) ToNativeSlice() []{{.MapKeyTypeName}} {
	items := make([]{{.MapKeyTypeName}}, 0, s.Len())
	s.Range(func(item {{.MapKeyTypeName}}) bool {
		items = append(items, item)
		return true
	})

	return items
}

`
const SliceTemplate string = `
////////////////
//// Slice /////
////////////////

// {{.VectorTypeName}}Slice is a slice type backed by a {{.VectorTypeName}}.
type {{.VectorTypeName}}Slice struct {
	vector      *{{.VectorTypeName}}
	start, stop int
}

// New{{.VectorTypeName}}Slice returns a new New{{.VectorTypeName}}Slice containing the items provided in items.
func New{{.VectorTypeName}}Slice(items ...{{.TypeName}}) *{{.VectorTypeName}}Slice {
	return &{{.VectorTypeName}}Slice{vector: empty{{.VectorTypeName}}.Append(items...), start: 0, stop: len(items)}
}

// Len returns the length of s.
func (s *{{.VectorTypeName}}Slice) Len() int {
	return s.stop - s.start
}

// Get returns the element at position i.
func (s *{{.VectorTypeName}}Slice) Get(i int) {{.TypeName}} {
	if i < 0 || s.start+i >= s.stop {
		panic("Index out of bounds")
	}

	return s.vector.Get(s.start + i)
}

// Set returns a new slice with the element at position i set to item.
func (s *{{.VectorTypeName}}Slice) Set(i int, item {{.TypeName}}) *{{.VectorTypeName}}Slice {
	if i < 0 || s.start+i >= s.stop {
		panic("Index out of bounds")
	}

	return s.vector.Set(s.start+i, item).Slice(s.start, s.stop)
}

// Append returns a new slice with item(s) appended to it.
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

// Slice returns a {{.VectorTypeName}}Slice that refers to all elements [start,stop) in s.
func (s *{{.VectorTypeName}}Slice) Slice(start, stop int) *{{.VectorTypeName}}Slice {
	assertSliceOk(start, stop, s.stop-s.start)
	return &{{.VectorTypeName}}Slice{vector: s.vector, start: s.start + start, stop: s.start + stop}
}

// Range calls f repeatedly passing it each element in s in order as argument until either
// all elements have been visited or f returns false.
func (s *{{.VectorTypeName}}Slice) Range(f func({{.TypeName}}) bool) {
	var currentNode []{{.TypeName}}
	for i := uint(s.start); i < uint(s.stop); i++ {
		if i&shiftBitMask == 0 || i == uint(s.start) {
			currentNode = s.vector.sliceFor(uint(i))
		}

		if !f(currentNode[i&shiftBitMask]) {
			return
		}
	}
}

`
const VectorTemplate string = `
//////////////
/// Vector ///
//////////////

// A {{.VectorTypeName}} is an ordered persistent/immutable collection of items corresponding roughly
// to the use cases for a slice.
type {{.VectorTypeName}} struct {
	tail  []{{.TypeName}}
	root  commonNode
	len   uint
	shift uint
}

var empty{{.VectorTypeName}}Tail = make([]{{.TypeName}}, 0)
var empty{{.VectorTypeName}} *{{.VectorTypeName}} = &{{.VectorTypeName}}{root: emptyCommonNode, shift: shiftSize, tail: empty{{.VectorTypeName}}Tail}

// New{{.VectorTypeName}} returns a new {{.VectorTypeName}} containing the items provided in items.
func New{{.VectorTypeName}}(items ...{{.TypeName}}) *{{.VectorTypeName}} {
	return empty{{.VectorTypeName}}.Append(items...)
}

// Get returns the element at position i.
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

// Set returns a new vector with the element at position i set to item.
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

// Append returns a new vector with item(s) appended to it.
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

// Slice returns a {{.VectorTypeName}}Slice that refers to all elements [start,stop) in v.
func (v *{{.VectorTypeName}}) Slice(start, stop int) *{{.VectorTypeName}}Slice {
	assertSliceOk(start, stop, v.Len())
	return &{{.VectorTypeName}}Slice{vector: v, start: start, stop: stop}
}

// Len returns the length of v.
func (v *{{.VectorTypeName}}) Len() int {
	return int(v.len)
}

// Range calls f repeatedly passing it each element in v in order as argument until either
// all elements have been visited or f returns false.
func (v *{{.VectorTypeName}}) Range(f func({{.TypeName}}) bool) {
	var currentNode []{{.TypeName}}
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
func (v *{{.VectorTypeName}}) ToNativeSlice() []{{.TypeName}} {
	result := make([]{{.TypeName}}, 0, v.len)
	for i := uint(0); i < v.len; i += nodeSize {
		result = append(result, v.sliceFor(i)...)
	}

	return result
}

`
const commentsNotWantedInGeneratedCode string = `
// peds -maps "FooMap<int, string>;BarMap<int16, int32>"
//      -sets "FooSet<mypackage.MyType>"
//      -vectors "FooVec<io.Bar>"
//      -imports "io;github.com/my/mypackage"
//      -package mycontainers
//      -file mycontainers_gen.go
`
