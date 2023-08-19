# The Rye Programming Language

Rye is a functional programming language for those of us who don't use Haskell, who don't read whitepapers on category theory, and who don't really know what a monad is.

#### Functional programming for the rest of us
Rye encourages and facilitates the basic principles of functional programming:

- *Syntactic sugar for short functions*
```
square := _ * _
abs := _ if _ > 0 else -_
```

- *Immutability by default*
```
x := 42
x = "something else"    // Error!
```

- *map and filter as keywords*
```
numbers := 1..5                     // [1,2,3,4]
evens := numbers where _ % 2 == 0   // [2, 4]
doubled := numbers map _ * 2        // [2,4,6,8]
```

- *`then` keyword (pipe operator)*
```
vowels := {"a", "e", "i", "o", "u"}
s := "functional programming"
    then split(_, "")
    where !(_ in vowels)
    then join(_, "")            // "fnctnl prgrmmng"
```

Though functional programming is encouraged, Rye doesn't prevent imperative programming:
```
var i := 0
while i < 10 {
    print(i)
    i += 1
}
```

The syntax, with a few notable exceptions, should be familiar to anyone with experience in commonly used languages, especially Python and, to a lesser extent, JavaScript and Go. 

#### Rye is <u>not</u> production ready
While this interpreter can run some non-trivial Rye programs either as `.ry` files or in the REPL, it is merely a proof-of-concept and is far from production ready.

Some planned but unimplemented features include:
- a type checker and more robust type system
- `match` expression
- asynchronous programming
- list and set comprehensions
- tuples and multiple returns

This tree-walk interpreter is a hobby project that was hacked together over a series of weekends and, as a result, has some rough edges. It crashes sometimes, does zero optimization, and has no useful parsing errors. 

## Getting started
Compiling the interpreter requires Go (Golang) 1.18 or greater.

Once you've cloned the repo, in the root directory just run:
``` 
go build .
```

To open the REPL, just run:
```
./rye
```
 (for Mac/Linux, Windows may be different)

To run a file, create a file in the same directory named "hello.ry" and paste the following:
```
print("Hello, world!")
```

And run it like this:
```
./rye hello.ry
```

Once you've got the interpreter compiled, feel free to explore the `/examples` directory!

## Language Reference

#### Primitive types
- `Integer`
- `Float`
- `Bool`
- `String`
- `Result`
- `Null`

#### The `Result` type
The `Result` type is inspired by the [Icon programming language](https://en.wikipedia.org/wiki/Icon_(programming_language)). 

It has two values, `success` and `fail`, and is meant as a more semantic alternative to values representing the failure of some process, like JavaScript's `Array.find` returning `undefined` or `Array.findIndex` returning `-1`.

It also can catch "soft" errors that shouldn't crash a program. 

Some examples:
- out-of-bounds errors
```
fruits := ["apple", "banana"]
fruits[0]   // "apple"
fruits[4]   // fail
```
- accessing non-existent object fields
```
user := {
    name: "John",
    age: 32
}

user.age            // 32
user.address.city   // fail
```
- `map` and `where` applied to something other than a collection
```
[-2, 5, 0] where _ > 0      // [5]
-250 where _ > 0            // fail

"not a list" then uppercase // "NOT A LIST"
"not a list" map uppercase  // fail
```
- functions used only for side effects
```
x := print("hello")
x == success            // true
```
- postfix conditionals without an `else`
```
a := 7
b := 4
"less than" if a < b else "greater"     // "greater"
"equal" if a == b                       // fail
```
- failed type conversions
```
Int("1.0")          // 1
Int("two")          // fail
```
- type errors
```
4 > "three"         // fail
```
- division by zero

The postfix `?` operator converts any value to a `Result` (helpful for distinguishing falsy values like `false`, `null`, `""`, and `0` from `fail`), and the `|` operator can provide a fallback value to an expression that may fail.
```
stuff := [-1, 2.5, "apple", true, 4]
isNumber := Float(_)? or Int(_)?
numbers := stuff where isNumber         // [-1, 2.5, 4]

inverse := 1 / _ | 0
inverse(2)              // 0.5
inverse(0)              // 0
```


In the yet-to-be-implemented type system, `?` will represent the union of a type and `fail`, so addition (`+`) will have the type signature of `(Float, Float) -> Float`, but division (`/`) will have the signature of `(Float, Float) -> Float?` since division is an operation that can fail when the second argument is `0`.

#### Operators

Math: `^` (exponentiation), `/`, `*`, `+`, `-`

Cardinality (length): `#`

Logic: `and`, `or`, `!`

Result: `|`, `?`

Conditional: `if`, `unless`, `else`

#### Declaration and assignment

The `:=` operator declares a new (immutable) variable, lexically scoped. In order to create a mutable variable, add `var` before the variable name.
```
x := 42
x = 43                  // error!

var y := "cookies"
y = cake                // no error
```

#### Functions

Functions are first-class values and are created with the `=>` operator.
```
max := (a, b) => a if a > b else b
```

Longer functions may have implicit returns.
```
max := (xs) => {
    var mx := xs[0]
    
    for x <- xs {
        if x > mx: mx = x
    }
    
    mx
}
```

Rye allows closures.
```
createAdder := a => b => a + b
add100 := createAdder(100)
print(add100(7))        // 107

createToggle := () => {
    var state := false
    
    return () => {
        state = !state
        return state
    }
}

myToggle := createToggle()

print(myToggle())       // true
print(myToggle())       // false
print(myToggle())       // true
```

Lists and objects may be destructured in a function's parameters.
```
foo := ([a, b]) => a + b
head([10, 20, 50])         // 30

bob := {
    name: "Bob", 
    favoriteColor: "blue"
}
sayHello := ({name}) => 
    print("Hello, " + name + "!")
sayHello(bob)               // Hello, Bob!
```

#### Collection types
- `List`
```
[1, 2, 3]
```
- `Set`
```
{"apple", "banana", "cherry"}
```
- `Object`
```
{
    foo: "bar",
    baz: true
}
```

`List` and `Set` can be used with `map` and `where`, as well as the cardinality (length) operator `#` (inspired by [Lua](https://www.lua.org/manual/5.4/manual.html#3.4.7)), and the `in` keyword. 
```
fruits := {"apple", "banana", "cherry"}
"apple" in fruits       // true
#fruits                 // 3

someNumbers := [12, 37, 9, -5]
#someNumbers            // 4
```

Items can be added an removed from a set with the `add` and `remove` utils.
```
add({"foo", "bar"}, "baz")      // {"foo", "bar", "baz"}
remove({"foo", "bar"}, "bar")   // {"foo"}
```

Utils also exist for `union`, `intersection`, and `difference`.
```
odds := Set(..10 where _ % 2 == 1)
multiplesOf3 := Set(..10 where _ % 3 == 0)

union(odds, multiplesOf3)           // {6, 5, 7, 9, 1, 3, 0}
intersection(odds, multiplesOf3)    // {3, 9}
difference(odds, multiplesOf3)      // {1, 5, 7}
```

The `Set` and `List` constructors are idempotent.

Ranges are lists created with the `..` operator.
```
xs := 5..10         // [5, 6, 7, 8, 9]
ys := ..4           // [0, 1, 2, 3]
```

Lists can be accessed with brackets, and objects can be accessed with brackets or `.`. Negative indices count backwards from the end of a list.
```
list := [3, 5, 7]
list[2]             // 7
list[-1]            // 7

obj := {a: 100}
obj.a               // 100
obj["a"]            // 100
```
The `..` operator can also be used to access slices of a list.
```
ns := 1..10
ns[3..7]        // [4, 5, 6, 7]
ns[5..]         // [6, 7, 8, 9]
```


#### Underscore functions (`_`)
Undescore functions are a shorthand for defining unary functions consisting of a single expression. The following are equivalent:
```
double := x => x * 2
double := _ * 2
```

Aside from declarations consisting of a single expression, underscore functions can only appear to the right of the `map`, `where`, and `then` keywords. The following are equivalent:
```
square := _ * _
..5 map square          // [0, 1, 4, 9, 16]
..5 map _ * _           // [0, 1, 4, 9, 16]
..5 map x => x * x      // [0, 1, 4, 9, 16]
```

Underscore functions are inspired by [a similar function shorthand in Scala](https://docs.scala-lang.org/scala3/book/fun-anonymous-functions.html).

#### Compound expressions with `map`, `where`, and `then`
With the `map` and `filter` higher-order functions being such a core part of functional programming (and also one of the most accessible for those who are new to FP), Rye elevates them to keywords. 

In combination with underscore functions, the result is a syntax that's easily readable even for those unfamiliar with thinking in terms of higher-order functions.
```
multiplesOf7 := ..100 where _ % 7 == 0

evenOdd := ..100 map "even" if _ % 2 == 0 else "odd"

mailingList := users where _.subscribed

emails := mailingList map _.email
```

The `then` keyword is Rye's equivalent of the pipe operator (`|>`), as seen in [Elixir](https://hexdocs.pm/elixir/Kernel.html#%7C%3E/2) and others. It is a readable way to compose functions, applying the result of the expression on the left-hand side to the function on the right-hand side. `then` expressions can be chained with more `then`s or with `map` or `where` to combine a series of operations into a single expression.
```
reverseWords := str => 
    split(str, " ")
        then reverse
        then join(_, " ")    
    
// equivalent to:
//  reverseWords := str => join(reverse(split(str, " ")), " ")
    
reverseWords("The quick brown fox")     // "fox brown quick The"
```


#### Control flow

The bodies of statements with `if`, `unless`, `else`, `while`, `until`, and `for` can begin with a `:` for one-line expressions, or can be wrapped in brackets.


- `if`/`unless` statement
```
if x > y: print("greater")

unless x < 0:
    print("positive")
    
if y == x {
    print("equal")
}
```
- `if`/`unless` expression
```
print("greater") if x > y
print("positive") unless x < 0
print("equal" if x == y else "not equal")
```
- `while`/`until` loops
```
var i := 0
until i == 10: 
    i += 1
    
var j := 1
while j < 50 {
    print(j)
    j *= 2
}
```
- `for` loops
```
for i <- [1,3,5,7]: print(i)

for i <- 1..20 {
    print(i)
}
```

#### The `index` keyword
The `index` keyword is a convenient way to use both the items and the index when iterating. It can be used in the body of a `for` statement or on the right-hand side of a `map` or `where` expression.
```
groceries := ["eggs", "bacon", "milk"]
for g <- groceries:
    print(index+1, g)

// 1 eggs
// 2 bacon
// 3 milk

sales := [342, 541, 492, 387, 421]
sales 
    map "decreasing" if _ > sales[index-1] else "increasing"
    then print
// ["increasing", "decreasing", "increasing", "increasing", "decreasing"]
```


    

