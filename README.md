Type safe persistent/immutable data structures for Go.

An experiment in how close to generics that code generation can take you.

Right now there's a vector, the plan is to also add a map and set.

Inspired by Clojures data structures and the work done in Pyrsistent for python.

Go feedback
-----------
* Generics would have made this a more pleasant experience both
  for the implementor of the library and the end user.
* Operator functions or similar would have made it possible to
  closer align it with the syntax of the built in types but
  is probably just syntactic sugar that would make the language
  trickier to read and understand overall.
* Go really needs a set built in to the language (if to align
  with the current slice and map) or as part of the standard
  library if generics is implemented into the language.
* There should be stronger support for immutability. Both in the
  languages by providing something like this library for immutable
  collections and in the runtime to make frequent allocation and
  garbage collection of short lived small objects super fast.

Run tests
---------
make test