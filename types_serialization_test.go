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
		assert.Equal(t, string(v), `{"typeId":0,"value":123}`)
	}

	v, err = VMValueNewFloat(3.2).ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, string(v), `{"typeId":1,"value":3.2}`)
	}

	v, err = VMValueNewStr("asd").ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, string(v), `{"typeId":2,"value":"asd"}`)
	}

	v, err = VMValueNewUndefined().ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, string(v), `{"typeId":3}`)
	}

	v, err = VMValueNewNull().ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, string(v), `{"typeId":4}`)
	}

	v, err = VMValueNewComputed("1 + this.x + d10").ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, string(v), `{"typeId":5,"value":{"expr":"1 + this.x + d10"}}`)
	}

	vm, _ := NewVMWithStore(nil)
	err = vm.Run(`func a(x) { return 5 }; a`)
	if assert.NoError(t, err) {
		ret := vm.Ret
		v, err = ret.ToJSON()
		assert.Equal(t, string(v), `{"typeId":8,"value":{"expr":"return 5 ","name":"a","params":["x"]}}`)
	}

	v, err = na(ni(1), nf(2.0), ns("test")).ToJSON()
	if assert.NoError(t, err) {
		assert.Equal(t, string(v), `{"typeId":6,"value":{"list":[{"typeId":0,"value":1},{"typeId":1,"value":2},{"typeId":2,"value":"test"}]}}`)
	}

	// 	递归检测
	v1 := na(ni(1), nf(2.0), ns("test"))
	ad, _ := v1.ReadArray()
	ad.List = append(ad.List, v1)
	v, err = v1.ToJSON()
	assert.Error(t, err)

	vm, _ = NewVMWithStore(nil)
	err = vm.Run(`ceil`)
	if assert.NoError(t, err) {
		ret := vm.Ret
		v, err = ret.ToJSON()
		assert.Equal(t, string(v), `{"typeId":9,"value":{"name":"ceil"}}`)
	}
}

func TestLoads(t *testing.T) {
	var err error
	var v *VMValue

	v, err = VMValueFromJSON([]byte(`{"typeId":0,"value":123}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeInt)
		assert.Equal(t, v.Value, int64(123))
	}

	v, err = VMValueFromJSON([]byte(`{"typeId":1,"value":3.2}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeFloat)
		assert.Equal(t, v.Value, float64(3.2))
	}

	v, err = VMValueFromJSON([]byte(`{"typeId":2,"value":"asd"}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeString)
		assert.Equal(t, v.Value, "asd")
	}

	v, err = VMValueFromJSON([]byte(`{"typeId":3}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeUndefined)
	}

	v, err = VMValueFromJSON([]byte(`{"typeId":4}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeNull)
	}

	v, err = VMValueFromJSON([]byte(`{"typeId":5,"value":{"expr":"1 + this.x + d10"}}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeComputedValue)
		assert.Equal(t, v.Value.(*ComputedData).Expr, "1 + this.x + d10")
	}

	v, err = VMValueFromJSON([]byte(`{"typeId":8,"value":{"expr":"return 5 ","name":"a","params":["x"]}}`))
	if assert.NoError(t, err) {
		assert.Equal(t, v.TypeId, VMTypeFunction)
		fd, _ := v.ReadFunctionData()
		assert.Equal(t, fd.Expr, "return 5 ")
		assert.Equal(t, fd.Name, "a")
		assert.Equal(t, fd.Params, []string{"x"})
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
