---
title: "override a struct method"
date: 2016-11-01
tags: go
---

```notebox
**NOTE** There is no object-oriented Java-like overriding in Go.
In this post I explain what I call overriding struct methods in Go.
```

There are times when we want to make a small modification to a complex struct.
Instead of changing the original struct, we could create a new struct and embed the existing implementation inside it.

Let's look at an example. Let's say we are using the Person struct in a package defined like this:

```go
package person

type Person struct{}

func (Person) Greet() string {
	return "hello"
}

func (Person) Bye() string {
	return "bye"
}
```

This is how we use this package:

```go
func main() {
	p := person.Person{}
	fmt.Println(p.Greet(), p.Bye())
}
```
```output
hello bye
```

Okay. What if we need the Greet func to return something else?
We don't want to modify the struct directly because it is in another package we don't maintain.
And we don't want to have to rewrite all the other methods.

To do that, we create a new struct like this:

```go
// FrenchPerson embeds person.Person
type FrenchPerson struct{
	person.Person
}

// Greet returns "bonjour"
func (FrenchPerson) Greet() string {
	return "bonjour"
}
```

Finally, let's test this new method:

```go
func main() {
	p1 := Person{}
	fmt.Println(p1.Greet(), p1.Bye())

	p2 := FrenchPerson{}
	fmt.Println(p2.Greet(), p2.Bye())
}
```
```output
hello bye
bonjour bye
```
We are calling Greet and Bye as before.
However, this time the overriden Greet func is called.

## Real world example

I find this particularly useful for overriding methods in an external dependency.

Take for example the [blackfriday](https://github.com/russross/blackfriday) markdown rendering package.
It defines a Renderer interface with 31 functions:

```go
type Renderer interface {
	// block-level callbacks
	BlockCode(out *bytes.Buffer, text []byte, lang string)
	BlockQuote(out *bytes.Buffer, text []byte)
	BlockHtml(out *bytes.Buffer, text []byte)
	Header(out *bytes.Buffer, text func() bool, level int, id string)
	HRule(out *bytes.Buffer)
	List(out *bytes.Buffer, text func() bool, flags int)
	ListItem(out *bytes.Buffer, text []byte, flags int)
	Paragraph(out *bytes.Buffer, text func() bool)
	Table(out *bytes.Buffer, header []byte, body []byte, columnData []int)
	TableRow(out *bytes.Buffer, text []byte)
	TableHeaderCell(out *bytes.Buffer, text []byte, flags int)
	TableCell(out *bytes.Buffer, text []byte, flags int)
	Footnotes(out *bytes.Buffer, text func() bool)
	FootnoteItem(out *bytes.Buffer, name, text []byte, flags int)
	TitleBlock(out *bytes.Buffer, text []byte)

	// Span-level callbacks
	AutoLink(out *bytes.Buffer, link []byte, kind int)
	CodeSpan(out *bytes.Buffer, text []byte)
	DoubleEmphasis(out *bytes.Buffer, text []byte)
	Emphasis(out *bytes.Buffer, text []byte)
	Image(out *bytes.Buffer, link []byte, title []byte, alt []byte)
	LineBreak(out *bytes.Buffer)
	Link(out *bytes.Buffer, link []byte, title []byte, content []byte)
	RawHtmlTag(out *bytes.Buffer, tag []byte)
	TripleEmphasis(out *bytes.Buffer, text []byte)
	StrikeThrough(out *bytes.Buffer, text []byte)
	FootnoteRef(out *bytes.Buffer, ref []byte, id int)

	// Low-level callbacks
	Entity(out *bytes.Buffer, entity []byte)
	NormalText(out *bytes.Buffer, text []byte)

	// Header and footer
	DocumentHeader(out *bytes.Buffer)
	DocumentFooter(out *bytes.Buffer)

	GetFlags() int
}
```

For this blog, all I want to do is to override the BlockCode func and keep everything else unchanged.
I used the same technique as the above example. This is what I did:

```go
type Renderer struct {
	*blackfriday.Html
}

func (options *Renderer) BlockCode(out *bytes.Buffer,
	text []byte, lang string) {
	// ... new implementation ...
}
```

The code for my blog engine is [here](https://github.com/siadat/blgo).


