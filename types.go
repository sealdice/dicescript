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
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type VMValueType int

const (
	VMTypeInt            VMValueType = 0
	VMTypeFloat          VMValueType = 1
	VMTypeString         VMValueType = 2
	VMTypeUndefined      VMValueType = 3
	VMTypeNull           VMValueType = 4
	VMTypeComputedValue  VMValueType = 5
	VMTypeArray          VMValueType = 6
	VMTypeDict           VMValueType = 7
	VMTypeFunction       VMValueType = 8
	VMTypeNativeFunction VMValueType = 9
	VMTypeNativeObject   VMValueType = 10

	// 内部对象
	vmTypeLocal  VMValueType = 20
	vmTypeGlobal VMValueType = 21
)

var binOperator = []func(*VMValue, *Context, *VMValue) *VMValue{
	(*VMValue).OpAdd,
	(*VMValue).OpSub,
	(*VMValue).OpMultiply,
	(*VMValue).OpDivide,
	(*VMValue).OpModulus,
	(*VMValue).OpPower,

	(*VMValue).OpCompLT,
	(*VMValue).OpCompLE,
	(*VMValue).OpCompEQ,
	(*VMValue).OpCompNE,
	(*VMValue).OpCompGE,
	(*VMValue).OpCompGT,
}

type RollExtraFlags struct {
	DiceMinMode         bool   // 骰子以最小值结算，用于获取下界
	DiceMaxMode         bool   // 以最大值结算 获取上界
	DisableLoadVarname  bool   // 不允许加载变量，这是为了防止遇到 .r XXX 被当做属性读取，而不是“由于XXX，骰出了”
	IgnoreDiv0          bool   // 当div0时暂不报错
	DefaultDiceSideNum  int64  // 默认骰子面数
	DefaultDiceSideExpr string // 默认骰子面数
	PrintBytecode       bool   // 执行时打印字节码
}

type Context struct {
	parser         *Parser
	subThreadDepth int
	attrs          *ValueMap
	upCtx          *Context
	//subThread      *Context // 用于执行子句

	/** 全局变量 */
	globalNames *ValueMap

	code      []ByteCode
	codeIndex int

	stack []VMValue
	top   int

	NumOpCount int64 // 算力计数
	//CocFlagVarPrefix string // 解析过程中出现，当VarNumber开启时有效，可以是困难极难常规大成功

	Flags RollExtraFlags // 标记
	Error error          // 报错信息

	Ret       *VMValue // 返回值
	RestInput string   // 剩余字符串
	Matched   string   // 匹配的字符串

	//lastDetails := []string{}
	//lastDetailsLeft := []string{}
	//calcDetail := ""

	/* 如果返回值为true，那么不作其他处理 */
	ValueStoreHookFunc func(ctx *Context, name string, v *VMValue) (solved bool)

	ValueStoreFunc func(name string, v *VMValue)
	ValueLoadFunc  func(name string) *VMValue
}

func (e *Context) Init() {
	e.attrs = &ValueMap{}
	e.globalNames = &ValueMap{}
}

func (e *Context) loadInnerVar(name string) *VMValue {
	return builtinValues[name]
}

func (ctx *Context) LoadNameGlobal(name string, isRaw bool) *VMValue {
	var loadFunc func(name string) *VMValue
	if loadFunc == nil {
		loadFunc = ctx.ValueLoadFunc
	}

	// 检测全局表
	if loadFunc != nil {
		val := loadFunc(name)
		if val != nil {
			if !isRaw && val.TypeId == VMTypeComputedValue {
				val = val.ComputedExecute(ctx)
				if ctx.Error != nil {
					return nil
				}
			}
			return val
		}
	}
	//else {
	//	ctx.Error = errors.New("未设置 ValueLoadFunc，无法获取变量")
	//	return nil
	//}

	// 检测内置变量/函数检查
	val := ctx.loadInnerVar(name)
	if val == nil {
		val = VMValueNewUndefined()
	}
	if !isRaw && val.TypeId == VMTypeComputedValue {
		val = val.ComputedExecute(ctx)
		if ctx.Error != nil {
			return nil
		}
	}
	return val
}

func (ctx *Context) LoadNameLocal(name string, isRaw bool) *VMValue {
	//if ctx.currentThis != nil {
	//	return ctx.currentThis.AttrGet(ctx, name)
	//} else {
	//if ctx.subThreadDepth >= 1 {
	ret, exists := ctx.attrs.Load(name)
	if !exists {
		ret = VMValueNewUndefined()
	}
	if !isRaw && ret.TypeId == VMTypeComputedValue {
		ret = ret.ComputedExecute(ctx)
		if ctx.Error != nil {
			return nil
		}
	}
	return ret
	//}
	//}
}

func (ctx *Context) LoadName(name string, isRaw bool) *VMValue {
	//fmt.Println("!!!!!!", name)
	// 先local再global
	curCtx := ctx
	for {
		ret := curCtx.LoadNameLocal(name, isRaw)

		if curCtx.Error != nil {
			ctx.Error = curCtx.Error
			return nil
		}
		if ret.TypeId != VMTypeUndefined {
			return ret
		}
		if curCtx.upCtx == nil {
			break
		} else {
			curCtx = curCtx.upCtx
		}
	}

	return ctx.LoadNameGlobal(name, isRaw)
	//if ctx.ValueLoadFunc != nil {
	//	ret := ctx.ValueLoadFunc(name)
	//	if ctx.Error != nil {
	//		return nil
	//	}
	//	if ret != nil {
	//		if !isRaw && ret.TypeId == VMTypeComputedValue {
	//			ret = ret.ComputedExecute(ctx)
	//			if ctx.Error != nil {
	//				return nil
	//			}
	//		}
	//		return ret
	//	}
	//}
	//return VMValueNewUndefined()
}

// StoreName 储存变量
func (ctx *Context) StoreName(name string, v *VMValue) {
	if _, ok := ctx.globalNames.Load(name); ok {
		ctx.StoreNameGlobal(name, v)
	} else {
		ctx.StoreNameLocal(name, v)
	}
}

func (ctx *Context) StoreNameLocal(name string, v *VMValue) {
	//fmt.Println("XXXXXX", name, v)
	ctx.attrs.Store(name, v.Clone())
}

func (ctx *Context) StoreNameGlobal(name string, v *VMValue) {
	storeFunc := ctx.ValueStoreFunc
	if storeFunc != nil {
		storeFunc(name, v.Clone())
	} else {
		ctx.Error = errors.New("未设置 ValueStoreNameFunc，无法储存变量")
		return
	}

}

type VMValue struct {
	TypeId VMValueType `json:"typeId"`
	Value  interface{} `json:"value"`
	//ExpiredTime int64       `json:"expiredTime"`
}

type VMDictValue VMValue

type ArrayData struct {
	List []*VMValue
}

type DictData struct {
	Dict *ValueMap
}

type ComputedData struct {
	Expr string

	/* 缓存数据 */
	Attrs     *ValueMap
	code      []ByteCode
	codeIndex int
}

type FunctionData struct {
	Expr   string
	Name   string
	Params []string

	/* 缓存数据 */
	code      []ByteCode
	codeIndex int
	//ctx       *Context
}

type NativeFunctionData struct {
	Name   string
	Params []string

	/* 缓存数据 */
	NativeFunc func(ctx *Context, params []*VMValue) *VMValue
}

type NativeObjectData interface {
	AttrSet(name string, v *VMValue)
	AttrGet(name string) *VMValue
	ItemSet(name string, v *VMValue)
	ItemGet(name string, v *VMValue) *VMValue
	ToString() string
}

func (v *VMValue) Clone() *VMValue {
	vNew := &VMValue{TypeId: v.TypeId, Value: v.Value}
	// TODO: 针对特定类型，进行Value的处理，不过大多数时候应该够用
	switch v.TypeId {
	}
	return vNew
}

func (v *VMValue) AsBool() bool {
	switch v.TypeId {
	case VMTypeInt:
		return v.Value != int64(0)
	case VMTypeString:
		return v.Value != ""
	case VMTypeNull, VMTypeUndefined:
		return false
	//case VMTypeComputedValue:
	//	vd := v.Value.(*VMComputedValueData)
	//	return vd.BaseValue.AsBool()
	default:
		return false
	}
}

func (v *VMValue) ToString() string {
	if v == nil {
		return "NIL"
	}
	switch v.TypeId {
	case VMTypeInt:
		return strconv.FormatInt(v.Value.(int64), 10)
	case VMTypeFloat:
		return strconv.FormatFloat(v.Value.(float64), 'f', -1, 64)
	case VMTypeString:
		return v.Value.(string)
	case VMTypeUndefined:
		return "undefined"
	case VMTypeNull:
		return "null"
	case VMTypeArray:
		s := "["
		arr, _ := v.ReadArray()
		for index, i := range arr.List {
			if i.TypeId == VMTypeArray {
				s += "[...]"
			} else if i.TypeId == VMTypeDict {
				s += "{...}"
			} else {
				s += i.ToRepr()
			}
			if index != len(arr.List)-1 {
				s += ", "
			}
		}
		s += "]"
		return s
	case VMTypeComputedValue:
		cd, _ := v.ReadComputed()
		return "&(" + cd.Expr + ")"
	case VMTypeDict:
		dd, _ := v.ReadDictData()
		items := []string{}

		dd.Dict.Range(func(key string, value *VMValue) bool {
			txt := ""
			if value.TypeId == VMTypeArray {
				txt = "[...]"
			} else if value.TypeId == VMTypeDict {
				txt = "{...}"
			} else {
				txt = value.ToRepr()
			}
			items = append(items, fmt.Sprintf("'%s': %s", key, txt))
			return true
		})
		return "{" + strings.Join(items, ", ") + "}"
	case VMTypeFunction:
		cd, _ := v.ReadFunctionData()
		return "function " + cd.Name
	case VMTypeNativeFunction:
		cd, _ := v.ReadNativeFunctionData()
		return "nfunction " + cd.Name
	default:
		return "a value"
	}
}

func (v *VMValue) ToRepr() string {
	if v == nil {
		return "NIL"
	}
	switch v.TypeId {
	case VMTypeString:
		// TODO: 检测其中是否有"
		return "'" + v.ToString() + "'"
	case VMTypeInt, VMTypeFloat, VMTypeUndefined, VMTypeNull, VMTypeArray, VMTypeComputedValue, VMTypeDict, VMTypeFunction, VMTypeNativeFunction:
		return v.ToString()
	default:
		return "<a value>"
	}
}

func (v *VMValue) ReadInt() (int64, bool) {
	if v.TypeId == VMTypeInt {
		return v.Value.(int64), true
	}
	return 0, false
}

func (v *VMValue) ReadFloat() (float64, bool) {
	if v.TypeId == VMTypeFloat {
		return v.Value.(float64), true
	}
	return 0, false
}

func (v *VMValue) ReadString() (string, bool) {
	if v.TypeId == VMTypeString {
		return v.Value.(string), true
	}
	return "", false
}

func (v *VMValue) ReadArray() (*ArrayData, bool) {
	if v.TypeId == VMTypeArray {
		return v.Value.(*ArrayData), true
	}
	return nil, false
}

func (v *VMValue) ReadComputed() (*ComputedData, bool) {
	if v.TypeId == VMTypeComputedValue {
		return v.Value.(*ComputedData), true
	}
	return nil, false
}

func (v *VMValue) ReadDictData() (*DictData, bool) {
	if v.TypeId == VMTypeDict {
		return v.Value.(*DictData), true
	}
	return nil, false
}

func (v *VMValue) MustReadDictData() *DictData {
	if v.TypeId == VMTypeDict {
		return v.Value.(*DictData)
	}
	panic("bad type")
}

func (v *VMValue) MustReadInt() int64 {
	val, ok := v.ReadInt()
	if ok {
		return val
	}
	panic("bad type")
}

func (v *VMValue) ReadFunctionData() (*FunctionData, bool) {
	if v.TypeId == VMTypeFunction {
		return v.Value.(*FunctionData), true
	}
	return nil, false
}

func (v *VMValue) ReadNativeFunctionData() (*NativeFunctionData, bool) {
	if v.TypeId == VMTypeNativeFunction {
		return v.Value.(*NativeFunctionData), true
	}
	return nil, false
}

func (v *VMValue) OpAdd(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			val := v.Value.(int64) + v2.Value.(int64)
			return VMValueNewInt(val)
		case VMTypeFloat:
			val := float64(v.Value.(int64)) + v2.Value.(float64)
			return VMValueNewFloat(val)
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			val := v.Value.(float64) + float64(v2.Value.(int64))
			return VMValueNewFloat(val)
		case VMTypeFloat:
			val := v.Value.(float64) + v2.Value.(float64)
			return VMValueNewFloat(val)
		}
	case VMTypeString:
		switch v2.TypeId {
		case VMTypeString:
			val := v.Value.(string) + v2.Value.(string)
			return VMValueNewStr(val)
		}
	case VMTypeArray:
		switch v2.TypeId {
		case VMTypeArray:
			arr, _ := v.ReadArray()
			arr2, _ := v2.ReadArray()

			length := len(arr.List) + len(arr2.List)
			if length > 512 {
				ctx.Error = errors.New("不能一次性创建过长的数组")
				return nil
			}

			arrFinal := make([]*VMValue, len(arr.List)+len(arr2.List))
			copy(arrFinal, arr.List)
			for index, i := range arr2.List {
				arrFinal[len(arr.List)+index] = i
			}
			return VMValueNewArray(arrFinal...)
		}
	}

	return nil
}

func (v *VMValue) OpSub(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			val := v.Value.(int64) - v2.Value.(int64)
			return VMValueNewInt(val)
		case VMTypeFloat:
			val := float64(v.Value.(int64)) - v2.Value.(float64)
			return VMValueNewFloat(val)
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			val := v.Value.(float64) - float64(v2.Value.(int64))
			return VMValueNewFloat(val)
		case VMTypeFloat:
			val := v.Value.(float64) - v2.Value.(float64)
			return VMValueNewFloat(val)
		}
	}

	return nil
}

func (v *VMValue) OpMultiply(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			// TODO: 溢出，均未考虑溢出
			val := v.Value.(int64) * v2.Value.(int64)
			return VMValueNewInt(val)
		case VMTypeFloat:
			val := float64(v.Value.(int64)) * v2.Value.(float64)
			return VMValueNewFloat(val)
		case VMTypeArray:
			return v2.ArrayRepeatTimesEx(ctx, v)
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			val := v.Value.(float64) * float64(v2.Value.(int64))
			return VMValueNewFloat(val)
		case VMTypeFloat:
			val := v.Value.(float64) * v2.Value.(float64)
			return VMValueNewFloat(val)
		}
	case VMTypeArray:
		return v.ArrayRepeatTimesEx(ctx, v2)
	}

	return nil
}

func (v *VMValue) OpDivide(ctx *Context, v2 *VMValue) *VMValue {
	// TODO: 被除数为0
	setDivideZero := func() {
		ctx.Error = errors.New("被除数被0")
	}

	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			if v2.Value.(int64) == 0 {
				setDivideZero()
				return nil
			}
			val := v.Value.(int64) / v2.Value.(int64)
			return VMValueNewInt(val)
		case VMTypeFloat:
			if v2.Value.(float64) == 0 {
				setDivideZero()
				return nil
			}
			val := float64(v.Value.(int64)) / v2.Value.(float64)
			return VMValueNewFloat(val)
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			if v2.Value.(int64) == 0 {
				setDivideZero()
				return nil
			}
			val := v.Value.(float64) / float64(v2.Value.(int64))
			return VMValueNewFloat(val)
		case VMTypeFloat:
			if v2.Value.(float64) == 0 {
				setDivideZero()
				return nil
			}
			val := v.Value.(float64) / v2.Value.(float64)
			return VMValueNewFloat(val)
		}
	}

	return nil
}

func (v *VMValue) OpModulus(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			val := v.Value.(int64) % v2.Value.(int64)
			return VMValueNewInt(val)
		}
	}

	return nil
}

func (v *VMValue) OpPower(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			val := int64(math.Pow(float64(v.Value.(int64)), float64(v2.Value.(int64))))
			return VMValueNewInt(val)
		case VMTypeFloat:
			val := math.Pow(float64(v.Value.(int64)), v2.Value.(float64))
			return VMValueNewFloat(val)
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			val := math.Pow(v.Value.(float64), float64(v2.Value.(int64)))
			return VMValueNewFloat(val)
		case VMTypeFloat:
			val := math.Pow(v.Value.(float64), v2.Value.(float64))
			return VMValueNewFloat(val)
		}
	}

	return nil
}

func boolToVMValue(v bool) *VMValue {
	var val int64
	if v {
		val = 1
	}
	return VMValueNewInt(val)
}

func (v *VMValue) OpCompLT(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			return boolToVMValue(v.Value.(int64) < v2.Value.(int64))
		case VMTypeFloat:
			return boolToVMValue(float64(v.Value.(int64)) < v2.Value.(float64))
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			return boolToVMValue(v.Value.(float64) < float64(v2.Value.(int64)))
		case VMTypeFloat:
			return boolToVMValue(v.Value.(float64) < v2.Value.(float64))
		}
	}

	return nil
}

func (v *VMValue) OpCompLE(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			return boolToVMValue(v.Value.(int64) <= v2.Value.(int64))
		case VMTypeFloat:
			return boolToVMValue(float64(v.Value.(int64)) <= v2.Value.(float64))
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			return boolToVMValue(v.Value.(float64) <= float64(v2.Value.(int64)))
		case VMTypeFloat:
			return boolToVMValue(v.Value.(float64) <= v2.Value.(float64))
		}
	}

	return nil
}

func (v *VMValue) OpCompEQ(ctx *Context, v2 *VMValue) *VMValue {
	if v == v2 {
		return VMValueNewInt(1)
	}
	if v.TypeId == v2.TypeId {
		return boolToVMValue(v.Value == v2.Value)
	}

	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeFloat:
			return boolToVMValue(float64(v.Value.(int64)) == v2.Value.(float64))
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			return boolToVMValue(v.Value.(float64) == float64(v2.Value.(int64)))
		}
	}

	return VMValueNewInt(0)
}

func (v *VMValue) OpCompNE(ctx *Context, v2 *VMValue) *VMValue {
	ret := v.OpCompEQ(ctx, v2)
	return boolToVMValue(!ret.AsBool())
}

func (v *VMValue) OpCompGE(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			return boolToVMValue(v.Value.(int64) >= v2.Value.(int64))
		case VMTypeFloat:
			return boolToVMValue(float64(v.Value.(int64)) >= v2.Value.(float64))
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			return boolToVMValue(v.Value.(float64) >= float64(v2.Value.(int64)))
		case VMTypeFloat:
			return boolToVMValue(v.Value.(float64) >= v2.Value.(float64))
		}
	}

	return nil
}

func (v *VMValue) OpCompGT(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			return boolToVMValue(v.Value.(int64) > v2.Value.(int64))
		case VMTypeFloat:
			return boolToVMValue(float64(v.Value.(int64)) > v2.Value.(float64))
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			return boolToVMValue(v.Value.(float64) > float64(v2.Value.(int64)))
		case VMTypeFloat:
			return boolToVMValue(v.Value.(float64) > v2.Value.(float64))
		}
	}

	return nil
}

func (v *VMValue) OpPositive() *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		return VMValueNewInt(v.Value.(int64))
	case VMTypeFloat:
		return VMValueNewFloat(v.Value.(float64))
	}
	return nil
}

func (v *VMValue) OpNegation() *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		return VMValueNewInt(-v.Value.(int64))
	case VMTypeFloat:
		return VMValueNewFloat(-v.Value.(float64))
	}
	return nil
}

func (v *VMValue) AttrSet(name string, val *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeComputedValue:
		cd, _ := v.ReadComputed()
		if cd.Attrs == nil {
			cd.Attrs = &ValueMap{}
		}
		cd.Attrs.Store(name, val.Clone())
		return val
	case VMTypeDict:
		d := (*VMDictValue)(v)
		d.Store(name, val)
		return val
	}

	return nil
}

func (v *VMValue) AttrGet(ctx *Context, name string) *VMValue {
	switch v.TypeId {
	case VMTypeComputedValue:
		cd, _ := v.ReadComputed()
		var ret *VMValue
		if cd.Attrs != nil {
			ret, _ = cd.Attrs.Load(name)
		}
		if ret == nil {
			ret = VMValueNewUndefined()
		}
		return ret
	case VMTypeDict:
		a := (*VMDictValue)(v)
		ret, _ := a.Load(name)
		if ret == nil {
			ret = VMValueNewUndefined()
		}
		return ret
	case vmTypeGlobal:
		// 加载全局变量
		ret := ctx.LoadNameGlobal(name, false)
		if ret == nil {
			ret = VMValueNewUndefined()
		}
		return ret
	case vmTypeLocal:
		ret := ctx.LoadNameLocal(name, false)
		if ret == nil {
			ret = VMValueNewUndefined()
		}
		return ret
	}

	return nil
}

func (v *VMValue) ItemGet(ctx *Context, index *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeArray:
		if index.TypeId != VMTypeInt {
			ctx.Error = errors.New(fmt.Sprintf("类型错误: 数字下标必须为数字，不能为 %s", index.GetTypeName()))
		} else {
			return v.ArrayItemGet(ctx, index.MustReadInt())
		}
	case VMTypeDict:
		if key, err := index.AsDictKey(); err != nil {
			ctx.Error = err
		} else {
			val, _ := (*VMDictValue)(v).Load(key)
			return val
		}
	case VMTypeUndefined, VMTypeNull:
		ctx.Error = errors.New("此类型无法取下标")
	}
	return nil
}

func (v *VMValue) ItemSet(ctx *Context, index *VMValue, val *VMValue) bool {
	switch v.TypeId {
	case VMTypeArray:
		if index.TypeId != VMTypeInt {
			ctx.Error = errors.New(fmt.Sprintf("类型错误: 数字下标必须为数字，不能为 %s", index.GetTypeName()))
		} else {
			return v.ArrayItemSet(ctx, index.MustReadInt(), val)
		}
	case VMTypeDict:
		if key, err := index.AsDictKey(); err != nil {
			ctx.Error = err
		} else {
			(*VMDictValue)(v).Store(key, val)
			return true
		}
	default:
		ctx.Error = errors.New("此类型无法赋值下标")
	}
	return false
}

func (v *VMValue) CallFunc(ctx *Context, name string, values []*VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeArray:
		switch name {
		case "kh":
			return v.ArrayFuncKeepHigh(ctx)
		case "kl":
			return v.ArrayFuncKeepLow(ctx)
		}
	}
	return VMValueNewUndefined()
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

func getClampRealIndex(ctx *Context, index int64, length int64) int64 {
	if index < 0 {
		// 负数下标支持
		index = length + index
	}
	if index < 0 {
		index = 0
	}

	if index > length {
		index = length
	}
	return index
}

func getRealIndex(ctx *Context, index int64, length int64) int64 {
	if index < 0 {
		// 负数下标支持
		index = length + index
	}
	if index >= length || index < 0 {
		ctx.Error = errors.New("无法获取此下标")
	}
	return index
}

func (v *VMValue) GetSlice(ctx *Context, a int64, b int64, step int64) *VMValue {
	length := v.Length(ctx)
	if ctx.Error != nil {
		return nil
	}

	_a := getClampRealIndex(ctx, a, length)
	_b := getClampRealIndex(ctx, b, length)

	if _a > _b {
		_a = _b
	}

	switch v.TypeId {
	case VMTypeString:
		str, _ := v.ReadString()
		newArr := str[_a:_b]
		return VMValueNewStr(newArr)
	case VMTypeArray:
		arr, _ := v.ReadArray()
		newArr := arr.List[_a:_b]
		return VMValueNewArray(newArr...)
	default:
		ctx.Error = errors.New("这个类型无法取得分片")
		return nil
	}
}

func (v *VMValue) Length(ctx *Context) int64 {
	var length int64

	switch v.TypeId {
	case VMTypeArray:
		arr, _ := v.ReadArray()
		length = int64(len(arr.List))
	case VMTypeString:
		str, _ := v.ReadString()
		length = int64(len(str))
	default:
		ctx.Error = errors.New("这个类型无法取得分片")
		return 0
	}

	return length
}

func (v *VMValue) GetSliceEx(ctx *Context, a *VMValue, b *VMValue) *VMValue {
	if a.TypeId == VMTypeUndefined {
		a = VMValueNewInt(0)
	}

	length := v.Length(ctx)
	if ctx.Error != nil {
		return nil
	}

	if b.TypeId == VMTypeUndefined {
		b = VMValueNewInt(length)
	}

	valA, ok := a.ReadInt()
	if !ok {
		ctx.Error = errors.New("第一个值类型错误")
		return nil
	}

	valB, ok := b.ReadInt()
	if !ok {
		ctx.Error = errors.New("第二个值类型错误")
		return nil
	}

	return v.GetSlice(ctx, valA, valB, 1)
}

func (v *VMValue) SetSlice(ctx *Context, a int64, b int64, step int64, val *VMValue) bool {
	arr, ok := v.ReadArray()
	if !ok {
		ctx.Error = errors.New("这个类型无法赋值分片")
		return false
	}
	arr2, ok := val.ReadArray()
	if !ok {
		ctx.Error = errors.New("val 的类型必须是一个列表")
		return false
	}
	length := int64(len(arr.List))
	_a := getClampRealIndex(ctx, a, length)
	_b := getClampRealIndex(ctx, b, length)

	if _a > _b {
		_a = _b
	}

	offset := len(arr2.List) - int(_b-_a)
	newArr := make([]*VMValue, len(arr.List)+offset)

	for i := int64(0); i < _a; i++ {
		newArr[i] = arr.List[i]
	}

	for i := 0; i < len(arr2.List); i++ {
		newArr[int(_a)+i] = arr2.List[i]
	}

	for i := int(_b) + offset; i < len(newArr); i++ {
		newArr[i] = arr.List[i-offset]
	}

	arr.List = newArr
	return true
}

func (v *VMValue) SetSliceEx(ctx *Context, a *VMValue, b *VMValue, val *VMValue) bool {
	if a.TypeId == VMTypeUndefined {
		a = VMValueNewInt(0)
	}

	arr, ok := v.ReadArray()
	if !ok {
		ctx.Error = errors.New("这个类型无法赋值分片")
		return false
	}

	if b.TypeId == VMTypeUndefined {
		b = VMValueNewInt(int64(len(arr.List)))
	}

	valA, ok := a.ReadInt()
	if !ok {
		ctx.Error = errors.New("第一个值类型错误")
		return false
	}

	valB, ok := b.ReadInt()
	if !ok {
		ctx.Error = errors.New("第二个值类型错误")
		return false
	}

	return v.SetSlice(ctx, valA, valB, 1, val)
}

func (v *VMValue) ArrayRepeatTimesEx(ctx *Context, times *VMValue) *VMValue {
	switch times.TypeId {
	case VMTypeInt:
		times, _ := times.ReadInt()
		ad, _ := v.ReadArray()
		length := int64(len(ad.List)) * times

		if length > 512 {
			ctx.Error = errors.New("不能一次性创建过长的数组")
			return nil
		}

		arr := make([]*VMValue, length)

		for i := int64(0); i < length; i++ {
			arr[i] = ad.List[int(i)%len(ad.List)].Clone()
		}
		return VMValueNewArray(arr...)
	}
	return nil
}

func (v *VMValue) GetTypeName() string {
	switch v.TypeId {
	case VMTypeInt:
		return "int64"
	case VMTypeFloat:
		return "float64"
	case VMTypeString:
		return "str"
	case VMTypeUndefined:
		return "undefined"
	case VMTypeNull:
		return "none"
	case VMTypeComputedValue:
		return "computed"
	case VMTypeArray:
		return "array"
	case VMTypeFunction:
		//return "function"
	}
	return "unknown"
}

func (v *VMValue) ComputedExecute(ctx *Context) *VMValue {
	cd, _ := v.ReadComputed()

	vm := NewVM()
	vm.Flags = ctx.Flags
	if cd.Attrs == nil {
		cd.Attrs = &ValueMap{}
	}
	vm.attrs = cd.Attrs

	vm.ValueStoreFunc = ctx.ValueStoreFunc
	vm.ValueLoadFunc = ctx.ValueLoadFunc
	vm.subThreadDepth = ctx.subThreadDepth + 1
	vm.upCtx = ctx
	vm.NumOpCount = ctx.NumOpCount + 100
	ctx.NumOpCount = vm.NumOpCount // 防止无限递归
	if vm.NumOpCount > 30000 {
		vm.Error = errors.New("允许算力上限")
		ctx.Error = vm.Error
		return nil
	}

	if cd.code == nil {
		_ = vm.Run(cd.Expr)
		cd.code = vm.code
		cd.codeIndex = vm.codeIndex
	} else {
		vm.code = cd.code
		vm.codeIndex = cd.codeIndex
		vm.parser.Evaluate()
	}

	if vm.Error != nil {
		ctx.Error = vm.Error
		return nil
	}

	var ret *VMValue
	if vm.top != 0 {
		ret = vm.stack[vm.top-1].Clone()
	} else {
		ret = VMValueNewUndefined()
	}

	ctx.NumOpCount = vm.NumOpCount
	return ret
}

func (v *VMValue) FuncInvoke(ctx *Context, params []*VMValue) *VMValue {
	// TODO: 先复制computed代码修改，后续重构

	vm := NewVM()
	cd, _ := v.ReadFunctionData()
	vm.attrs = &ValueMap{}

	// 设置参数
	if len(cd.Params) != len(params) {
		ctx.Error = errors.New("调用参数个数与函数定义不符")
		return nil
	}
	for index, i := range cd.Params {
		//if index >= len(params) {
		//	break
		//}
		vm.attrs.Store(i, params[index])
	}

	vm.Flags = ctx.Flags
	//vm.Flags.PrintBytecode = false
	vm.ValueStoreFunc = ctx.ValueStoreFunc
	vm.ValueLoadFunc = ctx.ValueLoadFunc
	vm.subThreadDepth = ctx.subThreadDepth + 1
	vm.upCtx = ctx
	vm.NumOpCount = ctx.NumOpCount + 100
	ctx.NumOpCount = vm.NumOpCount // 防止无限递归
	if vm.NumOpCount > 30000 {
		vm.Error = errors.New("允许算力上限")
		ctx.Error = vm.Error
		return nil
	}

	if cd.code == nil {
		_ = vm.Run(cd.Expr)
		cd.code = vm.code
		cd.codeIndex = vm.codeIndex
	} else {
		vm.code = cd.code
		vm.codeIndex = cd.codeIndex
		vm.parser.Evaluate()
	}

	if vm.Error != nil {
		ctx.Error = vm.Error
		return nil
	}

	var ret *VMValue
	if vm.top != 0 {
		ret = vm.stack[vm.top-1].Clone()
	} else {
		ret = VMValueNewUndefined()
	}

	ctx.NumOpCount = vm.NumOpCount
	vm.attrs = &ValueMap{} // 清空
	return ret
}

func (v *VMValue) FuncInvokeNative(ctx *Context, params []*VMValue) *VMValue {
	cd, _ := v.ReadNativeFunctionData()

	// 设置参数
	if len(cd.Params) != len(params) {
		ctx.Error = errors.New("调用参数个数与函数定义不符")
		return nil
	}
	ret := cd.NativeFunc(ctx, params)

	if ctx.Error != nil {
		ctx.Error = ctx.Error
		return nil
	}

	if ret == nil {
		ret = VMValueNewUndefined()
	}
	return ret
}

func (v *VMValue) AsDictKey() (string, error) {
	if v.TypeId == VMTypeString || v.TypeId == VMTypeInt || v.TypeId == VMTypeFloat {
		return v.ToString(), nil
	} else {
		return "", errors.New(fmt.Sprintf("类型错误: 字典键只能为字符串或数字，不支持 %s", v.GetTypeName()))
	}
}

func VMValueNewInt(i int64) *VMValue {
	// TODO: 小整数可以处理为不可变对象，且一直停留在内存中，就像python那样。这可以避免很多内存申请
	return &VMValue{TypeId: VMTypeInt, Value: i}
}

func VMValueNewFloat(i float64) *VMValue {
	return &VMValue{TypeId: VMTypeFloat, Value: i}
}

func VMValueNewStr(s string) *VMValue {
	return &VMValue{TypeId: VMTypeString, Value: s}
}

func VMValueNewUndefined() *VMValue {
	return &VMValue{TypeId: VMTypeUndefined}
}

func vmValueNewLocal() *VMValue {
	return &VMValue{TypeId: vmTypeLocal}
}

func vmValueNewGlobal() *VMValue {
	return &VMValue{TypeId: vmTypeGlobal}
}

func VMValueNewNull() *VMValue {
	return &VMValue{TypeId: VMTypeNull}
}

func VMValueNewArrayRaw(data []*VMValue) *VMValue {
	return &VMValue{TypeId: VMTypeArray, Value: &ArrayData{data}}
}

func VMValueNewArray(values ...*VMValue) *VMValue {
	var data []*VMValue
	for _, i := range values {
		data = append(data, i)
	}

	return &VMValue{TypeId: VMTypeArray, Value: &ArrayData{data}}
}

func VMValueNewDict(data *ValueMap) *VMDictValue {
	if data == nil {
		data = &ValueMap{}
	}
	return &VMDictValue{TypeId: VMTypeDict, Value: &DictData{data}}
}

func VMValueNewDictWithArray(arr []*VMValue) (*VMDictValue, error) {
	data := &ValueMap{}
	for i := 0; i < len(arr); i += 2 {
		kName, err := arr[0].AsDictKey()
		if err != nil {
			return nil, err
		}
		data.Store(kName, arr[1])
	}
	return &VMDictValue{TypeId: VMTypeDict, Value: &DictData{data}}, nil
}

func VMValueNewComputedRaw(computed *ComputedData) *VMValue {
	return &VMValue{TypeId: VMTypeComputedValue, Value: computed}
}

func VMValueNewComputed(expr string) *VMValue {
	return &VMValue{TypeId: VMTypeComputedValue, Value: &ComputedData{
		Expr: expr,
	}}
}

func VMValueNewFunctionRaw(computed *FunctionData) *VMValue {
	return &VMValue{TypeId: VMTypeFunction, Value: computed}
}

func VMValueNewNativeFunction(data *NativeFunctionData) *VMValue {
	return &VMValue{TypeId: VMTypeNativeFunction, Value: data}
}
