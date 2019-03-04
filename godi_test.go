package godi

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type A struct {
	i int
}

func TestGoDIBindSingleton(t *testing.T) {
	var di = New()
	di.BindSingleton("A", Maker(func(args ...interface{}) (interface{}, error) {
		return &A{i: args[0].(int)}, nil
	}))
	var r, err = di.Make("A", 0)
	assert.Nil(t, err)
	assert.Equal(t, r.(*A).i, 0)
	r.(*A).i = 1
	var r2 interface{}
	r2, err = di.Make("A", 1)
	assert.Nil(t, err)
	assert.Equal(t, r.(*A), r2.(*A), "both injected instances should be the same")
	assert.Equal(t, 1, r.(*A).i, "a.i should be 1")
}

func TestGoDIBindMust(t *testing.T) {
	var di = New()
	di.BindSingleton("A", Maker(func(args ...interface{}) (interface{}, error) {
		return &A{i: args[0].(int)}, nil
	}))
	var r = di.MustMake("A", 1).(*A)
	assert.Equal(t, 1, r.i)
	var r2 = di.MustMake("A").(*A)
	assert.Equal(t, r2, r)
}

func TestGoDIBindMustErr(t *testing.T) {
	var di = New()
	var e = errors.New("")
	defer func() {
		if err := recover(); err != nil {
			assert.Equal(t, e, err)
			return
		}
		assert.False(t, true, "Did not panic")
	}()
	di.BindSingleton("A", Maker(func(args ...interface{}) (interface{}, error) {
		return nil, e
	}))
	var _ = di.MustMake("A", 1).(*A)
	assert.False(t, true, "should have panicked")
}

func TestGoDIBindMustErrNotExist(t *testing.T) {
	var di = New()
	defer func() {
		if err := recover(); err != nil {
			assert.Equal(t, ErrDependencyNotFound, err)
			return
		}
		assert.False(t, true, "Did not panic")
	}()
	var _ = di.MustMake("A", 1).(*A)
	assert.False(t, true, "should have panicked")
}

func TestGoDIBind(t *testing.T) {
	var di = New()
	di.Bind("A", Maker(func(args ...interface{}) (interface{}, error) {
		return &A{i: args[0].(int)}, nil
	}))
	var r, err = di.Make("A", 0)
	assert.Nil(t, err)
	var r2 interface{}
	r2, err = di.Make("A", 1)
	assert.Nil(t, err)
	assert.NotEqual(t, r.(*A), r2.(*A), "both injected instances should not be the same")
}

func TestErrNotExist(t *testing.T) {
	var di = New()
	var _, err = di.Make("A", 0)
	assert.NotNil(t, err)
}

func TestSingletonParallel(t *testing.T) {
	var di = New()
	var mut = &sync.Mutex{}
	var ran int
	var changed bool
	di.BindSingleton("A", Maker(func(args ...interface{}) (interface{}, error) {
		mut.Lock()
		defer mut.Unlock()
		ran++
		return &A{i: args[0].(int)}, nil
	}))

	for i := 0; i < 1000; i++ {
		i := i
		t.Run(
			fmt.Sprintf("%d", i),
			func(t *testing.T) {
				t.Parallel()
				mut.Lock()
				defer mut.Unlock()
				if i > 500 && !changed {
					di.BindSingleton("A", Maker(func(args ...interface{}) (interface{}, error) {
						mut.Lock()
						defer mut.Unlock()
						ran++
						return &A{i: args[0].(int)}, nil
					}))
					di.Make("A", i)
					changed = true
					require.True(t, ran > 1)
				} else if i > 500 && changed {
					di.Make("A", i)
					require.True(t, ran > 1)
				} else {
					di.Make("A", i)
					require.Equal(t, 1, ran)
				}
			},
		)
	}
}
