package peds_testing

import "testing"

func TestLenOfNewMap(t *testing.T) {
	m := NewStringIntMap()
	assertEqual(t, 0, m.Len())

	m2 := NewStringIntMap(StringIntMapItem{Key: "a", Value: 1})
	assertEqual(t, 1, m2.Len())

	m3 := NewStringIntMap(StringIntMapItem{Key: "a", Value: 1}, StringIntMapItem{Key: "b", Value: 2})
	assertEqual(t, 2, m3.Len())
}

func TestLoadAndStore(t *testing.T) {
	m := NewStringIntMap()

	m2 := m.Store("a", 1)
	assertEqual(t, 0, m.Len())
	assertEqual(t, 1, m2.Len())

	v, ok := m.Load("a")
	assertEqual(t, 0, v)
	assertEqualBool(t, false, ok)

	v, ok = m2.Load("a")
	assertEqual(t, 1, v)
	assertEqualBool(t, true, ok)
}

func TestLoadAndStoreIntKey(t *testing.T) {
	m := NewIntStringMap()

	m2 := m.Store(1, "")
	v, _ := m.Load(2)
	assertEqualString(t, "", v)

	v, _ = m2.Load(1)
	assertEqualString(t, "", v)
}

/* TODO:- Delete
        - Benchmarks insert and access
        - Constructor from native map?
        - Improve parsing of specs to allow white spaces etc.
        - Dynamic sizing of backing vector depending on size of the map (which thresholds?)
        - More tests, store and load from larger structures
        - ToNativeMap() function (and ToNativeSlice for vectors?)
        - Custom imports?
        - Non comparable types cannot be used as keys (should be detected during compilation)
   	    - Custom hash function?
*/
