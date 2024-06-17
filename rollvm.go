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
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func NewVM() *Context {
	// 创建parser
	p := &Context{}
	p.Init()
	return p
}

// RunExpr 注: 最后不一定叫这个名字，这个函数作用是，即使当前vm被占用，也能执行语句，是为了指令hack而服务的
func (ctx *Context) RunExpr(value string, useUpCtxLocal bool) (*VMValue, error) {
	val := NewFunctionValRaw(&FunctionData{
		Expr:      value,
		Name:      "",
		Params:    nil,
		code:      nil,
		codeIndex: 0,
	})

	v := val.FuncInvokeRaw(ctx, nil, useUpCtxLocal)
	return v, ctx.Error
}

// GetErrorText 主要用于js，因为ctx.Error是数组，在js那边不被当做正常的异常对象，所以会报错
func (ctx *Context) GetErrorText() string {
	if ctx.Error != nil {
		return ctx.Error.Error()
	}
	return ""
}

func (ctx *Context) Parse(value string) error {
	// 检测是否正在执行，正在执行则使用新的上下文
	if ctx.IsRunning {
		return errors.New("正在执行中，无法执行新的语句")
	}

	p := newParser("", []byte(value))
	ctx.parser = p
	d := p.cur.data
	// p.debug = true

	// 初始化指令栈，默认指令长度512条，会自动增长
	d.code = make([]ByteCode, 512)
	d.codeIndex = 0
	d.Config = ctx.Config
	ctx.Error = nil
	ctx.NumOpCount = 0
	ctx.detailCache = ""

	// 开始解析，编译字节码
	if ctx.Config.ParseExprLimit != 0 {
		p.maxExprCnt = ctx.Config.ParseExprLimit
	}
	_, err := p.parse(nil)
	if err != nil {
		ctx.Error = err
		return err
	}

	ctx.code = p.cur.data.code
	ctx.codeIndex = p.cur.data.codeIndex

	return nil
}

// IsCalculateExists 只有表达式被解析后，才能被调用，暂不考虑存在invoke指令的情况
func (ctx *Context) IsCalculateExists() bool {
	for _, i := range ctx.code {
		switch i.T {
		case typeDice, typeDiceDC, typeDiceWod, typeDiceFate, typeDiceCocBonus, typeDiceCocPenalty:
			return true
		case typeAdd, typeSubtract, typeMultiply, typeDivide, typeModulus, typeExponentiation:
			return true
		case typeInvoke, typeInvokeSelf:
			return true
		}
	}
	return false
}

// IsV1IfCompatibleExists 是否存在v1的fstring-if兼容指令
func (ctx *Context) IsV1IfCompatibleExists() bool {
	for _, i := range ctx.code {
		switch i.T {
		case typeV1IfMark:
			return true
		}
	}
	return false
}

func (ctx *Context) RunAfterParsed() error {
	ctx.V1IfCompatibleCount = 0
	ctx.IsComputedLoaded = false
	// 以下为eval
	ctx.evaluate()
	if ctx.Error != nil {
		return ctx.Error
	}

	// 获取结果
	if ctx.top != 0 {
		ctx.Ret = &ctx.stack[ctx.top-1]
	} else {
		ctx.Ret = NewNullVal()
	}

	// 给出VM解析完句子后的剩余文本
	offset := ctx.parser.pt.offset
	matched := strings.TrimRightFunc(string(ctx.parser.data[:offset]), func(r rune) bool {
		return unicode.IsSpace(r)
	})
	ctx.Matched = matched
	ctx.RestInput = string(ctx.parser.data[len(matched):])
	return nil
}

// Run 执行给定语句
func (ctx *Context) Run(value string) error {
	if err := ctx.Parse(value); err != nil {
		return err
	}
	return ctx.RunAfterParsed()
}

type spanByBegin []BufferSpan

func (a spanByBegin) Len() int           { return len(a) }
func (a spanByBegin) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a spanByBegin) Less(i, j int) bool { return a[i].Begin < a[j].Begin }

type spanByEnd []BufferSpan

func (a spanByEnd) Len() int           { return len(a) }
func (a spanByEnd) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a spanByEnd) Less(i, j int) bool { return a[i].End < a[j].End }

// getE5 := func() error {
//	return errors.New("E5: 超出单指令允许算力，不予计算")
// }

func (ctx *Context) makeDetailStr(details []BufferSpan) string {
	if ctx.Config.CustomMakeDetailFunc != nil {
		return ctx.Config.CustomMakeDetailFunc(ctx, details, ctx.parser.data)
	}
	detailResult := ctx.parser.data

	curPoint := IntType(-1) // nolint
	lastEnd := IntType(-1)  // nolint

	var m []struct {
		begin IntType
		end   IntType
		spans []BufferSpan
	}

	for _, i := range details {
		// fmt.Println("?", i, lastEnd)
		if i.Begin > lastEnd {
			curPoint = i.Begin
			m = append(m, struct {
				begin IntType
				end   IntType
				spans []BufferSpan
			}{begin: curPoint, end: i.End, spans: []BufferSpan{i}})
		} else {
			m[len(m)-1].spans = append(m[len(m)-1].spans, i)
			if i.End > m[len(m)-1].end {
				m[len(m)-1].end = i.End
			}
		}

		if i.End > lastEnd {
			lastEnd = i.End
		}
	}

	for i := len(m) - 1; i >= 0; i-- {
		// for i := 0; i < len(m); i++ {
		item := m[i]
		size := len(item.spans)
		sort.Sort(spanByEnd(item.spans))
		last := item.spans[size-1]

		subDetailsText := ""
		if size > 1 {
			// 次级结果，如 (10d3)d5 中，此处为10d3的结果
			// 例如 (10d3)d5=63[(10d3)d5=...,10d3=19]
			for j := 0; j < len(item.spans)-1; j++ {
				span := item.spans[j]
				subDetailsText += "," + string(detailResult[span.Begin:span.End]) + "=" + span.Ret.ToString()
			}
		}

		exprText := string(detailResult[item.begin:item.end])

		var r []byte
		r = append(r, detailResult[:item.begin]...)

		// 主体结果部分，如 (10d3)d5=63[(10d3)d5=63=2+2+2+5+2+5+5+4+1+3+4+1+4+5+4+3+4+5+2,10d3=19]
		detail := "[" + exprText + "=" + last.Ret.ToString()
		if last.Text != "" {
			detail += "=" + last.Text
		}
		detail += subDetailsText + "]"

		r = append(r, ([]byte)(last.Ret.ToString()+detail)...)
		r = append(r, detailResult[item.end:]...)
		detailResult = r
	}

	return string(detailResult)
}

func (ctx *Context) evaluate() {
	ctx.top = 0
	ctx.stack = make([]VMValue, 1000)
	ctx.IsRunning = true
	stack := ctx.stack
	defer func() {
		ctx.IsRunning = false // 如果程序崩掉，不过halt
	}()

	e := ctx
	// ctx := &e.Context
	var details []BufferSpan
	numOpCountAdd := func(count IntType) bool {
		e.NumOpCount += count
		if ctx.Config.OpCountLimit > 0 && e.NumOpCount > ctx.Config.OpCountLimit {
			ctx.Error = errors.New("允许算力上限")
			return true
		}
		return false
	}

	diceStateIndex := -1
	var diceStates []struct {
		times    IntType // 次数，如 2d10，times为2
		isKeepLH IntType // 为1对应取低个数，为2对应取高个数，3为丢弃低个数，4为丢弃高个数
		lowNum   IntType
		highNum  IntType
		min      *IntType
		max      *IntType
	}

	diceInit := func() {
		diceStateIndex += 1
		data := struct {
			times    IntType // 次数，如 2d10，times为2
			isKeepLH IntType // 为1对应取低个数，为2对应取高个数
			lowNum   IntType
			highNum  IntType
			min      *IntType
			max      *IntType
		}{
			times: 1,
		}

		if diceStateIndex >= len(diceStates) {
			diceStates = append(diceStates, data)
		} else {
			// 其实我不太清楚这样是否对效率有提升。。
			diceStates[diceStateIndex] = data
		}
	}

	var wodState struct {
		pool      IntType
		points    IntType
		threshold IntType
		isGE      bool
	}

	wodInit := func() {
		wodState.pool = 1
		wodState.points = 10   // 面数，默认d10
		wodState.threshold = 8 // 成功线，默认9
		wodState.isGE = true
	}

	var dcState struct {
		pool   IntType
		points IntType
	}

	dcInit := func() {
		dcState.pool = 1    // 骰数，默认1
		dcState.points = 10 // 面数，默认d10
	}

	solveDetail := func() {
		if ctx.subThreadDepth != 0 {
			return
		}
		sort.Sort(spanByBegin(details))
		ctx.DetailSpans = details
	}

	var lastPop *VMValue
	stackPop := func() *VMValue {
		v := &e.stack[e.top-1]
		e.top -= 1
		lastPop = v
		return v
	}

	stackPop2 := func() (*VMValue, *VMValue) {
		v2, v1 := stackPop(), stackPop()
		lastPop = v1
		return v1, v2
	}

	stackPopN := func(num IntType) []*VMValue {
		var data []*VMValue
		for i := IntType(0); i < num; i++ {
			data = append(data, stackPop().Clone()) // 复制一遍规避栈问题
		}
		for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
			data[i], data[j] = data[j], data[i]
		}
		if num >= 1 {
			lastPop = data[0]
		}
		return data
	}

	stackPush := func(v *VMValue) {
		e.stack[e.top] = *v
		e.top += 1
	}

	var fstrBlockStack [5]int
	var fstrBlockIndex int

	startTime := time.Now().UnixMilli()
	for opIndex := 0; opIndex < e.codeIndex; opIndex += 1 {
		numOpCountAdd(1)

		if ctx.Error == nil && e.top == len(stack) {
			ctx.Error = errors.New("执行栈到达溢出线")
		}

		if ctx.Error != nil {
			return
		}

		code := e.code[opIndex]
		cIndex := fmt.Sprintf("%d/%d", opIndex+1, e.codeIndex)
		if ctx.Config.PrintBytecode {
			var subThread string
			if ctx.subThreadDepth != 0 {
				subThread = fmt.Sprintf("  S%d", ctx.subThreadDepth)
			}
			fmt.Printf("!!! %-20s %s %dms%s\n", code.CodeString(), cIndex, time.Now().UnixMilli()-startTime, subThread)
		}

		switch code.T {
		case typePushIntNumber:
			stack[e.top].TypeId = VMTypeInt
			stack[e.top].Value = code.Value
			e.top++
		case typePushFloatNumber:
			stack[e.top].TypeId = VMTypeFloat
			stack[e.top].Value = code.Value
			e.top++
		case typePushString:
			s := code.Value.(string)
			stack[e.top].TypeId = VMTypeString
			stack[e.top].Value = s
			e.top++
		case typePushArray:
			num := code.Value.(IntType)
			stackPush(NewArrayVal(stackPopN(num)...))
		case typePushDict:
			num := code.Value.(IntType)
			items := stackPopN(num * 2)
			dict, err := NewDictValWithArray(items...)
			if err != nil {
				e.Error = err
				return
			}
			stackPush(dict.V())
		case typePushComputed, typePushFunction:
			val := code.Value.(*VMValue)
			stackPush(val)
		case typePushNull:
			stackPush(NewNullVal())
		case typePushThis:
			stackPush(vmValueNewLocal())
		// case typePushGlobal:
		//	stackPush(vmValueNewGlobal())

		case typePushRange:
			a, b := stackPop2()
			_a, ok1 := a.ReadInt()
			_b, ok2 := b.ReadInt()
			if !(ok1 && ok2) {
				ctx.Error = errors.New("左右两个区间必须都是数字类型")
				return
			}

			step := IntType(1)
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
				arr[index] = NewIntVal(i)
				index++
				if i == _b {
					break
				}
			}
			stackPush(NewArrayVal(arr...))
		case typePushLast:
			if lastPop == nil {
				ctx.Error = errors.New("非法调用指令 push.last")
				return
			}
			stackPush(lastPop)
		case typePushDefaultExpr:
			// 创建一个函数对象，然后调用它
			if ctx.Config.DefaultDiceSideExpr != "" {
				var val *VMValue

				// 检查缓存
				if ctx.Config.defaultDiceSideExprCacheFunc != nil {
					fd, ok := ctx.Config.defaultDiceSideExprCacheFunc.ReadFunctionData()
					if ok {
						if fd.Expr == ctx.Config.DefaultDiceSideExpr {
							val = ctx.Config.defaultDiceSideExprCacheFunc
						}
					}
				}

				if val == nil {
					val = NewFunctionValRaw(&FunctionData{
						Expr:      ctx.Config.DefaultDiceSideExpr,
						Name:      "",
						Params:    nil,
						code:      nil,
						codeIndex: 0,
					})
					ctx.Config.defaultDiceSideExprCacheFunc = val
				}

				v := val.FuncInvoke(ctx, nil)
				if ctx.Error != nil {
					return
				}
				stackPush(v)
			} else {
				stackPush(NewIntVal(100))
			}

		case typeLogicAnd:
			a, b := stackPop2()
			if !a.AsBool() {
				stackPush(a)
			} else {
				stackPush(b)
			}

		case typeInvoke:
			paramsNum := code.Value.(IntType)
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
				ctx.Error = errors.New("类型错误: 无法调用，必须是一个函数")
			}

		case typeItemGet:
			itemIndex := stackPop()
			obj := stackPop()
			ret := obj.ItemGet(ctx, itemIndex)
			if ctx.Error != nil {
				return
			}
			if ret == nil {
				ret = NewNullVal()
			}
			stackPush(ret)
		case typeItemSet:
			val := stackPop()       // 右值
			itemIndex := stackPop() // 下标
			obj := stackPop()       // 数组 / 对象
			obj.ItemSet(ctx, itemIndex, val.Clone())
			if ctx.Error != nil {
				return
			}
		case typeAttrSet:
			attrVal, obj := stackPop2()
			attrName := code.Value.(string)

			ret := obj.AttrSet(ctx, attrName, attrVal.Clone())
			if ctx.Error == nil && ret == nil {
				ctx.Error = errors.New("不支持的类型：当前变量无法用.来设置属性")
			}
			if ctx.Error != nil {
				return
			}
		case typeAttrGet:
			obj := stackPop()
			attrName := code.Value.(string)
			ret := obj.AttrGet(ctx, attrName)
			if ctx.Error != nil {
				return
			}
			if ret == nil {
				ctx.Error = errors.New("不支持的类型：当前变量无法用.来取属性")
				return
			}
			stackPush(ret)
		case typeSliceGet:
			step := stackPop() // step
			if step.TypeId != VMTypeNull {
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
		case typeSliceSet:
			val := stackPop()
			step := stackPop() // step
			if step.TypeId != VMTypeNull {
				ctx.Error = errors.New("尚不支持分片步长")
				return
			}

			a, b := stackPop2()
			obj := stackPop()
			obj.SetSliceEx(ctx, a, b, val)
			if ctx.Error != nil {
				return
			}

		case typeReturn:
			solveDetail()
			ctx.IsRunning = false
			return
		case typeHalt:
			solveDetail()
			ctx.IsRunning = false
			return

		case typeLoadFormatString:
			num := int(code.Value.(IntType))

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
		case typeLoadName, typeLoadNameRaw, typeLoadNameWithDetail:
			name := code.Value.(string)
			val := ctx.LoadName(name, typeLoadNameRaw == code.T, true)
			if ctx.Error != nil {
				return
			}
			if typeLoadNameWithDetail == code.T {
				details[len(details)-1].Tag = "load"
				details[len(details)-1].Ret = val
				details[len(details)-1].Text = ""
			}
			if ctx.Config.HookFuncValueLoadOverwrite != nil {
				val = ctx.Config.HookFuncValueLoadOverwrite(name, val, &details[len(details)-1])
			}
			stackPush(val)

		case typeStoreName:
			v := e.stack[e.top-1].Clone()
			name := code.Value.(string)

			ctx.StoreName(name, v, true)
			if ctx.Error != nil {
				return
			}

		case typeJe, typeJeDup:
			v := stackPop()
			if v.AsBool() {
				opIndex += int(code.Value.(IntType))
				if code.T == typeJeDup {
					stackPush(v)
				}
			}
		case typeJne:
			t := stackPop()
			if !t.AsBool() {
				opIndex += int(code.Value.(IntType))
			}
		case typeJmp:
			opIndex += int(code.Value.(IntType))
		case typePop:
			stackPop()
		case typePopN:
			stackPopN(code.Value.(IntType))

		case typeAdd, typeSubtract, typeMultiply, typeDivide, typeModulus, typeExponentiation, typeNullCoalescing,
			typeCompLT, typeCompLE, typeCompEQ, typeCompNE, typeCompGE, typeCompGT,
			typeBitwiseAnd, typeBitwiseOr:
			// 所有二元运算符
			v1, v2 := stackPop2()
			opFunc := binOperator[code.T-typeAdd]
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

		case typePositive, typeNegation:
			v := stackPop()
			var ret *VMValue
			if code.T == typePositive {
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

		case typeDiceInit:
			diceInit()
		case typeDiceSetTimes:
			v := stackPop()
			times, ok := v.ReadInt()
			if !ok || times <= 0 {
				ctx.Error = errors.New("骰点次数不为正整数")
				return
			}
			diceStates[diceStateIndex].times = times
		case typeDiceSetKeepLowNum:
			v := stackPop()
			diceStates[diceStateIndex].isKeepLH = 1
			diceStates[diceStateIndex].lowNum, _ = v.ReadInt()
		case typeDiceSetKeepHighNum:
			v := stackPop()
			diceStates[diceStateIndex].isKeepLH = 2
			diceStates[diceStateIndex].highNum, _ = v.ReadInt()
		case typeDiceSetDropLowNum:
			v := stackPop()
			diceStates[diceStateIndex].isKeepLH = 3
			diceStates[diceStateIndex].lowNum, _ = v.ReadInt()
		case typeDiceSetDropHighNum:
			v := stackPop()
			diceStates[diceStateIndex].isKeepLH = 4
			diceStates[diceStateIndex].highNum, _ = v.ReadInt()
		case typeDiceSetMin:
			v := stackPop()
			i, _ := v.ReadInt()
			diceStates[diceStateIndex].min = &i
		case typeDiceSetMax:
			v := stackPop()
			i, _ := v.ReadInt()
			diceStates[diceStateIndex].max = &i
		case typeDetailMark:
			span := code.Value.(BufferSpan)
			details = append(details, span)
		case typeDice:
			diceState := diceStates[diceStateIndex]

			val := stackPop()
			bInt, ok := val.ReadInt()
			if !ok || bInt <= 0 {
				ctx.Error = errors.New("骰子面数不为正整数")
				return
			}
			if ok && (diceState.isKeepLH == 1 || diceState.isKeepLH == 3) && diceState.lowNum <= 0 {
				ctx.Error = errors.New("骰子取低个数不为正整数")
				return
			}
			if ok && (diceState.isKeepLH == 2 || diceState.isKeepLH == 4) && diceState.highNum <= 0 {
				ctx.Error = errors.New("骰子取高个数不为正整数")
				return
			}

			numOpCountAdd(diceState.times)
			if ctx.Error != nil {
				return
			}

			num, detail := RollCommon(ctx.randSrc, diceState.times, bInt, diceState.min, diceState.max, diceState.isKeepLH, diceState.lowNum, diceState.highNum)
			diceStateIndex -= 1

			ret := NewIntVal(num)
			details[len(details)-1].Ret = ret
			details[len(details)-1].Text = detail
			details[len(details)-1].Tag = "dice"
			stackPush(ret)

		case typeDiceFate:
			sum, detail := RollFate(ctx.randSrc)
			ret := NewIntVal(sum)
			details[len(details)-1].Ret = ret
			details[len(details)-1].Text = detail
			details[len(details)-1].Tag = "dice-fate"
			stackPush(ret)

		case typeDiceCocBonus, typeDiceCocPenalty:
			t := stackPop()
			diceNum := t.MustReadInt()

			if numOpCountAdd(diceNum) {
				return
			}

			isBonus := code.T == typeDiceCocBonus
			r, detailText := RollCoC(ctx.randSrc, isBonus, diceNum)
			ret := NewIntVal(r)
			details[len(details)-1].Ret = ret
			details[len(details)-1].Text = detailText
			if isBonus {
				details[len(details)-1].Tag = "dice-coc-bonus"
			} else {
				details[len(details)-1].Tag = "dice-coc-penalty"
			}
			stackPush(ret)

		case typeWodSetInit:
			// WOD 系列
			wodInit()
		case typeWodSetPoints:
			v := stackPop()
			// if v.TypeId != VMTypeInt {
			//   // ...
			// }
			wodState.points = v.MustReadInt()
		case typeWodSetThreshold:
			v := stackPop()
			wodState.threshold = v.MustReadInt()
			wodState.isGE = true
		case typeWodSetThresholdQ:
			v := stackPop()
			wodState.threshold = v.MustReadInt()
			wodState.isGE = false
		case typeWodSetPool:
			v := stackPop()
			wodState.pool = v.MustReadInt()
		case typeDiceWod:
			v := stackPop() // 加骰线

			// 变量检查
			if !wodCheck(ctx, v.MustReadInt(), wodState.pool, wodState.points, wodState.threshold) {
				return
			}

			num, _, _, detailText := RollWoD(ctx.randSrc, v.MustReadInt(), wodState.pool, wodState.points, wodState.threshold, wodState.isGE)
			ret := NewIntVal(num)
			details[len(details)-1].Ret = ret
			details[len(details)-1].Text = detailText
			details[len(details)-1].Tag = "dice-wod"
			stackPush(ret)

		case typeDCSetInit:
			// Double Cross
			dcInit()
		case typeDCSetPool:
			v := stackPop()
			dcState.pool = v.MustReadInt()
		case typeDCSetPoints:
			v := stackPop()
			dcState.points = v.MustReadInt()
		case typeDiceDC:
			v := stackPop() // 暴击值 / 也可以理解为加骰线
			if !doubleCrossCheck(ctx, v.MustReadInt(), dcState.pool, dcState.points) {
				return
			}
			success, _, _, detailText := RollDoubleCross(nil, v.MustReadInt(), dcState.pool, dcState.points)
			ret := NewIntVal(success)
			details[len(details)-1].Ret = ret
			details[len(details)-1].Text = detailText
			details[len(details)-1].Tag = "dice-dc"
			stackPush(ret)

		case typeFStringBlockPush:
			if fstrBlockIndex >= 4 {
				ctx.Error = errors.New("字符串模板嵌套层数过多")
				return
			}
			fstrBlockStack[fstrBlockIndex] = e.top
			fstrBlockIndex += 1
		case typeFStringBlockPop:
			newTop := fstrBlockStack[fstrBlockIndex-1]
			var v *VMValue
			if newTop != e.top {
				v = stackPop()
			}
			e.top = newTop
			fstrBlockIndex -= 1
			if v != nil {
				stackPush(v)
			} else {
				stackPush(NewStrVal(""))
			}
		case typeV1IfMark:
			// 满足条件: 首先在fstring中，其次栈里目前有东西
			if fstrBlockIndex > 0 {
				newTop := fstrBlockStack[fstrBlockIndex-1]
				if newTop != e.top {
					stackPush(NewStrVal("")) // 填入空字符串，模拟v1行为
					ctx.V1IfCompatibleCount += 1
				}
			}

		case typeStSetName:
			stName, stVal := stackPop2()
			if e.Config.CallbackSt != nil {
				name, _ := stName.ReadString()
				e.Config.CallbackSt("set", name, stVal.Clone(), nil, "", "")
			}
		case typeStModify:
			stName, stVal := stackPop2()
			stInfo := code.Value.(StInfo)

			if e.Config.CallbackSt != nil {
				name, _ := stName.ReadString()
				e.Config.CallbackSt("mod", name, stVal.Clone(), nil, stInfo.Op, stInfo.Text)
			}
		case typeStX0:
			stName, stVal := stackPop2()
			if e.Config.CallbackSt != nil {
				name, _ := stName.ReadString()
				e.Config.CallbackSt("set.x0", name, stVal.Clone(), nil, "", "")
			}
		case typeStX1:
			stVal := stackPop()
			stExtra := stackPop()
			stName := stackPop()
			if e.Config.CallbackSt != nil {
				name, _ := stName.ReadString()
				e.Config.CallbackSt("set.x1", name, stVal.Clone(), stExtra.Clone(), "", "")
			}
		}
	}
}

func (ctx *Context) GetAsmText() string {
	ret := ""
	ret += "=== VM Code ===\n"
	for index, i := range ctx.code {
		if index >= ctx.codeIndex {
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

func GetAsmText(code []ByteCode, codeIndex int) string {
	ret := ""
	ret += "=== VM Code ===\n"
	for index, i := range code {
		if index >= codeIndex {
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
