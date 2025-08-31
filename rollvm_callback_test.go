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
