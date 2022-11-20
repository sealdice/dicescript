package dicescript

import "errors"

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
