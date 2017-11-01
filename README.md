Statically type safe persistent/immutable/functional data structures for Go.

Inspired by Clojures data structures and the work done in
[Pyrsistent](https://www.github.com/tobgu/pyrsistent) for Python.

This is an experiment in how close to generics that code generation can take
you. There's currently a vector, a slice, a map and a set implemented.

## What's a persistent data structure?
Despite their name persistent data structures usually don't refer to
data structures stored on disk. Instead they are immutable data
structures which are copy-on-write. Eg. whenever you mutate a persistent
data structure a new one is created that contains the data in the
original plus the mutation while the original is left untouched.

To make this reasonably efficient, "structural sharing" is used between
the data structures. Here's a good introduction to how this is done in
Clojure:
http://hypirion.com/musings/understanding-persistent-vector-pt-1

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

## Examples

There are a couple of generated example collections in
[examples/collections.go](https://github.com/tobgu/peds/blob/master/examples/collections.go).

The `go:generate` command used can be found in [examples/types.go](https://github.com/tobgu/peds/blob/master/examples/types.go).

This illustrates the core usage pattern:
```
//go:generate peds -vectors=IntVector<int> -pkg=my_collections -file=collections/my_collections_gen.go

// Create a new vector
v := my_collections.NewIntVector(1, 2, 3)

// Create a copy of v with the first element set to 55, v is left untouched.
v2 := v.Set(0, 55)
```

## Godoc

#### Generic types
https://godoc.org/github.com/tobgu/peds/internal/generic_types

#### Generated examples
https://godoc.org/github.com/tobgu/peds/examples

## Experiences

There's an [experience report](https://github.com/tobgu/peds/blob/master/experience_report.md) based on the implementation of this library.

## Caveats
* Even though the data structures are immutable by most means it is not
  possible to use them as keys in hash maps for example. This is because
  they internally make use of slices, which are not comparable in Go.

## Possible improvements
* Investigate implementing the Map as a CHAMP tree.
* Introspection of the contained types possible to
  refine the hash functions?
* Get rid of Python dependency for developing peds (not needed to build or use peds).

## Regenerate templates and run tests
`make test`

## Want to contribute?
Great! Write an issue and let the discussions begin! Then file a PR!
