// Package examples contains a couple of examples of generated peds collections
package examples

// Person is a custom example type that represents a person
type Person struct {
	name    string
	ssn     string
	address string
}

//go:generate peds -maps="PersonBySsn<string,Person>" -sets="Persons<Person>" -vectors="IntVector<int>" -file=collections.go -pkg=examples

// Required for go generate it seems
func f() {
}