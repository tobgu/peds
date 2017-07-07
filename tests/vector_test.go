package peds_testing

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

///////////////
/// Helpers ///
///////////////

func assertEqual(t *testing.T, expected, actual int) {
	if expected != actual {
		_, file, line, _ := runtime.Caller(1)
		t.Errorf("Error %s, line %d. Expected: %d, actual: %d", file, line, expected, actual)
	}
}

func assertEqualString(t *testing.T, expected, actual string) {
	if expected != actual {
		_, file, line, _ := runtime.Caller(1)
		t.Errorf("Error %s, line %d. Expected: %d, actual: %d", file, line, expected, actual)
	}
}

func assertEqualBool(t *testing.T, expected, actual bool) {
	if expected != actual {
		_, file, line, _ := runtime.Caller(1)
		t.Errorf("Error %s, line %d. Expected: %d, actual: %d", file, line, expected, actual)
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

func inputArray(start, size int) []int {
	result := make([]int, 0, size)
	for i := start; i < start+size; i++ {
		result = append(result, i)
	}

	return result
}

var testSizes = []int{0, 1, 20, 32, 33, 50, 500, 32 * 32, 32*32 + 1, 10000, 32 * 32 * 32, 32*32*32 + 1}

/////////////
/// Array ///
/////////////

func TestPropertiesOfNewArray(t *testing.T) {
	for _, l := range testSizes {
		t.Run(fmt.Sprintf("NewArray %d", l), func(t *testing.T) {
			arr := NewIntVector(inputArray(0, l)...)
			assertEqual(t, arr.Len(), l)
			for i := 0; i < l; i++ {
				assertEqual(t, i, arr.Get(i))
			}
		})
	}
}

func TestSetItem(t *testing.T) {
	for _, l := range testSizes {
		t.Run(fmt.Sprintf("Set %d", l), func(t *testing.T) {
			arr := NewIntVector(inputArray(0, l)...)
			for i := 0; i < l; i++ {
				newArr := arr.Set(i, -i)
				assertEqual(t, -i, newArr.Get(i))
				assertEqual(t, i, arr.Get(i))
			}
		})
	}
}

func TestAppend(t *testing.T) {
	for _, l := range testSizes {
		arr := NewIntVector(inputArray(0, l)...)
		t.Run(fmt.Sprintf("Append %d", l), func(t *testing.T) {
			for i := 0; i < 70; i++ {
				newArr := arr.Append(inputArray(l, i)...)
				assertEqual(t, i+l, newArr.Len())
				for j := 0; j < i+l; j++ {
					assertEqual(t, j, newArr.Get(j))
				}

				// Original array is unchanged
				assertEqual(t, l, arr.Len())
			}
		})
	}
}

func TestArraySetOutOfBoundsNegative(t *testing.T) {
	defer assertPanic(t, "Index out of bounds")
	NewIntVector(inputArray(0, 10)...).Set(-1, 0)
}

func TestArraySetOutOfBoundsBeyondEnd(t *testing.T) {
	defer assertPanic(t, "Index out of bounds")
	NewIntVector(inputArray(0, 10)...).Set(10, 0)
}

func TestArrayGetOutOfBoundsNegative(t *testing.T) {
	defer assertPanic(t, "Index out of bounds")
	NewIntVector(inputArray(0, 10)...).Get(-1)
}

func TestArrayGetOutOfBoundsBeyondEnd(t *testing.T) {
	defer assertPanic(t, "Index out of bounds")
	NewIntVector(inputArray(0, 10)...).Get(10)
}

func TestArraySliceOutOfBounds(t *testing.T) {
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
			NewIntVector(inputArray(0, 10)...).Slice(s.start, s.stop)
		})
	}
}

//////////////
/// Vector ///
//////////////

func TestIteration(t *testing.T) {
	input := inputArray(0, 10000)
	arr := NewIntVector(input...)
	iter := arr.Iter()
	dst := make([]int, 0, 10000)
	for elem, ok := iter.Next(); ok; elem, ok = iter.Next() {
		dst = append(dst, elem)
	}

	assertEqual(t, len(input), len(dst))
	assertEqual(t, input[0], dst[0])
	assertEqual(t, input[5000], dst[5000])
	assertEqual(t, input[9999], dst[9999])
}

/////////////
/// Slice ///
/////////////

func TestSliceIndexes(t *testing.T) {
	arr := NewIntVector(inputArray(0, 1000)...)
	slice := arr.Slice(0, 10)
	assertEqual(t, 1000, arr.Len())
	assertEqual(t, 10, slice.Len())
	assertEqual(t, 0, slice.Get(0))
	assertEqual(t, 9, slice.Get(9))

	// Slice of slice
	slice2 := slice.Slice(3, 7)
	assertEqual(t, 10, slice.Len())
	assertEqual(t, 4, slice2.Len())
	assertEqual(t, 3, slice2.Get(0))
	assertEqual(t, 6, slice2.Get(3))
}

func TestSliceCreation(t *testing.T) {
	sliceLen := 10000
	slice := NewIntVectorSlice(inputArray(0, sliceLen)...)
	assertEqual(t, slice.Len(), sliceLen)
	for i := 0; i < sliceLen; i++ {
		assertEqual(t, i, slice.Get(i))
	}
}

func TestSliceSet(t *testing.T) {
	array := NewIntVector(inputArray(0, 1000)...)
	slice := array.Slice(10, 100)
	slice2 := slice.Set(5, 123)

	// Underlying array and original slice should remain unchanged. New slice updated
	// in the correct position
	assertEqual(t, 15, array.Get(15))
	assertEqual(t, 15, slice.Get(5))
	assertEqual(t, 123, slice2.Get(5))
}

func TestSliceAppendInTheMiddleOfBackingArray(t *testing.T) {
	array := NewIntVector(inputArray(0, 100)...)
	slice := array.Slice(0, 50)
	slice2 := slice.Append(inputArray(0, 10)...)

	// New slice
	assertEqual(t, 60, slice2.Len())
	assertEqual(t, 0, slice2.Get(50))
	assertEqual(t, 9, slice2.Get(59))

	// Original slice
	assertEqual(t, 50, slice.Len())

	// Original array
	assertEqual(t, 100, array.Len())
	assertEqual(t, 50, array.Get(50))
	assertEqual(t, 59, array.Get(59))
}

func TestSliceAppendAtTheEndOfBackingArray(t *testing.T) {
	array := NewIntVector(inputArray(0, 100)...)
	slice := array.Slice(0, 100)
	slice2 := slice.Append(inputArray(0, 10)...)

	// New slice
	assertEqual(t, 110, slice2.Len())
	assertEqual(t, 0, slice2.Get(100))
	assertEqual(t, 9, slice2.Get(109))

	// Original slice
	assertEqual(t, 100, slice.Len())
}

func TestSliceAppendAtMiddleToEndOfBackingArray(t *testing.T) {
	array := NewIntVector(inputArray(0, 100)...)
	slice := array.Slice(0, 50)
	slice2 := slice.Append(inputArray(0, 100)...)

	// New slice
	assertEqual(t, 150, slice2.Len())
	assertEqual(t, 0, slice2.Get(50))
	assertEqual(t, 99, slice2.Get(149))

	// Original slice
	assertEqual(t, 50, slice.Len())
}

func TestSliceIteration(t *testing.T) {
	arr := NewIntVector(inputArray(0, 1000)...)
	iter := arr.Slice(5, 200).Iter()
	dst := make([]int, 0)
	for elem, ok := iter.Next(); ok; elem, ok = iter.Next() {
		dst = append(dst, elem)
	}

	assertEqual(t, 195, len(dst))
	assertEqual(t, 5, dst[0])
	assertEqual(t, 55, dst[50])
	assertEqual(t, 199, dst[194])
}

func TestSliceSetOutOfBoundsNegative(t *testing.T) {
	defer assertPanic(t, "Index out of bounds")
	NewIntVector(inputArray(0, 10)...).Slice(2, 5).Set(-1, 0)
}

func TestSliceSetOutOfBoundsBeyondEnd(t *testing.T) {
	defer assertPanic(t, "Index out of bounds")
	NewIntVector(inputArray(0, 10)...).Slice(2, 5).Set(4, 0)
}

func TestSliceGetOutOfBoundsNegative(t *testing.T) {
	defer assertPanic(t, "Index out of bounds")
	NewIntVector(inputArray(0, 10)...).Slice(2, 5).Get(-1)
}

func TestSliceGetOutOfBoundsBeyondEnd(t *testing.T) {
	defer assertPanic(t, "Index out of bounds")
	NewIntVector(inputArray(0, 10)...).Slice(2, 5).Get(4)
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
			NewIntVector(inputArray(0, 10)...).Slice(2, 5).Slice(s.start, s.stop)
		})
	}
}

// TODO:
// - Document public types and methods
// - Expand README with examples

//////////////////////
///// Benchmarks /////
//////////////////////

// Used to avoid that the compiler optimizes the code away
var result int

func runIteration(b *testing.B, size int) {
	arr := NewIntVector(inputArray(0, size)...)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		iter := arr.Iter()
		for value, ok := iter.Next(); ok; value, ok = iter.Next() {
			result += value
		}

	}
}

func BenchmarkLargeIteration(b *testing.B) {
	runIteration(b, 100000)
}

func BenchmarkSmallIteration(b *testing.B) {
	runIteration(b, 10)
}
