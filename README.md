Statically type safe persistent/immutable/functional data structures for Go.

An experiment in how close to generics that code generation can take
you.

There's a vector, a slice, a map and a set.

Inspired by Clojures data structures and the work done in
[Pyrsistent](https://www.github.com/tobgu/pyrsistent) for Python.

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
* Get rid of Python requirement.

Regenerate templates and run tests
----------------------------------
`make test`
