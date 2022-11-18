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
	assert.True(t, valueEqual(funcCeil(vm, []*VMValue{nf(1.1)}), ni(2)))
	assert.True(t, valueEqual(funcCeil(vm, []*VMValue{nf(1.0)}), ni(1)))

	assert.True(t, valueEqual(funcRound(vm, []*VMValue{nf(1.6)}), ni(2)))
	funcRound(vm, []*VMValue{ni(1)})
	assert.Error(t, vm.Error)
	vm.Error = nil

	assert.True(t, valueEqual(funcFloor(vm, []*VMValue{nf(1.6)}), ni(1)))
	funcFloor(vm, []*VMValue{ni(1)})
	assert.Error(t, vm.Error)
	vm.Error = nil
}

func TestNativeFunctionConvert(t *testing.T) {
	vm := NewVM()
	assert.True(t, valueEqual(funcInt(vm, []*VMValue{nf(1.1)}), ni(1)))
	assert.True(t, valueEqual(funcInt(vm, []*VMValue{ni(1)}), ni(1)))
	assert.True(t, valueEqual(funcInt(vm, []*VMValue{ns("1")}), ni(1)))

	funcInt(vm, []*VMValue{ns("xx")})
	assert.Error(t, vm.Error)
	vm.Error = nil

	funcInt(vm, []*VMValue{na()})
	assert.Error(t, vm.Error)
	vm.Error = nil

	// float
	assert.True(t, valueEqual(funcFloat(vm, []*VMValue{nf(1.1)}), nf(1.1)))
	assert.True(t, valueEqual(funcFloat(vm, []*VMValue{ni(1)}), nf(1.0)))
	assert.True(t, valueEqual(funcFloat(vm, []*VMValue{ns("1")}), nf(1.0)))

	funcFloat(vm, []*VMValue{ns("xx")})
	assert.Error(t, vm.Error)
	vm.Error = nil

	funcFloat(vm, []*VMValue{na()})
	assert.Error(t, vm.Error)
	vm.Error = nil

	// str
	assert.True(t, valueEqual(funcStr(vm, []*VMValue{nf(1.1)}), ns("1.1")))
	assert.True(t, valueEqual(funcStr(vm, []*VMValue{na(ni(1), ni(2))}), ns("[1, 2]")))
	assert.True(t, valueEqual(funcStr(vm, []*VMValue{na(na(), ni(2))}), ns("[[...], 2]")))
}
