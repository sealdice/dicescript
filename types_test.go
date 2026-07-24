/*
  Copyright 2022 fy <fy0748@gmail.com>

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package dicescript

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
)

type compareTestData []struct {
	v1       *VMValue
	v2       *VMValue
	excepted *VMValue
}

func valueEqual(a *VMValue, b *VMValue) bool {
	return ValueEqual(a, b, false)
}

var ni = NewIntVal
var nf = NewFloatVal
var ns = NewStrVal
var na = NewArrayVal
var nd = NewDictValWithArrayMust

func TestContextConfigurationAndGlobals(t *testing.T) {
	vm := NewVM()

	vm.subThreadDepth = 3
	assert.Equal(t, 3, vm.Depth())

	cfg := RollConfig{
		IgnoreDiv0:            true,
		DefaultDiceSideExpr:   "42",
		EnableDiceDoubleCross: true,
	}
	vm.SetConfig(&cfg)
	assert.True(t, vm.Config.IgnoreDiv0)
	cfg.IgnoreDiv0 = false
	assert.True(t, vm.Config.IgnoreDiv0)

	var storedName string
	var storedVal *VMValue
	vm.GlobalValueStoreFunc = func(name string, v *VMValue) {
		storedName = name
		storedVal = v
	}
	vm.globalNames.Store("globalVar", NewNullVal())
	vm.StoreName("globalVar", ni(99), false)
	assert.Equal(t, "globalVar", storedName)
	assert.True(t, valueEqual(storedVal, ni(99)))
	assert.Nil(t, vm.Error)

	vm.GlobalValueStoreFunc = nil
	vm.Error = nil
	vm.StoreNameGlobal("missing", ni(1))
	if assert.NotNil(t, vm.Error) {
		assert.Contains(t, vm.Error.Error(), "ValueStore")
	}

	src := rand.PCGSource{}
	src.Seed(123)
	vm.RandSrc = &src
	seed, err := vm.GetCurSeed()
	assert.NoError(t, err)
	assert.NotEmpty(t, seed)
}

func TestCompare(t *testing.T) {
	ctx := NewVM()

	// lt 小于
	var compLTTest = compareTestData{
		// int, int
		{ni(0), ni(0), ni(0)}, // 0 < 0, false
		{ni(0), ni(2), ni(1)}, // 0 < 2, true
		{ni(2), ni(0), ni(0)}, // 2 < 0, false

		// int float
		{ni(0), nf(0), ni(0)}, // 0 < 0, false
		{ni(0), nf(2), ni(1)}, // 0 < 2, true
		{ni(2), nf(0), ni(0)}, // 2 < 0, false

		// float int
		{nf(0), ni(0), ni(0)}, // 0 < 0, false
		{nf(0), ni(2), ni(1)}, // 0 < 2, true
		{nf(2), ni(0), ni(0)}, // 2 < 0, false

		// float float
		{nf(0), nf(0), ni(0)}, // 0 < 0, false
		{nf(0), nf(2), ni(1)}, // 0 < 2, true
		{nf(2), nf(0), ni(0)}, // 2 < 0, false

		// int str
		{ni(0), ns("2"), nil}, // 0 < '2', ERR
	}

	for _, i := range compLTTest {
		r := (*VMValue).OpCompLT(i.v1, ctx, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("CompareLE(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// le 小于等于
	var compLETest = compareTestData{
		// int, int
		{ni(0), ni(0), ni(1)}, // 0 <= 0, true
		{ni(0), ni(2), ni(1)}, // 0 <= 2, true
		{ni(2), ni(0), ni(0)}, // 2 <= 0, false

		// int float
		{ni(0), nf(0), ni(1)}, // 0 <= 0, true
		{ni(0), nf(2), ni(1)}, // 0 <= 2, true
		{ni(2), nf(0), ni(0)}, // 2 <= 0, false

		// float int
		{nf(0), ni(0), ni(1)}, // 0 <= 0, true
		{nf(0), ni(2), ni(1)}, // 0 <= 2, true
		{nf(2), ni(0), ni(0)}, // 2 <= 0, false

		// float float
		{nf(0), nf(0), ni(1)}, // 0 <= 0, true
		{nf(0), nf(2), ni(1)}, // 0 <= 2, true
		{nf(2), nf(0), ni(0)}, // 2 <= 0, false

		// int str
		{ni(0), ns("2"), nil}, // 0 <= '2', ERR
	}

	for _, i := range compLETest {
		r := (*VMValue).OpCompLE(i.v1, ctx, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("CompareLE(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// ge 大于等于
	var compGETest = compareTestData{
		// int, int
		{ni(0), ni(0), ni(1)}, // 0 >= 0, true
		{ni(0), ni(2), ni(0)}, // 0 >= 2, false
		{ni(2), ni(0), ni(1)}, // 2 >= 0, true

		// int float
		{ni(0), nf(0), ni(1)}, // 0 >= 0, true
		{ni(0), nf(2), ni(0)}, // 0 >= 2, false
		{ni(2), nf(0), ni(1)}, // 2 >= 0, true

		// float int
		{nf(0), ni(0), ni(1)}, // 0 >= 0, true
		{nf(0), ni(2), ni(0)}, // 0 >= 2, false
		{nf(2), ni(0), ni(1)}, // 2 >= 0, true

		// float float
		{nf(0), nf(0), ni(1)}, // 0 >= 0, true
		{nf(0), nf(2), ni(0)}, // 0 >= 2, false
		{nf(2), nf(0), ni(1)}, // 2 >= 0, true

		// int str
		{ni(0), ns("2"), nil}, // 0 >= '2', ERR
	}

	for _, i := range compGETest {
		r := (*VMValue).OpCompGE(i.v1, ctx, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("CompareGE(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// gt 大于
	var compGTTest = compareTestData{
		// int, int
		{ni(0), ni(0), ni(0)}, // 0 > 0, false
		{ni(0), ni(2), ni(0)}, // 0 > 2, false
		{ni(2), ni(0), ni(1)}, // 2 > 0, true

		// int float
		{ni(0), nf(0), ni(0)}, // 0 > 0, false
		{ni(0), nf(2), ni(0)}, // 0 > 2, false
		{ni(2), nf(0), ni(1)}, // 2 > 0, true

		// float int
		{nf(0), ni(0), ni(0)}, // 0 > 0, false
		{nf(0), ni(2), ni(0)}, // 0 > 2, false
		{nf(2), ni(0), ni(1)}, // 2 > 0, true

		// float float
		{nf(0), nf(0), ni(0)}, // 0 > 0, false
		{nf(0), nf(2), ni(0)}, // 0 > 2, false
		{nf(2), nf(0), ni(1)}, // 2 > 0, true

		// int str
		{ni(0), ns("2"), nil}, // 0 > '2', ERR
	}

	for _, i := range compGTTest {
		r := (*VMValue).OpCompGT(i.v1, ctx, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("CompareGT(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// EQ
	theSame := ni(123)
	var compEQTest = compareTestData{
		{theSame, theSame, ni(1)},
		// int, int
		{ni(0), ni(0), ni(1)},  // 0 == 0, true
		{ni(-1), ni(1), ni(0)}, // -1 == 1, false
		// int, float
		{ni(0), nf(0), ni(1)}, // 0 == 0, true
		{ni(0), nf(1), ni(0)}, // 0 == 1, false
		// float, int
		{nf(1), ni(0), ni(0)}, // 1 == 0, false
		// int, str
		{ni(0), ns(""), ni(0)}, // 0 == '', false
	}
	for _, i := range compEQTest {
		r := (*VMValue).OpCompEQ(i.v1, ctx, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("CompareEQ(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	var compEQTest2 = compareTestData{
		{na(ni(1), ni(2), ni(3)), na(ni(1), ni(2), ni(3)), ni(1)},        // [1,2,3] == [1,2,3] true
		{na(ni(1), ni(2)), na(ni(1), ni(2), ni(3)), ni(0)},               // [1,2] == [1,2,3] false
		{na(ni(1), ni(2), ni(3)), na(ni(1), ni(2), ni(3), ni(4)), ni(0)}, // [1,2,3] == [1,2,3,4] false
		{na(ni(1), ni(2), ni(3)), na(ni(1), ni(2), ni(4)), ni(0)},        // [1,2,3] == [1,2,4] false

		{nd(ns("a"), ni(1)).V(), nd(ns("a"), ni(1)).V(), ni(1)},                 // {'a':1} == {'a':1} true
		{nd(ns("a"), ni(1)).V(), nd(ns("a"), ni(2)).V(), ni(0)},                 // {'a':1} == {'a':2} false
		{nd(ns("a"), ni(1)).V(), nd(ns("a"), ni(1), ns("b"), ni(2)).V(), ni(0)}, // {'a':1} == {'a':1,'b':2} false
	}

	for _, i := range compEQTest2 {
		r := (*VMValue).OpCompEQ(i.v1, ctx, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("CompareEQ2(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}
}

func TestPositiveAndNegative(t *testing.T) {
	assert.True(t, valueEqual((*VMValue).OpPositive(ni(1)), ni(1)))
	assert.True(t, valueEqual((*VMValue).OpPositive(nf(1.2)), nf(1.2)))
	assert.True(t, valueEqual((*VMValue).OpNegation(ni(1)), ni(-1)))
	assert.True(t, valueEqual((*VMValue).OpNegation(nf(1.2)), nf(-1.2)))
}

func TestAdditive(t *testing.T) {
	ctx := NewVM()
	// + add
	var addTest = compareTestData{
		// int, int
		{ni(1), ni(2), ni(3)}, // 1+2=3
		// int, float
		{ni(1), nf(2), nf(3)}, // 1+2=3
		// float, int
		{nf(1), ni(2), nf(3)}, // 1+2=3
		// float, flaot
		{nf(1), nf(2), nf(3)}, // 1+2=3
		// str, str
		{ns("aa"), ns("bb"), ns("aabb")}, // 'aa'+'bb'='aabb'
		//
		{na(ni(1), ni(2)), na(ni(3)), na(ni(1), ni(2), ni(3))},
	}

	for _, i := range addTest {
		r := (*VMValue).OpAdd(i.v1, ctx, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("OpAdd(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// - sub
	var subTest = compareTestData{
		// int, int
		{ni(3), ni(2), ni(1)}, // 3-2=1
		// int, float
		{ni(3), nf(2), nf(1)}, // 3-2=1
		// float, int
		{nf(3), ni(2), nf(1)}, // 3-2=1
		// float, flaot
		{nf(3), nf(2), nf(1)}, // 3-2=1
	}

	for _, i := range subTest {
		r := (*VMValue).OpSub(i.v1, ctx, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("OpSub(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// * multiply
	var subMul = compareTestData{
		// int, int
		{ni(3), ni(2), ni(6)}, // 3*2=6
		// int, float
		{ni(3), nf(2), nf(6)}, // 3*2=6
		// float, int
		{nf(3), ni(2), nf(6)}, // 3*2=6
		// float, flaot
		{nf(3), nf(2), nf(6)}, // 3*2=6
		// arr int
		{na(ni(1), ni(2)), ni(2), na(ni(1), ni(2), ni(1), ni(2))},
		// int arr
		{ni(2), na(ni(1), ni(2)), na(ni(1), ni(2), ni(1), ni(2))},
	}

	for _, i := range subMul {
		r := (*VMValue).OpMultiply(i.v1, ctx, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("Mul(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// * div
	var divTest = compareTestData{
		// int, int
		{ni(3), ni(2), ni(1)}, // 3/2=1
		// int, float
		{ni(3), nf(2), nf(1.5)}, // 3/2=1.5
		// float, int
		{nf(3), ni(2), nf(1.5)}, // 3/2=1.5
		// float, flaot
		{nf(3), nf(2), nf(1.5)}, // 3/2=1.5
		// TODO: 被除数为0
	}

	for _, i := range divTest {
		r := (*VMValue).OpDivide(i.v1, ctx, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("Div(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// * mod
	var modTest = compareTestData{
		// int, int
		{ni(2), ni(3), ni(2)}, // 2%3=2
		// int, float
		{ni(3), nf(2), nil},
		// TODO: 被除数为0
	}

	for _, i := range modTest {
		r := (*VMValue).OpModulus(i.v1, ctx, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("Mod(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// ** power
	var powerTest = compareTestData{
		// int, int
		{ni(2), ni(3), ni(8)}, // 2^3=8
		// int, float
		{ni(3), nf(4), nf(81)},
		// float, float
		{nf(3), nf(4), nf(81)},
		// float, int
		{nf(3), ni(4), nf(81)},
	}

	for _, i := range powerTest {
		r := (*VMValue).OpPower(i.v1, ctx, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("Power(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}
}

func TestAttrGet(t *testing.T) {
	vm := NewVM()
	err := vm.Run("&a = d + 1; &a.d = 2; &a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret.AttrGet(vm, "d"), ni(2)))
		assert.True(t, valueEqual(vm.Ret.AttrGet(vm, "a"), NewNullVal()))
	}
}

func TestDictProto(t *testing.T) {
	vm := NewVM()
	err := vm.Run("_d1 = {'a':1, 'b':2}")
	assert.NoError(t, err)
	err = vm.Run("_d2 = {'__proto__': _d1}")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret.AttrGet(vm, "a"), ni(1)))
		assert.True(t, valueEqual(vm.Ret.AttrGet(vm, "b"), ni(2)))
	}
}

func TestNativeObject(t *testing.T) {
	vm := NewVM()
	var slot *VMValue
	od := &NativeObjectData{
		Name: "obj1",
		AttrSet: func(ctx *Context, name string, v *VMValue) {
			slot = v
		},
		AttrGet: func(ctx *Context, name string) *VMValue {
			return slot
		},
		ItemGet: func(ctx *Context, index *VMValue) *VMValue {
			return slot
		},
		ItemSet: func(ctx *Context, index *VMValue, v *VMValue) {
			slot = v
		},
		DirFunc: func(ctx *Context) []*VMValue {
			return []*VMValue{ns("x")}
		},
	}
	v := NewNativeObjectVal(od)
	assert.True(t, valueEqual(v.AttrGet(vm, "a"), NewNullVal()))
	v.AttrSet(vm, "a", ni(1))
	assert.True(t, valueEqual(v.AttrGet(vm, "a"), ni(1)))

	assert.True(t, valueEqual(v.ItemGet(vm, ni(0)), ni(1)))
	v.AttrSet(vm, "a", ni(2))
	assert.True(t, valueEqual(v.ItemGet(vm, ni(0)), ni(2)))

	ret := funcDir(vm, nil, []*VMValue{v})
	assert.Equal(t, ret.ToString(), "['x']")
}

func TestNativeObjectComputedRead(t *testing.T) {
	vm := NewVM()
	slot := NewComputedVal("1+2")
	od := &NativeObjectData{
		Name: "obj1",
		AttrGet: func(ctx *Context, name string) *VMValue {
			return slot
		},
		ItemGet: func(ctx *Context, index *VMValue) *VMValue {
			return slot
		},
	}
	v := NewNativeObjectVal(od)
	assert.True(t, valueEqual(v.AttrGet(vm, "a"), ni(3)))
	assert.True(t, valueEqual(v.ItemGet(vm, ni(0)), ni(3)))
}

func TestAsBool(t *testing.T) {
	assert.Equal(t, ni(1).AsBool(), true)
	assert.Equal(t, ni(0).AsBool(), false)

	assert.Equal(t, nf(1.1).AsBool(), true)
	assert.Equal(t, nf(0.0).AsBool(), false)

	assert.Equal(t, ns("1").AsBool(), true)
	assert.Equal(t, ns("").AsBool(), false)

	assert.Equal(t, NewNullVal().AsBool(), false)

	assert.Equal(t, NewComputedVal("d10").AsBool(), true)
	assert.Equal(t, NewComputedVal("").AsBool(), false)

	assert.Equal(t, na(ns("1")).AsBool(), true)
	assert.Equal(t, na().AsBool(), false)

	assert.Equal(t, nd(ns("1"), ns("1")).V().AsBool(), true)
	assert.Equal(t, nd().V().AsBool(), false)

	vm := NewVM()
	_ = vm.Run("func a() {}")
	assert.Equal(t, vm.Ret.AsBool(), true)

	assert.Equal(t, builtinValues["toStr"].AsBool(), true)
}

func TestVMValueRepresentationsAndTypeReads(t *testing.T) {
	function := NewFunctionValRaw(&FunctionData{Name: "f"})
	nativeFunction := NewNativeFunctionVal(&NativeFunctionData{Name: "nf"})
	nativeObject := NewNativeObjectVal(&NativeObjectData{Name: "obj"})
	invalid := &VMValue{TypeId: VMValueType(999)}

	assert.Equal(t, "function f", function.ToString())
	assert.Equal(t, "nfunction nf", nativeFunction.ToString())
	assert.Equal(t, "nobject obj", nativeObject.ToString())
	assert.Equal(t, "a value", invalid.ToString())
	assert.Equal(t, "<a value>", invalid.ToRepr())
	assert.Equal(t, "NIL", (*VMValue)(nil).ToString())
	assert.Equal(t, "NIL", (*VMValue)(nil).ToRepr())

	_, ok := ni(1).ReadString()
	assert.False(t, ok)
	_, ok = ni(1).ReadArray()
	assert.False(t, ok)
	_, ok = ni(1).ReadComputed()
	assert.False(t, ok)
	_, ok = ni(1).ReadFunctionData()
	assert.False(t, ok)
	_, ok = ni(1).ReadNativeFunctionData()
	assert.False(t, ok)
	_, ok = ni(1).ReadNativeObjectData()
	assert.False(t, ok)

	assert.Panics(t, func() { ni(1).MustReadDictData() })
	assert.Panics(t, func() { ni(1).MustReadArray() })
	assert.Panics(t, func() { ns("1").MustReadInt() })
	assert.Panics(t, func() { ns("1").MustReadFloat() })

	typeNames := map[*VMValue]string{
		ni(1):               "int",
		nf(1):               "float",
		ns("x"):             "str",
		NewNullVal():        "null",
		NewComputedVal("1"): "computed",
		na():                "array",
		function:            "function",
		nativeFunction:      "nfunction",
		nativeObject:        "nobject",
		invalid:             "unknown",
	}
	for value, want := range typeNames {
		assert.Equal(t, want, value.GetTypeName())
	}
}

func TestVMValueArithmeticErrorCases(t *testing.T) {
	for name, test := range map[string]struct {
		left  *VMValue
		right *VMValue
	}{
		"int by float zero":   {left: ni(2), right: nf(0)},
		"float by int zero":   {left: nf(2), right: ni(0)},
		"float by float zero": {left: nf(2), right: nf(0)},
	} {
		t.Run(name, func(t *testing.T) {
			ctx := NewVM()
			assert.Nil(t, test.left.OpDivide(ctx, test.right))
			assert.Error(t, ctx.Error)
		})
	}

	ctx := NewVM()
	ctx.Config.IgnoreDiv0 = true
	left := ni(7)
	assert.Same(t, left, left.OpDivide(ctx, ni(0)))
	assert.NoError(t, ctx.Error)

	ctx = NewVM()
	assert.Nil(t, ni(7).OpModulus(ctx, ni(0)))
	assert.Error(t, ctx.Error)

	unsupported := NewNullVal()
	assert.Nil(t, unsupported.OpSub(ctx, ni(1)))
	assert.Nil(t, unsupported.OpMultiply(ctx, ni(1)))
	assert.Nil(t, unsupported.OpDivide(ctx, ni(1)))
	assert.Nil(t, unsupported.OpBitwiseAnd(ctx, ni(1)))
	assert.Nil(t, unsupported.OpBitwiseOr(ctx, ni(1)))
	assert.Nil(t, unsupported.OpPositive())
}

func TestVMValueItemAndSliceErrorCases(t *testing.T) {
	ctx := NewVM()
	assert.Nil(t, na(ni(1)).ItemGet(ctx, ns("0")))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	assert.Nil(t, nd().V().ItemGet(ctx, na()))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	assert.Nil(t, ns("abc").ItemGet(ctx, ns("0")))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	native := NewNativeObjectVal(&NativeObjectData{
		ItemGet: func(*Context, *VMValue) *VMValue { return nil },
		ItemSet: func(*Context, *VMValue, *VMValue) {},
	})
	assert.True(t, valueEqual(native.ItemGet(ctx, ni(0)), NewNullVal()))
	assert.True(t, native.ItemSet(ctx, ni(0), ni(1)))

	ctx = NewVM()
	assert.False(t, na().ItemSet(ctx, ns("0"), ni(1)))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	assert.False(t, nd().V().ItemSet(ctx, na(), ni(1)))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	assert.False(t, NewNullVal().ItemSet(ctx, ni(0), ni(1)))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	assert.Equal(t, IntType(0), getClampRealIndex(ctx, -10, 3))
	assert.Equal(t, IntType(3), getClampRealIndex(ctx, 10, 3))
	assert.True(t, valueEqual(na(ni(1), ni(2)).GetSlice(ctx, 2, 1, 1), na()))
	assert.Equal(t, IntType(0), nd().V().Length(ctx))
	assert.NoError(t, ctx.Error)

	ctx = NewVM()
	assert.Zero(t, ni(1).Length(ctx))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	assert.Nil(t, na().GetSliceEx(ctx, ns("bad"), NewNullVal()))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	assert.Nil(t, na().GetSliceEx(ctx, ni(0), ns("bad")))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	assert.False(t, ni(1).SetSlice(ctx, 0, 1, 1, na()))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	assert.False(t, na().SetSlice(ctx, 0, 1, 1, ni(1)))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	array := na(ni(1), ni(2))
	assert.True(t, array.SetSlice(ctx, 2, 1, 1, na(ni(3))))
	assert.True(t, valueEqual(array, na(ni(1), ni(3), ni(2))))

	ctx = NewVM()
	assert.False(t, ni(1).SetSliceEx(ctx, NewNullVal(), NewNullVal(), na()))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	assert.False(t, na().SetSliceEx(ctx, ns("bad"), NewNullVal(), na()))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	assert.False(t, na().SetSliceEx(ctx, ni(0), ns("bad"), na()))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	assert.Nil(t, na(ni(1)).ArrayRepeatTimesEx(ctx, ni(513)))
	assert.Error(t, ctx.Error)
	assert.Nil(t, na(ni(1)).ArrayRepeatTimesEx(NewVM(), nf(2)))
}

func TestNativeFunctionDefaultsAndErrors(t *testing.T) {
	definition := &NativeFunctionData{
		Name:     "withDefault",
		Params:   []string{"value"},
		Defaults: []*VMValue{ni(9)},
		NativeFunc: func(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
			assert.True(t, valueEqual(params[0], ni(9)))
			return nil
		},
	}
	fn := NewNativeFunctionVal(definition)
	ctx := NewVM()
	assert.True(t, valueEqual(fn.FuncInvokeNative(ctx, nil), NewNullVal()))
	assert.NoError(t, ctx.Error)

	ctx = NewVM()
	fn = NewNativeFunctionVal(&NativeFunctionData{
		Params: []string{"required"},
		NativeFunc: func(*Context, *VMValue, []*VMValue) *VMValue {
			t.Fatal("native function must not run with the wrong arity")
			return nil
		},
	})
	assert.Nil(t, fn.FuncInvokeNative(ctx, nil))
	assert.Error(t, ctx.Error)

	ctx = NewVM()
	fn = NewNativeFunctionVal(&NativeFunctionData{
		NativeFunc: func(ctx *Context, _ *VMValue, _ []*VMValue) *VMValue {
			ctx.Error = assert.AnError
			return ni(1)
		},
	})
	assert.Nil(t, fn.FuncInvokeNative(ctx, nil))
	assert.ErrorIs(t, ctx.Error, assert.AnError)
}

func TestInvalidDictionaryKeys(t *testing.T) {
	_, err := na().AsDictKey()
	assert.Error(t, err)
	_, err = NewDictValWithArray(na(), ni(1))
	assert.Error(t, err)
	assert.Panics(t, func() { NewDictValWithArrayMust(na(), ni(1)) })
}

func TestContextLifecycleAndLoadEdgeCases(t *testing.T) {
	ctx := NewVM()
	ctx.DetailSpans = []BufferSpan{}
	ctx.detailCache = "cached detail"
	assert.Equal(t, "cached detail", ctx.GetDetailText())

	seed, err := (&Context{}).GetCurSeed()
	assert.NoError(t, err)
	assert.NotEmpty(t, seed)
	seeded := &Context{Seed: seed}
	seeded.Init()
	assert.NotNil(t, seeded.RandSrc)

	ctx = NewVM()
	detail := &BufferSpan{Ret: ni(1)}
	ctx.Config.HookValueLoadPost = func(_ *Context, _ string, _ *VMValue, _ func(*VMValue) *VMValue, _ *BufferSpan) *VMValue {
		return ni(2)
	}
	value := ctx.solveLoadPostAndComputed("value", ni(1), false, detail)
	assert.True(t, valueEqual(value, ni(2)))
	assert.True(t, valueEqual(detail.Ret, ni(2)))
	value = ctx.solveLoadPostAndComputed("value", ni(1), false, nil)
	assert.True(t, valueEqual(value, ni(2)))

	ctx = NewVM()
	ctx.Config.HookValueLoadPre = func(_ *Context, name string) (string, *VMValue) {
		assert.Equal(t, "original", name)
		return "renamed", ni(7)
	}
	assert.True(t, valueEqual(ctx.LoadName("original", false, true), ni(7)))
	assert.True(t, valueEqual(ctx.LoadNameLocal("missing", false), NewNullVal()))
	assert.True(t, valueEqual(ctx.LoadNameGlobal("missing", false), NewNullVal()))
}

func TestCustomDiceRegistrationValidation(t *testing.T) {
	ctx := NewVM()
	assert.Error(t, ctx.RegCustomDice("X", nil))
	assert.Error(t, ctx.RegCustomDice("[", func(*Context, []string, any) (*VMValue, string, error) {
		return nil, "", nil
	}))
	assert.Error(t, ctx.RegCustomDiceParser(nil, func(*Context, []string, any) (*VMValue, string, error) {
		return nil, "", nil
	}))
	assert.Error(t, ctx.RegCustomDiceParser(func(*Context, *CustomDiceStream) (*CustomDiceParseResult, error) {
		return nil, nil
	}, nil))
}
