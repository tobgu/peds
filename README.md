Statically type safe persistent/immutable data structures for Go.

An experiment in how close to generics that code generation can take
you.

There's a vector, a slice, a map and a set.

Inspired by Clojures data structures and the work done in Pyrsistent for python.

## Installation
`go get github.com/tobgu/peds/cmd/peds`

## Usage
```
Generate statically type safe code for persistent data structures.

USAGE
peds

FLAGS        EXAMPLE
  -file      path/to/file.go
  -imports   import1;import2
  -maps      Map1<int,string>;Map2<float,int>
  -pkg       package_name
  -sets      Set1<int>
  -vectors   Vec1<int>
```

## Godoc
TODO

## Experiences

There's a separate experience report based on the discoveries made when
implementing this library in `experience_report.md`.

## Caveats
* Even though the data structures are immutable by most means it is not
  possible to use them as keys in hash maps for example. This is because
  they internally make use of slices, which are not comparable in Go.

## Possible improvements
* Investigate implementing the Map as a CHAMP tree.
* Introspection of the contained types possible to
  refine the hash functions?
* Get rid of Python requirement.

Run tests
---------
make test
