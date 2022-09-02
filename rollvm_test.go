package dicescript

import "testing"

func simpleExecute(t *testing.T, expr string, ret *VMValue) *Context {
	vm := NewVM()
	err := vm.Run(expr)
	if err != nil {
		t.Errorf("VM Error")
	}
	if !valueEqual(vm.Ret, ret) {
		t.Errorf("VM Error")
	}
	return vm
}

func TestSimpleRun(t *testing.T) {
	simpleExecute(t, "1+1", ni(2))
	simpleExecute(t, "2.0+1", nf(3))
	simpleExecute(t, ".5+1", nf(1.5))
}

func TestDice(t *testing.T) {
	// 语法可用性测试(并不做验算)
	simpleExecute(t, "4d1", ni(4))

	simpleExecute(t, "4d1k", ni(1))
	simpleExecute(t, "4d1k1", ni(1))
	simpleExecute(t, "4d1kh", ni(1))
	simpleExecute(t, "4d1kh1", ni(1))

	simpleExecute(t, "4d1q", ni(1))
	simpleExecute(t, "4d1q1", ni(1))
	simpleExecute(t, "4d1kl1", ni(1))

	simpleExecute(t, "4d1dl", ni(3))
	simpleExecute(t, "4d1dl1", ni(3))

	simpleExecute(t, "4d1dl", ni(3))
	simpleExecute(t, "4d1dl1", ni(3))

	simpleExecute(t, "4d1dh", ni(3))
	simpleExecute(t, "4d1dh1", ni(3))

	// 这种情况报个错如何？
	simpleExecute(t, "4d1k5", ni(4))
}

func TestUnsupportedOperandType(t *testing.T) {
	vm := NewVM()
	err := vm.Run("2 % 3.1")
	if err == nil {
		t.Errorf("VM Error")
	}
}

func TestBytecodeToString(t *testing.T) {
	ops := []ByteCode{
		{TypePushIntNumber, int64(1)},
		{TypePushFloatNumber, float64(1.2)},
		{TypePushString, "abc"},

		{TypeAdd, nil},
		{TypeSubtract, nil},
		{TypeMultiply, nil},
		{TypeDivide, nil},
		{TypeModulus, nil},
		{TypeExponentiation, nil},

		{TypeCompLT, nil},
		{TypeCompLE, nil},
		{TypeCompEQ, nil},
		{TypeCompNE, nil},
		{TypeCompGE, nil},
		{TypeCompGT, nil},

		{TypeBitwiseAnd, nil},
		{TypeBitwiseOr, nil},
		{TypeNop, nil},
	}

	for _, i := range ops {
		if i.CodeString() == "" {
			t.Errorf("Not work: %d", i.T)
		}
	}
}

func TestWriteCodeOverflow(t *testing.T) {
	vm := NewVM()
	for i := 0; i < 8193; i++ {
		vm.parser.WriteCode(TypeNop, nil)
	}
	if !vm.parser.checkStackOverflow() {
		t.Errorf("Failed")
	}
}
