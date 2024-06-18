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
