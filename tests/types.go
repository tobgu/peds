package peds_testing

// Types for testing.
// Defined in a separate file from the tests since they are not picked up
// if defined in a files with a *_test.go name.
type Foo uint
type Bar float64

// NOTE: The awkward quoting below is just to test that white spaces in the type specifications are ignored.
//       If you stay away from using white space the quoting should not be required.
//go:generate peds "-vectors=\"FooVector<Foo>; IntVector<int >\"" "-maps=\"StringIntMap<string, int>;IntStringMap<int,string>\"" "-sets=\"FooSet<Foo>; IntSet< int>\"" -pkg=peds_testing -file=types_gen.go

//  go generate seems to require a function in the file that contains the generation expression...
func f() {
}
