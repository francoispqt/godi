package godi

import (
	"errors"
	"sync"
)

// ErrDependencyNotFound is the error returned if the Maker is not found
var ErrDependencyNotFound = errors.New("Dependency does not exist")

// Injecter is the interface representing the dependency injecter
type Injecter interface {
	Make(interface{}) (interface{}, bool)
	BindSingleton(interface{}, Maker)
	Bind(interface{}, Maker)
}

// Container is the structure representing the dependency injecter
type Container struct {
	store sync.Map
}

// Maker is the function returning the instance of a dependency or an error
// it takes a list of arguments passed to the Container.Make method
type Maker func(...interface{}) (interface{}, error)

// New returns a new instance of Container
func New() *Container {
	return &Container{
		store: sync.Map{},
	}
}

// Make looks for the Maker function for the key k in the store and calls it with the given args
// if no Maker exist with the key k, it returns an ErrDependencyNotFound error
func (di *Container) Make(k interface{}, args ...interface{}) (interface{}, error) {
	if v, ok := di.store.Load(k); ok {
		return v.(Maker)(args...)
	}
	return nil, ErrDependencyNotFound
}

// MustMake looks for the Maker function for the key k in the store and calls it with the given args
// If an error happens, it panics
func (di *Container) MustMake(k interface{}, args ...interface{}) interface{} {
	if v, ok := di.store.Load(k); ok {
		var r, err = v.(Maker)(args...)
		if err != nil {
			panic(err)
		}
		return r
	}
	panic(ErrDependencyNotFound)
}

// BindSingleton adds a singleton Maker for the key k to the Container's store.
// It will always return the same instance returned by the first call to the Maker function.
func (di *Container) BindSingleton(k interface{}, f func(args ...interface{}) (interface{}, error)) *Container {
	var r interface{}
	var err error
	var mut = sync.Mutex{}
	di.store.Store(k, Maker(func(args ...interface{}) (interface{}, error) {
		mut.Lock()
		if r == nil && err == nil {
			r, err = f(args...)
		}
		mut.Unlock()
		return r, err
	}))
	return di
}

// Bind adds a Maker for the key k to the Container's store.
func (di *Container) Bind(k interface{}, f func(args ...interface{}) (interface{}, error)) *Container {
	di.store.Store(k, Maker(f))
	return di
}
