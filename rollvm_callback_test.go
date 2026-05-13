package dicescript

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalValueLoadOverwrite(t *testing.T) {
	vm := NewVM()
	vm.GlobalValueLoadOverwriteFunc = func(name string, curVal *VMValue) *VMValue {
		if curVal == nil {
			return NewIntVal(123)
		}
		return curVal
	}

	err := vm.Run("测试")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(123)))
	}

	err = vm.Run("toStr")
	if assert.NoError(t, err) {
		assert.Equal(t, vm.Ret.TypeId, VMTypeNativeFunction)
	}
}

func TestHookFuncValueLoadOverwrite(t *testing.T) {
	vm := NewVM()
	vm.Config.HookValueLoadPost = func(ctx *Context, name string, curVal *VMValue, doCompute func(v *VMValue) *VMValue, detail *BufferSpan) *VMValue {
		doCompute(curVal)
		if ctx.Error != nil {
			return nil
		}
		return ni(123)
	}

	err := vm.Run("测试")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(123)))
	}
}

func TestHookValueStore(t *testing.T) {
	t.Run("local overwrite", func(t *testing.T) {
		vm := NewVM()
		calls := 0
		vm.Config.HookValueStore = func(ctx *Context, name string, v *VMValue) (*VMValue, bool) {
			calls++
			assert.Same(t, vm, ctx)
			assert.Equal(t, "hp", name)
			assert.True(t, valueEqual(v, ni(1)))
			return ni(2), false
		}

		vm.StoreName("hp", ni(1), true)

		actual, ok := vm.Attrs.Load("hp")
		if assert.True(t, ok) {
			assert.True(t, valueEqual(actual, ni(2)))
		}
		assert.Equal(t, 1, calls)
	})

	t.Run("global overwrite", func(t *testing.T) {
		vm := NewVM()
		vm.globalNames.Store("hp", ni(0))

		var storedName string
		var storedVal *VMValue
		vm.GlobalValueStoreFunc = func(name string, v *VMValue) {
			storedName = name
			storedVal = v
		}
		vm.Config.HookValueStore = func(ctx *Context, name string, v *VMValue) (*VMValue, bool) {
			assert.Same(t, vm, ctx)
			assert.Equal(t, "hp", name)
			assert.True(t, valueEqual(v, ni(1)))
			return ni(3), false
		}

		vm.StoreName("hp", ni(1), true)

		assert.Equal(t, "hp", storedName)
		if assert.NotNil(t, storedVal) {
			assert.True(t, valueEqual(storedVal, ni(3)))
		}
		_, ok := vm.Attrs.Load("hp")
		assert.False(t, ok)
		assert.NoError(t, vm.Error)
	})

	t.Run("solved skips remaining storage", func(t *testing.T) {
		vm := NewVM()
		vm.globalNames.Store("hp", ni(0))

		stored := false
		vm.GlobalValueStoreFunc = func(name string, v *VMValue) {
			stored = true
		}
		vm.Config.HookValueStore = func(ctx *Context, name string, v *VMValue) (*VMValue, bool) {
			assert.Same(t, vm, ctx)
			assert.Equal(t, "hp", name)
			assert.True(t, valueEqual(v, ni(1)))
			return nil, true
		}

		vm.StoreName("hp", ni(1), true)

		assert.False(t, stored)
		_, ok := vm.Attrs.Load("hp")
		assert.False(t, ok)
		assert.NoError(t, vm.Error)
	})
}

func TestCustomDetailSpanRewrite(t *testing.T) {
	vm := NewVM()
	vm.Attrs.Store("x", ni(5))
	vm.Attrs.Store("a", NewComputedVal("4d1"))

	type callInfo struct {
		tag   string
		root  bool
		value string
	}
	var calls []callInfo

	vm.Config.CustomDetailSpanRewriteFunc = func(ctx *Context, defaultDetail string, span BufferSpan, isRoot bool, dataBuffer []byte, parsedOffset int) string {
		calls = append(calls, callInfo{tag: span.Tag, root: isRoot, value: defaultDetail})
		switch span.Tag {
		case "load":
			return "LOAD<" + defaultDetail + ">"
		case "load.computed":
			return "COMPUTED<" + defaultDetail + ">"
		default:
			return defaultDetail
		}
	}

	err := vm.Run("x")
	if !assert.NoError(t, err) {
		return
	}

	detail := vm.GetDetailText()
	assert.Equal(t, "5LOAD<>", detail)

	err = vm.Run("a")
	if !assert.NoError(t, err) {
		return
	}

	detail = vm.GetDetailText()
	assert.Equal(t, "4COMPUTED<[a=4[4d1=1+1+1+1]=4]>", detail)

	var loadSeen, computedSeen bool
	for _, c := range calls {
		switch c.tag {
		case "load":
			loadSeen = true
		case "load.computed":
			computedSeen = true
		}
	}
	assert.True(t, loadSeen)
	assert.True(t, computedSeen)
}
