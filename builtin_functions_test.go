package dicescript

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNativeFunctionCall(t *testing.T) {
	vm := NewVM()
	err := vm.Run("ceil(1.2)")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("ceil('')")
	assert.Error(t, err)
}

func TestNativeFunctionFloat(t *testing.T) {
	vm := NewVM()
	assert.True(t, valueEqual(funcCeil(vm, nil, []*VMValue{nf(1.1)}), ni(2)))
	assert.True(t, valueEqual(funcCeil(vm, nil, []*VMValue{nf(1.0)}), ni(1)))

	assert.True(t, valueEqual(funcRound(vm, nil, []*VMValue{nf(1.6)}), ni(2)))
	assert.True(t, valueEqual(funcRound(vm, nil, []*VMValue{ni(2)}), ni(2)))
	funcRound(vm, nil, []*VMValue{ns("1.6")})
	assert.Error(t, vm.Error)
	vm.Error = nil

	assert.True(t, valueEqual(funcFloor(vm, nil, []*VMValue{nf(1.6)}), ni(1)))
	assert.True(t, valueEqual(funcFloor(vm, nil, []*VMValue{ni(1)}), ni(1)))
	funcFloor(vm, nil, []*VMValue{ns("1.6")})
	assert.Error(t, vm.Error)
	vm.Error = nil
}

func TestNativeFunctionConvert(t *testing.T) {
	vm := NewVM()
	assert.True(t, valueEqual(funcInt(vm, nil, []*VMValue{nf(1.1)}), ni(1)))
	assert.True(t, valueEqual(funcInt(vm, nil, []*VMValue{ni(1)}), ni(1)))
	assert.True(t, valueEqual(funcInt(vm, nil, []*VMValue{ns("1")}), ni(1)))

	funcInt(vm, nil, []*VMValue{ns("xx")})
	assert.Error(t, vm.Error)
	vm.Error = nil

	funcInt(vm, nil, []*VMValue{na()})
	assert.Error(t, vm.Error)
	vm.Error = nil

	// float
	assert.True(t, valueEqual(funcFloat(vm, nil, []*VMValue{nf(1.1)}), nf(1.1)))
	assert.True(t, valueEqual(funcFloat(vm, nil, []*VMValue{ni(1)}), nf(1.0)))
	assert.True(t, valueEqual(funcFloat(vm, nil, []*VMValue{ns("1")}), nf(1.0)))

	funcFloat(vm, nil, []*VMValue{ns("xx")})
	assert.Error(t, vm.Error)
	vm.Error = nil

	funcFloat(vm, nil, []*VMValue{na()})
	assert.Error(t, vm.Error)
	vm.Error = nil

	// str
	assert.True(t, valueEqual(funcStr(vm, nil, []*VMValue{nf(1.1)}), ns("1.1")))
	assert.True(t, valueEqual(funcStr(vm, nil, []*VMValue{na(ni(1), ni(2))}), ns("[1, 2]")))
	assert.True(t, valueEqual(funcStr(vm, nil, []*VMValue{na(na(), ni(2))}), ns("[[], 2]")))
}

func TestNativeFunctionLoad(t *testing.T) {
	vm := NewVM()
	err := vm.Run("val = '123'; load('val')")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("123")))
	}

	vm = NewVM()
	err = vm.Run("load('load')")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, builtinValues["load"]))
	}
}

func TestNativeFunctionTypeId(t *testing.T) {
	vm := NewVM()
	err := vm.Run("typeId(1)")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(IntType(VMTypeInt))))
	}
}

func TestNativeFunctionStore(t *testing.T) {
	vm := NewVM()
	err := vm.Run("store('test', 123); test")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(123)))
	}
}

func TestNativeFunctionBool(t *testing.T) {
	vm := NewVM()
	assert.True(t, valueEqual(funcBool(vm, nil, []*VMValue{ni(1)}), ni(1)))
	assert.True(t, valueEqual(funcBool(vm, nil, []*VMValue{ni(0)}), ni(0)))
	assert.True(t, valueEqual(funcBool(vm, nil, []*VMValue{ns("hello")}), ni(1)))
	assert.True(t, valueEqual(funcBool(vm, nil, []*VMValue{ns("")}), ni(0)))
	assert.True(t, valueEqual(funcBool(vm, nil, []*VMValue{NewNullVal()}), ni(0)))
	assert.True(t, valueEqual(funcBool(vm, nil, []*VMValue{na(ni(1))}), ni(1)))
	assert.True(t, valueEqual(funcBool(vm, nil, []*VMValue{na()}), ni(0)))
}

func TestNativeFunctionRepr(t *testing.T) {
	vm := NewVM()
	assert.True(t, valueEqual(funcRepr(vm, nil, []*VMValue{ns("hello")}), ns("'hello'")))
	assert.True(t, valueEqual(funcRepr(vm, nil, []*VMValue{ni(123)}), ns("123")))
	assert.True(t, valueEqual(funcRepr(vm, nil, []*VMValue{nf(1.5)}), ns("1.5")))
	assert.True(t, valueEqual(funcRepr(vm, nil, []*VMValue{NewNullVal()}), ns("null")))
}

func TestNativeFunctionLoadRaw(t *testing.T) {
	vm := NewVM()
	err := vm.Run("val = '456'; loadRaw('val')")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("456")))
	}

	vm = NewVM()
	err = vm.Run("loadRaw(123)")
	assert.Error(t, err)
}

func TestNativeFunctionAbs(t *testing.T) {
	vm := NewVM()
	assert.True(t, valueEqual(funcAbs(vm, nil, []*VMValue{ni(-5)}), ni(5)))
	assert.True(t, valueEqual(funcAbs(vm, nil, []*VMValue{ni(5)}), ni(5)))
	assert.True(t, valueEqual(funcAbs(vm, nil, []*VMValue{nf(-3.5)}), nf(3.5)))
	assert.True(t, valueEqual(funcAbs(vm, nil, []*VMValue{nf(3.5)}), nf(3.5)))

	funcAbs(vm, nil, []*VMValue{ns("test")})
	assert.Error(t, vm.Error)
	vm.Error = nil
}
