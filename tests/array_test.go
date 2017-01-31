package peds_testing

import (
	"testing"
	"fmt"
	"runtime"
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

func inputArray(start, size int) []int {
	result := make([]int, 0, size)
	for i := start; i < start + size; i++ {
		result = append(result, i)
	}

	return result
}

var testSizes = []int{0,1,20,32,33,50,500,32*32,32*32+1,10000,32*32*32,32*32*32+1}

/////////////
/// Array ///
/////////////

func TestPropertiesOfNewArray(t *testing.T) {
	for _, l := range testSizes {
		t.Run(fmt.Sprintf("NewArray %d", l), func(t *testing.T) {
			arr :=  NewIntArray(inputArray(0, l)...)
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
			arr :=  NewIntArray(inputArray(0, l)...)
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
		arr :=  NewIntArray(inputArray(0, l)...)
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

/////////////////
/// Iteration ///
/////////////////
func TestIteration(t *testing.T) {
	input := inputArray(0, 10000)
	arr := NewIntArray(input...)
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
	arr := NewIntArray(inputArray(0, 1000)...)
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
	slice :=  NewIntSlice(inputArray(0, sliceLen)...)
	assertEqual(t, slice.Len(), sliceLen)
	for i := 0; i < sliceLen; i++ {
		assertEqual(t, i, slice.Get(i))
	}
}

func TestSliceSet(t *testing.T) {
	array := NewIntArray(inputArray(0, 1000)...)
	slice := array.Slice(10, 100)
	slice2 := slice.Set(5, 123)

	// Underlying array and original slice should remain unchanged. New slice updated
	// in the correct position
	assertEqual(t, 15, array.Get(15))
	assertEqual(t, 15, slice.Get(5))
	assertEqual(t, 123, slice2.Get(5))
}

func TestSliceAppendInTheMiddleOfBackingArray(t *testing.T) {
	array := NewIntArray(inputArray(0, 100)...)
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
	array := NewIntArray(inputArray(0, 100)...)
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
	array := NewIntArray(inputArray(0, 100)...)
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
	arr := NewIntArray(inputArray(0, 1000)...)
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


// TODO:
// - Test error cases, index out of bounds, etc.
// - Document public methods

//////////////////////
///// Benchmarks /////
//////////////////////

// Used to avoid that the compiler optimizes the code away
var result int

func runIteration(b *testing.B, size int) {
	arr := NewIntArray(inputArray(0, size)...)
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
