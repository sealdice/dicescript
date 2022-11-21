package dicescript

import (
	"errors"
)

func (d *VMDictValue) V() *VMValue {
	return (*VMValue)(d)
}

func (d *VMDictValue) Store(key string, value *VMValue) {
	if dd, ok := d.V().ReadDictData(); ok {
		dd.Dict.Store(key, value)
	}
}

// Load value为变量的值，ok代表是否找到变量
func (d *VMDictValue) Load(key string) (value *VMValue, ok bool) {
	if dd, ok := d.V().ReadDictData(); ok {
		return dd.Dict.Load(key)
	}
	return nil, false
}

func (d *VMDictValue) ToString() string {
	return d.V().ToString()
}

func (v *VMValue) ArrayItemGet(ctx *Context, index int64) *VMValue {
	if v.TypeId == VMTypeArray {
		arr, _ := v.ReadArray()
		index = getRealIndex(ctx, index, int64(len(arr.List)))
		if ctx.Error != nil {
			return nil
		}
		return arr.List[index]
	}
	ctx.Error = errors.New("此类型无法取下标")
	return nil
}

func (v *VMValue) ArrayItemSet(ctx *Context, index int64, val *VMValue) bool {
	if v.TypeId == VMTypeArray {
		arr, _ := v.ReadArray()
		index = getRealIndex(ctx, index, int64(len(arr.List)))
		if ctx.Error != nil {
			return false
		}
		arr.List[index] = val.Clone()
		return true
	}
	ctx.Error = errors.New("此类型无法赋值下标")
	return false
}

func (v *VMValue) ArrayFuncKeepHigh(ctx *Context) *VMValue {
	arr, _ := v.ReadArray()

	var maxFloat float64 // 次函数最大上限为flaot64上限
	isFloat := false
	isFirst := true

	for _, i := range arr.List {
		switch i.TypeId {
		case VMTypeInt:
			if isFirst {
				isFirst = false
				maxFloat = float64(i.Value.(int64))
			} else {
				val := float64(i.Value.(int64))
				if val > maxFloat {
					maxFloat = val
				}
			}
		case VMTypeFloat:
			isFloat = true
			if isFirst {
				isFirst = false
				maxFloat = i.Value.(float64)
			} else {
				val := i.Value.(float64)
				if val > maxFloat {
					maxFloat = val
				}
			}
		}
	}

	if isFloat {
		return VMValueNewFloat(maxFloat)
	} else {
		return VMValueNewInt(int64(maxFloat))
	}
}

func (v *VMValue) ArrayFuncKeepLow(ctx *Context) *VMValue {
	arr, _ := v.ReadArray()

	var maxFloat float64 // 次函数最大上限为flaot64上限
	isFloat := false
	isFirst := true

	for _, i := range arr.List {
		switch i.TypeId {
		case VMTypeInt:
			if isFirst {
				isFirst = false
				maxFloat = float64(i.Value.(int64))
			} else {
				val := float64(i.Value.(int64))
				if val < maxFloat {
					maxFloat = val
				}
			}
		case VMTypeFloat:
			isFloat = true
			if isFirst {
				isFirst = false
				maxFloat = i.Value.(float64)
			} else {
				val := i.Value.(float64)
				if val < maxFloat {
					maxFloat = val
				}
			}
		}
	}

	if isFloat {
		return VMValueNewFloat(maxFloat)
	} else {
		return VMValueNewInt(int64(maxFloat))
	}
}
