/*
  Copyright [2022] fy0748@gmail.com

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

type VMValueType int

const (
	VMTypeNumber        VMValueType = 0 // float
	VMTypeString        VMValueType = 1
	VMTypeNone          VMValueType = 2
	VMTypeComputedValue VMValueType = 4
	VMTypeArray         VMValueType = 5
)

type CodeType uint8

const (
	TypePushNumber CodeType = iota
	TypePushString
	TypeNegation
	TypeAdd
	TypeSubtract
	TypeMultiply
	TypeDivide
	TypeModulus
	TypeExponentiation
	TypeDiceUnary
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

	TypeCompLT
	TypeCompLE
	TypeCompEQ
	TypeCompNE
	TypeCompGE
	TypeCompGT

	TypeJmp
	TypeJe
	TypeJne

	TypeBitwiseAnd
	TypeBitwiseOr
	TypeLogicAnd
	TypeLogicOr
)

type ByteCode struct {
	T     CodeType
	Value interface{}
}

type RollExtraFlags struct {
	BigFailDiceOn      bool
	DisableLoadVarname bool  // 不允许加载变量，这是为了防止遇到 .r XXX 被当做属性读取，而不是“由于XXX，骰出了”
	IgnoreDiv0         bool  // 当div0时暂不报错
	DefaultDiceSideNum int64 // 默认骰子面数
}

type RollContext struct {
	Code []ByteCode
	Top  int

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
