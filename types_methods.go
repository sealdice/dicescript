package dicescript

import (
	"errors"
	"math/rand"
)

func funcComputedCompute(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	return this.ComputedExecute(ctx, nil)
}

func funcArrayKeepLow(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	isAllInt, ret := this.ArrayFuncKeepLow(ctx, params[0].MustReadInt())
	if isAllInt {
		return NewIntVal(IntType(ret))
	} else {
		return NewFloatVal(ret)
	}
}

func funcArrayKeepHigh(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	isAllInt, ret := this.ArrayFuncKeepHigh(ctx, params[0].MustReadInt())
	if isAllInt {
		return NewIntVal(IntType(ret))
	} else {
		return NewFloatVal(ret)
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
		return NewIntVal(IntType(sumNum))
	} else {
		return NewFloatVal(sumNum)
	}
}

func funcArrayLen(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	arr, _ := this.ReadArray()
	return NewIntVal(IntType(len(arr.List)))
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
	newArr := NewArrayVal(arr.List...)
	funcArrayShuttle(ctx, newArr, []*VMValue{})
	arr, _ = newArr.ReadArray()

	if val, ok := params[0].ReadInt(); ok {
		arr.List = arr.List[:val]
		return newArr
	} else {
		ctx.Error = errors.New("(arr.randSize)类型不符")
		return nil
	}
}

func funcArrayPop(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	arr, _ := this.ReadArray()
	if len(arr.List) >= 1 {
		val := arr.List[len(arr.List)-1]
		arr.List = arr.List[:len(arr.List)-1]
		return val
	}
	return NewNullVal()
}

func funcArrayShift(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	arr, _ := this.ReadArray()
	if len(arr.List) >= 1 {
		val := arr.List[0]
		arr.List = arr.List[1:]
		return val
	}
	return NewNullVal()
}

func funcArrayPush(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	arr, _ := this.ReadArray()
	arr.List = append(arr.List, params[0])
	return this
}

func funcDictKeys(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	d := this.MustReadDictData()
	var arr []*VMValue
	d.Dict.Range(func(key string, value *VMValue) bool {
		arr = append(arr, NewStrVal(key))
		return true
	})
	return NewArrayValRaw(arr)
}

func funcDictValues(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	d := this.MustReadDictData()
	var arr []*VMValue
	d.Dict.Range(func(key string, value *VMValue) bool {
		arr = append(arr, value)
		return true
	})
	return NewArrayValRaw(arr)
}

func funcDictItems(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	d := this.MustReadDictData()
	var arr []*VMValue
	d.Dict.Range(func(key string, value *VMValue) bool {
		arr = append(arr, NewArrayVal(NewStrVal(key), value))
		return true
	})
	return NewArrayValRaw(arr)
}

func funcDictLen(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	d := this.MustReadDictData()
	return NewIntVal(IntType(d.Dict.Length()))
}

var builtinProto = map[VMValueType]*VMDictValue{
	VMTypeComputedValue: NewDictValWithArrayMust(
		NewStrVal("compute"), nnf(&ndf{"Computed.compute", []string{}, nil, nil, nil}),
	),
	VMTypeArray: NewDictValWithArrayMust(
		NewStrVal("kh"), nnf(&ndf{"Array.kh", []string{"num"}, []*VMValue{NewIntVal(1)}, nil, funcArrayKeepHigh}),
		NewStrVal("kl"), nnf(&ndf{"Array.kl", []string{"num"}, []*VMValue{NewIntVal(1)}, nil, funcArrayKeepLow}),
		NewStrVal("sum"), nnf(&ndf{"Array.sum", []string{}, nil, nil, funcArraySum}),
		NewStrVal("len"), nnf(&ndf{"Array.len", []string{}, nil, nil, funcArrayLen}),
		NewStrVal("shuffle"), nnf(&ndf{"Array.shuffle", []string{}, nil, nil, funcArrayShuttle}),
		NewStrVal("rand"), nnf(&ndf{"Array.rand", []string{}, nil, nil, funcArrayRand}),
		NewStrVal("randSize"), nnf(&ndf{"Array.randSize", []string{"num"}, nil, nil, funcArrayRandSize}),
		NewStrVal("pop"), nnf(&ndf{"Array.pop", []string{}, nil, nil, funcArrayPop}),
		NewStrVal("shift"), nnf(&ndf{"Array.shift", []string{}, nil, nil, funcArrayShift}),
		NewStrVal("push"), nnf(&ndf{"Array.push", []string{"value"}, nil, nil, funcArrayPush}),
	),
	VMTypeDict: NewDictValWithArrayMust(
		NewStrVal("keys"), nnf(&ndf{"Dict.keys", []string{}, nil, nil, funcDictKeys}),
		NewStrVal("values"), nnf(&ndf{"Dict.values", []string{}, nil, nil, funcDictValues}),
		NewStrVal("items"), nnf(&ndf{"Dict.items", []string{}, nil, nil, funcDictItems}),
		NewStrVal("len"), nnf(&ndf{"Dict.len", []string{}, nil, nil, funcDictLen}),
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
		return NewFunctionValRaw(fd2)
	case VMTypeNativeFunction:
		fd, _ := funcDef.ReadNativeFunctionData()

		// 完成clone
		_fd := *fd
		fd2 := &_fd

		fd2.Self = v.Clone()
		return NewNativeFunctionVal(fd2)
	}
	return nil
}

func _init2() bool {
	// 因循环引用问题无法在上面声明
	funcCompute := nnf(&ndf{"Computed.compute", []string{}, nil, nil, funcComputedCompute})
	builtinProto[VMTypeComputedValue].Store("compute", funcCompute)
	return false
}

var _ = _init2()
