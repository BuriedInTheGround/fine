# fine

<p align="center"><img alt="This is fine." width="500" src="https://user-images.githubusercontent.com/26801023/145432387-284bc704-eefc-4212-a383-5d2d2335be29.png"></p>

[![Go Reference](https://pkg.go.dev/badge/interrato.dev/fine.svg)](https://pkg.go.dev/interrato.dev/fine)
[![Go Report Card](https://goreportcard.com/badge/interrato.dev/fine)](https://goreportcard.com/report/interrato.dev/fine)

A Finite State Machine [Go](https://go.dev/) library, kept simple.

**Note: the implementation is currently ongoing. The public API interface may
change.**

## Installation

```bash
go get interrato.dev/fine
```

## Concepts

### States

A *state* is a description of the status of the depicted system. Every *state*
has type `string`.

### Transitions

A *transition* is an association between an *event* and an *action*. An *event*
has always type `string`.

Note: a *transition* does not necessarily imply a change of the system state.

### Actions

An *action* is the result of an *event* happening on the depicted system. An
*action* can have one of the following types.
- `string`
- `func() string`
- `func(args ...interface{}) string`
- `func()`
- `func(args ...interface{})`

When an action has one of the first three types, it causes a change of the
system state.

#### Lifecycle actions

A *lifecycle action* is a special kind of *action* that runs automatically in
specific instants of the system lifecycle. *Lifecycle actions* differ in which
types they can have from normal ones. The following are the valid types for
*lifecycle actions*.
- `func()`
- `func(this *fine.FSM)`
- `func(metadata fine.Metadata)`
- `func(this *fine.FSM, metadata fine.Metadata)`

Two *lifecycle actions* are available, characterized by the `@enter` and
`@exit` *events*. These *actions* run when the system enters a new state and
when the system leaves a state, respectively.

Important note: when using the `this *fine.FSM` parameter, the code **must**
run inside a goroutine, or a deadlock would otherwise happen.

##### Metadata

The `fine.Metadata` type is simply a struct which contains the following
information:
- the `From` field with type `string`: the previous *state* from which the
transition to a different state started;
- the `To` field with type `string`: the new *state* where the transition
will end;
- the `Event` field with type `string`: the name of the *action* that caused
the state change;
- the `Args` field with type `[]interface{}`: the parameters that were passed
to the *action*.

## Code example

The following example shows the usage of the library in a simple situation. The
considered system is a simple turnstile with a state diagram like [this
one](https://commons.wikimedia.org/wiki/File:Turnstile_state_machine_colored.svg).
```go
package main

import (
    "fmt"

    "interrato.dev/fine"
)

func main() {
    turnstile := fine.Machine("locked", fine.States{
        "locked": {
            "pay":  "unlocked",
            "push": nil,
        },
        "unlocked": {
            "pay":  nil,
            "push": "locked",
        },
    })

    unsubscribe := turnstile.Subscribe(func(state string) {
        fmt.Printf("The turnstile is now %s.\n", state)
    })

    turnstile.Do("pay")
    turnstile.Do("push")

    unsubscribe()
}
```

The output of the previous code is as follows.
```bash
$ go run ./examples/turnstile
The turnstile is now locked.
The turnstile is now unlocked.
The turnstile is now locked.
```
