package dicescript

import (
	"errors"
	"math/rand"
)

func funcArrayKeepLow(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	isAllInt, ret := this.ArrayFuncKeepLow(ctx, params[0].MustReadInt())
	if isAllInt {
		return VMValueNewInt(IntType(ret))
	} else {
		return VMValueNewFloat(ret)
	}
}

func funcArrayKeepHigh(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	isAllInt, ret := this.ArrayFuncKeepHigh(ctx, params[0].MustReadInt())
	if isAllInt {
		return VMValueNewInt(IntType(ret))
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
		return VMValueNewInt(IntType(sumNum))
	} else {
		return VMValueNewFloat(sumNum)
	}
}

func funcArrayLen(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	arr, _ := this.ReadArray()
	return VMValueNewInt(IntType(len(arr.List)))
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
	if len(arr.List) >= 1 {
		val := arr.List[len(arr.List)-1]
		arr.List = arr.List[:len(arr.List)-1]
		return val
	}
	return VMValueNewUndefined()
}

func funcArrayShift(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	arr, _ := this.ReadArray()
	if len(arr.List) >= 1 {
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

func funcDictKeys(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	d := this.MustReadDictData()
	var arr []*VMValue
	d.Dict.Range(func(key string, value *VMValue) bool {
		arr = append(arr, VMValueNewStr(key))
		return true
	})
	return VMValueNewArrayRaw(arr)
}

func funcDictValues(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	d := this.MustReadDictData()
	var arr []*VMValue
	d.Dict.Range(func(key string, value *VMValue) bool {
		arr = append(arr, value)
		return true
	})
	return VMValueNewArrayRaw(arr)
}

func funcDictItems(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	d := this.MustReadDictData()
	var arr []*VMValue
	d.Dict.Range(func(key string, value *VMValue) bool {
		arr = append(arr, VMValueNewArray(VMValueNewStr(key), value))
		return true
	})
	return VMValueNewArrayRaw(arr)
}

func funcDictLen(ctx *Context, this *VMValue, params []*VMValue) *VMValue {
	d := this.MustReadDictData()
	var size IntType
	d.Dict.Range(func(key string, value *VMValue) bool {
		size++
		return true
	})
	return VMValueNewInt(size)
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
		VMValueNewStr("push"), nnf(&ndf{"Array.push", []string{"value"}, nil, nil, funcArrayPush}),
	),
	VMTypeDict: VMValueMustNewDictWithArray(
		VMValueNewStr("keys"), nnf(&ndf{"Dict.keys", []string{}, nil, nil, funcDictKeys}),
		VMValueNewStr("values"), nnf(&ndf{"Dict.values", []string{}, nil, nil, funcDictValues}),
		VMValueNewStr("items"), nnf(&ndf{"Dict.items", []string{}, nil, nil, funcDictItems}),
		VMValueNewStr("len"), nnf(&ndf{"Dict.len", []string{}, nil, nil, funcDictLen}),
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
