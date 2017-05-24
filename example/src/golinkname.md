---
title: "go:linkname compiler directive"
date: 2017-02-10
tags: go
---

## What

This is a short post about the go:linkname compiler directive.
The documentation [explains](https://golang.org/cmd/compile/#hdr-Compiler_Directives):

> The //go:linkname directive instructs the compiler to use “importpath.name”
> as the object file symbol name for the variable or function declared as
> “localname” in the source code.  Because this directive can subvert the type
> system and package modularity, it is only enabled in files that have
> imported "unsafe".

## Why

This directive is not used very often in practice.
I advise against using it.
As the doc says it can "subvert the type system and package modularity".
However, it is a good idea to understand what it does when reading the standard library.
It is used in the standard library to access unexported functions of another package.
This quote [explains](https://groups.google.com/d/msg/golang-nuts/JMA9Pk8LkDg/QJAV2vITDwAJ) the motivation behind it:

> It's primarily a hack that lets certain functions live in the runtime 
> package, and access runtime internals, but still pretend that they are 
> unexported functions of some different package.

## Example

Let's look at an example.
Here is the definition of the time.Now function.

```go
func Now() Time {
  sec, nsec, mono := now()
  t := unixTime(sec, nsec)
  t.setMono(int64(mono))
  return t
}
```

It calls the now function. Let's look at its definition.
```go
// Provided by package runtime.
func now() (sec int64, nsec int32, mono uint64)
```

The now function has no body.
This is valid syntax as [described in the spec](https://golang.org/ref/spec#Function_declarations).
It is used when a function is defined elsewhere, e.g., assembly or in another Go package using go:linkname.
In the case of the now function, the function is defined and linked in the runtime package.

```go
//go:linkname time_now time.now
func time_now() (sec int64, nsec int32, mono uint64) {
  sec, nsec = walltime()
  return sec, nsec, uint64(nanotime() - startNano + 1)
}
```

Notice the directive is linking a local function (time_now) to time.now.
The first argument is the local name, and the second argument is the import path of the linked function.
Here is another example:
```go
//go:linkname hellofunc a/b/pkg2.hola
```

In our time.Now example, the logic for getting the current time is performed in the runtime package.
<!--But runtime.time_now is hidden in the time package.-->
Without the go:linkname directive, runtime would have had to export time_now.
The API is designed so that users should fetch the current time using time.Now,
and not a function in runtime.

## Notes

- the directive comment could be in either two packages.
- to compile a package A with a function that is linked and defined in package B,
we need to import package B to allow the compiler/linker find the definition.
- we cannot compile a program with go:linkname using go-build command,
because go-build enables the -complete compiler flag.
With this flag, the compiler does not allow declaration of functions without body.
It expects all functions to be defined in the package or idiomatically exported/imported from another package.

To see more examples of go:linkname see the runtime package or [see this example](https://github.com/siadat/golinkname-test). 
