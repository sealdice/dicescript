package dicescript

import (
	"errors"
	"math"
	"strconv"
)

func funcCeil(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	v, ok := params[0].ReadFloat()
	if ok {
		return VMValueNewInt(int64(math.Ceil(v)))
	} else {
		ctx.Error = errors.New("类型错误: 只能是float")
	}
	return nil
}

func funcRound(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	v, ok := params[0].ReadFloat()
	if ok {
		return VMValueNewInt(int64(math.Round(v)))
	} else {
		ctx.Error = errors.New("类型错误: 只能是float")
	}
	return nil
}

func funcFloor(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	v, ok := params[0].ReadFloat()
	if ok {
		return VMValueNewInt(int64(math.Floor(v)))
	} else {
		ctx.Error = errors.New("类型错误: 只能是float")
	}
	return nil
}

func funcInt(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	switch params[0].TypeId {
	case VMTypeInt:
		return params[0]
	case VMTypeFloat:
		v, _ := params[0].ReadFloat()
		return VMValueNewInt(int64(v))
	case VMTypeString:
		s, _ := params[0].ReadString()
		val, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			return VMValueNewInt(val)
		} else {
			ctx.Error = errors.New("值错误: 无法进行 int() 转换: " + s)
		}
	default:
		ctx.Error = errors.New("类型错误: 只能是数字类型")
	}
	return nil
}

func funcFloat(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	switch params[0].TypeId {
	case VMTypeInt:
		v, _ := params[0].ReadInt()
		return VMValueNewFloat(float64(v))
	case VMTypeFloat:
		return params[0]
	case VMTypeString:
		s, _ := params[0].ReadString()
		val, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return VMValueNewFloat(val)
		} else {
			ctx.Error = errors.New("值错误: 无法进行 float() 转换: " + s)
		}
	default:
		ctx.Error = errors.New("类型错误: 只能是数字类型")
	}
	return nil
}

func funcStr(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	return VMValueNewStr(params[0].ToString())
}

func funcDir(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	return VMValueNewStr(params[0].ToString())
}

var nnf = VMValueNewNativeFunction

type ndf = NativeFunctionData

var builtinValues = map[string]*VMValue{
	"ceil":  nnf(&ndf{"ceil", []string{"value"}, nil, nil, funcCeil}),
	"floor": nnf(&ndf{"floor", []string{"value"}, nil, nil, funcFloor}),
	"round": nnf(&ndf{"round", []string{"value"}, nil, nil, funcRound}),
	"int":   nnf(&ndf{"int", []string{"value"}, nil, nil, funcInt}),
	"float": nnf(&ndf{"float", []string{"value"}, nil, nil, funcFloat}),
	"str":   nnf(&ndf{"str", []string{"value"}, nil, nil, funcStr}),
}

//

func funcArrayKeepLow(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	isAllInt, ret := this.ArrayFuncKeepLow(ctx, params[0].MustReadInt())
	if isAllInt {
		return VMValueNewInt(int64(ret))
	} else {
		return VMValueNewFloat(ret)
	}
}

func funcArrayKeepHigh(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	isAllInt, ret := this.ArrayFuncKeepHigh(ctx, params[0].MustReadInt())
	if isAllInt {
		return VMValueNewInt(int64(ret))
	} else {
		return VMValueNewFloat(ret)
	}
}

func funcArraySum(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	arr, _ := this.ReadArray()

	isAllInt := true
	sumNum := float64(0)
	for _, i := range arr.List {
		switch i.TypeId {
		case VMTypeInt:
			sumNum += float64(i.MustReadInt())
		case VMTypeFloat:
			isAllInt = false
			sumNum += i.MustReadFloat()
		}
	}

	if isAllInt {
		return VMValueNewInt(int64(sumNum))
	} else {
		return VMValueNewFloat(sumNum)
	}
}

var builtinProto = map[VMValueType]*VMDictValue{
	VMTypeArray: VMValueMustNewDictWithArray(
		VMValueNewStr("kh"), nnf(&ndf{"Array.kh", []string{"num"}, []*VMValue{VMValueNewInt(1)}, nil, funcArrayKeepHigh}),
		VMValueNewStr("kl"), nnf(&ndf{"Array.kl", []string{"num"}, []*VMValue{VMValueNewInt(1)}, nil, funcArrayKeepLow}),
		VMValueNewStr("sum"), nnf(&ndf{"Array.sum", []string{}, nil, nil, funcArraySum}),
	),
}

func getBindMethod(v *VMValue, funcDef *VMValue) *VMValue {
	switch funcDef.TypeId {
	case VMTypeFunction:
		fd, _ := funcDef.ReadFunctionData()

		// 完成clone
		_fd := *fd
		fd2 := &_fd

		fd2.Self = v.Clone()
		return VMValueNewFunctionRaw(fd2)
	case VMTypeNativeFunction:
		fd, _ := funcDef.ReadNativeFunctionData()

		// 完成clone
		_fd := *fd
		fd2 := &_fd

		fd2.Self = v.Clone()
		return VMValueNewNativeFunction(fd2)
	}
	return nil
}

//func getBindMethod(name string, v *VMValue, params []string, nativeFunc NativeFunctionDef) *VMValue {
//	return nnf(&NativeFunctionData{name, params, v.Clone(), nativeFunc})
//}
