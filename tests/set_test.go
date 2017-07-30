package peds_testing

import "testing"

func TestSetAdd(t *testing.T) {
	s := NewFooSet()
	assertEqual(t, 0, s.Len())

	s2 := s.Add(1)
	assertEqualBool(t, false, s.Contains(1))
	assertEqualBool(t, true, s2.Contains(1))
	assertEqualBool(t, false, s2.Contains(2))

}

func TestSetDelete(t *testing.T) {
	s := NewFooSet(1, 2, 3)
	assertEqual(t, 3, s.Len())
	assertEqualBool(t, true, s.Contains(1))

	s2 := s.Delete(1)
	assertEqual(t, 2, s2.Len())
	assertEqualBool(t, false, s2.Contains(1))

	assertEqualBool(t, true, s.Contains(1))
}

func TestSetIsSubset(t *testing.T) {
	t.Run("Empty sets are subsets of empty sets", func(t *testing.T) {
		assertEqualBool(t, true, NewFooSet().IsSubset(NewFooSet()))
	})

	t.Run("Empty sets are subsets of non empty sets", func(t *testing.T) {
		assertEqualBool(t, true, NewFooSet().IsSubset(NewFooSet(1, 2, 3)))
	})

	t.Run("Equal non-empty sets are subsets of each other", func(t *testing.T) {
		assertEqualBool(t, true, NewFooSet(1, 2).IsSubset(NewFooSet(1, 2)))
	})

	t.Run("Strict subset", func(t *testing.T) {
		assertEqualBool(t, true, NewFooSet(1, 2).IsSubset(NewFooSet(1, 2, 3)))
	})

	t.Run("Overlapping but not subset", func(t *testing.T) {
		assertEqualBool(t, false, NewFooSet(1, 2).IsSubset(NewFooSet(2, 3)))
	})

	t.Run("Non overlapping", func(t *testing.T) {
		assertEqualBool(t, false, NewFooSet(1, 2).IsSubset(NewFooSet(3, 4)))
	})
}
