package dicescript

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypesMethodComputedCompute(t *testing.T) {
	_attrs := &ValueMap{}
	_attrs.Store("x", ni(1))
	c := NewComputedValRaw(&ComputedData{Expr: "this.x + 10", Attrs: _attrs})

	vm := NewVM()
	ret := funcComputedCompute(vm, c, nil)

	assert.Equal(t, ret.ToString(), "11")
}

func TestTypesMethodArraySum(t *testing.T) {
	d := NewArrayVal(ni(1), nf(2.2), ni(3))
	v := funcArraySum(nil, d, nil)
	assert.Equal(t, v.ToString(), "6.2")
}

func TestTypesMethodArrayShuttle(t *testing.T) {
	d := NewArrayVal(ni(1), ni(2), ni(3), ni(4))
	v := funcArrayShuttle(nil, d, nil)
	// 不知道怎么写测试，因为总是有概率打乱后与原本一致
	assert.Equal(t, v.Length(nil), IntType(4))
}

func TestTypesMethodArrayRand(t *testing.T) {
	d := NewArrayVal(ni(1), ni(1), ni(1), ni(1))
	v := funcArrayRand(nil, d, nil)
	assert.Equal(t, v.MustReadInt(), IntType(1))
}

func TestTypesMethodArrayRandSize(t *testing.T) {
	d := NewArrayVal(ni(1), ni(1), ni(1), ni(1))
	v := funcArrayRandSize(nil, d, []*VMValue{ni(1)})
	assert.Equal(t, v.Length(nil), IntType(1))
}

func TestTypesMethodDictKeys(t *testing.T) {
	d := NewDictValWithArrayMust(ns("a"), ni(1), ns("b"), ni(2))
	v := funcDictKeys(nil, d.V(), nil)
	assert.True(t, valueEqual(v, na(ns("a"), ns("b"))) || valueEqual(v, na(ns("b"), ns("a"))))
}

func TestTypesMethodDictValues(t *testing.T) {
	d := NewDictValWithArrayMust(ns("a"), ni(1), ns("b"), ni(2))
	v := funcDictValues(nil, d.V(), nil)
	assert.True(t, valueEqual(v, na(ni(1), ni(2))) || valueEqual(v, na(ni(2), ni(1))))
}

func TestTypesMethodDictItems(t *testing.T) {
	d := NewDictValWithArrayMust(ns("a"), ni(1), ns("b"), ni(2))
	v := funcDictItems(nil, d.V(), nil)
	assert.True(t, valueEqual(v, na(na(ns("a"), ni(1)), na(ns("b"), ni(2)))) ||
		valueEqual(v, na(na(ns("b"), ni(2)), na(ns("a"), ni(1)))))
}

func TestTypesMethodDictLen(t *testing.T) {
	d := NewDictValWithArrayMust(ns("a"), ni(1), ns("b"), ni(2))
	v := funcDictLen(nil, d.V(), nil)
	assert.Equal(t, v.MustReadInt(), IntType(2))
}

func TestTypesMethodDictHas(t *testing.T) {
	vm := NewVM()
	d := NewDictValWithArrayMust(ns("a"), ni(1), ns("b"), ni(2))

	v := funcDictHas(vm, d.V(), []*VMValue{ns("a")})
	assert.True(t, valueEqual(v, ni(1)))

	v = funcDictHas(vm, d.V(), []*VMValue{ns("c")})
	assert.True(t, valueEqual(v, ni(0)))
}

func TestTypesMethodDictGet(t *testing.T) {
	vm := NewVM()
	c := NewComputedValRaw(&ComputedData{Expr: "1 + 2"})
	d := NewDictValWithArrayMust(ns("a"), ni(1), ns("c"), c)

	v := funcDictGet(vm, d.V(), []*VMValue{ns("a"), ni(9)})
	assert.True(t, valueEqual(v, ni(1)))

	v = funcDictGet(vm, d.V(), []*VMValue{ns("c"), ni(9)})
	assert.True(t, valueEqual(v, ni(3)))

	v = funcDictGet(vm, d.V(), []*VMValue{ns("missing"), ni(9)})
	assert.True(t, valueEqual(v, ni(9)))
}

func TestTypesMethodDictGetRaw(t *testing.T) {
	vm := NewVM()
	c := NewComputedValRaw(&ComputedData{Expr: "1 + 2"})
	d := NewDictValWithArrayMust(ns("a"), ni(1), ns("c"), c)

	v := funcDictGetRaw(vm, d.V(), []*VMValue{ns("a"), ni(9)})
	assert.True(t, valueEqual(v, ni(1)))

	v = funcDictGetRaw(vm, d.V(), []*VMValue{ns("c"), ni(9)})
	assert.Equal(t, VMTypeComputedValue, v.TypeId)

	v = funcDictGetRaw(vm, d.V(), []*VMValue{ns("missing"), ni(9)})
	assert.True(t, valueEqual(v, ni(9)))
}

func TestTypesMethodErrorAndEmptyCases(t *testing.T) {
	ctx := NewVM()
	assert.Nil(t, funcArrayRandSize(ctx, na(ni(1)), []*VMValue{ns("1")}))
	assert.Error(t, ctx.Error)

	assert.True(t, valueEqual(funcArrayPop(NewVM(), na(), nil), NewNullVal()))
	assert.True(t, valueEqual(funcArrayShift(NewVM(), na(), nil), NewNullVal()))

	d := NewDictVal(nil)
	for name, call := range map[string]func(*Context) *VMValue{
		"has": func(ctx *Context) *VMValue {
			return funcDictHas(ctx, d.V(), []*VMValue{na()})
		},
		"get": func(ctx *Context) *VMValue {
			return funcDictGet(ctx, d.V(), []*VMValue{na()})
		},
		"getRaw": func(ctx *Context) *VMValue {
			return funcDictGetRaw(ctx, d.V(), []*VMValue{na()})
		},
	} {
		t.Run(name+" rejects invalid key", func(t *testing.T) {
			ctx := NewVM()
			assert.Nil(t, call(ctx))
			assert.Error(t, ctx.Error)
		})
	}

	assert.True(t, valueEqual(funcDictGet(NewVM(), d.V(), []*VMValue{ns("missing")}), NewNullVal()))
	assert.True(t, valueEqual(funcDictGetRaw(NewVM(), d.V(), []*VMValue{ns("missing")}), NewNullVal()))
}

func TestGetBindMethodCopiesFunction(t *testing.T) {
	receiver := na(ni(1))
	definition := &FunctionData{Name: "method", Params: []string{"x"}}
	bound := getBindMethod(receiver, NewFunctionValRaw(definition))

	data, ok := bound.ReadFunctionData()
	assert.True(t, ok)
	assert.NotSame(t, definition, data)
	assert.Nil(t, definition.Self)
	assert.True(t, valueEqual(receiver, data.Self))
	assert.Nil(t, getBindMethod(receiver, ni(1)))
}
