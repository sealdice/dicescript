package dicescript

import (
	"errors"
	"math"
	"strconv"
)

func funcCeil(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	if params[0].TypeId == VMTypeInt {
		return params[0]
	}
	v, ok := params[0].ReadFloat()
	if ok {
		return NewIntVal(IntType(math.Ceil(v)))
	} else {
		ctx.Error = errors.New("(ceil)类型错误: 只能是数字类型")
	}
	return nil
}

func funcRound(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	if params[0].TypeId == VMTypeInt {
		return params[0]
	}
	v, ok := params[0].ReadFloat()
	if ok {
		return NewIntVal(IntType(math.Round(v)))
	} else {
		ctx.Error = errors.New("(round)类型错误: 只能是数字类型")
	}
	return nil
}

func funcFloor(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	if params[0].TypeId == VMTypeInt {
		return params[0]
	}
	v, ok := params[0].ReadFloat()
	if ok {
		return NewIntVal(IntType(math.Floor(v)))
	} else {
		ctx.Error = errors.New("(floor)类型错误: 只能是数字类型")
	}
	return nil
}

func funcAbs(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	v := params[0]
	switch v.TypeId {
	case VMTypeInt:
		val := v.MustReadInt()
		if val < 0 {
			return NewIntVal(-val)
		}
		return v
	case VMTypeFloat:
		val := v.MustReadFloat()
		if val < 0 {
			return NewFloatVal(-val)
		}
		return v
	}

	ctx.Error = errors.New("(abs)类型错误: 参数必须为int或float")
	return nil
}

func funcBool(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	v := params[0]
	if v.AsBool() {
		return NewIntVal(1)
	}
	return NewIntVal(0)
}

func funcInt(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	switch params[0].TypeId {
	case VMTypeInt:
		return params[0]
	case VMTypeFloat:
		v, _ := params[0].ReadFloat()
		return NewIntVal(IntType(v))
	case VMTypeString:
		s, _ := params[0].ReadString()
		val, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			return NewIntVal(IntType(val))
		} else {
			ctx.Error = errors.New("(int)值错误: 无法进行 int() 转换: " + s)
		}
	default:
		ctx.Error = errors.New("(int)类型错误: 只能是数字类型")
	}
	return nil
}

func funcFloat(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	switch params[0].TypeId {
	case VMTypeInt:
		v, _ := params[0].ReadInt()
		return NewFloatVal(float64(v))
	case VMTypeFloat:
		return params[0]
	case VMTypeString:
		s, _ := params[0].ReadString()
		val, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return NewFloatVal(val)
		} else {
			ctx.Error = errors.New("(float)值错误: 无法进行 float() 转换: " + s)
		}
	default:
		ctx.Error = errors.New("(float)类型错误: 只能是数字类型")
	}
	return nil
}

func funcStr(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	return NewStrVal(params[0].ToString())
}

func funcRepr(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	return NewStrVal(params[0].ToRepr())
}

func funcTypeId(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	return NewIntVal(IntType(params[0].TypeId))
}

func funcLoad(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	v := params[0]
	if v.TypeId != VMTypeString {
		ctx.Error = errors.New("(load)类型错误: 参数类型必须为str")
		return nil
	}

	name := v.Value.(string)
	val := ctx.LoadName(name, false, true)
	if ctx.Error != nil {
		return nil
	}

	if ctx.Config.HookFuncValueLoadOverwrite != nil {
		val = ctx.Config.HookFuncValueLoadOverwrite(ctx, name, val, nil)
	}

	return val.Clone()
}

func funcDir(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	typeId := params[0].TypeId
	var arr []*VMValue
	if v, ok := builtinProto[typeId]; ok {
		v.Range(func(key string, value *VMValue) bool {
			arr = append(arr, NewStrVal(key))
			return true
		})
	}
	if typeId == VMTypeNativeObject {
		v := params[0]
		d, _ := v.ReadNativeObjectData()
		if d.DirFunc != nil {
			arr = append(arr, d.DirFunc(ctx)...)
		}
	}
	return NewArrayValRaw(arr)
}

//
// func funcHelp(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
//	// 函数名，参数，说明
//	return NewStrVal(params[0].ToString())
// }

var nnf = NewNativeFunctionVal

type ndf = NativeFunctionData

var builtinValues = map[string]*VMValue{
	"ceil":  nnf(&ndf{"ceil", []string{"value"}, nil, nil, funcCeil}),
	"floor": nnf(&ndf{"floor", []string{"value"}, nil, nil, funcFloor}),
	"round": nnf(&ndf{"round", []string{"value"}, nil, nil, funcRound}),
	"abs":   nnf(&ndf{"abs", []string{"value"}, nil, nil, funcAbs}),

	"int":   nnf(&ndf{"int", []string{"value"}, nil, nil, funcInt}),
	"float": nnf(&ndf{"float", []string{"value"}, nil, nil, funcFloat}),
	"str":   nnf(&ndf{"str", []string{"value"}, nil, nil, funcStr}),
	"bool":  nnf(&ndf{"bool", []string{"value"}, nil, nil, funcBool}),

	"repr": nnf(&ndf{"repr", []string{"value"}, nil, nil, funcRepr}),
	"load": nnf(&ndf{"load", []string{"value"}, nil, nil, nil}),

	// TODO: roll()

	// 要不要进行权限隔绝？
	"dir": nnf(&ndf{"dir", []string{"value"}, nil, nil, funcDir}),
	// "help": nnf(&ndf{"help", []string{"value"}, nil, nil, funcHelp}),
	"typeId": nnf(&ndf{"typeId", []string{"value"}, nil, nil, funcTypeId}),
}

func _init() bool {
	// 因循环引用问题无法在上面声明
	nfd, _ := builtinValues["load"].ReadNativeFunctionData()
	nfd.NativeFunc = funcLoad
	return false
}

var _ = _init()
