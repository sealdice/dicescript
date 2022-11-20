package dicescript

import (
	"errors"
	"strconv"
)

type ParserData struct {
	counterStack  []int64  // f-string 嵌套计数，在解析时中起作用
	varnameStack  []string // 另一个解析用栈
	jmpStack      []int64
	breakStack    []int64 // break用，用时创建
	continueStack []int64 // continue用，用时创建
	codeStack     []struct {
		code  []ByteCode
		index int
	}
}

func (pd *ParserData) init() {
	pd.counterStack = []int64{}
	pd.varnameStack = []string{}
	pd.jmpStack = []int64{} // 不复用counterStack的原因是在 ?: 算符中两个都有用到
	pd.codeStack = []struct {
		code  []ByteCode
		index int
	}{} // 用于处理函数
}

func (e *Parser) checkStackOverflow() bool {
	if e.Error != nil {
		return true
	}
	if e.codeIndex >= len(e.code) {
		need := len(e.code) * 2
		if need <= 8192 {
			newCode := make([]ByteCode, need)
			copy(newCode, e.code)
			e.code = newCode
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

	c := &e.code[e.codeIndex]
	c.T = T
	c.Value = value
	e.codeIndex += 1
}

func (e *Parser) LMark() {
	e.WriteCode(TypeLeftValueMark, nil)
}

func (e *Parser) AddOp(operator CodeType) {
	var val interface{} = nil
	if operator == TypeJne || operator == TypeJmp {
		val = int64(0)
	}
	e.WriteCode(operator, val)
}

func (e *Parser) AddLoadName(value string) {
	e.WriteCode(TypeLoadName, value)
}

func (e *Parser) PushIntNumber(value string) {
	val, _ := strconv.ParseInt(value, 10, 64)
	e.WriteCode(TypePushIntNumber, int64(val))
}

func (e *Parser) PushStr(value string) {
	e.WriteCode(TypePushString, value)
}

func (e *Parser) PushArray(value int64) {
	e.WriteCode(TypePushArray, value)
}

func (e *Parser) PushDict(value int64) {
	e.WriteCode(TypePushDict, value)
}

func (e *Parser) PushUndefined() {
	e.WriteCode(TypePushUndefined, nil)
}

func (e *Parser) PushThis() {
	e.WriteCode(TypePushThis, nil)
}

func (e *Parser) PushGlobal() {
	e.WriteCode(TypePushGlobal, nil)
}

func (e *Parser) AddFormatString(value string, num int64) {
	//e.PushStr(value)
	e.WriteCode(TypeLoadFormatString, num) // num
}

func (e *Parser) PushFloatNumber(value string) {
	val, _ := strconv.ParseFloat(value, 64)
	e.WriteCode(TypePushFloatNumber, float64(val))
}

func (e *Parser) AddStore(text string) {
	e.WriteCode(TypeStoreName, text)
}

func (e *Parser) AddStoreGlobal(text string) {
	e.WriteCode(TypeStoreNameGlobal, text)
}

func (e *Parser) AddStoreLocal(text string) {
	e.WriteCode(TypeStoreNameLocal, text)
}

func (e *Parser) NamePush(test string) {
	e.varnameStack = append(e.varnameStack, test)
}

func (e *Parser) NamePop() string {
	last := len(e.varnameStack) - 1
	val := e.varnameStack[last]
	e.varnameStack = e.varnameStack[:last]
	return val
}

func (e *Parser) OffsetPush() {
	e.jmpStack = append(e.jmpStack, int64(e.codeIndex)-1)
}

func (p *Parser) ContinuePush() {
	if p.continueStack == nil {
		p.continueStack = []int64{}
	}
	p.AddOp(TypeJmp)
	p.continueStack = append(p.continueStack, int64(p.codeIndex)-1)
}

func (p *Parser) ContinueSet(offsetB int) {
	if p.continueStack != nil {
		for _, codeIndex := range p.continueStack {
			lastB := len(p.jmpStack) - 1 - offsetB
			jmpIndex := p.jmpStack[lastB]
			// 试出来的，这个是对的，那么也许while那个是错的？？还是说因为while最后多push了一个jmp呢？
			p.code[codeIndex].Value = -(int64(codeIndex) - jmpIndex)
		}
	}
}

func (p *Parser) BreakSet() {
	if p.breakStack != nil {
		for _, codeIndex := range p.breakStack {
			p.code[codeIndex].Value = int64(p.codeIndex) - codeIndex - 1
		}
	}
}

func (p *Parser) BreakPush() {
	if p.breakStack == nil {
		p.breakStack = []int64{}
	}
	p.AddOp(TypeJmp)
	p.breakStack = append(p.breakStack, int64(p.codeIndex)-1)
}

func (e *Parser) OffsetPopAndSet() {
	last := len(e.jmpStack) - 1
	codeIndex := e.jmpStack[last]
	e.jmpStack = e.jmpStack[:last]
	e.code[codeIndex].Value = int64(int64(e.codeIndex) - codeIndex - 1)
	//fmt.Println("XXXX", e.Code[codeIndex], "|", e.Top, codeIndex)
}

func (e *Parser) OffsetPopN(num int) {
	last := len(e.jmpStack) - num
	e.jmpStack = e.jmpStack[:last]
}

func (e *Parser) OffsetJmpSetX(offsetA int, offsetB int, rev bool) {
	lastA := len(e.jmpStack) - 1 - offsetA
	lastB := len(e.jmpStack) - 1 - offsetB

	codeIndex := e.jmpStack[lastA]
	jmpIndex := e.jmpStack[lastB]

	if rev {
		e.code[codeIndex].Value = -(int64(e.codeIndex) - jmpIndex - 1)
	} else {
		e.code[codeIndex].Value = int64(e.codeIndex) - jmpIndex - 1
	}
}

func (e *Parser) CounterPush() {
	e.counterStack = append(e.counterStack, 0)
}

func (e *Parser) CounterAdd(offset int64) {
	last := len(e.counterStack) - 1
	if last != -1 {
		e.counterStack[last] += offset
	}
}

func (e *Parser) CounterPop() int64 {
	last := len(e.counterStack) - 1
	num := e.counterStack[last]
	e.counterStack = e.counterStack[:last]
	return num
}

func (e *Parser) AddInvokeMethod(name string, paramsNum int64) {
	e.WriteCode(TypePushIntNumber, paramsNum)
	e.WriteCode(TypeInvokeSelf, name)
}

func (e *Parser) AddInvoke(paramsNum int64) {
	//e.WriteCode(TypePushIntNumber, paramsNum)
	e.WriteCode(TypeInvoke, paramsNum)
}

func (p *Parser) AddStoreComputed(name string, text string) {
	code, length := p.CodePop()
	val := VMValueNewComputedRaw(&ComputedData{
		Expr:      text,
		code:      code,
		codeIndex: length,
	})

	p.WriteCode(TypePushComputed, val)
	p.WriteCode(TypeStoreName, name)
}

func (p *Parser) AddStoreFunction(name string, paramsReversed []string, text string) {
	code, length := p.CodePop()

	// 翻转一次
	for i, j := 0, len(paramsReversed)-1; i < j; i, j = i+1, j-1 {
		paramsReversed[i], paramsReversed[j] = paramsReversed[j], paramsReversed[i]
	}

	val := VMValueNewFunctionRaw(&FunctionData{
		Expr:      text,
		Name:      name,
		Params:    paramsReversed,
		code:      code,
		codeIndex: length,
	})

	p.WriteCode(TypePushFuction, val)
	p.WriteCode(TypeStoreName, name)
}

func (p *Parser) AddAttrSet(objName string, attr string, isRaw bool) {
	if isRaw {
		p.WriteCode(TypeLoadNameRaw, objName)
	} else {
		p.WriteCode(TypeLoadName, objName)
	}
	p.WriteCode(TypeAttrSet, attr)
}

func (p *Parser) CodePush() {
	p.codeStack = append(p.codeStack, struct {
		code  []ByteCode
		index int
	}{code: p.code, index: p.codeIndex})
	p.code = make([]ByteCode, 256)
	p.codeIndex = 0
}

func (p *Parser) CodePop() ([]ByteCode, int) {
	lastCode, lastIndex := p.code, p.codeIndex

	last := len(p.codeStack) - 1
	info := p.codeStack[last]
	p.codeStack = p.codeStack[:last]
	p.code = info.code
	p.codeIndex = info.index
	return lastCode, lastIndex
}
