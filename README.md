# GoDI
GoDI is a small dependency injection package. 
It provides simple API to create a dependency injection container and add non-singleton or singleton dependency makers to the container. 

It is very useful for testing when you need to inject mocks for example. 

## Get started

Get GoDI
```shell
go get github.com/francoispqt/godi
```

Quick example:
```go
package main

import (
	"io"

	"github.com/francoispqt/godi"
)

type SomeWriter struct{}

func (s *SomeWriter) Write(b []byte) (int, error) {
	return 0, nil
}

var myDependencyKey = struct{}{}

func main() {
    // create the DI container
    var DI = godi.New()
    // add a resolver
	DI.Bind(myDependencyKey, func(args ...interface{}) (interface{}, error) {
		return &SomeWriter{}, nil
    })
    // get an instance
    // will panic if dependency does not exist or resolver returns an error
	_ = DI.MustMake(myDependencyKey).(io.Writer)
}
```

## New 
New returns a new empty DI container.
```go
var DI = godi.New()
```

## *Container.Bind
Bind method adds a dependency maker to the container. The args passed to the Maker function are the args passed to the `Make` or `MustMake` method of the container.
```go
DI.Bind(someKey, func(args ...interface{}) (interface{}, error) {
		return &SomeDependency{args[0].(string)}, nil
})
var r = DI.MustMake(someKey, "foo").(*SomeDependency)
```

## *Container.BindSingleton
BindSingleton method adds a dependency maker to the container. 
The maker function is called only once, further calls will return the same results. 
The args passed to the Maker function are the args passed to the `Make` or `MustMake` method of the container.
```go
DI.BindSingleton(someKey, func(args ...interface{}) (interface{}, error) {
		return &SomeDependency{args[0].(string)}, nil
})
var r = DI.MustMake(someKey, "foo").(*SomeDependency)
var r2 = DI.MustMake(someKey).(*SomeDependency)

fmt.Print(r == r2) // true
```

## *Container.Make
Make method calls the dependency maker for the given key. It returns the result and an error. The error will be `ErrDependencyNotFound` if the dependency maker is not found in the container, else if the Maker returns an error it will be bubbled up. Arguments passed to Make after the key are passed to the found Maker function.
```go
DI.Bind(someKey, func(args ...interface{}) (interface{}, error) {
        return &SomeDependency{
            k: args[0].(string), // "foo"
        }, nil
})
var r, err = DI.Make(someKey, "foo").(*SomeDependency)
```


## *Container.MustMake
Make method calls the dependency maker for the given key. It behaves like Make but panics if an error is encountered. Arguments passed to Make after the key are passed to the found Maker function.
```go
DI.Bind(someKey, func(args ...interface{}) (interface{}, error) {
        return &SomeDependency{
            k: args[0].(string), // "foo"
        }, nil
})
var r = DI.Make(someKey, "foo").(*SomeDependency)
```