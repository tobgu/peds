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

func inputSlice(start, size int) []int {
	result := make([]int, 0, size)
	for i := start; i < start+size; i++ {
		result = append(result, i)
	}

	return result
}

var testSizes = []int{0, 1, 20, 32, 33, 50, 500, 32 * 32, 32*32 + 1, 10000, 32 * 32 * 32, 32*32*32 + 1}

//////////////
/// Vector ///
//////////////

func TestPropertiesOfNewVector(t *testing.T) {
	for _, l := range testSizes {
		t.Run(fmt.Sprintf("NewVector %d", l), func(t *testing.T) {
			vec := NewIntVector(inputSlice(0, l)...)
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
			vec := NewIntVector(inputSlice(0, l)...)
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
		vec := NewIntVector(inputSlice(0, l)...)
		t.Run(fmt.Sprintf("Append %d", l), func(t *testing.T) {
			for i := 0; i < 70; i++ {
				newVec := vec.Append(inputSlice(l, i)...)
				assertEqual(t, i+l, newVec.Len())
				for j := 0; j < i+l; j++ {
					assertEqual(t, j, newVec.Get(j))
				}

				// Original vector is unchanged
				assertEqual(t, l, vec.Len())
			}
		})
	}
}

func TestVectorSetOutOfBoundsNegative(t *testing.T) {
	defer assertPanic(t, "Index out of bounds")
	NewIntVector(inputSlice(0, 10)...).Set(-1, 0)
}

func TestVectorSetOutOfBoundsBeyondEnd(t *testing.T) {
	defer assertPanic(t, "Index out of bounds")
	NewIntVector(inputSlice(0, 10)...).Set(10, 0)
}

func TestVectorGetOutOfBoundsNegative(t *testing.T) {
	defer assertPanic(t, "Index out of bounds")
	NewIntVector(inputSlice(0, 10)...).Get(-1)
}

func TestVectorGetOutOfBoundsBeyondEnd(t *testing.T) {
	defer assertPanic(t, "Index out of bounds")
	NewIntVector(inputSlice(0, 10)...).Get(10)
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
			NewIntVector(inputSlice(0, 10)...).Slice(s.start, s.stop)
		})
	}
}

//////////////
/// Vector ///
//////////////

func TestIteration(t *testing.T) {
	input := inputSlice(0, 10000)
	vec := NewIntVector(input...)
	iter := vec.Iter()
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
	vec := NewIntVector(inputSlice(0, 1000)...)
	slice := vec.Slice(0, 10)
	assertEqual(t, 1000, vec.Len())
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
	slice := NewIntVectorSlice(inputSlice(0, sliceLen)...)
	assertEqual(t, slice.Len(), sliceLen)
	for i := 0; i < sliceLen; i++ {
		assertEqual(t, i, slice.Get(i))
	}
}

func TestSliceSet(t *testing.T) {
	vector := NewIntVector(inputSlice(0, 1000)...)
	slice := vector.Slice(10, 100)
	slice2 := slice.Set(5, 123)

	// Underlying vector and original slice should remain unchanged. New slice updated
	// in the correct position
	assertEqual(t, 15, vector.Get(15))
	assertEqual(t, 15, slice.Get(5))
	assertEqual(t, 123, slice2.Get(5))
}

func TestSliceAppendInTheMiddleOfBackingVector(t *testing.T) {
	vector := NewIntVector(inputSlice(0, 100)...)
	slice := vector.Slice(0, 50)
	slice2 := slice.Append(inputSlice(0, 10)...)

	// New slice
	assertEqual(t, 60, slice2.Len())
	assertEqual(t, 0, slice2.Get(50))
	assertEqual(t, 9, slice2.Get(59))

	// Original slice
	assertEqual(t, 50, slice.Len())

	// Original vector
	assertEqual(t, 100, vector.Len())
	assertEqual(t, 50, vector.Get(50))
	assertEqual(t, 59, vector.Get(59))
}

func TestSliceAppendAtTheEndOfBackingVector(t *testing.T) {
	vector := NewIntVector(inputSlice(0, 100)...)
	slice := vector.Slice(0, 100)
	slice2 := slice.Append(inputSlice(0, 10)...)

	// New slice
	assertEqual(t, 110, slice2.Len())
	assertEqual(t, 0, slice2.Get(100))
	assertEqual(t, 9, slice2.Get(109))

	// Original slice
	assertEqual(t, 100, slice.Len())
}

func TestSliceAppendAtMiddleToEndOfBackingVector(t *testing.T) {
	vector := NewIntVector(inputSlice(0, 100)...)
	slice := vector.Slice(0, 50)
	slice2 := slice.Append(inputSlice(0, 100)...)

	// New slice
	assertEqual(t, 150, slice2.Len())
	assertEqual(t, 0, slice2.Get(50))
	assertEqual(t, 99, slice2.Get(149))

	// Original slice
	assertEqual(t, 50, slice.Len())
}

func TestSliceIteration(t *testing.T) {
	vec := NewIntVector(inputSlice(0, 1000)...)
	iter := vec.Slice(5, 200).Iter()
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
	NewIntVector(inputSlice(0, 10)...).Slice(2, 5).Set(-1, 0)
}

func TestSliceSetOutOfBoundsBeyondEnd(t *testing.T) {
	defer assertPanic(t, "Index out of bounds")
	NewIntVector(inputSlice(0, 10)...).Slice(2, 5).Set(4, 0)
}

func TestSliceGetOutOfBoundsNegative(t *testing.T) {
	defer assertPanic(t, "Index out of bounds")
	NewIntVector(inputSlice(0, 10)...).Slice(2, 5).Get(-1)
}

func TestSliceGetOutOfBoundsBeyondEnd(t *testing.T) {
	defer assertPanic(t, "Index out of bounds")
	NewIntVector(inputSlice(0, 10)...).Slice(2, 5).Get(4)
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
			NewIntVector(inputSlice(0, 10)...).Slice(2, 5).Slice(s.start, s.stop)
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
	vec := NewIntVector(inputSlice(0, size)...)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		iter := vec.Iter()
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
