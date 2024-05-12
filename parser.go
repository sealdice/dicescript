package dicescript

import (
	"errors"
	"strconv"
)

type ParserData struct {
	code      []ByteCode
	codeIndex int

	Config        RollConfig
	flagsStack    []RollConfig
	counterStack  []IntType // f-string 嵌套计数，在解析时中起作用
	varnameStack  []string  // 另一个解析用栈
	jmpStack      []IntType
	breakStack    []IntType // break，用时创建
	continueStack []IntType // continue用，用时创建
	loopInfo      []struct {
		continueIndex int
		breakIndex    int
	}
	loopLayer int // 当前loop层数
	codeStack []struct {
		code  []ByteCode
		index int
	}
}

type BufferSpan struct {
	begin IntType
	end   IntType
	ret   *VMValue
	text  string
}

func (pd *ParserData) init() {
	pd.counterStack = []IntType{}
	pd.varnameStack = []string{}
	pd.jmpStack = []IntType{} // 不复用counterStack的原因是在 ?: 算符中两个都有用到
	pd.codeStack = []struct {
		code  []ByteCode
		index int
	}{} // 用于处理函数
}

func (e *ParserData) LoopBegin() {
	e.loopLayer += 1
	e.loopInfo = append(e.loopInfo, struct {
		continueIndex int
		breakIndex    int
	}{continueIndex: len(e.continueStack), breakIndex: len(e.breakStack)})
}

func (e *ParserData) LoopEnd() {
	e.loopLayer -= 1
	info := e.loopInfo[len(e.loopInfo)-1]
	e.continueStack = e.continueStack[:info.continueIndex]
	e.breakStack = e.breakStack[:info.breakIndex]
	e.loopInfo = e.loopInfo[:len(e.loopInfo)-1]
}

func (e *ParserData) checkStackOverflow() bool {
	if e.codeIndex >= len(e.code) {
		need := len(e.code) * 2
		if need <= 8192 {
			newCode := make([]ByteCode, need)
			copy(newCode, e.code)
			e.code = newCode
		} else {
			//e.Error = errors.New("E1:指令虚拟机栈溢出，请不要发送过长的指令")
			return true
		}
	}
	return false
}

func (e *ParserData) WriteCode(T CodeType, value any) {
	if e.checkStackOverflow() {
		return
	}

	c := &e.code[e.codeIndex]
	c.T = T
	c.Value = value
	e.codeIndex += 1
}

func (p *ParserData) AddDiceDetail(begin IntType, end IntType) {
	p.WriteCode(typeDetailMark, BufferSpan{begin: begin, end: end})
}

func (e *ParserData) AddOp(operator CodeType) {
	var val interface{} = nil
	if operator == typeJne || operator == typeJmp {
		val = IntType(0)
	}
	e.WriteCode(operator, val)
}

func (e *ParserData) AddLoadName(value string) {
	e.WriteCode(typeLoadName, value)
}

func (e *ParserData) PushIntNumber(value string) {
	val, _ := strconv.ParseInt(value, 10, 64)
	e.WriteCode(typePushIntNumber, IntType(val))
}

func (e *ParserData) PushStr(value string) {
	e.WriteCode(typePushString, value)
}

func (e *ParserData) PushArray(value IntType) {
	e.WriteCode(typePushArray, value)
}

func (e *ParserData) PushDict(value IntType) {
	e.WriteCode(typePushDict, value)
}

func (e *ParserData) PushUndefined() {
	e.WriteCode(typePushUndefined, nil)
}

func (e *ParserData) PushThis() {
	e.WriteCode(typePushThis, nil)
}

func (e *ParserData) PushGlobal() {
	e.WriteCode(typePushGlobal, nil)
}

func (e *ParserData) AddFormatString(num IntType) {
	//e.PushStr(value)
	e.WriteCode(typeLoadFormatString, num) // num
}

func (e *ParserData) PushFloatNumber(value string) {
	val, _ := strconv.ParseFloat(value, 64)
	e.WriteCode(typePushFloatNumber, float64(val))
}

func (e *ParserData) AddStName() {
	e.WriteCode(typeStSetName, nil)
}

type StInfo struct {
	Op   string
	Text string
}

func (e *ParserData) AddStModify(op string, text string) {
	e.WriteCode(typeStModify, StInfo{op, text})
}

func (e *ParserData) AddStore(text string) {
	e.WriteCode(typeStoreName, text)
}

func (e *ParserData) AddStoreGlobal(text string) {
	e.WriteCode(typeStoreNameGlobal, text)
}

func (e *ParserData) AddStoreLocal(text string) {
	e.WriteCode(typeStoreNameLocal, text)
}

func (e *ParserData) NamePush(test string) {
	e.varnameStack = append(e.varnameStack, test)
}

func (e *ParserData) NamePop() string {
	last := len(e.varnameStack) - 1
	val := e.varnameStack[last]
	e.varnameStack = e.varnameStack[:last]
	return val
}

func (e *ParserData) OffsetPush() {
	e.jmpStack = append(e.jmpStack, IntType(e.codeIndex)-1)
}

func (p *ParserData) ContinuePush() error {
	if p.loopLayer > 0 {
		if p.continueStack == nil {
			p.continueStack = []IntType{}
		}
		p.AddOp(typeJmp)
		p.continueStack = append(p.continueStack, IntType(p.codeIndex)-1)
	} else {
		return errors.New("循环外不能放置continue")
	}
	return nil
}

func (p *ParserData) ContinueSet(offsetB int) {
	if p.continueStack != nil {
		info := p.loopInfo[len(p.loopInfo)-1]
		for _, codeIndex := range p.continueStack[info.continueIndex:] {
			lastB := len(p.jmpStack) - 1 - offsetB
			jmpIndex := p.jmpStack[lastB]
			// 试出来的，这个是对的，那么也许while那个是错的？？还是说因为while最后多push了一个jmp呢？
			p.code[codeIndex].Value = -(IntType(codeIndex) - jmpIndex)
		}
	}
}

func (p *ParserData) BreakSet() {
	if p.breakStack != nil {
		info := p.loopInfo[len(p.loopInfo)-1]
		for _, codeIndex := range p.breakStack[info.breakIndex:] {
			p.code[codeIndex].Value = IntType(p.codeIndex) - codeIndex - 1
		}
	}
}

func (p *ParserData) BreakPush() error {
	if p.loopLayer > 0 {
		if p.breakStack == nil {
			p.breakStack = []IntType{}
		}
		p.AddOp(typeJmp)
		p.breakStack = append(p.breakStack, IntType(p.codeIndex)-1)
		return nil
	} else {
		return errors.New("循环外不能放置break")
	}
}

func (e *ParserData) OffsetPopAndSet() {
	last := len(e.jmpStack) - 1
	codeIndex := e.jmpStack[last]
	e.jmpStack = e.jmpStack[:last]
	e.code[codeIndex].Value = IntType(IntType(e.codeIndex) - codeIndex - 1)
	//fmt.Println("XXXX", e.Code[codeIndex], "|", e.Top, codeIndex)
}

func (e *ParserData) OffsetPopN(num int) {
	last := len(e.jmpStack) - num
	e.jmpStack = e.jmpStack[:last]
}

func (e *ParserData) OffsetJmpSetX(offsetA int, offsetB int, rev bool) {
	lastA := len(e.jmpStack) - 1 - offsetA
	lastB := len(e.jmpStack) - 1 - offsetB

	codeIndex := e.jmpStack[lastA]
	jmpIndex := e.jmpStack[lastB]

	if rev {
		e.code[codeIndex].Value = -(IntType(e.codeIndex) - jmpIndex - 1)
	} else {
		e.code[codeIndex].Value = IntType(e.codeIndex) - jmpIndex - 1
	}
}

func (e *ParserData) CounterPush() {
	e.counterStack = append(e.counterStack, 0)
}

func (e *ParserData) CounterAdd(offset IntType) {
	last := len(e.counterStack) - 1
	if last != -1 {
		e.counterStack[last] += offset
	}
}

func (e *ParserData) CounterPop() IntType {
	last := len(e.counterStack) - 1
	num := e.counterStack[last]
	e.counterStack = e.counterStack[:last]
	return num
}

func (e *ParserData) FlagsPush() {
	e.flagsStack = append(e.flagsStack, e.Config)
}

func (e *ParserData) FlagsPop() {
	last := len(e.flagsStack) - 1
	e.Config = e.flagsStack[last]
	e.flagsStack = e.flagsStack[:last]
}

func (e *ParserData) AddInvokeMethod(name string, paramsNum IntType) {
	e.WriteCode(typePushIntNumber, paramsNum)
	e.WriteCode(typeInvokeSelf, name)
}

func (e *ParserData) AddInvoke(paramsNum IntType) {
	//e.WriteCode(typePushIntNumber, paramsNum)
	e.WriteCode(typeInvoke, paramsNum)
}

func (p *ParserData) AddStoreComputed(name string, text string) {
	code, length := p.CodePop()
	val := VMValueNewComputedRaw(&ComputedData{
		Expr:      text,
		code:      code,
		codeIndex: length,
	})

	p.WriteCode(typePushComputed, val)
	p.WriteCode(typeStoreName, name)
}

func (p *ParserData) AddStoreComputedOnStack(text string) {
	code, length := p.CodePop()
	val := VMValueNewComputedRaw(&ComputedData{
		Expr:      text,
		code:      code,
		codeIndex: length,
	})

	p.WriteCode(typePushComputed, val)
}

func (p *ParserData) AddStoreFunction(name string, paramsReversed []string, text string) {
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

	p.WriteCode(typePushFunction, val)
	if name != "" {
		p.WriteCode(typeStoreName, name)
	}
}

func (p *ParserData) AddAttrSet(objName string, attr string, isRaw bool) {
	if isRaw {
		p.WriteCode(typeLoadNameRaw, objName)
	} else {
		p.WriteCode(typeLoadName, objName)
	}
	p.WriteCode(typeAttrSet, attr)
}

func (p *ParserData) CodePush() {
	p.codeStack = append(p.codeStack, struct {
		code  []ByteCode
		index int
	}{code: p.code, index: p.codeIndex})
	p.code = make([]ByteCode, 256)
	p.codeIndex = 0
}

func (p *ParserData) CodePop() ([]ByteCode, int) {
	lastCode, lastIndex := p.code, p.codeIndex

	last := len(p.codeStack) - 1
	info := p.codeStack[last]
	p.codeStack = p.codeStack[:last]
	p.code = info.code
	p.codeIndex = info.index
	return lastCode, lastIndex
}
