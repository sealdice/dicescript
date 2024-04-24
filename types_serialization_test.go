package dicescript

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDumps(t *testing.T) {
	var err error
	var v []byte

	v, err = VMValueNewInt(123).ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"typeId":0,"value":123}`, string(v))
	}

	v, err = VMValueNewFloat(3.2).ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"typeId":1,"value":3.2}`, string(v))
	}

	v, err = VMValueNewStr("asd").ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"typeId":2,"value":"asd"}`, string(v))
	}

	v, err = VMValueNewUndefined().ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"typeId":3}`, string(v))
	}

	v, err = VMValueNewNull().ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"typeId":4}`, string(v))
	}

	v, err = VMValueNewComputed("1 + this.x + d10").ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"typeId":5,"value":{"expr":"1 + this.x + d10"}}`, string(v))
	}

	vm := NewVM()
	err = vm.Run(`func a(x) { return 5 }; a`)
	if assert.NoError(t, err) {
		ret := vm.Ret
		v, err = ret.ToJSON() // nolint
		assert.Equal(t, `{"typeId":8,"value":{"expr":"return 5 ","name":"a","params":["x"]}}`, string(v))
	}

	v, err = na(ni(1), nf(2.0), ns("test")).ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"typeId":6,"value":{"list":[{"typeId":0,"value":1},{"typeId":1,"value":2},{"typeId":2,"value":"test"}]}}`, string(v))
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
		assert.Equal(t, `{"typeId":9,"value":{"name":"ceil"}}`, string(v))
	}
}

func TestLoads(t *testing.T) {
	var err error
	var v *VMValue

	v, err = VMValueFromJSON([]byte(`{"typeId":0,"value":123}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeInt)
		assert.Equal(t, IntType(123), v.Value)
	}

	v, err = VMValueFromJSON([]byte(`{"typeId":1,"value":3.2}`))
	if assert.NoError(t, err) {
		assert.Equal(t, VMTypeFloat, v.TypeId)
		assert.Equal(t, float64(3.2), v.Value)
	}

	v, err = VMValueFromJSON([]byte(`{"typeId":2,"value":"asd"}`))
	if assert.NoError(t, err) {
		assert.Equal(t, VMTypeString, v.TypeId)
		assert.Equal(t, "asd", v.Value)
	}

	v, err = VMValueFromJSON([]byte(`{"typeId":3}`))
	if assert.NoError(t, err) {
		assert.Equal(t, VMTypeUndefined, v.TypeId)
	}

	v, err = VMValueFromJSON([]byte(`{"typeId":4}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeNull)
	}

	v, err = VMValueFromJSON([]byte(`{"typeId":5,"value":{"expr":"1 + this.x + d10"}}`))
	if assert.NoError(t, err) {
		assert.Equal(t, VMTypeComputedValue, v.TypeId)
		assert.Equal(t, "1 + this.x + d10", v.Value.(*ComputedData).Expr)
	}

	v, err = VMValueFromJSON([]byte(`{"typeId":8,"value":{"expr":"return 5 ","name":"a","params":["x"]}}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeFunction)
		fd, _ := v.ReadFunctionData()
		assert.Equal(t, "return 5 ", fd.Expr)
		assert.Equal(t, "a", fd.Name)
		assert.Equal(t, []string{"x"}, fd.Params)
	}

	v, err = VMValueFromJSON([]byte(`{"typeId":6,"value":{"list":[{"typeId":0,"value":1},{"typeId":1,"value":2},{"typeId":2,"value":"test"}]}}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeArray)
		ad, _ := v.ReadArray()
		assert.True(t, valueEqual(ad.List[0], ni(1)))
		assert.True(t, valueEqual(ad.List[1], nf(2.0)))
		assert.True(t, valueEqual(ad.List[2], ns("test")))
	}

	v, err = VMValueFromJSON([]byte(`{"typeId":9,"value":{"name":"ceil"}}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeNativeFunction)
		assert.True(t, valueEqual(v, builtinValues["ceil"]))
	}
}

func TestDumpsArray(t *testing.T) {
	v, err := VMValueNewArray(ni(1), ni(2), na(ni(3)), ni(4)).ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"typeId":6,"value":{"list":[{"typeId":0,"value":1},{"typeId":0,"value":2},{"typeId":6,"value":{"list":[{"typeId":0,"value":3}]}},{"typeId":0,"value":4}]}}`, string(v))
	}
}

func TestLoadsArray(t *testing.T) {
	v, err := VMValueFromJSON([]byte(`{"typeId":6,"value":{"list":[{"typeId":0,"value":1},{"typeId":0,"value":2},{"typeId":6,"value":{"list":[{"typeId":0,"value":3}]}},{"typeId":0,"value":4}]}}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeArray)
		assert.True(t, valueEqual(v, VMValueNewArray(ni(1), ni(2), na(ni(3)), ni(4))))
	}
}

func TestDumpsDict(t *testing.T) {
	m := VMValueNewDict(nil)
	m.Store("XXX", VMValueNewInt(222))

	v, err := m.V().ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"typeId":7,"value":{"dict":{"XXX":{"typeId":0,"value":222}}}}`, string(v))
	}
}

func TestLoadsDict(t *testing.T) {
	v, err := VMValueFromJSON([]byte(`{"typeId":7,"value":{"dict":{"XXX":{"typeId":0,"value":222}}}}`))
	if assert.NoError(t, err) {
		m := VMValueNewDict(nil)
		m.Store("XXX", VMValueNewInt(222))

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
	nVal := VMValueNewNativeObject(od)
	data, err := nVal.ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, `{"typeId":10,"value":{"name":"obj1"}}`, string(data))
	}

	v, err := VMValueFromJSON([]byte(`{"typeId":10,"value":{"name":"obj1"}}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeNativeObject)
		assert.Equal(t, v.Value.(*NativeObjectData).Name, "obj1")
	}
}
