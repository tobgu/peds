package peds_testing

import "testing"

func TestPropertiesOfNewMap(t *testing.T) {
	m := NewMyMap()
	assertEqual(t, 0, m.Len())

	m2 := NewMyMap(MyMapItem{Key: "a", Value: 1})
	assertEqual(t, 1, m2.Len())

	m3 := NewMyMap(MyMapItem{Key: "a", Value: 1}, MyMapItem{Key: "b", Value: 2})
	assertEqual(t, 2, m3.Len())
}

func TestLoadAndStore(t *testing.T) {
	m := NewMyMap()

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
