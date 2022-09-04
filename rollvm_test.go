package dicescript

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newVMWithStore(attrs map[string]*VMValue) (*Context, map[string]*VMValue) {
	vm := NewVM()
	if attrs == nil {
		attrs = map[string]*VMValue{}
	}

	vm.ValueStoreNameFunc = func(name string, v *VMValue) {
		attrs[name] = v
	}
	vm.ValueLoadNameFunc = func(name string) *VMValue {
		if val, ok := attrs[name]; ok {
			return val
		}
		return nil
	}
	return vm, attrs
}

func simpleExecute(t *testing.T, expr string, ret *VMValue) *Context {
	vm := NewVM()
	err := vm.Run(expr)
	if err != nil {
		fmt.Println(vm.GetAsmText())
		t.Errorf("VM Error: %s, %s", expr, err.Error())
		return vm
	}
	if !valueEqual(vm.Ret, ret) {
		fmt.Println(vm.GetAsmText())
		t.Errorf("not equal: %s %s", ret.ToString(), vm.Ret.ToString())
	}
	return vm
}

func TestSimpleRun(t *testing.T) {
	simpleExecute(t, "1+1", ni(2))
	simpleExecute(t, "2.0+1", nf(3))
	simpleExecute(t, ".5+1", nf(1.5))
}

func TestStr(t *testing.T) {
	simpleExecute(t, `""`, ns(""))
	simpleExecute(t, `''`, ns(""))
	simpleExecute(t, "``", ns(""))
	simpleExecute(t, "\x1e\x1e", ns(""))

	simpleExecute(t, "'123'", ns("123"))
	simpleExecute(t, "'12' + '3' ", ns("123"))
	simpleExecute(t, "`12{3}` ", ns("123"))
	simpleExecute(t, "`12{'3'}` ", ns("123"))
	simpleExecute(t, "`12{% 3 %}` ", ns("123"))
	simpleExecute(t, `"123"`, ns("123"))
	simpleExecute(t, "\x1e"+"12{% 3 %}"+"\x1e", ns("123"))

	simpleExecute(t, `"12\n3"`, ns("12\n3"))
	simpleExecute(t, `"12\r3"`, ns("12\r3"))
	simpleExecute(t, `"12\f3"`, ns("12\f3"))
	simpleExecute(t, `"12\t3"`, ns("12\t3"))
	simpleExecute(t, `"12\\3"`, ns("12\\3"))

	simpleExecute(t, `'12\n3'`, ns("12\n3"))
	simpleExecute(t, `'12\r3'`, ns("12\r3"))
	simpleExecute(t, `'12\f3'`, ns("12\f3"))
	simpleExecute(t, `'12\t3'`, ns("12\t3"))
	simpleExecute(t, `'12\\3'`, ns("12\\3"))

	simpleExecute(t, "\x1e"+`12\n3`+"\x1e", ns("12\n3"))
	simpleExecute(t, "\x1e"+`12\r3`+"\x1e", ns("12\r3"))
	simpleExecute(t, "\x1e"+`12\f3`+"\x1e", ns("12\f3"))
	simpleExecute(t, "\x1e"+`12\t3`+"\x1e", ns("12\t3"))
	simpleExecute(t, "\x1e"+`12\\3`+"\x1e", ns("12\\3"))

	simpleExecute(t, "`"+`12\n3`+"`", ns("12\n3"))
	simpleExecute(t, "`"+`12\r3`+"`", ns("12\r3"))
	simpleExecute(t, "`"+`12\f3`+"`", ns("12\f3"))
	simpleExecute(t, "`"+`12\t3`+"`", ns("12\t3"))
	simpleExecute(t, "`"+`12\\3`+"`", ns("12\\3"))

	// TODO: FIX
	//simpleExecute(t, `"12\"3"`, ns(`12"3`))
}

func TestEmptyInput(t *testing.T) {
	vm := NewVM()
	err := vm.Run("")
	if err == nil {
		t.Errorf("VM Error")
	}
}

func TestDice(t *testing.T) {
	// 语法可用性测试(并不做验算)
	simpleExecute(t, "4d1", ni(4))
	simpleExecute(t, "4D1", ni(4))

	simpleExecute(t, "4d1k", ni(1))
	simpleExecute(t, "4d1k1", ni(1))
	simpleExecute(t, "4d1kh", ni(1))
	simpleExecute(t, "4d1kh1", ni(1))

	simpleExecute(t, "4d1q", ni(1))
	simpleExecute(t, "4d1q1", ni(1))
	simpleExecute(t, "4d1kl(1)", ni(1))
	simpleExecute(t, "4d1kl1", ni(1))

	simpleExecute(t, "4d1dl", ni(3))
	simpleExecute(t, "4d1dl1", ni(3))

	simpleExecute(t, "4d1dl", ni(3))
	simpleExecute(t, "4d1dl1", ni(3))

	simpleExecute(t, "4d1dh", ni(3))
	simpleExecute(t, "4d1dh1", ni(3))

	// min max
	simpleExecute(t, "d20min20", ni(20))
	simpleExecute(t, "d20min30", ni(30)) // 与fvtt行为一致
	simpleExecute(t, "d20max1", ni(1))
	simpleExecute(t, "d20min30max1", ni(30)) // 同fvtt
	simpleExecute(t, "4d20k1min20", ni(20))

	// 优势
	simpleExecute(t, "d1优势", ni(1))
	simpleExecute(t, "d1劣势", ni(1))

	// 算力上限
	vm := NewVM()
	err := vm.Run("30001d20")
	if err == nil {
		t.Errorf("VM Error")
	}

	// 这种情况报个错如何？
	simpleExecute(t, "4d1k5", ni(4))
}

func TestUnsupportedOperandType(t *testing.T) {
	vm := NewVM()
	err := vm.Run("2 % 3.1")
	if err == nil {
		t.Errorf("VM Error: %s", err.Error())
	}
}

func TestValueStore1(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a=1")
	if err == nil {
		// 未设置 ValueStoreNameFunc，无法储存变量
		t.Errorf("VM Error: %s", err.Error())
	}

	vm = NewVM()
	err = vm.Run("bbb")
	if err == nil {
		t.Errorf("VM Error: %s", err.Error())
	}
}

func TestValueStore(t *testing.T) {
	attrs := map[string]*VMValue{}
	vm, _ := newVMWithStore(attrs)
	err := vm.Run("测试=1")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}

	vm, _ = newVMWithStore(attrs)
	err = vm.Run("测试   =   1")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}

	vm, _ = newVMWithStore(attrs)
	err = vm.Run("测试")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}

	vm, _ = newVMWithStore(attrs)
	err = vm.Run("CC")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}
	assert.True(t, valueEqual(vm.Ret, VMValueNewUndefined()))

	// 栈指针bug(两个变量实际都指向了栈的某一个位置，导致值相同)
	attrs = map[string]*VMValue{}
	vm, _ = newVMWithStore(attrs)
	err = vm.Run("b=1;d=2")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}
	assert.True(t, valueEqual(attrs["b"], ni(1)))
	assert.True(t, valueEqual(attrs["d"], ni(2)))
}

func TestIf(t *testing.T) {
	attrs := map[string]*VMValue{}
	vm, _ := newVMWithStore(attrs)
	err := vm.Run("if 0 { a = 2 } else if 2 { b = 1 } c= 1; ;;;;; d= 2;b")
	assert.NoError(t, err)
	assert.True(t, valueEqual(attrs["b"], ni(1)), attrs["b"])
	assert.True(t, valueEqual(attrs["c"], ni(1)), attrs["c"])
	assert.True(t, valueEqual(attrs["d"], ni(2)), attrs["d"])
	assert.True(t, attrs["a"] == nil)
}

func TestTernary(t *testing.T) {
	vm := NewVM()
	err := vm.Run("1 == 1 ? 2")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("1 == 1 ? 2 : 3")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("1 != 1 ? 2 : 3")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	vm, attrs := newVMWithStore(nil)
	attrs["a"] = VMValueNewInt64(1)
	err = vm.Run("a == 1 ? 'A', a == 2 ? 'B'")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("A")))
	}

	vm, _ = newVMWithStore(attrs)
	attrs["a"] = VMValueNewInt64(2)
	err = vm.Run("a == 1 ? 'A', a == 2 ? 'B', a == 3 ? 'C'")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("B")))
	}

	vm, _ = newVMWithStore(attrs)
	attrs["a"] = VMValueNewInt64(3)
	err = vm.Run("a == 1 ? 'A', a == 2 ? 'B', a == 3 ? 'C'")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("C")))
	}
}

func TestUnary(t *testing.T) {
	vm := NewVM()
	err := vm.Run("-1")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(-1)))
	}

	vm = NewVM()
	err = vm.Run("--1")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("-+1")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(-1)))
	}

	vm = NewVM()
	err = vm.Run("+-1")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(-1)))
	}

	vm = NewVM()
	err = vm.Run("-1.3")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, nf(-1.3)))
	}

	vm = NewVM()
	err = vm.Run("-'123'")
	assert.Error(t, err)
}

func TestRest(t *testing.T) {
	vm := NewVM()
	err := vm.Run("1 2")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
		assert.True(t, vm.RestInput == "2")
	}
}

func TestRecursion(t *testing.T) {
	vm, attrs := newVMWithStore(nil)
	err := vm.Run("&a = a + 1")
	assert.NoError(t, err)

	vm, _ = newVMWithStore(attrs)
	err = vm.Run("a")
	assert.Error(t, err) // 算力上限
}

func TestArray(t *testing.T) {
	vm := NewVM()
	err := vm.Run("[1,2,3]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, VMValueNewArray(ni(1), ni(2), ni(3))))
	}

	vm = NewVM()
	err = vm.Run("[1,3,2]kh")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	vm = NewVM()
	err = vm.Run("[1.2,2,3]kh")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, nf(3)))
	}

	vm = NewVM()
	err = vm.Run("[1,2.2,3]kh")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, nf(3)))
	}

	vm = NewVM()
	err = vm.Run("[1,3,2]kl")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("[2,3,1]kl")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("[1,3.1,2.1]kl")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, nf(1)))
	}

	vm = NewVM()
	err = vm.Run("[4.1,3.1,1]kl")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, nf(1)))
	}

	vm = NewVM()
	err = vm.Run("[1,2,3][1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("[1,2,3][-1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	vm = NewVM()
	err = vm.Run("[1,2,3][-4]")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("[1,2,3][4]")
	assert.Error(t, err)

	vm, _ = newVMWithStore(nil)
	err = vm.Run("a = [1,2,3]; a[1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm, _ = newVMWithStore(nil)
	err = vm.Run("b[1]")
	assert.Error(t, err)

	vm, _ = newVMWithStore(nil)
	err = vm.Run("b[0][0]")
	assert.Error(t, err)

	vm, _ = newVMWithStore(nil)
	err = vm.Run("[[1]][0][0]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm, _ = newVMWithStore(nil)
	err = vm.Run("([[2]])[0][0]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}
}

func TestComputed(t *testing.T) {
	vm, _ := newVMWithStore(nil)
	err := vm.Run("&a = d1+2; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	vm, _ = newVMWithStore(nil)
	err = vm.Run("&a = []+2; a")
	assert.Error(t, err)

	vm, _ = newVMWithStore(nil)
	err = vm.Run("&a = undefined; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, VMValueNewUndefined()))
	}

	vm, attrs := newVMWithStore(nil)
	err = vm.Run("&a = d1 + this.x")
	assert.NoError(t, err)

	vm, _ = newVMWithStore(attrs)
	err = vm.Run("a.x = 2")
	assert.NoError(t, err)

	vm, _ = newVMWithStore(attrs)
	err = vm.Run("a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	vm, _ = newVMWithStore(attrs)
	err = vm.Run("&a.x")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
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

		{TypeDiceInit, nil},
		{TypeDiceSetTimes, nil},
		{TypeDiceSetKeepLowNum, nil},
		{TypeDiceSetKeepHighNum, nil},
		{TypeDiceSetDropLowNum, nil},
		{TypeDiceSetDropHighNum, nil},
		{TypeDiceSetMin, nil},
		{TypeDiceSetMax, nil},

		{TypeJmp, int64(0)},
		{TypeJe, int64(0)},
		{TypeJne, int64(0)},
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

func TestGetASM(t *testing.T) {
	vm := NewVM()
	vm.Run("1+1")
	vm.GetAsmText()
}
