package peds_testing

import "testing"

func TestSetContains(t *testing.T) {
	s := NewFooSet()
	assertEqual(t, 0, s.Len())

	s2 := s.Add(1)
	assertEqualBool(t, true, s2.Contains(1))
	assertEqualBool(t, false, s2.Contains(2))

}
