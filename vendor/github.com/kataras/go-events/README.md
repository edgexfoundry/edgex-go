<p align="center">
  <img src="/logo.jpg" height="400">
  <br/>
</p>

[![build status](https://img.shields.io/github/workflow/status/kataras/go-events/CI/dev?style=for-the-badge)](https://github.com/kataras/go-events/actions) [![chat](https://img.shields.io/gitter/room/events/community.svg?color=cc2b5e&logo=gitter&style=for-the-badge)](https://gitter.im/events/community) [![donate](https://img.shields.io/badge/support-Go--Events-blue.svg?style=for-the-badge&logo=paypal)](https://iris-go.com/donate)
<br/>

Simple EventEmmiter for Go Programming Language. Inspired by [Nodejs EventEmitter](https://nodejs.org/api/events.html). For **Generics** support please wait until go1.18 stable is released, until then switch to [dev branch](https://github.com/kataras/go-events/tree/dev) instead.

Overview
------------
`New() EventEmmiter  // New returns a new, empty, EventEmmiter`


```go
// AddListener is an alias for .On(eventName, listener).
AddListener(EventName, ...Listener)
// Emit fires a particular event,
// Synchronously calls each of the listeners registered for the event named
// eventName, in the order they were registered,
// passing the supplied arguments to each.
Emit(EventName, ...interface{})
// EventNames returns an array listing the events for which the emitter has registered listeners.
// The values in the array will be strings.
EventNames() []EventName
// GetMaxListeners returns the max listeners for this emmiter
// see SetMaxListeners
GetMaxListeners() int
// ListenerCount returns the length of all registered listeners to a particular event
ListenerCount(EventName) int
// Listeners returns a copy of the array of listeners for the event named eventName.
Listeners(EventName) []Listener
// On registers a particular listener for an event, func receiver parameter(s) is/are optional
On(EventName, ...Listener)
// Once adds a one time listener function for the event named eventName.
// The next time eventName is triggered, this listener is removed and then invoked.
Once(EventName, ...Listener)
// RemoveAllListeners removes all listeners, or those of the specified eventName.
// Note that it will remove the event itself.
// Returns an indicator if event and listeners were found before the remove.
RemoveAllListeners(EventName) bool
// Clear removes all events and all listeners, restores Events to an empty value
Clear()
// SetMaxListeners obviously this function allows the MaxListeners
// to be decrease or increase. Set to zero for unlimited
SetMaxListeners(int)
// Len returns the length of all registered events
Len() int
```


```go
import "github.com/kataras/go-events"

// initialize a new EventEmmiter to use
e := events.New()

// register an event with name "my_event" and one listener
e.On("my_event", func(payload ...interface{}) {
  message := payload[0].(string)
  print(message) // prints "this is my payload"
})

// fire the 'my_event' event
e.Emit("my_event", "this is my payload")

```

Default/global EventEmmiter
```go

// register an event with name "my_event" and one listener to the global(package level) default EventEmmiter
events.On("my_event", func(payload ...interface{}) {
  message := payload[0].(string)
  print(message) // prints "this is my payload"
})

// fire the 'my_event' event
events.Emit("my_event", "this is my payload")

```

Remove an event

```go

events.On("my_event", func(payload ...interface{}) {
  // first listener...
},func (payload ...interface{}){
  // second listener...
})

println(events.Len()) // prints 1
println(events.ListenerCount("my_event")) // prints 2

// Remove our event, when/if we don't need this or we want to clear all of its listeners
events.RemoveAllListeners("my_event")

println(events.Len()) // prints 0
println(events.ListenerCount("my_event")) // prints 0


```
Installation
------------

The only requirement is the [Go Programming Language](https://golang.org/dl).

```bash
$ go get -u github.com/kataras/go-events
```


FAQ
------------

Explore [these questions](https://github.com/kataras/go-events/issues?go-events=label%3Aquestion) or navigate to the [community chat][Chat].

Versioning
------------

Current: v0.0.3

Read more about Semantic Versioning 2.0.0

 - http://semver.org/
 - https://en.wikipedia.org/wiki/Software_versioning
 - https://wiki.debian.org/UpstreamGuide#Releases_and_Versions

Contributing
------------

If you are interested in contributing to the go-events project, please prepare a PR.

License
------------

This project is licensed under the MIT License.

License can be found [here](LICENSE).
