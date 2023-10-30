package dicescript

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAssignSpace(t *testing.T) {
	// stmtAssign
	vm := NewVM()
	err := vm.Run("a = 1")
	//fmt.Println("xxx", vm.GetAsmText())
	if assert.NoError(t, err) {
		assert.True(t, vmValueEqual(vm, "a", ni(1)))
	}

	vm = NewVM()
	err = vm.Run("a=1")
	if assert.NoError(t, err) {
		assert.True(t, vmValueEqual(vm, "a", ni(1)))
	}

	vm = NewVM()
	err = vm.Run("a  =   1")
	if assert.NoError(t, err) {
		assert.True(t, vmValueEqual(vm, "a", ni(1)))
	}
}

func TestAssignComputedSpace(t *testing.T) {
	// stmtAssign
	// 允许 = 前后空格
	vm := NewVM()
	err := vm.Run("&a=1; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("&a = 1 ; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}
}

func TestAssignThisWithSpace(t *testing.T) {
	// stmtAssign
	// 允许语法: a   .  b  = 1
	vm := NewVM()
	err := vm.Run("this  .  xx = 1")
	if assert.NoError(t, err) {
		assert.Equal(t, vm.RestInput, "")
	}
}

func TestAssignComputedWithSpace(t *testing.T) {
	vm := NewVM()
	err := vm.Run("&a = 1")
	if assert.NoError(t, err) {
		err = vm.Run("&a.x = 2")
		if assert.NoError(t, err) {
			err = vm.Run("&a.x")
			if assert.NoError(t, err) {
				assert.True(t, valueEqual(vm.Ret, ni(2)))
			}
		}
	}
}

func TestAssignArrayWithSpace(t *testing.T) {
	// stmtAssign
	// 允许语法: [1,2,3]  [0] = 3  // 注: 暂未支持
	vm := NewVM()
	err := vm.Run("a = [1,2,3];  a[0] = 3")
	if assert.NoError(t, err) {
		assert.True(t, vm.RestInput == "")
	}
}
