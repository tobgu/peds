# Experience report
This sums up some of the key experiences and thoughts from this and
other projects that I would like to share as an experience report going
into the work of Go 2.
The different subjects described below are listed in the order of
importance to me. Most important on top, less important towards the
bottom.

## Background
I don't have excessive experience in Go but have used it for this as
well as a couple of other projects since about one year. There are
probably a ton of stuff that could be improved in this project, some
of which I'm aware (link to TODO) and some which I've not yet come to
realize. I'm happy to take suggestions and and PRs!

As for compiler construction I don't have much experience beyond the
basics so it may be the case that I make overly simplistic assumptions
in my comments below. If so, please bear with me and let me know.

### Why Go?
These are some of the reasons I choose to engage with the Go language
in the first place:
* Performance, natively compiled with a memory layout which lets you
  write fast, mechanically sympathetic, programs with low memory overhead.
* Quick compilation.
* Tooling, batteries included. Great profiling and diagnostics tooling
  comes built in.
* Garbage collected, makes it possible to move faster and focus on the
  business problem.
* A great, and growing, community.
* A large number of companies heavily invested in Go.

I specifically did not come to Go because of it being a "simple"
language (as written and read). The lack of expressiveness and
possibility for abstraction in the language was actually one of the
few things that made me reluctant to use Go in the first place.

I realise that having a small language is an enabler for some of the
compelling reasons of Go that I've listed above such as quick compile
times though.

## Generics
The collections in peds are of course a prime example of a data
structures that would benefit tremendously from generics in the
language.

### Writing the code
When starting the project the choice was between code generation to
make the collections statically type safe and more efficient and
interface{} to make life easier for the implementer (me) but throw
statical type safety and performance over board.
To me it was like choosing between two evils but in the end I opted
for the former because of personal preferences. It would also give
me the chance to explore the pains of writing a code generator.

I knew that I wanted peds to be a standalone binary for generating the
code as opposed to "generic" code that would then make use of a third
party library such as genny to generate the code.
This was mainly a decision based on usability since relying on a
third party library would be one more thing that the developer
using peds would have to learn. Also I wanted to have full control
over the inputs to the code generation, something that would not
be possible with any of the existing tools I looked at.

With that in mind the generic/templated code would have to reside within
the peds binary. To me that meant the template code would have to be
template strings (I choose go text templates since they are part of
the stdlib). But I (the library author) don't want to write all the
code within strings. That would mean no help from the IDE/editor or
a lot of other tooling that I'm used to as a Go programmer.
To solve this the generic code is written as plain Go code with special
("magic") names for the generic types and functions. This code is then
processed by a small Python script that turns it into a number of
template strings with suitable template variables that are then used
by the code generator. Peds then uses these template strings,
substitutes the template variables with data from the command line
options and outputs the generated code to a file and package of the
users choice.

All this is of course painful for the library author. Beyond the
initial setup I must keep track of which those magic names are
where substitution happens.
For the end user the experience should hopefully be fairly smooth
though. It's just one command that can be used together with
go generate.

### Access control
Another problem with code generation compared to built in generic
support is that there is no control of where the generated code ends
up and that, in the case of containers the generated code must be
aware of the types of the contained data.

In peds this is currently solved by allowing the user to specify
target file and package together with any imports that may be
needed within the generated code to access the contained types.
While it works it feels dirty. All functions, types and data
defined in the generated code is potentially at risk of being
accessed by code existing in the same package. For peds this weakens
the promise of immutability somewhat. For the promise to be held
only publicly accessible data and functions must be used.

### Requirements on the generic type
For the Map there is a requirement on the key, it must be hashable.
With code generation these kind of requirements
cannot be properly defined. Furthermore, even though all basic types
in Go already supports hashing (for use with the native map) this
functionality is not accessible outside the runtime package (I believe
this has it's reasons and that it has been done to reduce the number
of bugs related to code that does not fulfill the contract between
hash and equality, that said this is an example of decisions in Go
that sometimes make me feel like a third class citizens in the
language).

Hashing as implemented in peds right now is very basic. If the key type
is a recognized as a basic Go type (int, float, string, ...) a custom,
tailor made, hash function is applied to it. If not then the type is
printed to a string and that string is in turn hashed. Improvements to
the implementation are certainly possible here.

Ideally though I would like to delegate the hashing to the type that
is used as key. That would allow it to be implemented efficiently based
on knowledge that is simply not available to peds.

For this to happen with a generic-like some kind of constraint on the
generic type would have to be expressed.

### Concluding thoughts
To support the peds use case in Go two things are needed from a
generics like suggestion:

1. The ability to instantiate the container for arbitrary type(s)
2. The ability to restrict those types to such that are hashable and
   comparable for the map and set types.

The below is not a suggestion for how to finally implement generics but
rather an example that would fulfill the needs by peds.

```
peds.Map[K:Hashable, V]
```

Hashable would state that the type of the key needs to be hashable.
To me it seems like the current Go interfaces could serve as a perfect
base for these kind of specifications. While not strictly necessary
since compilation of the generic type would fail if the `Hash()`
function was missing, it provides a much clearer contract at no
cost for the client since Go interfaces are implemented implicitly
(which is great!).

```
type Hashable interface {
    Hash() uint64
}
```

This would allow the compiler to generate efficient code avoiding
the runtime indirection usually associated with interface functions
and potentially inline the code in `Hash()` of the key type.

I want to finish with my view on generics implementation. I'm
pro the templating approach where specific code is emitted for every
instance of the generic type/function and no boxing takes place.
Rust is an example of a language that uses this approach.

I believe boxing/type erasure (as done in Java for example) would undo
many possible use cases where generics could be used to write
"tight"/"performance critical"/"mechanically sympathetic" algorithms or
data structures.
In Java this may be OK since there's a JIT that can do all sorts of
optimizations in runtime while in Go that is not the case.
While I'm all for the mantra of avoiding premature optimization on the
application level this becomes less and less true the further down the
stack you go since you have less and less context.
With the language being in the bottom levels I think performance is key
to avoid restricting what can be done on higher levels.

With a templating approach the binary will be fatter for sure since more
code is generated. The Go binaries are already fairly fat though and to
me, adding a couple of extra % to that does not really matter.
The negative impact on the instruction cache because of code bloat
that some claim could be a problem should be tested in reality.
Furthermore, if the compiler/linker only generated code for the functions
used by the client code as opposed to the full API of the generic
type my gut feeling is that the code bloat should not be that dramatic
since in reality many client applications often only use a small subset
of the provided functionality. This would of course make the compiler
more complicated and potentially increase compile times but I trust
the compiler team to solve this. ;-)

## Immutability
With the strong focus on concurrency in Go I was surprised that
immutability is not focused on more as a means to write safe
concurrent code.

While the value semantics of basic types and structs alleviates
some of the problems it does not apply to slices and maps which are
reference types.

A comment in the code that a variable should be treated as immutable
is not a satisfactory alternative to true language support. I would,
for example, never trust code in an evolving application to adhere
to such a comment if I was debugging an issue (and I'm not alone
it seems, https://github.com/lukechampine/freeze).

The creation of peds is a way to show case immutable alternatives.
While I realise it may not fit into the core/stdlib of the language
I believe that better options at the language level to specify that
data cannot be changed is core.

In peds the vector is implemented as a tree of nodes which are 32
elements wide. Each such node is in itself immutable (as in that's
the intention) but I cannot describe this in the language.
This makes bugs in my implementation more likely because the compiler
cannot check that I adhere to my own rules. It also reduces the
possibility to communicate this information to anyone else or my future
self.

## Iteration

### The need for unification
The use of the `range` keyword to iterate over slices and maps is fine
I think. The only problem is that it's not possible to use with any
other data structures.
That's why people come up with all sorts of ideas for how to do it. The
principal ideas are described nicely here:
https://blog.kowalczyk.info/article/1Bkr/3-ways-to-iterate-in-go.html

In peds I finally settled on a callback based solution which closely
resembles that of the `sync.Map` that came with Go 1.9 for familiarity.

The implementation is similar to the below example:

```
type T []int

func (t T) Range(f func(int) bool) {
	for _, x := range t {
		ok := f(x)
		if !ok {
			return
		}
	}
}
```

I think it's a reasonable solution but it requires the user to define a
callback function, anonymous or named. It's also slower than using
`range` directly since it involves a function call, that the compiler
does not know to inline, for each step in the iteration.

What I would like to see is a unified way of iterating over any data
structure. This would allow users to learn how to iterate in Go once.

### for-loop says nothing
I applause that there is only one loop construct in Go (I don't want
a while loop, a do-while loop, ...) but I think the for-loop is over
used.

First of all I don't agree that it's simple:
* There are at least four places where you could make an error in a for
  loop:
   - the init statement
   - the condition expression
   - the post statement
   - the loop body
* Every time I see a for loop I have to look at all parts of it to
  determine what it actually does.

I think that higher level construct such as map, reduce, filter, etc.
are actually good things since they immediately tell me something about
the intention of the iteration. They also reduce the number of places
where mistakes can be made, often you just have to concern yourself with
the loop body.

## Sum types/tagged unions
In peds the vector is implemented as a wide tree consisting of
32-element wide nodes. Internal nodes contain references to lower
level nodes while the leafs contain the actual values contained
in the vector. A node can hence be either of two, and only two,
different types.
I could not come up with a good way of describing this. In the end I
decided to implement the nodes as `{}interface`, a type that in the
code is referred to as `commonNode`.

When accessing the nodes they are first type asserted back to their
original type. Which type to assert to is determined by helper
variables which keep track of where in the tree you are.

When inserting new nodes they are first casted to `commonNode`. There
is nothing in the type system preventing me from inserting anything
into a node. Something that would cause other parts of the code to
fail at runtime. Presumably always far away from the root cause
since the error would occur in the read path while the root cause was
introduced in the write.

A sum type containing either an internal node or a leaf node would
help in this situation since it would restrict the possible errors.
It would also document the tree structure nicely for later maintainers
something that cannot be said about the commonNode since `interface{}`
says nothing.

As an alternative to using the empty interface I experimented with
a poor mans sum type along the way. Like this:

```
type Node struct {
    nodes []Node
    data []GenericType
}
```

The idea was that only one of the fields in the struct would ever be
set depending on the type of node. The other field would be `nil`.
The pro of this type compared to the empty interface is that it's now
clearly stated which types make up a valid node.
The con is that it gives the illusion that it's OK to set both fields.
It also occupies more space than the empty interface and performs
slightly worse.
Because of the drawbacks described I decided to discard this experiment.

## Set
I can't believe there's not a built in generic set type in Go 1! Sets
are great for so many things and I really hate re-implementing them
as a `map[type]struct{}` for every type I need them in Go.

This is more of an experience in general with the Go language than
specific to the implementation of peds. These previous experiences went
into designing the set type in peds though (which closely mimics the
set API in python).

## Round up
I don't think any of the subjects listed above are novel, they have all
been discussed before in various forms. I do believe there's a reason
that they have been discussed before though. They would help to solve
a class of problems a lot more elegantly than what is possible in
Go today.

I also think that adding some of them (mainly generics) would make the
language more attractive to a large number of programmers who today
dismiss Go because of lack of expressiveness in the type system.
Adding generics would also open up for the creation of a ton of
libraries for data structures and algorithms that nobody bothers
to implement in Go today. This would make the language more useful in
areas and domains where it is not used that much today. While as an
application developer the need for generics may not arise that often,
I think it would allow a lot of applications to stand on the shoulders
of giants (great libraries).

Finally I would like to stress the importance (to me) of focusing on
performance and mechanical sympathy moving forward to avoid that Go
ends up with the scripting languages where anything performance critical
has to be implemented as an extension in a differnt language.
I really want to be able to use Go for everything, all the way from
the tightest loops and up!

Go is a great language with really nice runtime characteristics,
tooling and a nice deployment story. Lets build on that! :-)
