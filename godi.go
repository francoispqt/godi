package godi

import (
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

// ErrDependencyNotFoundMsg is the error message returned
// if the Maker is not found
var ErrDependencyNotFoundMsg = "Dependency '%v' does not exist"

// ErrDependencyNotFound is a custom error type for errors
// when dependency does not exist
type ErrDependencyNotFound error

// ErrDependencyNotFoundF returns an ErrDependencyNotFound error with the given
// dependency
func ErrDependencyNotFoundF(dep interface{}) ErrDependencyNotFound {
	return ErrDependencyNotFound(
		errors.Errorf(ErrDependencyNotFoundMsg, dep),
	)
}

// IsErrDependencyNotFound returns whether the given error is a
// DependencNotFound error
func IsErrDependencyNotFound(err error) bool {
	_, ok := err.(ErrDependencyNotFound)
	return ok
}

// Injecter is the interface representing the dependency injecter
type Injecter interface {
	Make(interface{}) (interface{}, bool)
	BindSingleton(interface{}, Maker)
	Bind(interface{}, Maker)
}

// Container is the structure representing the dependency injecter
type Container struct {
	valueStore *atomic.Value
	store      *atomic.Value
	mut        *sync.Mutex
}

// Maker is the function returning the instance of a dependency or an error
// it takes a list of arguments passed to the Container.Make method
type Maker func(...interface{}) (interface{}, error)

// New returns a new instance of Container
func New() *Container {
	var valueStore atomic.Value
	var valueM = make(map[interface{}]interface{})
	valueStore.Store(valueM)

	var v atomic.Value
	var m = make(map[interface{}]Maker, 0)
	v.Store(m)

	return &Container{
		valueStore: &valueStore,
		store:      &v,
		mut:        &sync.Mutex{},
	}
}

// Make looks for the Maker function for the key k in the store and calls it with the given args
// it returns an ErrDependencyNotFound error if no Maker exist with the key k, else if the Maker returns a non nil error it will bubble up.
func (di *Container) Make(k interface{}, args ...interface{}) (interface{}, error) {
	var valueStore = di.valueStore.Load().(map[interface{}]interface{})
	if v, ok := valueStore[k]; ok {
		return v, nil
	}

	var v = di.store.Load().(map[interface{}]Maker)

	if v, ok := v[k]; ok {
		return v(args...)
	}
	return nil, ErrDependencyNotFoundF(k)
}

// MustMake looks for the Maker function for the key k in the store and calls it with the given args
// It panics if an error happens.
func (di *Container) MustMake(k interface{}, args ...interface{}) interface{} {
	var valueStore = di.valueStore.Load().(map[interface{}]interface{})
	if v, ok := valueStore[k]; ok {
		return v
	}

	var v = di.store.Load().(map[interface{}]Maker)
	if v, ok := v[k]; ok {
		var r, err = v(args...)
		if err != nil {
			panic(err)
		}
		return r
	}
	panic(ErrDependencyNotFoundF(k))
}

// BindSingleton adds a singleton Maker for the key k to the Container's store.
// It will always return the same instance returned by the first call to the Maker function.
func (di *Container) BindSingleton(k interface{}, f func(args ...interface{}) (interface{}, error)) *Container {
	di.mut.Lock()
	defer di.mut.Unlock()

	// we recreate the maker store
	var v = di.store.Load().(map[interface{}]Maker)
	var nV = make(map[interface{}]Maker, len(v))
	for ok, ov := range v {
		nV[ok] = ov
	}

	// we recreate the value store but we omit the current key
	var valueStore = di.valueStore.Load().(map[interface{}]interface{})
	var nValueStore = make(map[interface{}]interface{}, len(valueStore))
	for ok, ov := range valueStore {
		if ok != k {
			nValueStore[ok] = ov
		}
	}
	di.valueStore.Store(nValueStore)

	var r interface{}
	var err error
	var mut = sync.Mutex{}
	nV[k] = Maker(func(args ...interface{}) (interface{}, error) {
		mut.Lock()
		defer mut.Unlock()

		if r == nil && err == nil {
			r, err = f(args...)
			if err == nil {
				di.mut.Lock()
				// now that it's been ran once, we replace the Maker to return the value
				// to avoid locking anything a get significantly better perfs
				var valueStore = di.valueStore.Load().(map[interface{}]interface{})
				var nValueStore = make(map[interface{}]interface{}, len(valueStore))
				for ok, ov := range valueStore {
					nValueStore[ok] = ov
				}

				nValueStore[k] = r
				di.valueStore.Store(nValueStore)
				di.mut.Unlock()
			}
		}
		return r, err
	})

	di.store.Store(nV)
	return di
}

// Bind adds a Maker for the key k to the Container's store.
func (di *Container) Bind(k interface{}, f func(args ...interface{}) (interface{}, error)) *Container {
	di.mut.Lock()
	defer di.mut.Unlock()

	// we recreate the value store but we omit the current key
	var valueStore = di.valueStore.Load().(map[interface{}]interface{})
	var nValueStore = make(map[interface{}]interface{}, len(valueStore))
	for ok, ov := range valueStore {
		if ok != k {
			nValueStore[ok] = ov
		}
	}
	di.valueStore.Store(nValueStore)

	var v = di.store.Load().(map[interface{}]Maker)
	var nV = make(map[interface{}]Maker, len(v))
	for ok, ov := range v {
		nV[ok] = ov
	}
	nV[k] = Maker(f)
	di.store.Store(nV)
	return di
}
