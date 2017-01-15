package peds_testing

import "testing"

func assertEqual(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected: %s, actual: %s", expected, actual)
	}
}

func TestPropertiesOfEmptyNewArray(t *testing.T) {
	arr := NewIntArray()
	assertEqual(t, 0, arr.Len())
}

func TestPropertiesOfNonEmptyNewArray(t *testing.T) {
	arr := NewIntArray(1, 2, 3)
	assertEqual(t, 3, arr.Len())
	assertEqual(t, 2, int(arr.Get(1)))
}

func TestSetDoesNotModifyOriginalArray(t *testing.T) {
	arr := NewIntArray(1, 2, 3)
	arr2 := arr.Set(1, 5)
	assertEqual(t, 2, arr.Get(1))
	assertEqual(t, 5, arr2.Get(1))
}

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
