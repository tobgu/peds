package peds_testing

// Types for testing.
// Defined in a separate file from the tests since they are not picked up
// if defined in a files with a *_test.go name.
type Foo uint
type Bar float64
//go:generate genny -in=../array.go -pkg=peds_testing -out=array_test_gen.go gen "Item=Foo,int"
//go:generate genny -in=../map.go -pkg=peds_testing -out=map_test_gen.go gen "Key=Foo Value=Bar"


//  go generate seems to require a function in the file that contains the generation expression...
func f() {
}
