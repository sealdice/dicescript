package dicescript

import (
	"fmt"
	"strconv"
)

type CodeType uint8

type ByteCode struct {
	T     CodeType
	Value interface{}
}

const (
	TypePushIntNumber CodeType = iota
	TypePushFloatNumber
	TypePushString
	TypePushArray
	TypePushDict
	TypePushRange
	TypePushComputed
	TypePushUndefined
	TypePushNone
	TypePushThis
	TypePushGlobal
	TypePushFunction
	TypePushLast

	TypeLoadFormatString
	TypeLoadName
	TypeLoadNameWithDetail
	TypeLoadNameRaw // 如遇到computed，这个版本不取出其内容
	TypeStoreName
	TypeStoreNameGlobal
	TypeStoreNameLocal

	TypeInvoke
	TypeInvokeSelf
	TypeItemGet
	TypeItemSet
	TypeAttrSet
	TypeGetAttr
	TypeSliceGet
	TypeSliceSet

	TypeAdd // 注意，修改顺序时一定要顺带修改下面的数组
	TypeSubtract
	TypeMultiply
	TypeDivide
	TypeModulus
	TypeExponentiation
	TypeNullCoalescing

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

	TypeNegation
	TypePositive

	TypeDiceInit
	TypeDiceSetTimes
	TypeDiceSetKeepLowNum
	TypeDiceSetKeepHighNum
	TypeDiceSetDropLowNum
	TypeDiceSetDropHighNum
	TypeDiceSetMin
	TypeDiceSetMax
	TypeDice

	TypeDiceCocPenalty
	TypeDiceCocBonus
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
	TypeHalt
	TypeDetailMark

	TypePop
	TypePopN

	TypeNop

	TypeJmp
	TypeJe
	TypeJne
	TypeJeDup
	TypeReturn
)

func (code *ByteCode) CodeString() string {
	switch code.T {
	case TypePushIntNumber:
		return "push.int " + strconv.FormatInt(code.Value.(int64), 10)
	case TypePushFloatNumber:
		return "push.flt " + strconv.FormatFloat(code.Value.(float64), 'f', 2, 64)
	case TypePushString:
		return "push.str " + code.Value.(string)
	case TypePushRange:
		return "push.range"
	case TypePushArray:
		return "push.arr " + strconv.FormatInt(code.Value.(int64), 10)
	case TypePushDict:
		return "push.dict " + strconv.FormatInt(code.Value.(int64), 10)
	case TypePushComputed:
		computed, _ := code.Value.(*VMValue).ReadComputed()
		return "push.computed " + computed.Expr
	case TypePushUndefined:
		return "push.undefined"
	case TypePushNone:
		return "push.none"
	case TypePushThis:
		return "push.this"
	case TypePushGlobal:
		return "push.global"
	case TypePushFunction:
		computed, _ := code.Value.(*VMValue).ReadFunctionData()
		return "push.func " + computed.Name

	case TypeInvoke:
		return "invoke " + strconv.FormatInt(code.Value.(int64), 10)

	case TypeInvokeSelf:
		return "invoke.self " + code.Value.(string)
	case TypeItemGet:
		return "item.get"
	case TypeItemSet:
		return "item.set"
	case TypeAttrSet:
		return "attr.set " + code.Value.(string)
	case TypeGetAttr:
		return "attr.get " + code.Value.(string)
	case TypeSliceGet:
		return "slice.get"
	case TypeSliceSet:
		return "slice.set"

	case TypeAdd:
		return "add"
	case TypeSubtract:
		return "sub"
	case TypeMultiply:
		return "mul"
	case TypeDivide:
		return "div"
	case TypeModulus:
		return "mod"
	case TypeExponentiation:
		return "pow"
	case TypeNullCoalescing:
		return "nullCoalescing"

	case TypeLogicAnd:
		return "and"
	case TypeLogicOr:
		return "or"

	case TypeBitwiseAnd:
		return "&"
	case TypeBitwiseOr:
		return "|"

	case TypeNegation:
		return "neg"
	case TypePositive:
		return "pos"

	case TypeDiceInit:
		return "dice.init"
	case TypeDiceSetTimes:
		return "dice.setTimes"
	case TypeDiceSetKeepLowNum:
		return "dice.setKeepLow"
	case TypeDiceSetKeepHighNum:
		return "dice.setKeepHigh"
	case TypeDiceSetDropLowNum:
		return "dice.setDropLow"
	case TypeDiceSetDropHighNum:
		return "dice.setDropHigh"
	case TypeDiceSetMin:
		return "dice.setMin"
	case TypeDiceSetMax:
		return "dice.setMax"
	case TypeDice:
		return "dice"

	case TypeDiceCocPenalty:
		return "coc.penalty"
	case TypeDiceCocBonus:
		return "coc.bonus"
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
		return "dc.setInit"
	case TypeDCSetPool:
		return "dc.setPool"
	case TypeDCSetPoints:
		return "dc.setPoints"
	case TypeDiceWod:
		return "dice.wod"
	case TypeLoadName:
		return "ld " + code.Value.(string)
	case TypeLoadNameWithDetail:
		return "ld.d " + code.Value.(string)
	case TypeLoadNameRaw:
		return "ld.raw " + code.Value.(string)
	case TypeLoadFormatString:
		return fmt.Sprintf("ld.fs %d", code.Value)
	case TypeStoreName:
		return fmt.Sprintf("store %s", code.Value)
	case TypeStoreNameGlobal:
		return fmt.Sprintf("store.global %s", code.Value)
	case TypeStoreNameLocal:
		return fmt.Sprintf("store.local %s", code.Value)
	case TypeHalt:
		return "halt"
	case TypeDetailMark:
		v := code.Value.(BufferSpan)
		return fmt.Sprintf("mark.detail %d, %d", v.begin, v.end)
	case TypeJmp:
		return fmt.Sprintf("jmp %d", code.Value)
	case TypeJe:
		return fmt.Sprintf("je %d", code.Value)
	case TypeJeDup:
		return fmt.Sprintf("je.dup %d", code.Value)
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
	case TypePushLast:
		return "push.last"
	case TypePop:
		return "pop"
	case TypePopN:
		return fmt.Sprintf("popn %d", code.Value)
	case TypeNop:
		return "nop"
	case TypeReturn:
		return "ret"
	}
	return ""
}
