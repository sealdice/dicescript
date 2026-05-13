//go:build js
// +build js

package main

import (
	//"regexp"
	//"strconv"

	"github.com/gopherjs/gopherjs/js"
	ds "github.com/sealdice/dicescript"
)

var scope = map[string]*ds.VMValue{}

func newVM(name string) *js.Object {
	actor := ds.NewDictVal(nil)
	actor.Store("力量", ds.NewIntVal(50))
	actor.Store("敏捷", ds.NewIntVal(60))
	actor.Store("智力", ds.NewIntVal(70))
	scope["actor"] = actor.V()

	vm := ds.NewVM()
	//vm.GlobalValueStoreFunc = func(name string, v *ds.VMValue) {
	//	scope[name] = v
	//}

	//re := regexp.MustCompile(`^_(\D+)(\d+)$`)
	//vm.GlobalValueLoadFunc = func(name string) *ds.VMValue {
	//	m := re.FindStringSubmatch(name)
	//	if len(m) > 1 {
	//		val, _ := strconv.ParseInt(m[2], 10, 64)
	//		return ds.VMValueNewInt(ds.IntType(val))
	//	}
	//
	//	if v, exists := actor.Load(name); exists {
	//		return v
	//	}
	//
	//	if val, ok := scope[name]; ok {
	//		return val
	//	}
	//	return nil
	//}

	return js.MakeFullWrapper(vm)
}

func main() {
	diceModule := map[string]interface{}{
		"newVMForPlaygournd": newVM,
		"newVM": func() *js.Object {
			vm := ds.NewVM()
			return js.MakeFullWrapper(vm)
		},
		"newConfig": func() *js.Object {
			return js.MakeFullWrapper(&ds.RollConfig{})
		},
		"newValueMap": func() *js.Object {
			return js.MakeFullWrapper(&ds.ValueMap{})
		},
		"vmNewInt": func(i ds.IntType) *js.Object {
			return js.MakeFullWrapper(ds.NewIntVal(i))
		},
		"vmNewFloat": func(i float64) *js.Object {
			return js.MakeFullWrapper(ds.NewFloatVal(i))
		},
		"vmNewStr": func(s string) *js.Object {
			return js.MakeFullWrapper(ds.NewStrVal(s))
		},
		//"vmNewArray":    js.MakeWrapper(newArray),
		"vmNewDict": func() *js.Object {
			return js.MakeFullWrapper(ds.NewDictVal(nil))
		},
		"help": "此项目的js绑定: https://github.com/sealdice/dice",
	}

	js.Module.Get("exports").Set("ds", diceModule)
}
