package peds_testing

import (
	"testing"
	"fmt"
)

func assertEqual(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected: %s, actual: %s", expected, actual)
	}
}

func inputArray(size int) []int {
	result := make([]int, 0, size)
	for i := 0; i < size; i++ {
		result = append(result, i)
	}

	return result
}


const maxSize = 2000

func TestPropertiesOfNewArray(t *testing.T) {
	for l := 0; l < maxSize; l++ {
		t.Run(fmt.Sprintf("NewArray %d", l), func(t *testing.T) {
			arr :=  NewIntArray(inputArray(l)...)
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
			arr :=  NewIntArray(inputArray(l)...)
			for i := 0; i < l; i++ {
				newArr := arr.Set(i, -i)
				assertEqual(t, -i, newArr.Get(i))
				assertEqual(t, i, arr.Get(i))
			}
		})
	}
}


// TODO:
// - Set on large arrays
// - Append on large arrays
// - Array with no tail
// - Error cases...

func TestAppendDoesNotModifyOriginalArray(t *testing.T) {
	arr := NewIntArray(1, 2, 3)
	arr2 := arr.Append(5)
	assertEqual(t, 3, arr.Len())
	assertEqual(t, 4, arr2.Len())
	assertEqual(t, 5, arr2.Get(3))
}

func TestSlicing(t *testing.T) {
	arr := NewIntArray(1, 2, 3)
	arr2 := arr.Slice(0, 2)
	assertEqual(t, 3, arr.Len())
	assertEqual(t, 2, arr2.Len())
	assertEqual(t, 1, arr2.Get(0))
	assertEqual(t, 2, arr2.Get(1))
}
