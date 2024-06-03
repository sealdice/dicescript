package dicescript

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDumps(t *testing.T) {
	var err error
	var v []byte

	v, err = NewIntVal(123).ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"t":0,"v":123}`, string(v))
	}

	v, err = NewFloatVal(3.2).ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"t":1,"v":3.2}`, string(v))
	}

	v, err = NewStrVal("asd").ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"t":2,"v":"asd"}`, string(v))
	}

	v, err = NewNullVal().ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"t":4}`, string(v))
	}

	v, err = NewComputedVal("1 + this.x + d10").ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"t":5,"v":{"expr":"1 + this.x + d10"}}`, string(v))
	}

	vm := NewVM()
	err = vm.Run(`func a(x) { return 5 }; a`)
	if assert.NoError(t, err) {
		ret := vm.Ret
		v, err = ret.ToJSON() // nolint
		assert.Equal(t, `{"t":8,"v":{"expr":"return 5 ","name":"a","params":["x"]}}`, string(v))
	}

	v, err = na(ni(1), nf(2.0), ns("test")).ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"t":6,"v":{"list":[{"t":0,"v":1},{"t":1,"v":2},{"t":2,"v":"test"}]}}`, string(v))
	}

	m := ValueMap{}
	m.Store("v2", ni(2))
	m.Store("v1", ni(1))
	v, err = NewDictVal(&m).V().ToJSON()
	if assert.NoError(t, err) {
		// 注: 反序列化的两个值顺序不是固定的
		//assert.Equal(t, `{"t":7,"v":{"dict":{"v1":{"t":0,"v":1},"v2":{"t":0,"v":2}}}}`, string(v))
		assert.True(t, string(v) == `{"t":7,"v":{"dict":{"v1":{"t":0,"v":1},"v2":{"t":0,"v":2}}}}` ||
			string(v) == `{"t":7,"v":{"dict":{"v2":{"t":0,"v":2},"v1":{"t":0,"v":1}}}}`)
	}

	// 	递归检测
	v1 := na(ni(1), nf(2.0), ns("test"))
	ad, _ := v1.ReadArray()
	ad.List = append(ad.List, v1)
	v, err = v1.ToJSON() // nolint
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run(`ceil`)
	if assert.NoError(t, err) {
		ret := vm.Ret
		v, err = ret.ToJSON() // nolint
		assert.Equal(t, `{"t":9,"v":{"name":"ceil"}}`, string(v))
	}
}

func TestLoads(t *testing.T) {
	var err error
	var v *VMValue

	v, err = VMValueFromJSON([]byte(`{"t":0,"v":123}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeInt)
		assert.Equal(t, IntType(123), v.Value)
	}

	v, err = VMValueFromJSON([]byte(`{"t":1,"v":3.2}`))
	if assert.NoError(t, err) {
		assert.Equal(t, VMTypeFloat, v.TypeId)
		assert.Equal(t, float64(3.2), v.Value)
	}

	v, err = VMValueFromJSON([]byte(`{"t":2,"v":"asd"}`))
	if assert.NoError(t, err) {
		assert.Equal(t, VMTypeString, v.TypeId)
		assert.Equal(t, "asd", v.Value)
	}

	v, err = VMValueFromJSON([]byte(`{"t":4}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeNull)
	}

	v, err = VMValueFromJSON([]byte(`{"t":5,"v":{"expr":"1 + this.x + d10"}}`))
	if assert.NoError(t, err) {
		assert.Equal(t, VMTypeComputedValue, v.TypeId)
		assert.Equal(t, "1 + this.x + d10", v.Value.(*ComputedData).Expr)
	}

	v, err = VMValueFromJSON([]byte(`{"t":8,"v":{"expr":"return 5 ","name":"a","params":["x"]}}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeFunction)
		fd, _ := v.ReadFunctionData()
		assert.Equal(t, "return 5 ", fd.Expr)
		assert.Equal(t, "a", fd.Name)
		assert.Equal(t, []string{"x"}, fd.Params)
	}

	v, err = VMValueFromJSON([]byte(`{"t":6,"v":{"list":[{"t":0,"v":1},{"t":1,"v":2},{"t":2,"v":"test"}]}}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeArray)
		ad, _ := v.ReadArray()
		assert.True(t, valueEqual(ad.List[0], ni(1)))
		assert.True(t, valueEqual(ad.List[1], nf(2.0)))
		assert.True(t, valueEqual(ad.List[2], ns("test")))
	}

	v, err = VMValueFromJSON([]byte(`{"t":9,"v":{"name":"ceil"}}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeNativeFunction)
		assert.True(t, valueEqual(v, builtinValues["ceil"]))
	}
}

func TestDumpsArray(t *testing.T) {
	v, err := NewArrayVal(ni(1), ni(2), na(ni(3)), ni(4)).ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"t":6,"v":{"list":[{"t":0,"v":1},{"t":0,"v":2},{"t":6,"v":{"list":[{"t":0,"v":3}]}},{"t":0,"v":4}]}}`, string(v))
	}
}

func TestLoadsArray(t *testing.T) {
	v, err := VMValueFromJSON([]byte(`{"t":6,"v":{"list":[{"t":0,"v":1},{"t":0,"v":2},{"t":6,"v":{"list":[{"t":0,"v":3}]}},{"t":0,"v":4}]}}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeArray)
		assert.True(t, valueEqual(v, NewArrayVal(ni(1), ni(2), na(ni(3)), ni(4))))
	}
}

func TestDumpsDict(t *testing.T) {
	m := NewDictVal(nil)
	m.Store("XXX", NewIntVal(222))

	v, err := m.V().ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"t":7,"v":{"dict":{"XXX":{"t":0,"v":222}}}}`, string(v))
	}
}

func TestLoadsDict(t *testing.T) {
	v, err := VMValueFromJSON([]byte(`{"t":7,"v":{"dict":{"XXX":{"t":0,"v":222}}}}`))
	if assert.NoError(t, err) {
		m := NewDictVal(nil)
		m.Store("XXX", NewIntVal(222))

		assert.Equal(t, v.TypeId, VMTypeDict)
		assert.True(t, valueEqual(v, m.V()))
	}
}

func TestNativeTypes(t *testing.T) {
	//vm := NewVM()
	var slot *VMValue
	od := &NativeObjectData{
		Name: "obj1",
		AttrSet: func(ctx *Context, name string, v *VMValue) {
			slot = v
		},
		AttrGet: func(ctx *Context, name string) *VMValue {
			return slot
		},
	}
	nVal := NewNativeObjectVal(od)
	data, err := nVal.ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"t":10,"v":{"name":"obj1"}}`, string(data))
	}

	v, err := VMValueFromJSON([]byte(`{"t":10,"v":{"name":"obj1"}}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeNativeObject)
		assert.Equal(t, v.Value.(*NativeObjectData).Name, "obj1")
	}
}
