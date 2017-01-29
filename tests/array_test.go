package peds_testing

import (
	"testing"
	"fmt"
)

func assertEqual(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected: %d, actual: %d", expected, actual)
	}
}

func inputArray(start, size int) []int {
	result := make([]int, 0, size)
	for i := start; i < start + size; i++ {
		result = append(result, i)
	}

	return result
}


const maxSize = 2000

func TestPropertiesOfNewArray(t *testing.T) {
	for l := 0; l < maxSize; l++ {
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
	for l := 0; l < maxSize; l++ {
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


// TODO:
// - Error cases...

func TestAppend(t *testing.T) {
	for l := 0; l < 1000; l++ {
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

func TestSlicing(t *testing.T) {
	arr := NewIntArray(1, 2, 3)
	arr2 := arr.Slice(0, 2)
	assertEqual(t, 3, arr.Len())
	assertEqual(t, 2, arr2.Len())
	assertEqual(t, 1, arr2.Get(0))
	assertEqual(t, 2, arr2.Get(1))
}
