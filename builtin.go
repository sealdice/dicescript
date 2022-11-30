package dicescript

import (
	"errors"
	"math"
	"math/rand"
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

func funcAbs(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	v := params[0]
	switch v.TypeId {
	case VMTypeInt:
		val := v.MustReadInt()
		if val < 0 {
			return VMValueNewInt(-val)
		}
		return v
	case VMTypeFloat:
		val := v.MustReadFloat()
		if val < 0 {
			return VMValueNewFloat(-val)
		}
		return v
	}

	ctx.Error = errors.New("类型错误: 参数必须为int或float")
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
	"abs":   nnf(&ndf{"abs", []string{"value"}, nil, nil, funcAbs}),
	// TODO: roll()
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

func funcArrayLen(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	arr, _ := this.ReadArray()
	return VMValueNewInt(int64(len(arr.List)))
}

func funcArrayShuttle(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	arr, _ := this.ReadArray()

	lst := arr.List
	for i := len(lst) - 1; i > 0; i-- { // Fisher–Yates shuffle
		j := rand.Intn(i + 1)
		lst[i], lst[j] = lst[j], lst[i]
	}
	return this
}

func funcArrayRand(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	arr, _ := this.ReadArray()
	return arr.List[rand.Intn(len(arr.List))]
}

func funcArrayRandSize(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	arr, _ := this.ReadArray()
	newArr := VMValueNewArray(arr.List...)
	funcArrayShuttle(ctx, newArr, []*VMValue{})
	arr, _ = newArr.ReadArray()

	if val, ok := params[0].ReadInt(); ok {
		arr.List = arr.List[:val]
		return newArr
	} else {
		ctx.Error = errors.New("类型不符")
		return nil
	}
}

func funcArrayPop(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	arr, _ := this.ReadArray()
	if len(arr.List) > 1 {
		val := arr.List[len(arr.List)-1]
		arr.List = arr.List[:len(arr.List)-1]
		return val
	}
	return VMValueNewUndefined()
}

func funcArrayShift(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	arr, _ := this.ReadArray()
	if len(arr.List) > 1 {
		val := arr.List[0]
		arr.List = arr.List[1:]
		return val
	}
	return VMValueNewUndefined()
}

func funcArrayPush(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	arr, _ := this.ReadArray()
	arr.List = append(arr.List, params[0])
	return this
}

var builtinProto = map[VMValueType]*VMDictValue{
	VMTypeArray: VMValueMustNewDictWithArray(
		VMValueNewStr("kh"), nnf(&ndf{"Array.kh", []string{"num"}, []*VMValue{VMValueNewInt(1)}, nil, funcArrayKeepHigh}),
		VMValueNewStr("kl"), nnf(&ndf{"Array.kl", []string{"num"}, []*VMValue{VMValueNewInt(1)}, nil, funcArrayKeepLow}),
		VMValueNewStr("sum"), nnf(&ndf{"Array.sum", []string{}, nil, nil, funcArraySum}),
		VMValueNewStr("len"), nnf(&ndf{"Array.len", []string{}, nil, nil, funcArrayLen}),
		VMValueNewStr("shuffle"), nnf(&ndf{"Array.shuffle", []string{}, nil, nil, funcArrayShuttle}),
		VMValueNewStr("rand"), nnf(&ndf{"Array.rand", []string{}, nil, nil, funcArrayRand}),
		VMValueNewStr("randSize"), nnf(&ndf{"Array.rand", []string{"num"}, nil, nil, funcArrayRandSize}),
		VMValueNewStr("pop"), nnf(&ndf{"Array.pop", []string{}, nil, nil, funcArrayPop}),
		VMValueNewStr("shift"), nnf(&ndf{"Array.shift", []string{}, nil, nil, funcArrayShift}),
		VMValueNewStr("push"), nnf(&ndf{"Array.shift", []string{"value"}, nil, nil, funcArrayPush}),
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
