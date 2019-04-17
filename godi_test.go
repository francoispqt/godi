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

	di.BindSingleton("B", Maker(func(args ...interface{}) (interface{}, error) {
		return &A{i: args[0].(int)}, nil
	}))
	var b interface{}
	b, err = di.Make("B", 1)
	assert.Nil(t, err)
	assert.Equal(t, b.(*A).i, 1)

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

	di.BindSingleton("B", Maker(func(args ...interface{}) (interface{}, error) {
		return &A{i: args[0].(int)}, nil
	}))
	var b = di.MustMake("B", 2).(*A)
	assert.Equal(t, 2, b.i)

	var r2 = di.MustMake("A", 1).(*A)
	assert.Equal(t, 1, r2.i)

	var b2 = di.MustMake("B").(*A)
	assert.Equal(t, 2, b2.i)

	di.BindSingleton("A", Maker(func(args ...interface{}) (interface{}, error) {
		return &A{i: args[0].(int)}, nil
	}))

	var r3 = di.MustMake("A", 2).(*A)
	assert.Equal(t, 2, r3.i)
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
			assert.True(t, IsErrDependencyNotFound(err.(error)))
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

	di.BindSingleton("AS", Maker(func(args ...interface{}) (interface{}, error) {
		return &A{i: args[0].(int)}, nil
	}))
	var s interface{}
	s, err = di.Make("AS", 1)
	assert.Nil(t, err)
	assert.True(t, s.(*A).i == 1)

	di.Bind("B", Maker(func(args ...interface{}) (interface{}, error) {
		return &A{i: args[0].(int)}, nil
	}))
	var b interface{}
	b, err = di.Make("B", 1)
	assert.Nil(t, err)
	assert.Equal(t, b.(*A).i, 1)

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

func TestMakeSingletonParallel(t *testing.T) {
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
				di.Make("A", 1)

				mut.Lock()
				defer mut.Unlock()

				if !changed {
					require.Equal(t, 1, ran)
				} else if i == 500 {
					changed = true
					di.BindSingleton("A", Maker(func(args ...interface{}) (interface{}, error) {
						mut.Lock()
						defer mut.Unlock()
						ran++
						return &A{i: args[0].(int)}, nil
					}))
				} else if changed {
					require.Equal(t, 2, ran)
				}
			},
		)
	}
}

func TestMustMakeSingletonParallel(t *testing.T) {
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
				di.Make("A", 1)

				mut.Lock()
				defer mut.Unlock()

				if !changed {
					require.Equal(t, 1, ran)
				} else if i == 500 {
					changed = true
					di.BindSingleton("A", Maker(func(args ...interface{}) (interface{}, error) {
						mut.Lock()
						defer mut.Unlock()
						ran++
						return &A{i: args[0].(int)}, nil
					}))
				} else if changed {
					require.Equal(t, 2, ran)
				}
			},
		)
	}
}
