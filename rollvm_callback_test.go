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

	err = vm.Run("str")
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
