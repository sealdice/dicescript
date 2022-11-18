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
	"sort"
	"strconv"
	"strings"
	"time"
)

func NewVM() *Context {
	// 创建parser
	p := &Parser{}
	p.ParserData.init()
	p.Context.Init()
	p.parser = p

	return &p.Context
}

func (ctx *Context) Run(value string) error {
	var err error

	// 初始化Parser，这里是分词过程
	p := ctx.parser
	p.Buffer = value
	err = p.Init()

	// 初始化指令栈，默认指令长度512条，会自动增长
	p.code = make([]ByteCode, 512)
	p.codeIndex = 0
	ctx.Error = nil
	ctx.NumOpCount = 0

	// 开始解析，编译字节码
	err = p.Parse()
	p.Execute()

	// 执行字节码
	p.Evaluate()
	if ctx.Error != nil {
		return ctx.Error
	}

	// 获取结果
	if ctx.top != 0 {
		ctx.Ret = &ctx.stack[ctx.top-1]
	} else {
		ctx.Ret = VMValueNewUndefined()
	}

	// 给出VM解析完句子后的剩余文本
	tks := p.Tokens()
	if len(tks) > 0 {
		// 注意，golang的string下标等同于[]byte下标，也就是说中文会被打断
		// parser里有一个[]rune类型的，但问题是他句尾带了一个endsymbol
		runeBuffer := []rune(value)
		lastToken := tks[len(tks)-1]
		ctx.RestInput = strings.TrimSpace(string(runeBuffer[lastToken.end:]))
		ctx.Matched = strings.TrimSpace(string(runeBuffer[:lastToken.end]))
	} else {
		ctx.RestInput = ""
		ctx.Matched = ""
	}

	return err
}

//getE5 := func() error {
//	return errors.New("E5: 超出单指令允许算力，不予计算")
//}

func Roll(dicePoints int64) int64 {
	if dicePoints == 0 {
		return 0
	}
	val := rand.Int63()%dicePoints + 1
	return val
}

func (e *Parser) Evaluate() {
	e.top = 0
	e.stack = make([]VMValue, 1000)
	stack := e.stack

	ctx := &e.Context
	//lastDetails := []string{}
	//lastDetailsLeft := []string{}
	//
	numOpCountAdd := func(count int64) bool {
		e.NumOpCount += count
		if e.NumOpCount > 30000 {
			ctx.Error = errors.New("允许算力上限")
			return true
		}
		return false
	}

	diceStateIndex := -1
	var diceStates []struct {
		times    int64 // 次数，如 2d10，times为2
		isKeepLH int64 // 为1对应取低个数，为2对应取高个数
		lowNum   int64
		highNum  int64
		min      *int64
		max      *int64
	}

	diceInit := func() {
		diceStateIndex += 1
		diceStates = append(diceStates, struct {
			times    int64 // 次数，如 2d10，times为2
			isKeepLH int64 // 为1对应取低个数，为2对应取高个数
			lowNum   int64
			highNum  int64
			min      *int64
			max      *int64
		}{
			times: 1,
		})
	}

	stackPop := func() *VMValue {
		v := &e.stack[e.top-1]
		e.top -= 1
		return v
	}

	stackPop2 := func() (*VMValue, *VMValue) {
		v2, v1 := stackPop(), stackPop()
		return v1, v2
	}

	stackPopN := func(num int64) []*VMValue {
		var data []*VMValue
		for i := int64(0); i < num; i++ {
			data = append(data, stackPop().Clone()) // 复制一遍规避栈问题
		}
		for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
			data[i], data[j] = data[j], data[i]
		}
		return data
	}

	stackPush := func(v *VMValue) {
		e.stack[e.top] = *v
		e.top += 1
	}

	startTime := time.Now().UnixMilli()
	for opIndex := 0; opIndex < e.codeIndex; opIndex += 1 {
		numOpCountAdd(1)
		if ctx.Error != nil {
			return
		}

		code := e.code[opIndex]
		cIndex := fmt.Sprintf("%d/%d", opIndex+1, e.codeIndex)
		if ctx.Flags.PrintBytecode {
			var subThread string
			if ctx.subThreadDepth != 0 {
				subThread = fmt.Sprintf("  S%d", ctx.subThreadDepth)
			}
			fmt.Printf("!!! %-20s %s %dms%s\n", code.CodeString(), cIndex, time.Now().UnixMilli()-startTime, subThread)
		}

		switch code.T {
		case TypePushIntNumber:
			stack[e.top].TypeId = VMTypeInt
			stack[e.top].Value = code.Value
			e.top++
		case TypePushFloatNumber:
			stack[e.top].TypeId = VMTypeFloat
			stack[e.top].Value = code.Value
			e.top++
		case TypePushString:
			s := code.Value.(string)
			unquote, err := strconv.Unquote(`"` + strings.ReplaceAll(s, `"`, `\"`) + `"`)
			if err != nil {
				unquote = s
			}
			stack[e.top].TypeId = VMTypeString
			stack[e.top].Value = unquote
			e.top++
		case TypePushArray:
			num := code.Value.(int64)
			stackPush(VMValueNewArray(stackPopN(num)...))
		case TypePushComputed, TypePushFuction:
			val := code.Value.(*VMValue)
			stackPush(val)
		case TypePushUndefined:
			stackPush(VMValueNewUndefined())
		case TypePushThis:
			//if ctx.currentThis != nil {
			//	stackPush(ctx.currentThis)
			//} else {
			stackPush(vmValueNewLocal())
			//}
		case TypePushGlobal:
			stackPush(vmValueNewGlobal())

		case TypePushRange:
			a, b := stackPop2()
			_a, ok1 := a.ReadInt()
			_b, ok2 := b.ReadInt()
			if !(ok1 && ok2) {
				ctx.Error = errors.New("左右两个区间必须都是数字类型")
				return
			}

			step := int64(1)
			length := _b - _a
			if length < 0 {
				step = -1
				length = -length
			}
			length += 1

			if length > 512 {
				ctx.Error = errors.New("不能一次性创建过长的数组")
				return
			}

			arr := make([]*VMValue, length)
			index := 0
			for i := _a; ; i += step {
				arr[index] = VMValueNewInt(i)
				index++
				if i == _b {
					break
				}
			}
			stackPush(VMValueNewArray(arr...))

		case TypeInvoke:
			paramsNum := code.Value.(int64)
			arr := stackPopN(paramsNum)
			funcObj := stackPop()

			if funcObj.TypeId == VMTypeFunction {
				ret := funcObj.FuncInvoke(ctx, arr)
				if ctx.Error != nil {
					return
				}
				stackPush(ret)
			} else if funcObj.TypeId == VMTypeNativeFunction {
				ret := funcObj.FuncInvokeNative(ctx, arr)
				if ctx.Error != nil {
					return
				}
				stackPush(ret)
			} else {
				ctx.Error = errors.New("无法调用")
			}

		case TypeInvokeSelf:
			paramsNum, _ := stackPop().ReadInt()
			arr := stackPopN(paramsNum)
			stackPush(arr[0].CallFunc(ctx, code.Value.(string), arr[1:]))
		case TypeGetItem:
			itemIndex, _ := stackPop().ReadInt() // 这有类型问题
			arr := stackPop()
			v := arr.ArrayGetItem(ctx, itemIndex)
			if ctx.Error != nil {
				return
			}
			stackPush(v)
		case TypeSetItem:
			val := stackPop()                    // 右值
			itemIndex, _ := stackPop().ReadInt() // 下标
			arr := stackPop()                    // 数组
			arr.ArraySetItem(ctx, itemIndex, val)
			if ctx.Error != nil {
				return
			}

		case TypeSetAttr:
			attrVal, obj := stackPop2()
			attrName := code.Value.(string)

			ret := obj.SetAttr(attrName, attrVal)
			if ctx.Error == nil && ret == nil {
				ctx.Error = errors.New("不支持的类型：当前变量无法用.来设置属性")
			}
			if ctx.Error != nil {
				return
			}
		case TypeGetAttr:
			obj := stackPop()
			attrName := code.Value.(string)
			ret := obj.GetAttr(ctx, attrName)
			if ctx.Error != nil {
				return
			}
			if ret == nil {
				ctx.Error = errors.New("不支持的类型：当前变量无法用.来取属性")
				return
			}
			stackPush(ret)
		case TypeSliceGet:
			step := stackPop() // step
			if step.TypeId != VMTypeUndefined {
				ctx.Error = errors.New("尚不支持分片步长")
				return
			}

			a, b := stackPop2()
			obj := stackPop()
			ret := obj.GetSliceEx(ctx, a, b)
			if ctx.Error != nil {
				return
			}
			stackPush(ret)
		case TypeSliceSet:
			val := stackPop()
			step := stackPop() // step
			if step.TypeId != VMTypeUndefined {
				ctx.Error = errors.New("尚不支持分片步长")
				return
			}

			a, b := stackPop2()
			obj := stackPop()
			obj.SetSliceEx(ctx, a, b, val)
			if ctx.Error != nil {
				return
			}

		case TypeReturn:
			return
		case TypeHalt:
			return

		case TypeLoadFormatString:
			num := int(code.Value.(int64))

			outStr := ""
			for index := 0; index < num; index++ {
				var val VMValue
				if e.top-num+index < 0 {
					e.Error = errors.New("E3:无效的表达式")
					return
				} else {
					val = stack[e.top-num+index]
				}
				outStr += val.ToString()
			}

			e.top -= num
			stack[e.top].TypeId = VMTypeString
			stack[e.top].Value = outStr
			e.top++
		case TypeLoadName, TypeLoadNameRaw:
			name := code.Value.(string)
			val := ctx.LoadName(name, TypeLoadNameRaw == code.T)
			if ctx.Error != nil {
				return
			}
			stackPush(val)

		case TypeStoreName:
			v := stackPop()
			name := code.Value.(string)

			ctx.StoreName(name, v)
			if ctx.Error != nil {
				return
			}

		case TypeJne:
			t := stackPop()
			if !t.AsBool() {
				opIndex += int(code.Value.(int64))
			}
		case TypeJmp:
			opIndex += int(code.Value.(int64))
		case TypePop:
			stackPop()

		case TypeAdd, TypeSubtract, TypeMultiply, TypeDivide, TypeModulus, TypeExponentiation,
			TypeCompLT, TypeCompLE, TypeCompEQ, TypeCompNE, TypeCompGE, TypeCompGT:
			// 所有二元运算符
			v1, v2 := stackPop2()
			opFunc := binOperator[code.T-TypeAdd]
			ret := opFunc(v1, ctx, v2)
			if ctx.Error == nil && ret == nil {
				// TODO: 整理所有错误类型
				opErr := fmt.Sprintf("这两种类型无法使用 %s 算符连接: %s, %s", code.CodeString(), v1.GetTypeName(), v2.GetTypeName())
				ctx.Error = errors.New(opErr)
			}
			if ctx.Error != nil {
				return
			}
			stackPush(ret)

		case TypePositive, TypeNegation:
			v := stackPop()
			var ret *VMValue
			if code.T == TypePositive {
				ret = v.OpPositive()
			} else {
				ret = v.OpNegation()
			}
			if ret == nil {
				// TODO: 整理所有错误类型
				opErr := fmt.Sprintf("此类型无法使用一元算符 %s: %s", code.CodeString(), v.GetTypeName())
				ctx.Error = errors.New(opErr)
			}
			if ctx.Error != nil {
				return
			}
			stackPush(ret)

		case TypeDiceInit:
			diceInit()
		case TypeDiceSetTimes:
			v := stackPop()
			diceStates[len(diceStates)-1].times, _ = v.ReadInt()
		case TypeDiceSetKeepLowNum:
			v := stackPop()
			diceStates[len(diceStates)-1].isKeepLH = 1
			diceStates[len(diceStates)-1].lowNum, _ = v.ReadInt()
		case TypeDiceSetKeepHighNum:
			v := stackPop()
			diceStates[len(diceStates)-1].isKeepLH = 2
			diceStates[len(diceStates)-1].highNum, _ = v.ReadInt()
		case TypeDiceSetDropLowNum:
			v := stackPop()
			diceStates[len(diceStates)-1].isKeepLH = 3
			diceStates[len(diceStates)-1].lowNum, _ = v.ReadInt()
		case TypeDiceSetMin:
			v := stackPop()
			i, _ := v.ReadInt()
			diceStates[len(diceStates)-1].min = &i
		case TypeDiceSetMax:
			v := stackPop()
			i, _ := v.ReadInt()
			diceStates[len(diceStates)-1].max = &i
		case TypeDiceSetDropHighNum:
			v := stackPop()
			diceStates[len(diceStates)-1].isKeepLH = 4
			diceStates[len(diceStates)-1].highNum, _ = v.ReadInt()
		case TypeDice:
			diceState := diceStates[len(diceStates)-1]
			var nums []int64
			val := stackPop()
			bInt, _ := val.ReadInt()

			numOpCountAdd(diceState.times)
			if ctx.Error != nil {
				return
			}

			for i := int64(0); i < diceState.times; i += 1 {
				oneDice := Roll(bInt)
				if diceState.max != nil {
					if oneDice > *diceState.max {
						oneDice = *diceState.max
					}
				}
				if diceState.min != nil {
					if oneDice < *diceState.min {
						oneDice = *diceState.min
					}
				}
				nums = append(nums, oneDice)
			}

			pickNum := diceState.times

			if diceState.isKeepLH != 0 {
				if diceState.isKeepLH == 1 || diceState.isKeepLH == 3 {
					pickNum = diceState.lowNum
					sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] }) // 从小到大
				} else {
					pickNum = diceState.highNum
					sort.Slice(nums, func(i, j int) bool { return nums[i] > nums[j] }) // 从大到小
				}
				if diceState.isKeepLH > 2 {
					pickNum = diceState.times - pickNum
				}
			}

			num := int64(0)
			for i := int64(0); i < pickNum; i++ {
				// 当取数大于上限 跳过
				if i >= int64(len(nums)) {
					continue
				}
				num += nums[i]
			}

			stackPush(VMValueNewInt(num))
		}
	}
}

func (e *Context) GetAsmText() string {
	ret := ""
	ret += "=== VM Code ===\n"
	for index, i := range e.code {
		if index >= e.codeIndex {
			break
		}
		s := i.CodeString()
		if s != "" {
			ret += s + "\n"
		} else {
			ret += "@raw: " + strconv.FormatInt(int64(i.T), 10) + "\n"
		}
	}
	ret += "=== VM Code End===\n"
	return ret
}
