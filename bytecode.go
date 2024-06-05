package dicescript

import (
	"fmt"
	"strconv"
)

type CodeType uint8

type ByteCode struct {
	T     CodeType
	Value any
}

const (
	typePushIntNumber CodeType = iota
	typePushFloatNumber
	typePushString
	typePushArray
	typePushDict
	typePushRange
	typePushComputed
	typePushNull
	typePushThis
	typePushGlobal
	typePushFunction
	typePushLast
	typePushDefaultExpr

	typeLoadFormatString
	typeLoadName
	typeLoadNameWithDetail
	typeLoadNameRaw // 如遇到computed，这个版本不取出其内容
	typeStoreName
	typeStoreNameGlobal
	typeStoreNameLocal

	typeInvoke
	typeInvokeSelf
	typeItemGet
	typeItemSet
	typeAttrSet
	typeGetAttr
	typeSliceGet
	typeSliceSet

	typeAdd // 注意，修改顺序时一定要顺带修改下面的数组
	typeSubtract
	typeMultiply
	typeDivide
	typeModulus
	typeExponentiation
	typeNullCoalescing

	typeCompLT
	typeCompLE
	typeCompEQ
	typeCompNE
	typeCompGE
	typeCompGT

	typeBitwiseAnd
	typeBitwiseOr
	typeLogicAnd
	typeLogicOr

	typeNegation
	typePositive

	typeDiceInit
	typeDiceSetTimes
	typeDiceSetKeepLowNum
	typeDiceSetKeepHighNum
	typeDiceSetDropLowNum
	typeDiceSetDropHighNum
	typeDiceSetMin
	typeDiceSetMax
	typeDice

	typeDiceCocPenalty
	typeDiceCocBonus
	typeDiceFate
	typeDiceWod
	typeWodSetInit       // 重置参数
	typeWodSetPool       // 设置骰池(骰数)
	typeWodSetPoints     // 面数
	typeWodSetThreshold  // 阈值 >=
	typeWodSetThresholdQ // 阈值 <=
	typeDiceDC
	typeDCSetInit
	typeDCSetPool   // 骰池
	typeDCSetPoints // 面数
	typeHalt
	typeDetailMark

	typePop
	typePopN

	typeNop

	typeJmp
	typeJe
	typeJne
	typeJeDup
	typeReturn

	typeStSetName
	typeStModify
	typeStX0
	typeStX1
)

func (code *ByteCode) CodeString() string {
	switch code.T {
	case typePushIntNumber:
		return "push.int " + strconv.FormatInt(int64(code.Value.(IntType)), 10)
	case typePushFloatNumber:
		return "push.flt " + strconv.FormatFloat(code.Value.(float64), 'f', 2, 64)
	case typePushString:
		return "push.str " + code.Value.(string)
	case typePushRange:
		return "push.range"
	case typePushArray:
		return "push.arr " + strconv.FormatInt(int64(code.Value.(IntType)), 10)
	case typePushDict:
		return "push.dict " + strconv.FormatInt(int64(code.Value.(IntType)), 10)
	case typePushComputed:
		computed, _ := code.Value.(*VMValue).ReadComputed()
		return "push.computed " + computed.Expr
	case typePushNull:
		return "push.null"
	case typePushThis:
		return "push.this"
	case typePushGlobal:
		return "push.global"
	case typePushFunction:
		computed, _ := code.Value.(*VMValue).ReadFunctionData()
		return "push.func " + computed.Name

	case typeInvoke:
		return "invoke " + strconv.FormatInt(int64(code.Value.(IntType)), 10)

	case typeInvokeSelf:
		return "invoke.self " + code.Value.(string)
	case typeItemGet:
		return "item.get"
	case typeItemSet:
		return "item.set"
	case typeAttrSet:
		return "attr.set " + code.Value.(string)
	case typeGetAttr:
		return "attr.get " + code.Value.(string)
	case typeSliceGet:
		return "slice.get"
	case typeSliceSet:
		return "slice.set"

	case typeAdd:
		return "add"
	case typeSubtract:
		return "sub"
	case typeMultiply:
		return "mul"
	case typeDivide:
		return "div"
	case typeModulus:
		return "mod"
	case typeExponentiation:
		return "pow"
	case typeNullCoalescing:
		return "nullCoalescing"

	case typeLogicAnd:
		return "and"
	case typeLogicOr:
		return "or"

	case typeBitwiseAnd:
		return "&"
	case typeBitwiseOr:
		return "|"

	case typeNegation:
		return "neg"
	case typePositive:
		return "pos"

	case typeDiceInit:
		return "dice.init"
	case typeDiceSetTimes:
		return "dice.setTimes"
	case typeDiceSetKeepLowNum:
		return "dice.setKeepLow"
	case typeDiceSetKeepHighNum:
		return "dice.setKeepHigh"
	case typeDiceSetDropLowNum:
		return "dice.setDropLow"
	case typeDiceSetDropHighNum:
		return "dice.setDropHigh"
	case typeDiceSetMin:
		return "dice.setMin"
	case typeDiceSetMax:
		return "dice.setMax"
	case typeDice:
		return "dice"

	case typeDiceCocPenalty:
		return "coc.penalty"
	case typeDiceCocBonus:
		return "coc.bonus"
	case typeDiceFate:
		return "dice.fate"
	case typeWodSetInit:
		return "wod.init"
	case typeWodSetPool:
		return "wod.pool"
	case typeWodSetPoints:
		return "wod.points"
	case typeWodSetThreshold:
		return "wod.threshold"
	case typeWodSetThresholdQ:
		return "wod.thresholdQ"
	case typeDiceDC:
		return "dice.dc"
	case typeDCSetInit:
		return "dc.setInit"
	case typeDCSetPool:
		return "dc.setPool"
	case typeDCSetPoints:
		return "dc.setPoints"
	case typeDiceWod:
		return "dice.wod"
	case typeLoadName:
		return "ld " + code.Value.(string)
	case typeLoadNameWithDetail:
		return "ld.d " + code.Value.(string)
	case typeLoadNameRaw:
		return "ld.raw " + code.Value.(string)
	case typeLoadFormatString:
		return fmt.Sprintf("ld.fs %d", code.Value)
	case typeStoreName:
		return fmt.Sprintf("store %s", code.Value)
	case typeStoreNameGlobal:
		return fmt.Sprintf("store.global %s", code.Value)
	case typeStoreNameLocal:
		return fmt.Sprintf("store.local %s", code.Value)
	case typeHalt:
		return "halt"
	case typeDetailMark:
		v := code.Value.(BufferSpan)
		return fmt.Sprintf("mark.detail %d, %d", v.begin, v.end)
	case typeJmp:
		return fmt.Sprintf("jmp %d", code.Value)
	case typeJe:
		return fmt.Sprintf("je %d", code.Value)
	case typeJeDup:
		return fmt.Sprintf("je.dup %d", code.Value)
	case typeJne:
		return fmt.Sprintf("jne %d", code.Value)
	case typeCompLT:
		return "comp.lt"
	case typeCompLE:
		return "comp.le"
	case typeCompEQ:
		return "comp.eq"
	case typeCompNE:
		return "comp.ne"
	case typeCompGE:
		return "comp.ge"
	case typeCompGT:
		return "comp.gt"
	case typePushLast:
		return "push.last"
	case typePushDefaultExpr:
		return "push.def_expr"
	case typePop:
		return "pop"
	case typePopN:
		return fmt.Sprintf("popn %d", code.Value)
	case typeNop:
		return "nop"
	case typeReturn:
		return "ret"

	case typeStSetName:
		return "st.set"
	case typeStModify:
		return fmt.Sprintf("st.mod %s", code.Value)
	case typeStX0:
		return "st.x0"
	case typeStX1:
		return "st.x1"
	}
	return ""
}
