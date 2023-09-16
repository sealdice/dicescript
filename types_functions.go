package dicescript

import (
	"errors"
	"sort"
)

func (d *VMDictValue) V() *VMValue {
	return (*VMValue)(d)
}

func (d *VMDictValue) Store(key string, value *VMValue) {
	if dd, ok := d.V().ReadDictData(); ok {
		dd.Dict.Store(key, value)
	}
}

func (d *VMDictValue) Range(callback func(key string, value *VMValue) bool) {
	if dd, ok := d.V().ReadDictData(); ok {
		dd.Dict.Range(callback)
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

func (v *VMValue) ArrayFuncKeepBase(ctx *Context, pickNum int64, orderType int) (isAllInt bool, ret float64) {
	arr, _ := v.ReadArray()
	var nums []float64

	isAllInt = true
	for _, i := range arr.List {
		switch i.TypeId {
		case VMTypeInt:
			nums = append(nums, float64(i.MustReadInt()))
		case VMTypeFloat:
			isAllInt = false
			nums = append(nums, i.MustReadFloat())
		}
	}

	if orderType == 0 {
		sort.Slice(nums, func(i, j int) bool { return nums[i] > nums[j] }) // 从大到小
	} else if orderType == 1 {
		sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] }) // 从小到大
	}

	num := float64(0)
	for i := int64(0); i < pickNum; i++ {
		// 当取数大于上限 跳过
		if i >= int64(len(nums)) {
			continue
		}
		num += nums[i]
	}

	return isAllInt, num
}

func (v *VMValue) ArrayFuncKeepHigh(ctx *Context, pickNum int64) (isAllInt bool, ret float64) {
	return v.ArrayFuncKeepBase(ctx, pickNum, 0)
}

func (v *VMValue) ArrayFuncKeepLow(ctx *Context, pickNum int64) (isAllInt bool, ret float64) {
	return v.ArrayFuncKeepBase(ctx, pickNum, 1)
}
