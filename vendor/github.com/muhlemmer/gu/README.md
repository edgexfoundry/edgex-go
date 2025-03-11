# Generic Utilities

[![Go Reference](https://pkg.go.dev/badge/github.com/muhlemmer/gu.svg)](https://pkg.go.dev/github.com/muhlemmer/gu)
[![Go](https://github.com/muhlemmer/gu/actions/workflows/go.yml/badge.svg)](https://github.com/muhlemmer/gu/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/muhlemmer/gu/branch/main/graph/badge.svg?token=I7UCR4XRV1)](https://codecov.io/gh/muhlemmer/gu)

GU is a collection of Generic Utility functions, using Type Parameters featured in Go 1.18 and later. I often found myself writing boilerplate code for slices, maps, poitners etc. Since 1.18 I started using generics in some of my repositories and found that some functions often are the same between projects. The repository is a collection of those (utiltity) functions.

Although the functions are pretty basic and *almost* don't justify putting them in a package, I share this code under the [unlicense](https://unlicense.org/), with the purpose:

- Make my own life easier when reusing boiler plate code;
- So that others can easily use these utilities;
- People who want to learn more about generics in Go can read the code;

## Features

There is no logic in which order I'm adding features. Ussualy when I see repetative code that can be generalized, it is dropped in here. Which means that there might be other utilities that seem to be missing. [Contributions](contributing) are welcome.

Below features link to pkg.go.dev documentation where examples can be found.

### Pointers

- [`Ptr`](https://pkg.go.dev/github.com/muhlemmer/gu#Ptr) allows for getting a direct pointer. For example from fuction returns: `t := gu.Ptr(time.Unix())` where `t := &time.Unix()` is illigal Go code.
- [`Value`](https://pkg.go.dev/github.com/muhlemmer/gu#Value) safely returns a value through a pointer. When the pointer is `nil`, the zero value is returned without panic.

### Slices

- [Transform](https://pkg.go.dev/github.com/muhlemmer/gu#InterfaceSlice) a slice  of any type into a slice of interface (`[]T` to `[]interface{}`).
- [Assert](https://pkg.go.dev/github.com/muhlemmer/gu#AssertInterfaces) a slice of interfaces to a slice of any type (`[]interface{}` to `[]T`).
- [Transform](https://pkg.go.dev/github.com/muhlemmer/gu#Transform) slices of similar types that implement the [Transformer](https://pkg.go.dev/github.com/muhlemmer/gu#Transformer) interface.

### Maps

- [Copy a map](https://pkg.go.dev/github.com/muhlemmer/gu#MapCopy).
- [Copy a map](https://pkg.go.dev/github.com/muhlemmer/gu#MapCopyKeys) by certain keys only.
- Check if two maps are [equal](https://pkg.go.dev/github.com/muhlemmer/gu#MapEqual).

## Contributing

Open for Pull Requests.

- In case of a bugfix, please clearly describe the issue and how to reproduce. Preferably a unit test that exposes the behaviour.
- A new feature should be properly documented (godoc), added to REAMDE.md and fully unit tested. If the function seems to be abstract an example needs to be provided in the testfile (`ExampleXxx()` format)
- All code needs to be `go fmt`ed

Please note the [unlicense](LICENSE): you forfait all copyright when contributing to this repository.
