package dicescript

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypesFuncDict(t *testing.T) {
	d := NewDictVal(nil)
	d.Store("a", ni(1))

	v, _ := d.Load("a")
	assert.True(t, valueEqual(v, ni(1)))

	v, ok := d.Load("b")
	assert.True(t, v == nil)
	assert.Equal(t, ok, false)

	d2 := (*VMDictValue)(ni(2))
	v, ok = d2.Load("a")
	assert.True(t, v == nil)
	assert.Equal(t, ok, false)
}

func TestTypesFuncDictToStr(t *testing.T) {
	data := &ValueMap{}
	data.Store("a", ni(1))
	d := NewDictVal(data)
	assert.Equal(t, d.ToString(), "{'a': 1}")
}

func TestTypesFuncDictRange(t *testing.T) {
	d := NewDictValWithArrayMust(ns("a"), ni(1), ns("b"), ni(2))
	seen := map[string]IntType{}
	d.Range(func(key string, value *VMValue) bool {
		seen[key] = value.MustReadInt()
		return true
	})
	assert.Equal(t, map[string]IntType{"a": 1, "b": 2}, seen)

	invalid := (*VMDictValue)(ni(2))
	invalid.Range(func(string, *VMValue) bool {
		t.Fatal("callback must not run for a non-dict value")
		return false
	})
}

func TestTypesFuncArray(t *testing.T) {
	vm := NewVM()
	arr := na(ni(1), ni(2), ni(3))
	assert.Equal(t, arr.ArrayItemGet(vm, 1), ni(2))

	ni(1).ArrayItemGet(vm, 1)
	assert.Error(t, vm.Error) // 此类型无法取下标

	vm = NewVM()
	arr = na(ni(1), ni(2), ni(3))
	arr.ArrayItemSet(vm, 1, ni(4))
	assert.Equal(t, arr.MustReadArray().List[1].MustReadInt(), IntType(4))

	vm = NewVM()
	arr = na(ni(1), ni(2), ni(3))
	arr.ArrayItemSet(vm, 3, ni(4))
	assert.Error(t, vm.Error)

	vm = NewVM()
	ni(1).ArrayItemSet(vm, 1, ni(2))
	assert.Error(t, vm.Error) // 此类型无法赋值下标
}
