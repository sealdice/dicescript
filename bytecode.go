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
	TypePushComputed
	TypePushUndefined
	TypePushNone
	TypePushThis
	TypePushFuction

	TypeLoadFormatString
	TypeLoadName
	TypeLoadNameRaw // 如遇到computed，这个版本不取出其内容
	TypeStoreName

	TypeInvoke
	TypeInvokeSelf
	TypeGetItem
	TypeSetItem
	TypeSetAttr
	TypeGetAttr

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
	TypeHalt
	TypeSwap
	TypeLeftValueMark
	TypeDiceSetK
	TypeDiceSetQ
	TypeClearDetail

	TypePop
	TypeNop

	TypeJmp
	TypeJe
	TypeJne
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
	case TypePushArray:
		return "push.arr " + strconv.FormatInt(code.Value.(int64), 10)
	case TypePushComputed:
		computed, _ := code.Value.(*VMValue).ReadComputed()
		return "push.computed " + computed.Expr
	case TypePushUndefined:
		return "push.undefined"
	case TypePushNone:
		return "push.none"
	case TypePushThis:
		return "push.this"
	case TypePushFuction:
		computed, _ := code.Value.(*VMValue).ReadFunctionData()
		return "push.func " + computed.Expr

	case TypeInvoke:
		return "invoke " + strconv.FormatInt(code.Value.(int64), 10)

	case TypeInvokeSelf:
		return "invoke.self " + code.Value.(string)
	case TypeGetItem:
		return "item.get"
	case TypeSetItem:
		return "item.set"
	case TypeSetAttr:
		return "attr.set " + code.Value.(string)
	case TypeGetAttr:
		return "attr.get " + code.Value.(string)

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

	case TypeDicePenalty:
		return "dice.penalty"
	case TypeDiceBonus:
		return "dice.bonus"
	case TypeDiceSetK:
		return "dice.setk"
	case TypeDiceSetQ:
		return "dice.setq"
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
	case TypeLoadName:
		return "ld " + code.Value.(string)
	case TypeLoadNameRaw:
		return "ld.raw " + code.Value.(string)
	case TypeLoadFormatString:
		return fmt.Sprintf("ld.fs %d", code.Value)
	case TypeStoreName:
		return fmt.Sprintf("store %s", code.Value)
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
	case TypeNop:
		return "nop"
	case TypePop:
		return "pop"
	case TypeReturn:
		return "ret"
	case TypeClearDetail:
		return "reset"
	}
	return ""
}
