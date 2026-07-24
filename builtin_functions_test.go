package dicescript

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.True(t, valueEqual(funcToInt(vm, nil, []*VMValue{nf(1.1)}), ni(1)))
	assert.True(t, valueEqual(funcToInt(vm, nil, []*VMValue{ni(1)}), ni(1)))
	assert.True(t, valueEqual(funcToInt(vm, nil, []*VMValue{ns("1")}), ni(1)))

	funcToInt(vm, nil, []*VMValue{ns("xx")})
	assert.Error(t, vm.Error)
	vm.Error = nil

	funcToInt(vm, nil, []*VMValue{na()})
	assert.Error(t, vm.Error)
	vm.Error = nil

	// float
	assert.True(t, valueEqual(funcToFloat(vm, nil, []*VMValue{nf(1.1)}), nf(1.1)))
	assert.True(t, valueEqual(funcToFloat(vm, nil, []*VMValue{ni(1)}), nf(1.0)))
	assert.True(t, valueEqual(funcToFloat(vm, nil, []*VMValue{ns("1")}), nf(1.0)))

	funcToFloat(vm, nil, []*VMValue{ns("xx")})
	assert.Error(t, vm.Error)
	vm.Error = nil

	funcToFloat(vm, nil, []*VMValue{na()})
	assert.Error(t, vm.Error)
	vm.Error = nil

	// str
	assert.True(t, valueEqual(funcToStr(vm, nil, []*VMValue{nf(1.1)}), ns("1.1")))
	assert.True(t, valueEqual(funcToStr(vm, nil, []*VMValue{na(ni(1), ni(2))}), ns("[1, 2]")))
	assert.True(t, valueEqual(funcToStr(vm, nil, []*VMValue{na(na(), ni(2))}), ns("[[], 2]")))
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
	assert.True(t, valueEqual(funcToBool(vm, nil, []*VMValue{ni(1)}), ni(1)))
	assert.True(t, valueEqual(funcToBool(vm, nil, []*VMValue{ni(0)}), ni(0)))
	assert.True(t, valueEqual(funcToBool(vm, nil, []*VMValue{ns("hello")}), ni(1)))
	assert.True(t, valueEqual(funcToBool(vm, nil, []*VMValue{ns("")}), ni(0)))
	assert.True(t, valueEqual(funcToBool(vm, nil, []*VMValue{NewNullVal()}), ni(0)))
	assert.True(t, valueEqual(funcToBool(vm, nil, []*VMValue{na(ni(1))}), ni(1)))
	assert.True(t, valueEqual(funcToBool(vm, nil, []*VMValue{na()}), ni(0)))
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

func TestNativeFunctionLoadRawAttr(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a = &(1+2); obj = {'a': &a, 'x': {'a': &a}}; typeId(loadRawAttr(obj, 'a'))")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(IntType(VMTypeComputedValue))))
	}

	vm = NewVM()
	err = vm.Run("a = &(1+2); obj = {'a': &a, 'x': {'a': &a}}; repr(loadRawAttr(loadRawItem(obj, 'x'), 'a'))")
	if assert.NoError(t, err) {
		repr, ok := vm.Ret.ReadString()
		if assert.True(t, ok) {
			assert.Equal(t, "&(1+2)", repr)
		}
	}

	vm = NewVM()
	err = vm.Run("a = &(1+2); obj = {'a': &a}; obj.a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}
}

func TestNativeFunctionLoadRawItem(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a = &(1+2); obj = {'a': &a}; typeId(loadRawItem(obj, 'a'))")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(IntType(VMTypeComputedValue))))
	}

	vm = NewVM()
	err = vm.Run("a = &(1+2); arr = [&a]; typeId(loadRawItem(arr, 0))")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(IntType(VMTypeComputedValue))))
	}
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

func TestNativeFunctionErrorPropagation(t *testing.T) {
	ctx := NewVM()
	assert.True(t, valueEqual(funcCeil(ctx, nil, []*VMValue{ni(1)}), ni(1)))

	ctx = NewVM()
	ctx.Attrs.Store("broken", NewComputedVal("("))
	assert.Nil(t, funcLoad(ctx, nil, []*VMValue{ns("broken")}))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	assert.Nil(t, funcLoadRawAttr(ctx, nil, []*VMValue{ni(1), ni(2)}))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	obj := NewNativeObjectVal(&NativeObjectData{AttrGet: func(ctx *Context, _ string) *VMValue {
		ctx.Error = assert.AnError
		return nil
	}})
	assert.Nil(t, funcLoadRawAttr(ctx, nil, []*VMValue{obj, ns("field")}))
	assert.ErrorIs(t, ctx.Error, assert.AnError)

	ctx = NewVM()
	assert.Nil(t, funcLoadRawItem(ctx, nil, []*VMValue{ni(1), ni(0)}))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	assert.Nil(t, funcStore(ctx, nil, []*VMValue{ni(1), ni(2)}))
	assert.Error(t, ctx.Error)
}
