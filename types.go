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
	"math"
	"strconv"
)

type VMValueType int

const (
	VMTypeInt64         VMValueType = 0
	VMTypeFloat64       VMValueType = 1
	VMTypeString        VMValueType = 2
	VMTypeNone          VMValueType = 4 // 这里错开是为了和旧版兼容
	VMTypeComputedValue VMValueType = 5
	VMTypeArray         VMValueType = 6
)

type CodeType uint8

const (
	TypePushIntNumber CodeType = iota
	TypePushFloatNumber
	TypePushString
	TypeNegation

	TypeDiceUnary

	TypeAdd // 注意，修改顺序时一定要顺带修改下面的数组
	TypeSubtract
	TypeMultiply
	TypeDivide
	TypeModulus
	TypeExponentiation

	TypeCompLT
	TypeCompLE
	TypeCompEQ
	TypeCompNE
	TypeCompGE
	TypeCompGT

	TypeBitwiseAnd
	TypeBitwiseOr
	TypeLogicAnd
	TypeLogicOr

	TypeDiceInit
	TypeDiceSetTimes
	TypeDiceSetKeepLowNum
	TypeDiceSetKeepHighNum
	TypeDiceSetDropLowNum
	TypeDiceSetDropHighNum
	TypeDice

	TypeDicePenalty
	TypeDiceBonus
	TypeDiceFate
	TypeDiceWod
	TypeWodSetInit       // 重置参数
	TypeWodSetPool       // 设置骰池(骰数)
	TypeWodSetPoints     // 面数
	TypeWodSetThreshold  // 阈值 >=
	TypeWodSetThresholdQ // 阈值 <=
	TypeDiceDC
	TypeDCSetInit
	TypeDCSetPool   // 骰池
	TypeDCSetPoints // 面数
	TypeLoadVarname
	TypeLoadFormatString
	TypeStore
	TypeHalt
	TypeSwap
	TypeLeftValueMark
	TypeDiceSetK
	TypeDiceSetQ
	TypeClearDetail

	TypePop

	TypeJmp
	TypeJe
	TypeJne
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

type ByteCode struct {
	T     CodeType
	Value interface{}
}

type RollExtraFlags struct {
	DiceMinMode        bool  // 骰子以最小值结算，用于获取下界
	DiceMaxMode        bool  // 以最大值结算 获取上界
	DisableLoadVarname bool  // 不允许加载变量，这是为了防止遇到 .r XXX 被当做属性读取，而不是“由于XXX，骰出了”
	IgnoreDiv0         bool  // 当div0时暂不报错
	DefaultDiceSideNum int64 // 默认骰子面数
}

type Context struct {
	parser *Parser

	Code      []ByteCode
	codeIndex int

	stack []VMValue
	top   int

	NumOpCount       int64  // 算力计数
	CocFlagVarPrefix string // 解析过程中出现，当VarNumber开启时有效，可以是困难极难常规大成功

	jmpStack     []int   // 跳转栈
	counterStack []int64 // f-string 嵌套计数，我记这个做什么？

	Flags RollExtraFlags // 标记
	Error error          // 报错信息

	Ret       *VMValue // 返回值
	RestInput string   // 剩余字符串
	Matched   string   // 匹配的字符串
}

func (e *Context) Init(stackLength int) {
	e.Code = make([]ByteCode, stackLength)
	e.jmpStack = []int{}
	e.counterStack = []int64{}
}

type VMValue struct {
	TypeId      VMValueType `json:"typeId"`
	Value       interface{} `json:"value"`
	ExpiredTime int64       `json:"expiredTime"`
}

func (v *VMValue) AsBool() bool {
	switch v.TypeId {
	case VMTypeInt64:
		return v.Value != int64(0)
	case VMTypeString:
		return v.Value != ""
	case VMTypeNone:
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
	if v.Value == nil {
		return "unknown"
	}
	switch v.TypeId {
	case VMTypeInt64:
		return strconv.FormatInt(v.Value.(int64), 10)
	case VMTypeFloat64:
		return strconv.FormatFloat(v.Value.(float64), 'f', 2, 64)
	case VMTypeString:
		return v.Value.(string)
	case VMTypeNone:
		return v.Value.(string)
	//case VMTypeComputedValue:
	//vd := v.Value.(*VMComputedValueData)
	//return vd.BaseValue.ToString() + "=> (" + vd.Expr + ")"
	default:
		return "a value"
	}
}

func (v *VMValue) ReadInt64() (int64, bool) {
	if v.TypeId == VMTypeInt64 {
		return v.Value.(int64), true
	}
	return 0, false
}

func (v *VMValue) ReadString() (string, bool) {
	if v.TypeId == VMTypeString {
		return v.Value.(string), true
	}
	return "", false
}

func (v *VMValue) OpAdd(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := v.Value.(int64) + v2.Value.(int64)
			return VMValueNewInt64(val)
		case VMTypeFloat64:
			val := float64(v.Value.(int64)) + v2.Value.(float64)
			return VMValueNewFloat64(val)
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := v.Value.(float64) + float64(v2.Value.(int64))
			return VMValueNewFloat64(val)
		case VMTypeFloat64:
			val := v.Value.(float64) + v2.Value.(float64)
			return VMValueNewFloat64(val)
		}
	case VMTypeString:
		switch v2.TypeId {
		case VMTypeString:
			val := v.Value.(string) + v2.Value.(string)
			return VMValueNewStr(val)
		}
	case VMTypeComputedValue:
		// TODO:
	case VMTypeArray:
		// TODO:
	}

	return nil
}

func (v *VMValue) OpSub(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := v.Value.(int64) - v2.Value.(int64)
			return VMValueNewInt64(val)
		case VMTypeFloat64:
			val := float64(v.Value.(int64)) - v2.Value.(float64)
			return VMValueNewFloat64(val)
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := v.Value.(float64) - float64(v2.Value.(int64))
			return VMValueNewFloat64(val)
		case VMTypeFloat64:
			val := v.Value.(float64) - v2.Value.(float64)
			return VMValueNewFloat64(val)
		}
	}

	return nil
}

func (v *VMValue) OpMultiply(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			// TODO: 溢出，均未考虑溢出
			val := v.Value.(int64) * v2.Value.(int64)
			return VMValueNewInt64(val)
		case VMTypeFloat64:
			val := float64(v.Value.(int64)) * v2.Value.(float64)
			return VMValueNewFloat64(val)
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := v.Value.(float64) * float64(v2.Value.(int64))
			return VMValueNewFloat64(val)
		case VMTypeFloat64:
			val := v.Value.(float64) * v2.Value.(float64)
			return VMValueNewFloat64(val)
		}
	}

	return nil
}

func (v *VMValue) OpDivide(ctx *Context, v2 *VMValue) *VMValue {
	// TODO: 被除数为0
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := v.Value.(int64) / v2.Value.(int64)
			return VMValueNewInt64(val)
		case VMTypeFloat64:
			val := float64(v.Value.(int64)) / v2.Value.(float64)
			return VMValueNewFloat64(val)
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := v.Value.(float64) / float64(v2.Value.(int64))
			return VMValueNewFloat64(val)
		case VMTypeFloat64:
			val := v.Value.(float64) / v2.Value.(float64)
			return VMValueNewFloat64(val)
		}
	}

	return nil
}

func (v *VMValue) OpModulus(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := v.Value.(int64) % v2.Value.(int64)
			return VMValueNewInt64(val)
		}
	}

	return nil
}

func (v *VMValue) OpPower(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := int64(math.Pow(float64(v.Value.(int64)), float64(v2.Value.(int64))))
			return VMValueNewInt64(val)
		case VMTypeFloat64:
			val := math.Pow(float64(v.Value.(int64)), v2.Value.(float64))
			return VMValueNewFloat64(val)
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := math.Pow(v.Value.(float64), float64(v2.Value.(int64)))
			return VMValueNewFloat64(val)
		case VMTypeFloat64:
			val := math.Pow(v.Value.(float64), v2.Value.(float64))
			return VMValueNewFloat64(val)
		}
	}

	return nil
}

func boolToVMValue(v bool) *VMValue {
	var val int64
	if v {
		val = 1
	}
	return VMValueNewInt64(val)
}

func (v *VMValue) OpCompLT(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(int64) < v2.Value.(int64))
		case VMTypeFloat64:
			return boolToVMValue(float64(v.Value.(int64)) < v2.Value.(float64))
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(float64) < float64(v2.Value.(int64)))
		case VMTypeFloat64:
			return boolToVMValue(v.Value.(float64) < v2.Value.(float64))
		}
	}

	return nil
}

func (v *VMValue) OpCompLE(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(int64) <= v2.Value.(int64))
		case VMTypeFloat64:
			return boolToVMValue(float64(v.Value.(int64)) <= v2.Value.(float64))
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(float64) <= float64(v2.Value.(int64)))
		case VMTypeFloat64:
			return boolToVMValue(v.Value.(float64) <= v2.Value.(float64))
		}
	}

	return nil
}

func (v *VMValue) OpCompEQ(ctx *Context, v2 *VMValue) *VMValue {
	if v == v2 {
		return VMValueNewInt64(1)
	}
	if v.TypeId == v2.TypeId {
		return boolToVMValue(v.Value == v2.Value)
	}

	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeFloat64:
			return boolToVMValue(float64(v.Value.(int64)) == v2.Value.(float64))
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(float64) == float64(v2.Value.(int64)))
		}
	}

	return VMValueNewInt64(0)
}

func (v *VMValue) OpCompNE(ctx *Context, v2 *VMValue) *VMValue {
	ret := v.OpCompEQ(ctx, v2)
	return boolToVMValue(!ret.AsBool())
}

func (v *VMValue) OpCompGE(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(int64) >= v2.Value.(int64))
		case VMTypeFloat64:
			return boolToVMValue(float64(v.Value.(int64)) >= v2.Value.(float64))
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(float64) >= float64(v2.Value.(int64)))
		case VMTypeFloat64:
			return boolToVMValue(v.Value.(float64) >= v2.Value.(float64))
		}
	}

	return nil
}

func (v *VMValue) OpCompGT(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(int64) > v2.Value.(int64))
		case VMTypeFloat64:
			return boolToVMValue(float64(v.Value.(int64)) > v2.Value.(float64))
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(float64) > float64(v2.Value.(int64)))
		case VMTypeFloat64:
			return boolToVMValue(v.Value.(float64) > v2.Value.(float64))
		}
	}

	return nil
}

func (v *VMValue) GetTypeName() string {
	switch v.TypeId {
	case VMTypeInt64:
		return "int64"
	case VMTypeFloat64:
		return "float64"
	case VMTypeString:
		return "str"
	case VMTypeNone:
		return "none"
	case VMTypeComputedValue:
		return "computed"
	case VMTypeArray:
		return "array"
	}
	return "unknown"
}

func VMValueNewInt64(i int64) *VMValue {
	// TODO: 小整数可以处理为不可变对象，且一直停留在内存中，就像python那样。这可以避免很多内存申请
	return &VMValue{TypeId: VMTypeInt64, Value: i}
}

func VMValueNewFloat64(i float64) *VMValue {
	return &VMValue{TypeId: VMTypeFloat64, Value: i}
}

func VMValueNewStr(s string) *VMValue {
	return &VMValue{TypeId: VMTypeString, Value: s}
}
