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
	"github.com/stretchr/testify/assert"
	"testing"
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

	assert.Equal(t, builtinValues["str"].AsBool(), true)
}
