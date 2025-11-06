package dicescript

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// 测试未覆盖的语法分支

func TestNullCoalescing(t *testing.T) {
	// ?? 运算符
	vm := NewVM()
	err := vm.Run("null ?? 5")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(5)))
	}

	vm = NewVM()
	err = vm.Run("10 ?? 5")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(10)))
	}

	vm = NewVM()
	err = vm.Run("a = null; a ?? 'default'")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("default")))
	}
}

func TestExponentiation(t *testing.T) {
	// ** 运算符
	vm := NewVM()
	err := vm.Run("2 ** 3")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(8)))
	}

	vm = NewVM()
	err = vm.Run("5 ** 2")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(25)))
	}

	vm = NewVM()
	err = vm.Run("2.0 ** 3.0")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, nf(8.0)))
	}
}

func TestUnaryPositive(t *testing.T) {
	// 一元正号 +
	vm := NewVM()
	err := vm.Run("+5")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(5)))
	}

	vm = NewVM()
	err = vm.Run("+3.14")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, nf(3.14)))
	}

	vm = NewVM()
	err = vm.Run("a = 10; +a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(10)))
	}
}

func TestReturnWithoutValue(t *testing.T) {
	// return 不带值
	vm := NewVM()
	err := vm.Run("func test() { return }")
	assert.NoError(t, err)

	vm = NewVM()
	err = vm.Run("func test() { return }; test()")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewNullVal()))
	}
}

func TestStringEscapes(t *testing.T) {
	// 测试字符串转义
	vm := NewVM()
	err := vm.Run("'\\'test\\''")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("'test'")))
	}

	vm = NewVM()
	err = vm.Run("\"\\\"test\\\"\"")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("\"test\"")))
	}

	vm = NewVM()
	err = vm.Run("`test \\{ test \\}`")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("test { test }")))
	}

	vm = NewVM()
	err = vm.Run("'test \\\\ slash'")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("test \\ slash")))
	}
}

func TestWoDDiceVariants(t *testing.T) {
	// WoD 骰子的 q 参数（阈值）
	vm := NewVM()
	vm.Config.EnableDiceWoD = true
	err := vm.Run("5a10q8")
	assert.NoError(t, err)
	// 只要能运行不报错即可，具体结果是随机的
}

func TestFStringEdgeCases(t *testing.T) {
	// f-string 的特殊情况
	vm := NewVM()
	err := vm.Run("a = 5; `value is {a}`")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("value is 5")))
	}
}
