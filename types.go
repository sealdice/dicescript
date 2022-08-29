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

	TypeExponentiation
	TypeDiceUnary

	TypeAdd // 注意，修改顺序时一定要顺带修改下面的数组
	TypeSubtract
	TypeMultiply
	TypeDivide
	TypeModulus

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
	TypeDiceSetPickLowNum
	TypeDiceSetPickHighNum
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

var binOperator = []func(*VMValue, *VMValue) *VMValue{
	(*VMValue).Add,
	(*VMValue).Sub,
	(*VMValue).Multiply,
	(*VMValue).Divide,
	(*VMValue).Modulus,

	(*VMValue).CompLT,
	(*VMValue).CompLE,
	(*VMValue).CompEQ,
	(*VMValue).CompNE,
	(*VMValue).CompGE,
	(*VMValue).CompGT,
}

type ByteCode struct {
	T     CodeType
	Value interface{}
}

type RollExtraFlags struct {
	MinDiceMode        bool  // 骰子以最小值结算，用于获取下界
	MaxDiceMode        bool  // 以最大值结算 获取上界
	DisableLoadVarname bool  // 不允许加载变量，这是为了防止遇到 .r XXX 被当做属性读取，而不是“由于XXX，骰出了”
	IgnoreDiv0         bool  // 当div0时暂不报错
	DefaultDiceSideNum int64 // 默认骰子面数
}

type RollContext struct {
	Code      []ByteCode
	CodeIndex int

	Stack []VMValue
	Top   int

	NumOpCount       int64  // 算力计数
	CocFlagVarPrefix string // 解析过程中出现，当VarNumber开启时有效，可以是困难极难常规大成功

	JmpStack     []int   // 跳转栈
	CounterStack []int64 // f-string 嵌套计数，我记这个做什么？

	flags RollExtraFlags // 标记
	Error error          // 报错信息
}

func (e *RollContext) Init(stackLength int) {
	e.Code = make([]ByteCode, stackLength)
	e.JmpStack = []int{}
	e.CounterStack = []int64{}
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
	switch v.TypeId {
	case VMTypeInt64:
		return strconv.FormatInt(v.Value.(int64), 10)
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

func (v *VMValue) Add(v2 *VMValue) *VMValue {
	// TODO: 先粗暴假设都是int，以后再改
	val := v.Value.(int64) + v2.Value.(int64)
	return &VMValue{VMTypeInt64, val, 0}
}

func (v *VMValue) Sub(v2 *VMValue) *VMValue {
	val := v.Value.(int64) - v2.Value.(int64)
	return &VMValue{VMTypeInt64, val, 0}
}

func (v *VMValue) Multiply(v2 *VMValue) *VMValue {
	val := v.Value.(int64) * v2.Value.(int64)
	return &VMValue{VMTypeInt64, val, 0}
}

func (v *VMValue) Divide(v2 *VMValue) *VMValue {
	val := v.Value.(int64) / v2.Value.(int64)
	return &VMValue{VMTypeInt64, val, 0}
}

func (v *VMValue) Modulus(v2 *VMValue) *VMValue {
	val := v.Value.(int64) % v2.Value.(int64)
	return &VMValue{VMTypeInt64, val, 0}
}

func (v *VMValue) CompLT(v2 *VMValue) *VMValue {
	var val int64
	ok := v.Value.(int64) < v2.Value.(int64)
	if ok {
		val = 1
	}
	return &VMValue{VMTypeInt64, val, 0}
}

func (v *VMValue) CompLE(v2 *VMValue) *VMValue {
	var val int64
	ok := v.Value.(int64) <= v2.Value.(int64)
	if ok {
		val = 1
	}
	return &VMValue{VMTypeInt64, val, 0}
}

func (v *VMValue) CompEQ(v2 *VMValue) *VMValue {
	var val int64
	ok := v.Value.(int64) == v2.Value.(int64)
	if ok {
		val = 1
	}
	return &VMValue{VMTypeInt64, val, 0}
}

func (v *VMValue) CompNE(v2 *VMValue) *VMValue {
	var val int64
	ok := v.Value.(int64) != v2.Value.(int64)
	if ok {
		val = 1
	}
	return &VMValue{VMTypeInt64, val, 0}
}

func (v *VMValue) CompGE(v2 *VMValue) *VMValue {
	var val int64
	ok := v.Value.(int64) >= v2.Value.(int64)
	if ok {
		val = 1
	}
	return &VMValue{VMTypeInt64, val, 0}
}

func (v *VMValue) CompGT(v2 *VMValue) *VMValue {
	var val int64
	ok := v.Value.(int64) > v2.Value.(int64)
	if ok {
		val = 1
	}
	return &VMValue{VMTypeInt64, val, 0}
}
