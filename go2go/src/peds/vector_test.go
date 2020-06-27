// Code generated by go2go; DO NOT EDIT.


//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:1
package peds

//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:1
import (
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:1
 "fmt"
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:1
 "runtime"
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:1
 "strings"
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:1
 "testing"
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:1
)

//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:13
func TestCanBeInstantiated(t *testing.T) {
	v := instantiate୦୦NewVector୦int(1, 2, 3)
	v2 := v.Append(4)
	if v2.Len() != 4 {
		t.Errorf("Expected %d != Expected 4", v2.Len())
	}

	for i := 0; i < 4; i++ {
		actual, expected := v2.Get(i), i+1
		if actual != expected {
			t.Errorf("Actual %d != Expected %d", actual, expected)
		}
	}
}

//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:32
func assertEqual(t *testing.T, expected, actual int) {
	if expected != actual {
		_, file, line, _ := runtime.Caller(1)
		t.Errorf("Error %s, line %d. Expected: %d, actual: %d", file, line, expected, actual)
	}
}

func assertEqualString(t *testing.T, expected, actual string) {
	if expected != actual {
		_, file, line, _ := runtime.Caller(1)
		t.Errorf("Error %s, line %d. Expected: %v, actual: %v", file, line, expected, actual)
	}
}

func assertEqualBool(t *testing.T, expected, actual bool) {
	if expected != actual {
		_, file, line, _ := runtime.Caller(1)
		t.Errorf("Error %s, line %d. Expected: %v, actual: %v", file, line, expected, actual)
	}
}

func assertPanic(t *testing.T, expectedMsg string) {
	if r := recover(); r == nil {
		_, _, line, _ := runtime.Caller(1)
		t.Errorf("Did not raise, line %d.", line)
	} else {
		msg := r.(string)
		if !strings.Contains(msg, expectedMsg) {
			t.Errorf("Msg '%s', did not contain '%s'", msg, expectedMsg)
		}
	}
}

func inputSlice(start, size int) []int {
	result := make([]int, 0, size)
	for i := start; i < start+size; i++ {
		result = append(result, i)
	}

	return result
}

var testSizes = []int{0, 1, 20, 32, 33, 50, 500, 32 * 32, 32*32 + 1, 10000, 32 * 32 * 32, 32*32*32 + 1}

//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:80
func TestPropertiesOfNewVector(t *testing.T) {
	for _, l := range testSizes {
		t.Run(fmt.Sprintf("NewVector %d", l), func(t *testing.T) {
			vec := instantiate୦୦NewVector୦int(inputSlice(0, l)...)
			assertEqual(t, vec.Len(), l)
			for i := 0; i < l; i++ {
				assertEqual(t, i, vec.Get(i))
			}
		})
	}
}

func TestSetItem(t *testing.T) {
	for _, l := range testSizes {
		t.Run(fmt.Sprintf("Set %d", l), func(t *testing.T) {
			vec := instantiate୦୦NewVector୦int(inputSlice(0, l)...)
			for i := 0; i < l; i++ {
				newArr := vec.Set(i, -i)
				assertEqual(t, -i, newArr.Get(i))
				assertEqual(t, i, vec.Get(i))
			}
		})
	}
}

func TestAppend(t *testing.T) {
	for _, l := range testSizes {
		vec := instantiate୦୦NewVector୦int(inputSlice(0, l)...)
		t.Run(fmt.Sprintf("Append %d", l), func(t *testing.T) {
			for i := 0; i < 70; i++ {
				newVec := vec.Append(inputSlice(l, i)...)
				assertEqual(t, i+l, newVec.Len())
				for j := 0; j < i+l; j++ {
					assertEqual(t, j, newVec.Get(j))
				}

//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:117
    assertEqual(t, l, vec.Len())
			}
		})
	}
}

func TestVectorSetOutOfBoundsNegative(t *testing.T) {
										defer assertPanic(t, "Index out of bounds")
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:124
 instantiate୦୦NewVector୦int(inputSlice(0, 10)...).Set(-1, 0)
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:126
}

func TestVectorSetOutOfBoundsBeyondEnd(t *testing.T) {
										defer assertPanic(t, "Index out of bounds")
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:129
 instantiate୦୦NewVector୦int(inputSlice(0, 10)...).Set(10, 0)
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:131
}

func TestVectorGetOutOfBoundsNegative(t *testing.T) {
										defer assertPanic(t, "Index out of bounds")
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:134
 instantiate୦୦NewVector୦int(inputSlice(0, 10)...).Get(-1)
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:136
}

func TestVectorGetOutOfBoundsBeyondEnd(t *testing.T) {
										defer assertPanic(t, "Index out of bounds")
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:139
 instantiate୦୦NewVector୦int(inputSlice(0, 10)...).Get(10)
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:141
}

func TestVectorSliceOutOfBounds(t *testing.T) {
	tests := []struct {
		start, stop int
		msg         string
	}{
		{-1, 3, "Invalid slice index"},
		{0, 11, "Slice bounds out of range"},
		{5, 3, "Invalid slice index"},
	}

	for _, s := range tests {
		t.Run(fmt.Sprintf("start=%d, stop=%d", s.start, s.stop), func(t *testing.T) {
												defer assertPanic(t, s.msg)
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:155
   instantiate୦୦NewVector୦int(inputSlice(0, 10)...).Slice(s.start, s.stop)
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:157
  })
	}
}

func TestCompleteIteration(t *testing.T) {
										input := inputSlice(0, 10000)
										dst := make([]int, 0, 10000)
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:163
 instantiate୦୦NewVector୦int(input...).Range(func(elem int) bool {
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:165
  dst = append(dst, elem)
		return true
	})

	assertEqual(t, len(input), len(dst))
	assertEqual(t, input[0], dst[0])
	assertEqual(t, input[5000], dst[5000])
	assertEqual(t, input[9999], dst[9999])
}

func TestCanceledIteration(t *testing.T) {
										input := inputSlice(0, 10000)
										count := 0
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:177
 instantiate୦୦NewVector୦int(input...).Range(func(elem int) bool {
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:179
  count++
		if count == 5 {
			return false
		}
		return true
	})

	assertEqual(t, 5, count)
}

//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:193
func TestSliceIndexes(t *testing.T) {
										vec := instantiate୦୦NewVector୦int(inputSlice(0, 1000)...)
										slice := vec.Slice(0, 10)
										assertEqual(t, 1000, vec.Len())
										assertEqual(t, 10, slice.Len())
										assertEqual(t, 0, slice.Get(0))
										assertEqual(t, 9, slice.Get(9))

//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:202
 slice2 := slice.Slice(3, 7)
	assertEqual(t, 10, slice.Len())
	assertEqual(t, 4, slice2.Len())
	assertEqual(t, 3, slice2.Get(0))
	assertEqual(t, 6, slice2.Get(3))
}

func TestSliceCreation(t *testing.T) {
	sliceLen := 10000
	slice := instantiate୦୦NewVectorSlice୦int(inputSlice(0, sliceLen)...)
	assertEqual(t, slice.Len(), sliceLen)
	for i := 0; i < sliceLen; i++ {
		assertEqual(t, i, slice.Get(i))
	}
}

func TestSliceSet(t *testing.T) {
										vector := instantiate୦୦NewVector୦int(inputSlice(0, 1000)...)
										slice := vector.Slice(10, 100)
										slice2 := slice.Set(5, 123)

//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:225
 assertEqual(t, 155, vector.Get(15))
	assertEqual(t, 15, slice.Get(5))
	assertEqual(t, 123, slice2.Get(5))
}

func TestSliceAppendInTheMiddleOfBackingVector(t *testing.T) {
										vector := instantiate୦୦NewVector୦int(inputSlice(0, 100)...)
										slice := vector.Slice(0, 50)
										slice2 := slice.Append(inputSlice(0, 10)...)

//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:236
 assertEqual(t, 60, slice2.Len())
										assertEqual(t, 0, slice2.Get(50))
										assertEqual(t, 9, slice2.Get(59))

//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:241
 assertEqual(t, 50, slice.Len())

//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:244
 assertEqual(t, 100, vector.Len())
	assertEqual(t, 50, vector.Get(50))
	assertEqual(t, 59, vector.Get(59))
}

func TestSliceAppendAtTheEndOfBackingVector(t *testing.T) {
										vector := instantiate୦୦NewVector୦int(inputSlice(0, 100)...)
										slice := vector.Slice(0, 100)
										slice2 := slice.Append(inputSlice(0, 10)...)

//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:255
 assertEqual(t, 110, slice2.Len())
										assertEqual(t, 0, slice2.Get(100))
										assertEqual(t, 9, slice2.Get(109))

//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:260
 assertEqual(t, 100, slice.Len())
}

func TestSliceAppendAtMiddleToEndOfBackingVector(t *testing.T) {
										vector := instantiate୦୦NewVector୦int(inputSlice(0, 100)...)
										slice := vector.Slice(0, 50)
										slice2 := slice.Append(inputSlice(0, 100)...)

//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:269
 assertEqual(t, 150, slice2.Len())
										assertEqual(t, 0, slice2.Get(50))
										assertEqual(t, 99, slice2.Get(149))

//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:274
 assertEqual(t, 50, slice.Len())
}

func TestSliceCompleteIteration(t *testing.T) {
	vec := instantiate୦୦NewVector୦int(inputSlice(0, 1000)...)
	dst := make([]int, 0)

	vec.Slice(5, 200).Range(func(elem int) bool {
		dst = append(dst, elem)
		return true
	})

	assertEqual(t, 195, len(dst))
	assertEqual(t, 5, dst[0])
	assertEqual(t, 55, dst[50])
	assertEqual(t, 199, dst[194])
}

func TestSliceCanceledIteration(t *testing.T) {
	vec := instantiate୦୦NewVector୦int(inputSlice(0, 1000)...)
	count := 0

	vec.Slice(5, 200).Range(func(elem int) bool {
		count++
		if count == 5 {
			return false
		}

		return true
	})

	assertEqual(t, 5, count)
}

func TestSliceSetOutOfBoundsNegative(t *testing.T) {
										defer assertPanic(t, "Index out of bounds")
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:309
 instantiate୦୦NewVector୦int(inputSlice(0, 10)...).Slice(2, 5).Set(-1, 0)
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:311
}

func TestSliceSetOutOfBoundsBeyondEnd(t *testing.T) {
										defer assertPanic(t, "Index out of bounds")
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:314
 instantiate୦୦NewVector୦int(inputSlice(0, 10)...).Slice(2, 5).Set(4, 0)
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:316
}

func TestSliceGetOutOfBoundsNegative(t *testing.T) {
										defer assertPanic(t, "Index out of bounds")
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:319
 instantiate୦୦NewVector୦int(inputSlice(0, 10)...).Slice(2, 5).Get(-1)
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:321
}

func TestSliceGetOutOfBoundsBeyondEnd(t *testing.T) {
										defer assertPanic(t, "Index out of bounds")
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:324
 instantiate୦୦NewVector୦int(inputSlice(0, 10)...).Slice(2, 5).Get(4)
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:326
}

func TestSliceSliceOutOfBounds(t *testing.T) {
	tests := []struct {
		start, stop int
		msg         string
	}{
		{-1, 3, "Invalid slice index"},
		{0, 4, "Slice bounds out of range"},
		{3, 2, "Invalid slice index"},
	}

	for _, s := range tests {
		t.Run(fmt.Sprintf("start=%d, stop=%d", s.start, s.stop), func(t *testing.T) {
												defer assertPanic(t, s.msg)
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:340
   instantiate୦୦NewVector୦int(inputSlice(0, 10)...).Slice(2, 5).Slice(s.start, s.stop)
//line /home/tobias/Development/go/peds/go2go/src/peds/vector_test.go2:342
  })
	}
}

func TestToNativeVector(t *testing.T) {
	lengths := []int{0, 1, 7, 32, 512, 1000}
	for _, length := range lengths {
		t.Run(fmt.Sprintf("length=%d", length), func(t *testing.T) {
			inputS := inputSlice(0, length)
			v := instantiate୦୦NewVector୦int(inputS...)

			outputS := v.ToNativeSlice()

			assertEqual(t, len(inputS), len(outputS))
			for i := range outputS {
				assertEqual(t, inputS[i], outputS[i])
			}
		})
	}
}
//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:34
func instantiate୦୦NewVector୦int(items ...int,

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:34
) *instantiate୦୦Vector୦int {

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:37
 tail := make([]int, 0)
	v := &instantiate୦୦Vector୦int{root: emptyCommonNode, shift: shiftSize, tail: tail}
	return v.Append(items...)
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:238
func instantiate୦୦NewVectorSlice୦int(items ...(int),) *instantiate୦୦VectorSlice୦int {
	return &instantiate୦୦VectorSlice୦int{vector: instantiate୦୦NewVector୦int(items...), start: 0, stop: len(items)}
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:240
type instantiate୦୦Vector୦int struct {
//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:27
 tail []int

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:28
 root  commonNode
										len   uint
										shift uint
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:43
func (v *instantiate୦୦Vector୦int,) Append(item ...int,

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:43
) *instantiate୦୦Vector୦int {
	result := v
	itemLen := uint(len(item))
	for insertOffset := uint(0); insertOffset < itemLen; {
		tailLen := result.len - result.tailOffset()
		tailFree := nodeSize - tailLen
		if tailFree == 0 {
			result = result.pushLeafNode(result.tail)
			result.tail = make([]int, 0)
			tailFree = nodeSize
			tailLen = 0
		}

		batchLen := uintMin(itemLen-insertOffset, tailFree)
		newTail := make([]int, 0, tailLen+batchLen)
		newTail = append(newTail, result.tail...)
		newTail = append(newTail, item[insertOffset:insertOffset+batchLen]...)
		result = &instantiate୦୦Vector୦int{root: result.root, tail: newTail, len: result.len + batchLen, shift: result.shift}
		insertOffset += batchLen
	}

	return result
}

func (v *instantiate୦୦Vector୦int,) tailOffset() uint {
	if v.len < nodeSize {
		return 0
	}

	return ((v.len - 1) >> shiftSize) << shiftSize
}

func (v *instantiate୦୦Vector୦int,) pushLeafNode(node []int,

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:75
) *instantiate୦୦Vector୦int {
										var newRoot commonNode
										newShift := v.shift

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:80
 if (v.len >> shiftSize) > (1 << v.shift) {
		newNode := newPath(v.shift, node)
		newRoot = commonNode([]commonNode{v.root, newNode})
		newShift = v.shift + shiftSize
	} else {
		newRoot = v.pushTail(v.shift, v.root, node)
	}

	return &instantiate୦୦Vector୦int{root: newRoot, tail: v.tail, len: v.len, shift: newShift}
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:99
func (v *instantiate୦୦Vector୦int,) pushTail(level uint, parent commonNode, tailNode []int,

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:99
) commonNode {
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

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:119
func (v *instantiate୦୦Vector୦int,) Len() int {
	return int(v.len)
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:124
func (v *instantiate୦୦Vector୦int,) Get(i int) int {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	return v.sliceFor(uint(i))[i&shiftBitMask]
}

func (v *instantiate୦୦Vector୦int,) sliceFor(i uint) []int {
	if i >= v.tailOffset() {
		return v.tail
	}

	node := v.root
	for level := v.shift; level > 0; level -= shiftSize {
		node = node.([]commonNode)[(i>>level)&shiftBitMask]
	}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:144
 return node.([]int)
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:149
func (v *instantiate୦୦Vector୦int,) Set(i int, item int,

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:149
) *instantiate୦୦Vector୦int {
	if i < 0 || uint(i) >= v.len {
		panic("Index out of bounds")
	}

	if uint(i) >= v.tailOffset() {
		newTail := make([]int, len(v.tail))
		copy(newTail, v.tail)
		newTail[i&shiftBitMask] = item
		return &instantiate୦୦Vector୦int{root: v.root, tail: newTail, len: v.len, shift: v.shift}
	}

	return &instantiate୦୦Vector୦int{root: v.doAssoc(v.shift, v.root, uint(i), item), tail: v.tail, len: v.len, shift: v.shift}
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:165
func (v *instantiate୦୦Vector୦int,) doAssoc(level uint, node commonNode, i uint, item int,

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:165
) commonNode {
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

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:182
func (v *instantiate୦୦Vector୦int,) Range(f func(int,

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:182
) bool) {
										var currentNode []int

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:184
 for i := uint(0); i < v.len; i++ {
		if i&shiftBitMask == 0 {
			currentNode = v.sliceFor(i)
		}

		if !f(currentNode[i&shiftBitMask]) {
			return
		}
	}
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:196
func (v *instantiate୦୦Vector୦int,) Slice(start, stop int) *instantiate୦୦VectorSlice୦int {
	assertSliceOk(start, stop, v.Len())
	return &instantiate୦୦VectorSlice୦int{vector: v, start: start, stop: stop}
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:203
func (v *instantiate୦୦Vector୦int,) ToNativeSlice() []int {
	result := make([]int, 0, v.len)
	for i := uint(0); i < v.len; i += nodeSize {
		result = append(result, v.sliceFor(i)...)
	}

	return result
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:210
type instantiate୦୦VectorSlice୦int struct {
//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:233
 vector *instantiate୦୦Vector୦int
	start, stop int
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:243
func (s *instantiate୦୦VectorSlice୦int,) Len() int {
	return s.stop - s.start
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:248
func (s *instantiate୦୦VectorSlice୦int,) Get(i int) (int) {
	if i < 0 || s.start+i >= s.stop {
		panic("Index out of bounds")
	}

	return s.vector.Get(s.start + i)
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:257
func (s *instantiate୦୦VectorSlice୦int,) Set(i int, item (int),) *instantiate୦୦VectorSlice୦int {
	if i < 0 || s.start+i >= s.stop {
		panic("Index out of bounds")
	}

	return s.vector.Set(s.start+i, item).Slice(s.start, s.stop)
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:266
func (s *instantiate୦୦VectorSlice୦int,) Append(items ...(int),) *instantiate୦୦VectorSlice୦int {
										newSlice := instantiate୦୦VectorSlice୦int{vector: s.vector, start: s.start, stop: s.stop + len(items)}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:272
 itemPos := 0
	for ; s.stop+itemPos < s.vector.Len() && itemPos < len(items); itemPos++ {
		newSlice.vector = newSlice.vector.Set(s.stop+itemPos, items[itemPos])
	}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:278
 newSlice.vector = newSlice.vector.Append(items[itemPos:]...)
	return &newSlice
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:283
func (s *instantiate୦୦VectorSlice୦int,) Slice(start, stop int) *instantiate୦୦VectorSlice୦int {
	assertSliceOk(start, stop, s.stop-s.start)
	return &instantiate୦୦VectorSlice୦int{vector: s.vector, start: s.start + start, stop: s.start + stop}
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:290
func (s *instantiate୦୦VectorSlice୦int,) Range(f func((int),) bool) {
	var currentNode [](int)
	for i := uint(s.start); i < uint(s.stop); i++ {
		if i&shiftBitMask == 0 || i == uint(s.start) {
			currentNode = s.vector.sliceFor(uint(i))
		}

		if !f(currentNode[i&shiftBitMask]) {
			return
		}
	}
}

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:301
var _ = fmt.Errorf
//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:301
var _ = runtime.BlockProfile

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:301
type _ strings.Builder

//line /home/tobias/Development/go/peds/go2go/src/peds/containers.go2:301
var _ = testing.AllocsPerRun
