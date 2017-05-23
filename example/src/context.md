---
title: "Context API explained"
date: 2017-01-21
tags: go
---

### Introduction

Let's start with a simple problem.
We have this program that performs an action every second.

```go
func Perform() {
    for {
        SomeFunction()
        time.Sleep(time.Second)
    }
}
```

And we call it like this

```go
go Perform()
```

The goal is to cancel the Perform function either explicitly or when a deadline is exceeded.
Context package was initially designed to implement exactly what we need; request cancelation and deadline.
Take a look at the context.Context interface:

```go
type Context interface {
    Done() <-chan struct{}
    Err() error
    Deadline() (deadline time.Time, ok bool)
    Value(key interface{}) interface{}
}
```

Notice that all the methods perform a query and get information:

- `ctx.Done()` return cancelation channel, which is used to check if context is canceled.
- `ctx.Err()` return cancelation reason (DeadlineExceeded or Canceled).
- `ctx.Deadline()` return deadline, if set.
- `ctx.Value(key)` return value for key.

This API raises a few questions.
Why does ctx.Done() return a channel? Why not a bool value?
Why is there no cancel method? How do we set a deadline?
What is ctx.Value(key) doing here?
To understand this API,
it is useful to know that it is designed to satisfy the following two requirements:

#### 1. Cancelation should be advisory

<!--
Starting a goroutine is easy. Simply insert "go" before a function call.
However, stopping a running goroutine is not as easy.
-->

A caller is not aware of the internals of the function it is invoking.
It should not interrupt or panic the callee.
It is the responsibility of every function to return on its own.

Instead of forcing a function to stop, the caller should *inform* it that its work is no longer needed.
Caller sends the information about cancelation and let the function decide how to deal with it.
For example, a function could clean up and return early
when it is informed that its work is no longer needed.

#### 2. Cancelation should be transitive

When canceling a function,
we need to also cancel all functions that are running on its behalf.
This means that the cancelation information
should be broadcast from caller down to all of its child functions.

### Create a context

The simplest way to create a context is using context.Background():

```go
ctx := context.Background()
```

context.Background() returns an empty context.
For cancelation to be advisory and transitive,
we should give each function the cancelation information as its first argument.
We change our program from

```go
go Perform()
```

to

```go
ctx := context.Background()
go Perform(ctx)
```

### Set a deadline

An empty context is useless.
We need to set a deadline or be able to cancel it.
However, the context.Context interface only defines query methods.
We are not able to modify its deadline.

The reason we cannot modify a context is that we want to prevent the Perform function to be able to modify or cancel the request.
The direction of the flow of information in context is strictly from parent to child.
For example, when a user closes a tab in their browser (parent), all the functions running behalf of that tab (child) should be canceled.

Therefore, we derive a new context with its deadline updated:

```go
ctx, cancel := context.WithDeadline(parentContext, time)
// or
ctx, cancel := context.WithTimeout(parentContext, duration)
```

Notice that cancel is returned as a separate value.
If ctx had a cancel method, child functions would have been able to cancel it.
Again, the API stricts the direction of the cancelation to only go down from parent to child.
In the special case where we need the child function to cancel the request, we will have to pass the cancel function as a separate argument explicitly.

Continuing with our example we will have

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
go Perform(ctx)
```

We can use cancel() to signal Perform that we don't need its work anymore.
In the next section we will see how Perform handles this signal.

### Check if context is canceled

The cancelation event should be broadcast down to all called functions.
Go channels have a property that make them suitable for this purpose;
receiving from a closed channel returns a zero value immediately.
This means that multiple functions could watch a channel until it is closed.
When it is closed they know that it was canceled.

The Done method returns a read-only channel that is closed on cancelation.
Here's a simple example for checking if the context is canceled.

```go
func Perform(ctx context.Context) {
    for {
        SomeFunction()

        select {
        case <-ctx.Done():
            // ctx is canceled
            return
        default:
            // ctx is not canceled, continue immediately
        }
    }
}
```

Notice that the select statement does not block.
It is because it has a default statement.
This causes the for loop to execute SomeFunction immediately.
We need to sleep for 1 second between each iteration.

```go
func Perform(ctx context.Context) {
    for {
        SomeFunction()

        select {
        case <-ctx.Done():
            // ctx is canceled
            return
        case <-time.After(time.Second):
            // wait for 1 second
        }
    }
}
```

When context is canceled, we find out the cause by calling ctx.Err().

```go
func Perform(ctx context.Context) error {
    for {
        SomeFunction()

        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(time.Second):
            // wait for 1 second
        }
    }
    return nil
}
```

This function has only two possible values:
context.DeadlineExceeded and context.Canceled.
ctx.Err() is expected to be called only *after* ctx.Done() is closed.
The result of ctx.Err() before ctx is canceled is not defined by the API.

If SomeFunction takes a long time, we could let it know about the cancelation as well.
We do that by passing ctx to it as its first argument.

```go
func Perform(ctx context.Context) error {
    for {
        SomeFunction(ctx)

        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(time.Second):
            // wait for 1 second
        }
    }
    return nil
}
```

### What is context.TODO()?
Similar to context.Background, another way of creating a context is
```go
ctx := context.TODO()
```
TODO function returns an empty context as well.
TODO is used while refactoring functions to support context.
We use it when a parent context is not available in that function yet.
All TODO contexts should eventually be replaced with another context.

### What is ctx.WithValue?
The most common usage of context is with handling cancelation in requests.
To achieve that, context is usually carried out during the lifetime of a request (e.g. as the first argument to all functions).

Another useful information that should be carried out during the life of a request is values such as user session and login information.
The context package makes it easy to store those values in context instances as well.
Because they share the same call path as the cancelation information.
To set a value we derive a context using context.WithValue

```go
ctx := context.WithValue(parentContext, key, value)
```

To retrieve this value from ctx or any context that is derived from it use

```go
value := ctx.Value(key)
```

### Other resources

I highly recommend the following two resources
for anyone who wants to understand the context package.

- [Cancellation, Context, and Plumbing](https://vimeo.com/115309491) (video) by Sameer Ajmani
- [Pipelines and cancellation](https://blog.golang.org/pipelines) (blog post) by Sameer Ajmani

### Conclusion

I hope this post helped the reader understand the context API a little better.
[Comment](https://www.reddit.com/r/golang/comments/5p7qnb/context_api_explained/),
email (siadat at gmail),
or [tweet me](https://twitter.com/sinasiadat) your suggestions and corrections.
