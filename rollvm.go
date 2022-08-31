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
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func (e *Parser) checkStackOverflow() bool {
	if e.Error != nil {
		return true
	}
	if e.CodeIndex >= len(e.Code) {
		need := len(e.Code) * 2
		if need <= 8192 {
			newCode := make([]ByteCode, need)
			copy(newCode, e.Code)
			e.Code = newCode
		} else {
			e.Error = errors.New("E1:指令虚拟机栈溢出，请不要发送过长的指令")
			return true
		}
	}
	return false
}

func (e *Parser) WriteCode(T CodeType, value interface{}) {
	if e.checkStackOverflow() {
		return
	}

	c := &e.Code[e.CodeIndex]
	c.T = T
	c.Value = value
	e.CodeIndex += 1
}

func (e *Parser) AddLeftValueMark() {
	e.WriteCode(TypeLeftValueMark, nil)
}

func (e *Parser) LMark() {
	e.WriteCode(TypeLeftValueMark, nil)
}

func (e *Parser) AddOperator(operator CodeType) {
	e.WriteCode(operator, nil)
}

func (e *Parser) PushIntNumber(value string) {
	val, _ := strconv.ParseInt(value, 10, 64)
	e.WriteCode(TypePushIntNumber, int64(val))
}

func (e *Parser) PushFloatNumber(value string) {
	val, _ := strconv.ParseFloat(value, 64)
	e.WriteCode(TypePushFloatNumber, float64(val))
}

type Runtime struct {
	parser    *Parser
	Flags     RollExtraFlags // 注: flag 之类还是不要写这，这样无法复用
	RestInput string
	Matched   string
	Ret       *VMValue
}

func (e *Runtime) Run(value string) error {
	var err error

	// 创建parser并初始化，这里是分词过程
	p := &Parser{Buffer: value}
	err = p.Init()

	// 初始化指令栈，默认指令长度512条
	p.RollContext.Init(512)

	// 开始解析，编译字节码
	err = p.Parse()
	p.RollContext.flags = e.Flags
	p.Execute()

	// 执行字节码
	p.Evaluate()
	if p.RollContext.Error != nil {
		return err
	}

	e.Ret = &p.RollContext.Stack[0]

	// 给出VM解析完句子后的剩余文本
	tks := p.Tokens()
	// 注意，golang的string下标等同于[]byte下标，也就是说中文会被打断
	// parser里有一个[]rune类型的，但问题是他句尾带了一个endsymbol
	runeBuffer := []rune(value)
	lastToken := tks[len(tks)-1]
	e.RestInput = strings.TrimSpace(string(runeBuffer[lastToken.end:]))
	e.Matched = strings.TrimSpace(string(runeBuffer[:lastToken.end]))

	return err
}

//getE5 := func() error {
//	return errors.New("E5: 超出单指令允许算力，不予计算")
//}

func DiceRoll64(dicePoints int64) int64 {
	if dicePoints == 0 {
		return 0
	}
	val := rand.Int63()%dicePoints + 1
	return val
}

func (e *Parser) Evaluate() {
	e.Top = 0
	e.Stack = make([]VMValue, 1000)
	stack := e.Stack

	//lastDetails := []string{}
	//lastDetailsLeft := []string{}
	//
	numOpCountAdd := func(count int64) bool {
		e.NumOpCount += count
		if e.NumOpCount > 30000 {
			return true
		}
		return false
	}

	diceStateIndex := -1
	var diceStates []struct {
		times    int64 // 次数，如 2d10，times为2
		isPickLH int64 // 为1对应取低个数，为2对应取高个数
		lowNum   int64
		highNum  int64
	}

	diceInit := func() {
		diceStateIndex += 1
		diceStates = append(diceStates, struct {
			times    int64 // 次数，如 2d10，times为2
			isPickLH int64 // 为1对应取低个数，为2对应取高个数
			lowNum   int64
			highNum  int64
		}{
			times: 1,
		})
	}

	stackPop := func() *VMValue {
		v := &e.Stack[e.Top-1]
		e.Top -= 1
		return v
	}

	stackPop2 := func() (*VMValue, *VMValue) {
		v2, v1 := stackPop(), stackPop()
		return v1, v2
	}

	stackPush := func(v *VMValue) {
		e.Stack[e.Top] = *v
		e.Top += 1
	}

	for opIndex := 0; opIndex < e.CodeIndex; opIndex += 1 {
		numOpCountAdd(1)
		code := e.Code[opIndex]
		cIndex := fmt.Sprintf("%d/%d", opIndex+1, e.CodeIndex)
		fmt.Println("!!!", code.CodeString(), time.Now().UnixMilli(), cIndex)

		switch code.T {
		case TypePushIntNumber:
			stack[e.Top].TypeId = VMTypeInt64
			stack[e.Top].Value = code.Value
			e.Top++
		case TypePushFloatNumber:
			stack[e.Top].TypeId = VMTypeFloat64
			stack[e.Top].Value = code.Value
			e.Top++
		case TypeDiceInit:
			diceInit()
		case TypeDiceSetTimes:
			v := stackPop()
			diceStates[len(diceStates)-1].times, _ = v.ReadInt64()
		case TypeDiceSetPickLowNum:
			v := stackPop()
			diceStates[len(diceStates)-1].isPickLH = 1
			diceStates[len(diceStates)-1].lowNum, _ = v.ReadInt64()
		case TypeDiceSetPickHighNum:
			v := stackPop()
			diceStates[len(diceStates)-1].isPickLH = 2
			diceStates[len(diceStates)-1].highNum, _ = v.ReadInt64()
		case TypeAdd, TypeSubtract, TypeMultiply, TypeDivide, TypeModulus,
			TypeCompLT, TypeCompLE, TypeCompEQ, TypeCompNE, TypeCompGE, TypeCompGT:
			// 所有二元运算符
			v1, v2 := stackPop2()
			opFunc := binOperator[code.T-TypeAdd]
			stackPush(opFunc(v1, v2))
		case TypeDice:
			diceState := diceStates[len(diceStates)-1]
			var nums []int64
			bInt := int64(100)
			for i := int64(0); i < diceState.times; i += 1 {
				if e.flags.MinDiceMode {
					nums = append(nums, bInt)
				} else {
					nums = append(nums, DiceRoll64(bInt))
				}
			}
		}
	}
}

func (code *ByteCode) CodeString() string {
	switch code.T {
	case TypePushIntNumber:
		return "push " + strconv.FormatInt(code.Value.(int64), 10)
	case TypePushFloatNumber:
		return "push " + strconv.FormatFloat(code.Value.(float64), 'f', 2, 64)
	case TypePushString:
		return "push.str " + code.Value.(string)
	case TypeAdd:
		return "add"
	case TypeNegation, TypeSubtract:
		return "sub"
	case TypeMultiply:
		return "mul"
	case TypeDivide:
		return "div"
	case TypeModulus:
		return "mod"
	case TypeExponentiation:
		return "pow"
	case TypeDice:
		return "dice"
	case TypeDicePenalty:
		return "dice.penalty"
	case TypeDiceBonus:
		return "dice.bonus"
	case TypeDiceSetK:
		return "dice.setk"
	case TypeDiceSetQ:
		return "dice.setq"
	case TypeDiceUnary:
		return "dice1"
	case TypeDiceFate:
		return "dice.fate"
	case TypeWodSetInit:
		return "wod.init"
	case TypeWodSetPool:
		return "wod.pool"
	case TypeWodSetPoints:
		return "wod.points"
	case TypeWodSetThreshold:
		return "wod.threshold"
	case TypeWodSetThresholdQ:
		return "wod.thresholdQ"
	case TypeDiceDC:
		return "dice.dc"
	case TypeDCSetInit:
		return "dice.setInit"
	case TypeDCSetPool:
		return "dice.setPool"
	case TypeDCSetPoints:
		return "dice.setPoints"
	case TypeDiceWod:
		return "dice.wod"
	case TypeLoadVarname:
		return "ld.v " + code.Value.(string)
	case TypeLoadFormatString:
		return fmt.Sprintf("ld.fs %d, %s", code.Value, "code.ValueStr")
	case TypeStore:
		return "store"
	case TypeHalt:
		return "halt"
	case TypeSwap:
		return "swap"
	case TypeLeftValueMark:
		return "mark.left"
	case TypeJmp:
		return fmt.Sprintf("jmp %d", code.Value)
	case TypeJe:
		return fmt.Sprintf("je %d", code.Value)
	case TypeJne:
		return fmt.Sprintf("jne %d", code.Value)
	case TypeCompLT:
		return "comp.lt"
	case TypeCompLE:
		return "comp.le"
	case TypeCompEQ:
		return "comp.eq"
	case TypeCompNE:
		return "comp.ne"
	case TypeCompGE:
		return "comp.ge"
	case TypeCompGT:
		return "comp.gt"
	case TypePop:
		return "pop"
	case TypeClearDetail:
		return "reset"
	}
	return ""
}
