package dicescript

import (
	"errors"
	"math"
	"strconv"
)

func funcCeil(ctx *Context, params []*VMValue) *VMValue {
	v, ok := params[0].ReadFloat64()
	if ok {
		return VMValueNewInt64(int64(math.Ceil(v)))
	} else {
		ctx.Error = errors.New("类型错误: 只能是float")
	}
	return nil
}

func funcRound(ctx *Context, params []*VMValue) *VMValue {
	v, ok := params[0].ReadFloat64()
	if ok {
		return VMValueNewInt64(int64(math.Round(v)))
	} else {
		ctx.Error = errors.New("类型错误: 只能是float")
	}
	return nil
}

func funcFloor(ctx *Context, params []*VMValue) *VMValue {
	v, ok := params[0].ReadFloat64()
	if ok {
		return VMValueNewInt64(int64(math.Floor(v)))
	} else {
		ctx.Error = errors.New("类型错误: 只能是float")
	}
	return nil
}

func funcInt(ctx *Context, params []*VMValue) *VMValue {
	switch params[0].TypeId {
	case VMTypeInt64:
		return params[0]
	case VMTypeFloat64:
		v, _ := params[0].ReadFloat64()
		return VMValueNewInt64(int64(v))
	case VMTypeString:
		s, _ := params[0].ReadString()
		val, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			return VMValueNewInt64(val)
		} else {
			ctx.Error = errors.New("值错误: 无法进行 int() 转换: " + s)
		}
	default:
		ctx.Error = errors.New("类型错误: 只能是数字类型")
	}
	return nil
}

func funcFloat(ctx *Context, params []*VMValue) *VMValue {
	switch params[0].TypeId {
	case VMTypeInt64:
		v, _ := params[0].ReadInt64()
		return VMValueNewFloat64(float64(v))
	case VMTypeFloat64:
		return params[0]
	case VMTypeString:
		s, _ := params[0].ReadString()
		val, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return VMValueNewFloat64(val)
		} else {
			ctx.Error = errors.New("值错误: 无法进行 float() 转换: " + s)
		}
	default:
		ctx.Error = errors.New("类型错误: 只能是数字类型")
	}
	return nil
}

func funcStr(ctx *Context, params []*VMValue) *VMValue {
	return VMValueNewStr(params[0].ToString())
}

var nnf = VMValueNewNativeFunction
var builtinValues = map[string]*VMValue{
	"ceil":  nnf(&NativeFunctionData{"ceil", []string{"value"}, funcCeil}),
	"floor": nnf(&NativeFunctionData{"floor", []string{"value"}, funcFloor}),
	"round": nnf(&NativeFunctionData{"round", []string{"value"}, funcRound}),
	"int":   nnf(&NativeFunctionData{"int", []string{"value"}, funcInt}),
	"float": nnf(&NativeFunctionData{"float", []string{"value"}, funcFloat}),
	"str":   nnf(&NativeFunctionData{"str", []string{"value"}, funcStr}),
}
