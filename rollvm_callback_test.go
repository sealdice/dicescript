package dicescript

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
