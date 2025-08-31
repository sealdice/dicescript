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
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/exp/rand"
)

type VMValueType int
type IntType int                        // :IntType
const IntTypeSize = strconv.IntSize / 8 // 只能为 4 或 8(32位/64位)

const (
	VMTypeInt            VMValueType = 0
	VMTypeFloat          VMValueType = 1
	VMTypeString         VMValueType = 2
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
	(*VMValue).OpNullCoalescing,

	(*VMValue).OpCompLT,
	(*VMValue).OpCompLE,
	(*VMValue).OpCompEQ,
	(*VMValue).OpCompNE,
	(*VMValue).OpCompGE,
	(*VMValue).OpCompGT,

	(*VMValue).OpBitwiseAnd,
	(*VMValue).OpBitwiseOr,
}

type RollConfig struct {
	EnableDiceWoD         bool // 启用WOD骰子语法，即XaYmZkNqM，X个数，Y加骰线，Z面数，N阈值(>=)，M阈值(<=)
	EnableDiceCoC         bool // 启用COC骰子语法，即bX/pX奖惩骰
	EnableDiceFate        bool // 启用Fate骰语法，即fX
	EnableDiceDoubleCross bool // 启用双十字骰语法，即XcY

	DisableBitwiseOp bool // 禁用位运算，用于st，如 &a=1d4
	DisableStmts     bool // 禁用语句语法(如if while等)，仅允许表达式
	DisableNDice     bool // 禁用Nd语法，即只能2d6这样写，不能写2d

	// 如果返回值为true，那么跳过剩下的储存流程。如果overwrite不为nil，对v进行覆盖。
	// 另注: 钩子函数中含有ctx的原因是可能在函数中进行调用，此时ctx会发生变化
	HookValueStore func(ctx *Context, name string, v *VMValue) (overwrite *VMValue, solved bool)
	// 如果overwrite不为nil，将结束值加载并使用overwrite值。如果为nil，将以newName为key进行加载
	HookValueLoadPre func(ctx *Context, name string) (newName string, overwrite *VMValue)
	// 读取后回调(返回值将覆盖之前读到的值。如果之前未读取到值curVal将为nil)，用户需要在里面调用doCompute保证结果正确
	HookValueLoadPost func(ctx *Context, name string, curVal *VMValue, doCompute func(curVal *VMValue) *VMValue, detail *BufferSpan) *VMValue

	// st回调，注意val和extra都经过clone，可以放心储存
	CallbackSt              func(_type string, name string, val *VMValue, extra *VMValue, op string, detail string)                 // st回调
	CustomMakeDetailFunc    func(ctx *Context, details []BufferSpan, dataBuffer []byte, parsedOffset int) string                    // 自定义计算过程
	CustomDetailRewriteFunc func(ctx *Context, curDetail string, detailSpan BufferSpan, dataBuffer []byte, parsedOffset int) string // 自定义单项detail重写

	ParseExprLimit               uint64   // 解析算力限制，防止构造特殊语句进行DOS攻击，0为无限，建议值1000万
	OpCountLimit                 IntType  // 算力限制，超过这个值会报错，0为无限，建议值30000
	DefaultDiceSideExpr          string   // 默认骰子面数
	defaultDiceSideExprCacheFunc *VMValue // expr的缓存函数

	PrintBytecode bool // 执行时打印字节码
	IgnoreDiv0    bool // 当div0时暂不报错

	DiceMinMode bool // 骰子以最小值结算，用于获取下界
	DiceMaxMode bool // 以最大值结算 获取上界
}

type customDiceItem struct {
	// expr     string
	// callback func(ctx *Context, groups []string) *VMValue
	// parse func(ctx *Context, p *parser)

	// 该怎样写呢？似乎将一些解析相关的struct暴露出去并不合适
	// 这里我有三个选项，第一种是创建一种更简单的语法进行编译：例如
	// nos 'd' expr 可以看成是一种简易正则
	// 选项二：
	// 将parser独立成包(关联点很少，并不困难)，同时将几个expr暴露出去
	// 这样用户直接与parser进行交互
	// 使用难度较大，但自由度也很大
	// 选项三：
	// 对文档流进行简易封装，用户可以使用GetNext获取下一个字符，最后用户需要告知消耗了几个字符，然后生成了什么
	// 自由度甚至更大，但需要额外封装一下 expr 等几个主要的语法节点供调用
	//
	// 值得额外注意的一点是，customDice实际上会参与两个阶段，即解析和执行
	// 解析阶段需要确认是否匹配，同时给出抓取文本的规则，接着执行时需要使用解析过程生成的数据
	// 这需要专门的数据结构来配合，回溯的问题也值得考虑
	//
	// 再补充一点，由于是字节码虚拟机，所以 'd' e:expr 这里对e的映射实际上没有意义，
	// 并不会得到e的值因为expr语句实际上的执行结果是写入字节码
	// 所以用户还需要能操作vm数据栈？还是搞点语法糖？
	// 我认为可以先考虑选项三，再从其基础上制作选项一
}

type Context struct {
	parser         *parser
	subThreadDepth int
	Attrs          *ValueMap
	UpCtx          *Context
	// subThread      *Context // 用于执行子句

	code      []ByteCode
	codeIndex int

	stack []VMValue
	top   int

	NumOpCount IntType // 算力计数
	// CocFlagVarPrefix string // 解析过程中出现，当VarNumber开启时有效，可以是困难极难常规大成功

	Config RollConfig // 标记
	Error  error      // 报错信息

	Ret              *VMValue // 返回值
	RestInput        string   // 剩余字符串
	Matched          string   // 匹配的字符串
	DetailSpans      []BufferSpan
	detailCache      string // 计算过程
	IsComputedLoaded bool

	Seed    []byte          // 随机种子，16个字节，即双uint64
	RandSrc *rand.PCGSource // 根据种子生成的source

	IsRunning      bool // 是否正在运行，Run时会置为true，halt时会置为false
	CustomDiceInfo []*customDiceItem

	forceSolveDetail bool // 一个辅助属性，用于computed时强制获取计算过程

	/** 全局变量 */
	globalNames *ValueMap

	// 全局scope的写入回调
	GlobalValueStoreFunc func(name string, v *VMValue)
	// 全局scope的读取回调
	GlobalValueLoadFunc func(name string) *VMValue
	// 全局scope的读取后回调(返回值将覆盖之前读到的值。如果之前未读取到值curVal将为nil)
	GlobalValueLoadOverwriteFunc func(name string, curVal *VMValue) *VMValue
}

func (ctx *Context) GetDetailText() string {
	if ctx.DetailSpans != nil {
		if ctx.detailCache != "" {
			return ctx.detailCache
		}
		ctx.detailCache = ctx.makeDetailStr(ctx.DetailSpans)
		return ctx.detailCache
	}
	return ""
}

func (ctx *Context) StackTop() int {
	return ctx.top
}

func (ctx *Context) Depth() int {
	return ctx.subThreadDepth
}

func (ctx *Context) SetConfig(cfg *RollConfig) {
	ctx.Config = *cfg
}

func (ctx *Context) Init() {
	ctx.Attrs = &ValueMap{}
	ctx.globalNames = &ValueMap{}
	ctx.detailCache = ""
	ctx.DetailSpans = nil
	if ctx.Seed != nil {
		s := rand.PCGSource{}
		_ = s.UnmarshalBinary(ctx.Seed)
		ctx.RandSrc = &s
	}
}

func (ctx *Context) GetCurSeed() ([]byte, error) {
	if ctx.RandSrc != nil {
		return ctx.RandSrc.MarshalBinary()
	}
	return randSource.MarshalBinary()
}

func (ctx *Context) loadInnerVar(name string) *VMValue {
	return builtinValues[name]
}

func (ctx *Context) solveLoadPostAndComputed(name string, val *VMValue, isRaw bool, detail *BufferSpan) *VMValue {
	withDetail := detail != nil

	// 计算真实结果
	doCompute := func(val *VMValue) *VMValue {
		if !isRaw && val.TypeId == VMTypeComputedValue {
			if withDetail {
				val = val.ComputedExecute(ctx, detail)
			} else {
				val = val.ComputedExecute(ctx, &BufferSpan{})
			}
			if ctx.Error != nil {
				return nil
			}
		}

		// 追加计算结果到detail
		if withDetail {
			detail.Ret = val
		}
		return val
	}

	if ctx.Config.HookValueLoadPost != nil {
		if detail != nil {
			oldRet := detail.Ret
			val = ctx.Config.HookValueLoadPost(ctx, name, val, doCompute, detail)
			if oldRet == detail.Ret {
				// 如果ret发生变化才修改，顺便修改detail中的结果为最终结果
				detail.Ret = val
			}
		} else {
			val = ctx.Config.HookValueLoadPost(ctx, name, val, doCompute, &BufferSpan{})
		}
	} else {
		val = doCompute(val)
	}
	return val
}

func (ctx *Context) LoadNameGlobalWithDetail(name string, isRaw bool, detail *BufferSpan) *VMValue {
	var loadFunc func(name string) *VMValue
	if loadFunc == nil {
		loadFunc = ctx.GlobalValueLoadFunc
	}

	// 检测全局表
	if loadFunc != nil {
		val := loadFunc(name)
		if val != nil {
			if !isRaw && val.TypeId == VMTypeComputedValue {
				val = val.ComputedExecute(ctx, detail)
				if ctx.Error != nil {
					return nil
				}
			}
			return val
		}
	}
	// else {
	//	ctx.Error = errors.New("未设置 GlobalValueLoadFunc，无法获取变量")
	//	return nil
	// }

	// 检测内置变量/函数检查
	val := ctx.loadInnerVar(name)
	if ctx.GlobalValueLoadOverwriteFunc != nil {
		val = ctx.GlobalValueLoadOverwriteFunc(name, val)
	}
	if val == nil {
		val = NewNullVal()
	}

	val = ctx.solveLoadPostAndComputed(name, val, isRaw, detail)
	return val
}

func (ctx *Context) LoadNameGlobal(name string, isRaw bool) *VMValue {
	return ctx.LoadNameGlobalWithDetail(name, isRaw, nil)
}

func (ctx *Context) LoadNameLocalWithDetail(name string, isRaw bool, detail *BufferSpan) *VMValue {
	// if ctx.currentThis != nil {
	//	return ctx.currentThis.AttrGet(ctx, name)
	// } else {
	// if ctx.subThreadDepth >= 1 {
	val, exists := ctx.Attrs.Load(name)
	if !exists {
		val = NewNullVal()
	}

	val = ctx.solveLoadPostAndComputed(name, val, isRaw, detail)
	return val
	// }
	// }
}

func (ctx *Context) LoadNameLocal(name string, isRaw bool) *VMValue {
	return ctx.LoadNameLocalWithDetail(name, isRaw, nil)
}

func (ctx *Context) LoadNameWithDetail(name string, isRaw bool, useHook bool, detail *BufferSpan) *VMValue {
	// 有个疑问，这里useHook真有用吗，什么情况下有用？
	if useHook && ctx.Config.HookValueLoadPre != nil {
		var overwrite *VMValue
		name, overwrite = ctx.Config.HookValueLoadPre(ctx, name)

		if overwrite != nil {
			// 使用弄进来的替代值进行计算
			return overwrite
		}
	}

	// 先local再global
	curCtx := ctx
	for {
		ret := curCtx.LoadNameLocalWithDetail(name, isRaw, detail)

		if curCtx.Error != nil {
			ctx.Error = curCtx.Error
			return nil
		}
		if ret.TypeId != VMTypeNull {
			return ret
		}
		if curCtx.UpCtx == nil {
			break
		} else {
			curCtx = curCtx.UpCtx
		}
	}

	return ctx.LoadNameGlobalWithDetail(name, isRaw, detail)
}

func (ctx *Context) LoadName(name string, isRaw bool, useHook bool) *VMValue {
	return ctx.LoadNameWithDetail(name, isRaw, useHook, nil)
}

// StoreName 储存变量
func (ctx *Context) StoreName(name string, v *VMValue, useHook bool) {
	if useHook && ctx.Config.HookValueStore != nil {
		overwrite, solved := ctx.Config.HookValueStore(ctx, name, v)
		if solved {
			return
		}
		if overwrite != nil {
			v = overwrite
		}
	}
	if _, ok := ctx.globalNames.Load(name); ok {
		ctx.StoreNameGlobal(name, v)
	} else {
		ctx.StoreNameLocal(name, v)
	}
}

func (ctx *Context) StoreNameLocal(name string, v *VMValue) {
	ctx.Attrs.Store(name, v)
}

func (ctx *Context) StoreNameGlobal(name string, v *VMValue) {
	storeFunc := ctx.GlobalValueStoreFunc
	if storeFunc != nil {
		storeFunc(name, v)
	} else {
		ctx.Error = errors.New("未设置 ValueStoreNameFunc，无法储存变量")
		return
	}
}

func (ctx *Context) RegCustomDice(s string, callback func(ctx *Context, groups []string) *VMValue) error {
	// re, err := regexp.Compile(s)
	// if err != nil {
	// 	return err
	// }
	// ctx.CustomDiceInfo = append(ctx.CustomDiceInfo, &customDiceItem{re, callback})
	return nil
}

type VMValue struct {
	TypeId VMValueType `json:"t"`
	Value  any         `json:"v"`
	// ExpiredTime int64       `json:"expiredTime"`
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
	Expr     string
	Name     string
	Params   []string
	Defaults []*VMValue

	/* 缓存数据 */
	Self      *VMValue // 若存在self，即为bound method
	code      []ByteCode
	codeIndex int
	// ctx       *Context
}

type NativeFunctionDef func(ctx *Context, this *VMValue, params []*VMValue) *VMValue

type NativeFunctionData struct {
	Name     string
	Params   []string
	Defaults []*VMValue

	/* 缓存数据 */
	Self       *VMValue // 若存在self，即为bound method
	NativeFunc NativeFunctionDef
}

type NativeObjectData struct {
	Name     string
	AttrSet  func(ctx *Context, name string, v *VMValue)
	AttrGet  func(ctx *Context, name string) *VMValue
	ItemSet  func(ctx *Context, index *VMValue, v *VMValue)
	ItemGet  func(ctx *Context, index *VMValue) *VMValue
	DirFunc  func(ctx *Context) []*VMValue
	ToString func(ctx *Context) string
}

func (v *VMValue) Clone() *VMValue {
	// switch v.TypeId {
	// case VMTypeDict, VMTypeArray:
	//	return v
	// default:
	return &VMValue{TypeId: v.TypeId, Value: v.Value}
	// }
}

func (v *VMValue) AsBool() bool {
	switch v.TypeId {
	case VMTypeInt:
		return v.Value != IntType(0)
	case VMTypeFloat:
		return v.Value != 0.0
	case VMTypeString:
		return v.Value != ""
	case VMTypeNull:
		return false
	case VMTypeComputedValue:
		vd := v.Value.(*ComputedData)
		return vd.Expr != ""
	case VMTypeArray:
		ad := v.MustReadArray()
		return len(ad.List) != 0
	case VMTypeDict:
		dd := v.MustReadDictData()
		return dd.Dict.Length() != 0
	case VMTypeFunction, VMTypeNativeFunction, VMTypeNativeObject:
		return true
	default:
		return false
	}
}

type recursionInfo struct {
	exists map[interface{}]bool
}

func (v *VMValue) ToString() string {
	ri := &recursionInfo{exists: map[interface{}]bool{}}
	return v.toStringRaw(ri)
}

func (v *VMValue) toStringRaw(ri *recursionInfo) string {
	if v == nil {
		return "NIL"
	}
	switch v.TypeId {
	case VMTypeInt:
		return strconv.FormatInt(int64(v.Value.(IntType)), 10)
	case VMTypeFloat:
		return strconv.FormatFloat(v.Value.(float64), 'f', -1, 64)
	case VMTypeString:
		return v.Value.(string)
	case VMTypeNull:
		return "null"
	case VMTypeArray:
		// 避免循环重复
		if _, exists := ri.exists[v.Value]; exists {
			return "[...]"
		}
		ri.exists[v.Value] = true

		s := "["
		arr, _ := v.ReadArray()
		for index, i := range arr.List {
			x := i.toReprRaw(ri)
			s += x
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
		// 避免循环重复
		if _, exists := ri.exists[v.Value]; exists {
			return "{...}"
		}
		ri.exists[v.Value] = true

		var items []string
		dd, _ := v.ReadDictData()
		dd.Dict.Range(func(key string, value *VMValue) bool {
			txt := value.toReprRaw(ri)
			// txt := ""
			// if value.TypeId == VMTypeArray {
			//	txt = "[...]"
			// } else if value.TypeId == VMTypeDict {
			//	txt = "{...}"
			// } else {
			//	txt = value.ToRepr()
			// }
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
	case VMTypeNativeObject:
		od, _ := v.ReadNativeObjectData()
		return "nobject " + od.Name
	default:
		return "a value"
	}
}

func (v *VMValue) toReprRaw(ri *recursionInfo) string {
	if v == nil {
		return "NIL"
	}
	switch v.TypeId {
	case VMTypeString:
		// TODO: 检测其中是否有"
		return "'" + v.toStringRaw(ri) + "'"
	case VMTypeInt, VMTypeFloat, VMTypeNull, VMTypeArray, VMTypeComputedValue, VMTypeDict, VMTypeFunction, VMTypeNativeFunction, VMTypeNativeObject:
		return v.toStringRaw(ri)
	default:
		return "<a value>"
	}
}

func (v *VMValue) ToRepr() string {
	ri := &recursionInfo{exists: map[any]bool{}}
	return v.toReprRaw(ri)
}

func (v *VMValue) ReadInt() (IntType, bool) {
	if v.TypeId == VMTypeInt {
		return v.Value.(IntType), true
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
	panic("错误: 不正确的类型")
}

func (v *VMValue) MustReadArray() *ArrayData {
	if ad, ok := v.ReadArray(); ok {
		return ad
	}
	panic("错误: 不正确的类型")
}

func (v *VMValue) MustReadInt() IntType {
	val, ok := v.ReadInt()
	if ok {
		return val
	}
	panic("错误: 不正确的类型")
}

func (v *VMValue) MustReadFloat() float64 {
	val, ok := v.ReadFloat()
	if ok {
		return val
	}
	panic("错误: 不正确的类型")
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

func (v *VMValue) ReadNativeObjectData() (*NativeObjectData, bool) {
	if v.TypeId == VMTypeNativeObject {
		return v.Value.(*NativeObjectData), true
	}
	return nil, false
}

func (v *VMValue) OpAdd(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			val := v.Value.(IntType) + v2.Value.(IntType)
			return NewIntVal(val)
		case VMTypeFloat:
			val := float64(v.Value.(IntType)) + v2.Value.(float64)
			return NewFloatVal(val)
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			val := v.Value.(float64) + float64(v2.Value.(IntType))
			return NewFloatVal(val)
		case VMTypeFloat:
			val := v.Value.(float64) + v2.Value.(float64)
			return NewFloatVal(val)
		}
	case VMTypeString:
		switch v2.TypeId {
		case VMTypeString:
			val := v.Value.(string) + v2.Value.(string)
			return NewStrVal(val)
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
			return NewArrayVal(arrFinal...)
		}
	}

	return nil
}

func (v *VMValue) OpSub(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			val := v.Value.(IntType) - v2.Value.(IntType)
			return NewIntVal(val)
		case VMTypeFloat:
			val := float64(v.Value.(IntType)) - v2.Value.(float64)
			return NewFloatVal(val)
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			val := v.Value.(float64) - float64(v2.Value.(IntType))
			return NewFloatVal(val)
		case VMTypeFloat:
			val := v.Value.(float64) - v2.Value.(float64)
			return NewFloatVal(val)
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
			val := v.Value.(IntType) * v2.Value.(IntType)
			return NewIntVal(val)
		case VMTypeFloat:
			val := float64(v.Value.(IntType)) * v2.Value.(float64)
			return NewFloatVal(val)
		case VMTypeArray:
			return v2.ArrayRepeatTimesEx(ctx, v)
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			val := v.Value.(float64) * float64(v2.Value.(IntType))
			return NewFloatVal(val)
		case VMTypeFloat:
			val := v.Value.(float64) * v2.Value.(float64)
			return NewFloatVal(val)
		}
	case VMTypeArray:
		return v.ArrayRepeatTimesEx(ctx, v2)
	}

	return nil
}

func (v *VMValue) OpDivide(ctx *Context, v2 *VMValue) *VMValue {
	setDivideZero := func() *VMValue {
		if ctx.Config.IgnoreDiv0 {
			return v
		}
		ctx.Error = errors.New("被除数为0")
		return nil
	}

	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			if v2.Value.(IntType) == 0 {
				return setDivideZero()
			}
			val := v.Value.(IntType) / v2.Value.(IntType)
			return NewIntVal(val)
		case VMTypeFloat:
			if v2.Value.(float64) == 0 {
				return setDivideZero()
			}
			val := float64(v.Value.(IntType)) / v2.Value.(float64)
			return NewFloatVal(val)
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			if v2.Value.(IntType) == 0 {
				return setDivideZero()
			}
			val := v.Value.(float64) / float64(v2.Value.(IntType))
			return NewFloatVal(val)
		case VMTypeFloat:
			if v2.Value.(float64) == 0 {
				return setDivideZero()
			}
			val := v.Value.(float64) / v2.Value.(float64)
			return NewFloatVal(val)
		}
	}

	return nil
}

func (v *VMValue) OpModulus(ctx *Context, v2 *VMValue) *VMValue {
	setDivideZero := func() {
		ctx.Error = errors.New("被除数被0")
	}

	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			if v2.Value.(IntType) == 0 {
				setDivideZero()
				return nil
			}
			val := v.Value.(IntType) % v2.Value.(IntType)
			return NewIntVal(val)
		}
	}

	return nil
}

func (v *VMValue) OpPower(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			val := IntType(math.Pow(float64(v.Value.(IntType)), float64(v2.Value.(IntType))))
			return NewIntVal(val)
		case VMTypeFloat:
			val := math.Pow(float64(v.Value.(IntType)), v2.Value.(float64))
			return NewFloatVal(val)
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			val := math.Pow(v.Value.(float64), float64(v2.Value.(IntType)))
			return NewFloatVal(val)
		case VMTypeFloat:
			val := math.Pow(v.Value.(float64), v2.Value.(float64))
			return NewFloatVal(val)
		}
	}

	return nil
}

func (v *VMValue) OpNullCoalescing(ctx *Context, v2 *VMValue) *VMValue {
	if v.TypeId == VMTypeNull {
		return v2
	} else {
		return v
	}
}

func boolToVMValue(v bool) *VMValue {
	var val IntType
	if v {
		val = 1
	}
	return NewIntVal(val)
}

func (v *VMValue) OpCompLT(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			return boolToVMValue(v.Value.(IntType) < v2.Value.(IntType))
		case VMTypeFloat:
			return boolToVMValue(float64(v.Value.(IntType)) < v2.Value.(float64))
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			return boolToVMValue(v.Value.(float64) < float64(v2.Value.(IntType)))
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
			return boolToVMValue(v.Value.(IntType) <= v2.Value.(IntType))
		case VMTypeFloat:
			return boolToVMValue(float64(v.Value.(IntType)) <= v2.Value.(float64))
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			return boolToVMValue(v.Value.(float64) <= float64(v2.Value.(IntType)))
		case VMTypeFloat:
			return boolToVMValue(v.Value.(float64) <= v2.Value.(float64))
		}
	}

	return nil
}

func (v *VMValue) OpCompEQ(ctx *Context, v2 *VMValue) *VMValue {
	return boolToVMValue(ValueEqual(v, v2, true))
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
			return boolToVMValue(v.Value.(IntType) >= v2.Value.(IntType))
		case VMTypeFloat:
			return boolToVMValue(float64(v.Value.(IntType)) >= v2.Value.(float64))
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			return boolToVMValue(v.Value.(float64) >= float64(v2.Value.(IntType)))
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
			return boolToVMValue(v.Value.(IntType) > v2.Value.(IntType))
		case VMTypeFloat:
			return boolToVMValue(float64(v.Value.(IntType)) > v2.Value.(float64))
		}
	case VMTypeFloat:
		switch v2.TypeId {
		case VMTypeInt:
			return boolToVMValue(v.Value.(float64) > float64(v2.Value.(IntType)))
		case VMTypeFloat:
			return boolToVMValue(v.Value.(float64) > v2.Value.(float64))
		}
	}

	return nil
}

func (v *VMValue) OpBitwiseAnd(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			return NewIntVal(v.Value.(IntType) & v2.Value.(IntType))
		}
	}
	return nil
}

func (v *VMValue) OpBitwiseOr(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		switch v2.TypeId {
		case VMTypeInt:
			return NewIntVal(v.Value.(IntType) | v2.Value.(IntType))
		}
	}
	return nil
}

func (v *VMValue) OpPositive() *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		return NewIntVal(v.Value.(IntType))
	case VMTypeFloat:
		return NewFloatVal(v.Value.(float64))
	}
	return nil
}

func (v *VMValue) OpNegation() *VMValue {
	switch v.TypeId {
	case VMTypeInt:
		return NewIntVal(-v.Value.(IntType))
	case VMTypeFloat:
		return NewFloatVal(-v.Value.(float64))
	}
	return nil
}

func (v *VMValue) AttrSet(ctx *Context, name string, val *VMValue) *VMValue {
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
	case VMTypeNativeObject:
		od, _ := v.ReadNativeObjectData()
		od.AttrSet(ctx, name, val)
		return val
	}

	return nil
}

// AttrGet 如果返回nil 说明不支持 . 取属性
func (v *VMValue) AttrGet(ctx *Context, name string) *VMValue {
	switch v.TypeId {
	case VMTypeComputedValue:
		cd, _ := v.ReadComputed()
		var ret *VMValue
		if cd.Attrs != nil {
			ret, _ = cd.Attrs.Load(name)
		}
		if ret == nil {
			ret = NewNullVal()
		}
		return ret
	case VMTypeDict:
		a := (*VMDictValue)(v)
		ret, _ := a.Load(name)
		if ret == nil {
			var ok bool
			p1 := v
			p1x := a

			for {
				if p1, ok = p1x.Load("__proto__"); ok && p1.TypeId == VMTypeDict {
					var exists bool
					p1x = (*VMDictValue)(p1)
					ret, exists = p1x.Load(name)

					if exists {
						break
					}
				} else {
					break
				}
			}

			// if Ret == nil {
			//	Ret = VMValueNewUndefined()
			// }
		}
		// TODO: 思考一下 Dict.keys 和 Dict.values 与 ArrtGet 的冲突
		if ret != nil {
			return ret
		}
	case vmTypeGlobal:
		// 加载全局变量
		ret := ctx.LoadNameGlobal(name, false)
		if ret == nil {
			ret = NewNullVal()
		}
		return ret
	case vmTypeLocal:
		ret := ctx.LoadNameLocal(name, false)
		if ret == nil {
			ret = NewNullVal()
		}
		return ret
	case VMTypeNativeObject:
		od, _ := v.ReadNativeObjectData()
		ret := od.AttrGet(ctx, name)
		if ret != nil {
			return ret
		}
	}

	proto := builtinProto[v.TypeId]
	if proto != nil {
		if method, ok := proto.Load(name); ok {
			return getBindMethod(v, method)
		}
	}

	// 给少数几个类明确设定为不支持，返回nil
	// 其他一律返回 undefined
	switch v.TypeId {
	case VMTypeInt, VMTypeFloat, VMTypeString, VMTypeNull:
		return nil
	}

	return NewNullVal()
}

func (v *VMValue) ItemGet(ctx *Context, index *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeArray:
		if index.TypeId != VMTypeInt {
			ctx.Error = fmt.Errorf("类型错误: 数字下标必须为数字，不能为 %s", index.GetTypeName())
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
	case VMTypeString:
		if index.TypeId != VMTypeInt {
			ctx.Error = fmt.Errorf("类型错误: 数字下标必须为数字，不能为 %s", index.GetTypeName())
		} else {
			str, _ := v.ReadString()
			rstr := []rune(str)

			rIndex := index.MustReadInt()
			_index := getClampRealIndex(ctx, rIndex, IntType(len(rstr)))

			newArr := string(rstr[_index : _index+1])
			return NewStrVal(newArr)
		}
	case VMTypeNativeObject:
		od, _ := v.ReadNativeObjectData()
		ret := od.ItemGet(ctx, index)
		if ret == nil {
			ret = NewNullVal()
		}
		return ret
	default:
		// case VMTypeUndefined, VMTypeNull:
		ctx.Error = errors.New("此类型无法取下标")
	}
	return nil
}

func (v *VMValue) ItemSet(ctx *Context, index *VMValue, val *VMValue) bool {
	switch v.TypeId {
	case VMTypeArray:
		if index.TypeId != VMTypeInt {
			ctx.Error = fmt.Errorf("类型错误: 数字下标必须为数字，不能为 %s", index.GetTypeName())
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
	case VMTypeNativeObject:
		od, _ := v.ReadNativeObjectData()
		od.ItemSet(ctx, index, val)
		if ctx.Error == nil {
			return true
		}
	default:
		ctx.Error = errors.New("此类型无法赋值下标")
	}
	return false
}

func getRealIndex(ctx *Context, index IntType, length IntType) IntType {
	if index < 0 {
		// 负数下标支持
		index = length + index
	}
	if index >= length || index < 0 {
		ctx.Error = errors.New("无法获取此下标")
	}
	return index
}

func getClampRealIndex(ctx *Context, index IntType, length IntType) IntType {
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

func (v *VMValue) GetSlice(ctx *Context, a IntType, b IntType, step IntType) *VMValue {
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
		newArr := string([]rune(str)[_a:_b])
		return NewStrVal(newArr)
	case VMTypeArray:
		arr, _ := v.ReadArray()
		newArr := arr.List[_a:_b]
		return NewArrayVal(newArr...)
	default:
		ctx.Error = errors.New("这个类型无法取得分片")
		return nil
	}
}

func (v *VMValue) Length(ctx *Context) IntType {
	var length IntType

	switch v.TypeId {
	case VMTypeArray:
		arr, _ := v.ReadArray()
		length = IntType(len(arr.List))
	case VMTypeDict:
		d := v.MustReadDictData()
		length = IntType(d.Dict.Length())
	case VMTypeString:
		str, _ := v.ReadString()
		length = IntType(len([]rune(str)))
	default:
		ctx.Error = errors.New("这个类型无法取得长度")
		return 0
	}

	return length
}

func (v *VMValue) GetSliceEx(ctx *Context, a *VMValue, b *VMValue) *VMValue {
	if a.TypeId == VMTypeNull {
		a = NewIntVal(0)
	}

	length := v.Length(ctx)
	if ctx.Error != nil {
		return nil
	}

	if b.TypeId == VMTypeNull {
		b = NewIntVal(length)
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

func (v *VMValue) SetSlice(ctx *Context, a, b, step IntType, val *VMValue) bool {
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
	length := IntType(len(arr.List))
	_a := getClampRealIndex(ctx, a, length)
	_b := getClampRealIndex(ctx, b, length)

	if _a > _b {
		_a = _b
	}

	offset := len(arr2.List) - int(_b-_a)
	newArr := make([]*VMValue, len(arr.List)+offset)

	for i := IntType(0); i < _a; i++ {
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
	if a.TypeId == VMTypeNull {
		a = NewIntVal(0)
	}

	arr, ok := v.ReadArray()
	if !ok {
		ctx.Error = errors.New("这个类型无法赋值分片")
		return false
	}

	if b.TypeId == VMTypeNull {
		b = NewIntVal(IntType(len(arr.List)))
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
		length := IntType(len(ad.List)) * times

		if length > 512 {
			ctx.Error = errors.New("不能一次性创建过长的数组")
			return nil
		}

		arr := make([]*VMValue, length)

		for i := IntType(0); i < length; i++ {
			arr[i] = ad.List[int(i)%len(ad.List)].Clone()
		}
		return NewArrayVal(arr...)
	}
	return nil
}

func (v *VMValue) GetTypeName() string {
	switch v.TypeId {
	case VMTypeInt:
		return "int"
	case VMTypeFloat:
		return "float"
	case VMTypeString:
		return "str"
	case VMTypeNull:
		return "null"
	case VMTypeComputedValue:
		return "computed"
	case VMTypeArray:
		return "array"
	case VMTypeFunction:
		return "function"
	case VMTypeNativeFunction:
		return "nfunction"
	case VMTypeNativeObject:
		return "nobject"
	}
	return "unknown"
}

func (v *VMValue) ComputedExecute(ctx *Context, detail *BufferSpan) *VMValue {
	cd, _ := v.ReadComputed()

	vm := NewVM()
	vm.Config = ctx.Config
	if cd.Attrs == nil {
		cd.Attrs = &ValueMap{}
	}
	vm.Attrs = cd.Attrs

	vm.GlobalValueStoreFunc = ctx.GlobalValueStoreFunc
	vm.GlobalValueLoadFunc = ctx.GlobalValueLoadFunc
	vm.GlobalValueLoadOverwriteFunc = ctx.GlobalValueLoadOverwriteFunc
	vm.subThreadDepth = ctx.subThreadDepth + 1
	vm.UpCtx = ctx
	vm.NumOpCount = ctx.NumOpCount + 100
	ctx.NumOpCount = vm.NumOpCount // 防止无限递归
	vm.RandSrc = ctx.RandSrc
	vm.forceSolveDetail = true
	if ctx.Config.OpCountLimit > 0 && vm.NumOpCount > vm.Config.OpCountLimit {
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
		vm.evaluate()
	}

	if vm.Error != nil {
		ctx.Error = vm.Error
		return nil
	}

	var ret *VMValue
	var detailText string
	if vm.top != 0 {
		ret = vm.stack[vm.top-1].Clone()
		if vm.parser == nil {
			// 兼容: 如果没有parser填充一个避免报错，不过会占用一些额外的内存
			vm.parser = &parser{data: []byte(cd.Expr)}
			vm.parser.pt.offset = len(vm.parser.data)
		}
		detailText = vm.makeDetailStr(vm.DetailSpans)
	} else {
		ret = NewNullVal()
	}

	ctx.NumOpCount = vm.NumOpCount
	ctx.IsComputedLoaded = true

	if detail != nil {
		// detail.Expr = cd.Expr
		detail.Tag = "load.computed"
		detail.Text = detailText
	}
	return ret
}

func (v *VMValue) FuncInvoke(ctx *Context, params []*VMValue) *VMValue {
	return v.FuncInvokeRaw(ctx, params, false)
}

func (v *VMValue) FuncInvokeRaw(ctx *Context, params []*VMValue, useUpCtxLocal bool) *VMValue {
	// TODO: 先复制computed代码修改，后续重构

	vm := NewVM()
	cd, _ := v.ReadFunctionData()
	if useUpCtxLocal {
		vm.Attrs = ctx.Attrs
	} else {
		vm.Attrs = &ValueMap{}
	}

	// 设置参数
	if len(cd.Params) != len(params) {
		ctx.Error = fmt.Errorf("调用参数个数与函数定义不符，需求%d，传入%d", len(cd.Params), len(params))
		return nil
	}
	for index, i := range cd.Params {
		// if index >= len(params) {
		//	break
		// }
		vm.Attrs.Store(i, params[index])
	}

	vm.Config = ctx.Config
	// vm.Config.PrintBytecode = false
	vm.GlobalValueStoreFunc = ctx.GlobalValueStoreFunc
	vm.GlobalValueLoadFunc = ctx.GlobalValueLoadFunc
	vm.GlobalValueLoadOverwriteFunc = ctx.GlobalValueLoadOverwriteFunc
	vm.subThreadDepth = ctx.subThreadDepth + 1
	vm.UpCtx = ctx
	vm.NumOpCount = ctx.NumOpCount + 100 // 递归视为消耗 + 100
	ctx.NumOpCount = vm.NumOpCount       // 防止无限递归
	vm.RandSrc = ctx.RandSrc
	if ctx.Config.OpCountLimit > 0 && vm.NumOpCount > vm.Config.OpCountLimit {
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
		vm.evaluate()
	}

	if vm.Error != nil {
		ctx.Error = vm.Error
		return nil
	}

	var ret *VMValue
	if vm.top != 0 {
		ret = vm.stack[vm.top-1].Clone()
	} else {
		ret = NewNullVal()
	}

	ctx.NumOpCount = vm.NumOpCount
	if !useUpCtxLocal {
		vm.Attrs = &ValueMap{} // 清空
	}
	ctx.IsComputedLoaded = true
	return ret
}

func (v *VMValue) FuncInvokeNative(ctx *Context, params []*VMValue) *VMValue {
	cd, _ := v.ReadNativeFunctionData()

	// 设置参数
	if cd.Defaults != nil {
		// 参数填充
		for i := 0; i < len(cd.Defaults); i++ {
			if cd.Defaults[i] != nil {
				if len(params) <= i {
					params = append(params, cd.Defaults[i])
				}
			}
		}
	}

	if len(cd.Params) != len(params) {
		ctx.Error = fmt.Errorf("调用参数个数与函数定义不符，需求%d，传入%d", len(cd.Params), len(params))
		return nil
	}
	ret := cd.NativeFunc(ctx, cd.Self, params)

	if ctx.Error != nil {
		return nil
	}

	if ret == nil {
		ret = NewNullVal()
	}
	return ret
}

func (v *VMValue) AsDictKey() (string, error) {
	if v.TypeId == VMTypeString || v.TypeId == VMTypeInt || v.TypeId == VMTypeFloat {
		return v.ToString(), nil
	} else {
		return "", fmt.Errorf("类型错误: 字典键只能为字符串或数字，不支持 %s", v.GetTypeName())
	}
}

func ValueEqual(a *VMValue, b *VMValue, autoConvert bool) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if a.TypeId == b.TypeId {
		switch a.TypeId {
		case VMTypeArray:
			arr1, _ := a.ReadArray()
			arr2, _ := b.ReadArray()
			if len(arr1.List) != len(arr2.List) {
				return false
			}
			for index, i := range arr1.List {
				if !ValueEqual(i, arr2.List[index], autoConvert) {
					return false
				}
			}
			return true
		case VMTypeDict:
			d1 := a.MustReadDictData()
			d2 := b.MustReadDictData()
			if len(d1.Dict.dirty) != len(d2.Dict.dirty) {
				return false
			}
			isSame := true
			d1.Dict.Range(func(key string, value *VMValue) bool {
				isEqual := ValueEqual(value, d2.Dict.MustLoad(key), autoConvert)
				if !isEqual {
					isSame = false
					return false
				}
				return true
			})
			return isSame
		case VMTypeComputedValue:
			c1, _ := a.ReadComputed()
			c2, _ := b.ReadComputed()
			return c1.Expr == c2.Expr
		case VMTypeNativeFunction:
			fd1, _ := a.ReadNativeFunctionData()
			fd2, _ := b.ReadNativeFunctionData()
			return reflect.ValueOf(fd1.NativeFunc).Pointer() == reflect.ValueOf(fd2.NativeFunc).Pointer()
		default:
			return a.Value == b.Value
		}
	} else {
		if autoConvert {
			switch a.TypeId {
			case VMTypeInt:
				switch b.TypeId {
				case VMTypeFloat:
					return float64(a.Value.(IntType)) == b.Value.(float64)
				}
			case VMTypeFloat:
				switch b.TypeId {
				case VMTypeInt:
					return a.Value.(float64) == float64(b.Value.(IntType))
				}
			}
		}
	}
	return false
}

func NewIntVal(i IntType) *VMValue {
	// TODO: 小整数可以处理为不可变对象，且一直停留在内存中，就像python那样。这可以避免很多内存申请
	return &VMValue{TypeId: VMTypeInt, Value: i}
}

func NewFloatVal(i float64) *VMValue {
	return &VMValue{TypeId: VMTypeFloat, Value: i}
}

func NewStrVal(s string) *VMValue {
	return &VMValue{TypeId: VMTypeString, Value: s}
}

func vmValueNewLocal() *VMValue {
	return &VMValue{TypeId: vmTypeLocal}
}

// func vmValueNewGlobal() *VMValue {
//	return &VMValue{TypeId: vmTypeGlobal}
// }

func NewNullVal() *VMValue {
	return &VMValue{TypeId: VMTypeNull}
}

func NewArrayValRaw(data []*VMValue) *VMValue {
	return &VMValue{TypeId: VMTypeArray, Value: &ArrayData{data}}
}

func NewArrayVal(values ...*VMValue) *VMValue {
	var data []*VMValue
	data = append(data, values...)
	return &VMValue{TypeId: VMTypeArray, Value: &ArrayData{data}}
}

func NewDictVal(data *ValueMap) *VMDictValue {
	if data == nil {
		data = &ValueMap{}
	}
	return &VMDictValue{TypeId: VMTypeDict, Value: &DictData{data}}
}

func NewDictValWithArray(arr ...*VMValue) (*VMDictValue, error) {
	data := &ValueMap{}
	for i := 0; i < len(arr); i += 2 {
		kName, err := arr[i].AsDictKey()
		if err != nil {
			return nil, err
		}
		data.Store(kName, arr[i+1])
	}
	return &VMDictValue{TypeId: VMTypeDict, Value: &DictData{data}}, nil
}

func NewDictValWithArrayMust(arr ...*VMValue) *VMDictValue {
	d, err := NewDictValWithArray(arr...)
	if err != nil {
		panic(err)
	}
	return d
}

func NewComputedValRaw(computed *ComputedData) *VMValue {
	return &VMValue{TypeId: VMTypeComputedValue, Value: computed}
}

func NewComputedVal(expr string) *VMValue {
	return &VMValue{TypeId: VMTypeComputedValue, Value: &ComputedData{
		Expr: expr,
	}}
}

func NewFunctionValRaw(computed *FunctionData) *VMValue {
	return &VMValue{TypeId: VMTypeFunction, Value: computed}
}

func NewNativeFunctionVal(data *NativeFunctionData) *VMValue {
	return &VMValue{TypeId: VMTypeNativeFunction, Value: data}
}

func NewNativeObjectVal(data *NativeObjectData) *VMValue {
	return &VMValue{TypeId: VMTypeNativeObject, Value: data}
}
