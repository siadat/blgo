---
title: "How I blog"
date: 2016-11-10
---

I wrote a small Go program called [blgo](https://github.com/siadat/blgo) to renders my posts
using [blackfriday](https://github.com/russross/blackfriday) markdown processor.
I also use [fsnotify](https://github.com/fsnotify/fsnotify) to watch templates and posts for changes.

There are many Javascript syntax highlighters.
I prefer not to use them.
They add a new dependency, increase page size, are slow, and cause the style to flick with a delay.

I customized the blackfriday.BlockCode func to use GoDoc's code renderer.
It makes my Go code blocks look similar to GoDoc, which is pleasing to my eyes.
Here is an example:

```go
// Print hello
fmt.Println("Hello")
```

Isn't that beautiful?


To render HTML files, I execute this on [my blog repository](https://github.com/siadat/siadat.github.io/):

```shell
$ blgo watch src/*.md
```

To preview the posts, I use the
[gofile](https://github.com/siadat/gofile)
command to serve the files locally:

```shell
$ gofile 8080
```

For more details on how I changed the BlockCode funct see my
post on [overriding a struct method](http://localhost:8080/post/curious-gopher-override.html).
