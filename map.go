package peds

import "github.com/cheekybits/genny/generic"

type Key generic.Type
type Value generic.Type

type KeyValueItem struct {
	key   Key
	value Value
}

/////////////////////
/// Backing array ///
/////////////////////

type KeyValueMap struct {
	tail  []KeyValueItem
	root  privateKeyValueNode
	len   uint
	shift uint
}

// The "private" prefix is there just for Genny to match on the type name "KeyValueItem"
// but we don't want to expose this type outside the package.

type privateKeyValueNode interface{}

var emptyKeyValueNode privateKeyValueNode = []privateKeyValueNode{}
var emptyKeyValueTail = make([]KeyValueItem, 0)
var emptyKeyValueArray *KeyValueMap = &KeyValueMap{root: emptyKeyValueNode, shift: privateKeyValueshift, tail: emptyKeyValueTail}

const privateKeyValueshift = 5
const privateKeyValueNodeSize = 32
const privateKeyValueBitMask = 0x1F

func NewKeyValueMap(items ...KeyValueItem) *KeyValueMap {
	return emptyKeyValueArray.mappend(items...)
}

func (m *KeyValueMap) get(i int) KeyValueItem {
	if i < 0 || uint(i) >= m.len {
		panic("Index out of bounds")
	}

	return m.arrayFor(uint(i))[i&privateKeyValueBitMask]
}

func (m *KeyValueMap) arrayFor(i uint) []KeyValueItem {
	if i >= m.tailOffset() {
		return m.tail
	}

	node := m.root
	for level := m.shift; level > 0; level -= privateKeyValueshift {
		node = node.([]privateKeyValueNode)[(i>>level)&privateKeyValueBitMask]
	}

	return node.([]KeyValueItem)
}

func (m *KeyValueMap) tailOffset() uint {
	if m.len < privateKeyValueNodeSize {
		return 0
	}

	return ((m.len - 1) >> privateKeyValueshift) << privateKeyValueshift
}

func (m *KeyValueMap) set(i int, item KeyValueItem) *KeyValueMap {
	if i < 0 || uint(i) >= m.len {
		panic("Index out of bounds")
	}

	if uint(i) >= m.tailOffset() {
		newTail := make([]KeyValueItem, len(m.tail))
		copy(newTail, m.tail)
		newTail[i&privateKeyValueBitMask] = item
		return &KeyValueMap{root: m.root, tail: newTail, len: m.len, shift: m.shift}
	}

	return &KeyValueMap{root: m.doAssoc(m.shift, m.root, uint(i), item), tail: m.tail, len: m.len, shift: m.shift}
}

func (m *KeyValueMap) doAssoc(level uint, node privateKeyValueNode, i uint, item KeyValueItem) privateKeyValueNode {
	if level == 0 {
		ret := make([]KeyValueItem, privateKeyValueNodeSize)
		copy(ret, node.([]KeyValueItem))
		ret[i&privateKeyValueBitMask] = item
		return ret
	}

	ret := make([]privateKeyValueNode, privateKeyValueNodeSize)
	copy(ret, node.([]privateKeyValueNode))
	subidx := (i >> level) & privateKeyValueBitMask
	ret[subidx] = m.doAssoc(level-privateKeyValueshift, ret[subidx], i, item)
	return ret
}

func newKeyValuePath(shift uint, node privateKeyValueNode) privateKeyValueNode {
	if shift == 0 {
		return node
	}

	return newKeyValuePath(shift-privateKeyValueshift, privateKeyValueNode([]privateKeyValueNode{node}))
}

func (m *KeyValueMap) pushTail(level uint, parent privateKeyValueNode, tailNode []KeyValueItem) privateKeyValueNode {
	subIdx := ((m.len - 1) >> level) & privateKeyValueBitMask
	parentNode := parent.([]privateKeyValueNode)
	ret := make([]privateKeyValueNode, subIdx+1)
	copy(ret, parentNode)
	var nodeToInsert privateKeyValueNode

	if level == privateKeyValueshift {
		nodeToInsert = tailNode
	} else if subIdx < uint(len(parentNode)) {
		nodeToInsert = m.pushTail(level-privateKeyValueshift, parentNode[subIdx], tailNode)
	} else {
		nodeToInsert = newKeyValuePath(level-privateKeyValueshift, tailNode)
	}

	ret[subIdx] = nodeToInsert
	return ret
}

func uintKeyValueMin(a, b uint) uint {
	if a < b {
		return a
	}

	return b
}

func (m *KeyValueMap) mappend(item ...KeyValueItem) *KeyValueMap {
	result := m
	itemLen := uint(len(item))
	for insertOffset := uint(0); insertOffset < itemLen; {
		tailLen := result.len - result.tailOffset()
		tailFree := privateKeyValueNodeSize - tailLen
		if tailFree == 0 {
			result = result.pushLeafNode(result.tail)
			result.tail = emptyKeyValueArray.tail
			tailFree = privateKeyValueNodeSize
			tailLen = 0
		}

		batchLen := uintKeyValueMin(itemLen-insertOffset, tailFree)
		newTail := make([]KeyValueItem, 0, tailLen+batchLen)
		newTail = append(newTail, result.tail...)
		newTail = append(newTail, item[insertOffset:insertOffset+batchLen]...)
		result = &KeyValueMap{root: result.root, tail: newTail, len: result.len + batchLen, shift: result.shift}
		insertOffset += batchLen
	}

	return result
}

func (m *KeyValueMap) pushLeafNode(node []KeyValueItem) *KeyValueMap {
	var newRoot privateKeyValueNode
	newShift := m.shift

	// Root overflow?
	if (m.len >> privateKeyValueshift) > (1 << m.shift) {
		newNode := newKeyValuePath(m.shift, node)
		newRoot = privateKeyValueNode([]privateKeyValueNode{m.root, newNode})
		newShift = m.shift + privateKeyValueshift
	} else {
		newRoot = m.pushTail(m.shift, m.root, node)
	}

	return &KeyValueMap{root: newRoot, tail: m.tail, len: m.len, shift: newShift}
}


////////////////////////
/// Public functions ///
////////////////////////

func (m *KeyValueMap) Len() int {
	return int(m.len)
}


func (m *KeyValueMap) Load(key Key) (value Value, ok bool) {
	var temp Value
	return temp, false
}

func (m *KeyValueMap) Store(key Key, value Value) *KeyValueMap {
	return &KeyValueMap{}
}

func (m *KeyValueMap) Delete(key Key) *KeyValueMap {
	return &KeyValueMap{}
}

func (m *KeyValueMap) Range(f func(key Key, value Value) bool) {
}
