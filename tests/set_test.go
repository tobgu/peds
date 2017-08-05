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

func TestSetIsSuperset(t *testing.T) {
	t.Run("Empty sets are supersets of empty sets", func(t *testing.T) {
		assertEqualBool(t, true, NewFooSet().IsSuperset(NewFooSet()))
	})

	t.Run("Empty sets are not supsets of non empty sets", func(t *testing.T) {
		assertEqualBool(t, false, NewFooSet().IsSuperset(NewFooSet(1, 2, 3)))
	})

	t.Run("Equal non-empty sets are supersets of each other", func(t *testing.T) {
		assertEqualBool(t, true, NewFooSet(1, 2).IsSuperset(NewFooSet(1, 2)))
	})

	t.Run("Strict superset", func(t *testing.T) {
		assertEqualBool(t, true, NewFooSet(1, 2, 3).IsSuperset(NewFooSet(1, 2)))
	})

	t.Run("Overlapping but not superset", func(t *testing.T) {
		assertEqualBool(t, false, NewFooSet(1, 2).IsSuperset(NewFooSet(2, 3)))
	})

	t.Run("Non overlapping", func(t *testing.T) {
		assertEqualBool(t, false, NewFooSet(1, 2).IsSuperset(NewFooSet(3, 4)))
	})
}

func assertSetsEqual(s1, s2 *FooSet) bool {
	return s1.Equals(s2)
}

func TestSetUnion(t *testing.T) {
	emptySet := NewFooSet()

	t.Run("Empty sets union empty set is an empty set", func(t *testing.T) {
		assertEqualBool(t, true, assertSetsEqual(emptySet, emptySet.Union(emptySet)))
	})

	t.Run("Empty sets union non empty set is non empty set", func(t *testing.T) {
		assertEqualBool(t, true, assertSetsEqual(NewFooSet(1, 2, 3), emptySet.Union(NewFooSet(1, 2, 3))))
	})

	t.Run("Non empty sets union non empty contains all elements from both sets", func(t *testing.T) {
		assertEqualBool(t, true, assertSetsEqual(NewFooSet(1, 2, 3), NewFooSet(1, 2).Union(NewFooSet(2, 3))))
	})

}

func TestSetDifference(t *testing.T) {
	emptySet := NewFooSet()
	nonEmptySet := NewFooSet(1, 2, 3)

	t.Run("Difference between empty sets are empty sets", func(t *testing.T) {
		assertEqualBool(t, true, assertSetsEqual(emptySet, emptySet.Difference(emptySet)))
	})

	t.Run("Difference between non empty set and empty set is the non empty set", func(t *testing.T) {
		assertEqualBool(t, true, assertSetsEqual(nonEmptySet, nonEmptySet.Difference(emptySet)))
	})

	t.Run("Difference between empty set and non empty set is the empty set", func(t *testing.T) {
		assertEqualBool(t, true, assertSetsEqual(emptySet, emptySet.Difference(nonEmptySet)))
	})

	t.Run("Difference results in all elements part of first set but not second", func(t *testing.T) {
		assertEqualBool(t, true, assertSetsEqual(NewFooSet(1), NewFooSet(1, 2).Difference(NewFooSet(2, 3))))
	})
}

func TestSetSymmetricDifference(t *testing.T) {
	emptySet := NewFooSet()
	nonEmptySet := NewFooSet(1, 2, 3)

	t.Run("Symmetric difference between empty sets are empty sets", func(t *testing.T) {
		assertEqualBool(t, true, assertSetsEqual(emptySet, emptySet.SymmetricDifference(emptySet)))
	})

	t.Run("Symmetric difference between non empty set and empty set is the non empty set", func(t *testing.T) {
		assertEqualBool(t, true, assertSetsEqual(nonEmptySet, nonEmptySet.SymmetricDifference(emptySet)))
	})

	t.Run("Symmetric difference between empty set and non empty set is the non empty set", func(t *testing.T) {
		assertEqualBool(t, true, assertSetsEqual(nonEmptySet, emptySet.SymmetricDifference(nonEmptySet)))
	})

	t.Run("Symmetric difference is all elements part of first or second set but not both", func(t *testing.T) {
		assertEqualBool(t, true, assertSetsEqual(NewFooSet(1, 3), NewFooSet(1, 2).SymmetricDifference(NewFooSet(2, 3))))
	})
}

func TestSetIntersection(t *testing.T) {
	emptySet := NewFooSet()
	nonEmptySet := NewFooSet(1, 2, 3)

	t.Run("Intersection between empty sets are empty sets", func(t *testing.T) {
		assertEqualBool(t, true, assertSetsEqual(emptySet, emptySet.Intersection(emptySet)))
	})

	t.Run("Intersection between non empty set and empty set is the empty set", func(t *testing.T) {
		assertEqualBool(t, true, assertSetsEqual(emptySet, nonEmptySet.Intersection(emptySet)))
	})

	t.Run("Intersection results in all elements part of first and second set", func(t *testing.T) {
		assertEqualBool(t, true, assertSetsEqual(NewFooSet(2), NewFooSet(1, 2).Intersection(NewFooSet(2, 3))))
	})
}

func TestSetToNativeSlice(t *testing.T) {
	set := NewIntSet(1, 2, 3)
	theSlice := set.ToNativeSlice()
	assertEqual(t, 3, len(theSlice))
	assertEqual(t, 1, theSlice[0])
	assertEqual(t, 2, theSlice[1])
	assertEqual(t, 3, theSlice[2])
}
