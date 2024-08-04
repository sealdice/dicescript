package dicescript

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func vmValueEqual(vm *Context, aKey string, bValue *VMValue) bool {
	return valueEqual(vm.Attrs.MustLoad(aKey), bValue)
}

func simpleExecute(t *testing.T, expr string, ret *VMValue) *Context {
	vm := NewVM()
	err := vm.Run(expr)
	if err != nil {
		t.Errorf("VM Error: %s, %s", expr, err.Error())
		return vm
	}
	if !valueEqual(vm.Ret, ret) {
		t.Errorf("not equal: %s %s", ret.ToString(), vm.Ret.ToString())
	}
	return vm
}

func TestValueDefineBool(t *testing.T) {
	simpleExecute(t, "true", ni(1))
	simpleExecute(t, "false", ni(0))
}

func TestValueDefineNumber(t *testing.T) {
	simpleExecute(t, "123", ni(123))
	simpleExecute(t, "1.2", nf(1.2))
}

func TestValueIdentifier(t *testing.T) {
	simpleExecute(t, "val", NewNullVal())
}

func TestSimpleRun(t *testing.T) {
	simpleExecute(t, "1+1", ni(2))
	simpleExecute(t, "2.0+1", nf(3))
	simpleExecute(t, ".5+1", nf(1.5))
}

func TestValueDefineStr(t *testing.T) {
	simpleExecute(t, `""`, ns(""))
	simpleExecute(t, `''`, ns(""))
	simpleExecute(t, "``", ns(""))
	simpleExecute(t, "\x1e\x1e", ns(""))

	simpleExecute(t, "'123'", ns("123"))
	simpleExecute(t, "'12' + '3' ", ns("123"))
	simpleExecute(t, "`12{3}` ", ns("123"))
	simpleExecute(t, "`12{3 }` ", ns("123"))
	simpleExecute(t, "`12{ 3}` ", ns("123"))
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

	simpleExecute(t, `"12\"3"`, ns(`12"3`))
	simpleExecute(t, `"\""`, ns(`"`))
	simpleExecute(t, `"\r"`, ns("\r"))
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
	vm.Config.OpCountLimit = 30000
	err := vm.Run("30001d20")
	if err == nil {
		t.Errorf("VM Error")
	}

	// 这种情况报个错如何？
	simpleExecute(t, "4d1k5", ni(4))
}

func TestVMMultiply(t *testing.T) {
	vm := NewVM()
	err := vm.Run("2*3")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(6)))
	}

	err = vm.Run("2*  3")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(6)))
	}

	err = vm.Run("2  *3")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(6)))
	}
}

func TestVMDivideModulus(t *testing.T) {
	vm := NewVM()
	err := vm.Run("3/2")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	err = vm.Run("3/  2.0")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, nf(1.5)))
	}

	err = vm.Run("3.0  /2")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, nf(1.5)))
	}
}

func TestDiceNoSpaceForModifier(t *testing.T) {
	vm := NewVM()
	err := vm.Run("3d1 k2")
	if assert.NoError(t, err) {
		// 注: 如果读取为3d1k2，值为2为错，读取3d1剩余文本k2为对
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}
}

func TestUnsupportedOperandType(t *testing.T) {
	vm := NewVM()
	err := vm.Run("2 % 3.1")
	if assert.Error(t, err) {
		// VM Error: 这两种类型无法使用 mod 算符连接: int64, float64
		assert.Equal(t, err.Error(), "这两种类型无法使用 mod 算符连接: int, float")
	}
}

func TestValueStore1(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a=1")
	if err != nil {
		// 未设置 ValueStoreNameFunc，无法储存变量
		t.Errorf("VM Error: %s", err.Error())
	}

	err = vm.Run("a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("bbb")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}
}

func TestValueStore(t *testing.T) {
	vm := NewVM()
	err := vm.Run("测试=1")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}

	vm = NewVM()
	err = vm.Run("测试   =   1")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}

	vm = NewVM()
	err = vm.Run("测试")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}

	vm = NewVM()
	err = vm.Run("CC")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}
	assert.True(t, valueEqual(vm.Ret, NewNullVal()))

	// 栈指针bug(两个变量实际都指向了栈的某一个位置，导致值相同)
	vm = NewVM()
	err = vm.Run("b=1;d=2")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}
	assert.True(t, vmValueEqual(vm, "b", ni(1)))
	assert.True(t, vmValueEqual(vm, "d", ni(2)))
}

func TestIfBasic(t *testing.T) {
	vm := NewVM()
	err := vm.Run("if 1 { a = 2 } ")
	if assert.NoError(t, err) {
		assert.True(t, vmValueEqual(vm, "a", ni(2)))
	}
}

func TestIf(t *testing.T) {
	vm := NewVM()
	err := vm.Run("if 0 { a = 2 } else if 2 { b = 1 } c= 1; ;;;;; d= 2;b")
	assert.NoError(t, err)
	assert.True(t, vmValueEqual(vm, "b", ni(1)))
	assert.True(t, vmValueEqual(vm, "c", ni(1)))
	assert.True(t, vmValueEqual(vm, "d", ni(2)))

	_, exists := vm.Attrs.Load("a")
	assert.True(t, !exists)
}

//

func TestStatementLines(t *testing.T) {
	vm := NewVM()
	err := vm.Run("i = 0 ;; i = 3")
	assert.NoError(t, err)
	assert.True(t, vmValueEqual(vm, "i", ni(3)))

	vm = NewVM()
	err = vm.Run("i = 0 ;    ;  ; i = 3")
	assert.NoError(t, err)
	assert.True(t, vmValueEqual(vm, "i", ni(3)))

	vm = NewVM()
	err = vm.Run("i = 0 if 1 { i = 3 }")
	assert.NoError(t, err)
	assert.Equal(t, " if 1 { i = 3 }", vm.RestInput)

	vm = NewVM()
	err = vm.Run("i = 0; if 1 { i = 3 }")
	assert.NoError(t, err)
	assert.True(t, vmValueEqual(vm, "i", ni(3)))

	vm = NewVM()
	err = vm.Run("i = 0   ;if 1 { i = 3 }")
	assert.NoError(t, err)
	assert.True(t, vmValueEqual(vm, "i", ni(3)))

	vm = NewVM()
	err = vm.Run("i = 0;  if 1 { i = 3 }")
	assert.NoError(t, err)
	assert.True(t, vmValueEqual(vm, "i", ni(3)))
}

func TestKeywords(t *testing.T) {
	vm := NewVM()
	err := vm.Run("while123")
	assert.NoError(t, err)
	assert.True(t, vm.RestInput == "")

	keywords := []string{
		"while", "if", "else", "continue", "break", "func",
	}

	suffixBad := []string{
		"", "=", "#", ";", "=1", " ", " =1", "!", "\"", "%", "^", "&", "*", "(", ")", "/", "+", "-", ".", ".aa",
		"[", "]", "[1]", ":", "<", ">", "?",
	}

	for _, i := range keywords {
		for _, j := range suffixBad {
			vm := NewVM()
			err = vm.Run(i + j)
			assert.Errorf(t, err, i+j)
		}
	}

	vm = NewVM()
	err = vm.Run("return 1")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}
}

func TestWhile(t *testing.T) {
	vm := NewVM()
	err := vm.Run("i = 0; while i<5 { i=i+1 }")
	assert.NoError(t, err)
	assert.True(t, vm.NumOpCount < 100)

	vm = NewVM()
	vm.Config.OpCountLimit = 30000
	err = vm.Run("i = 0; while 1 {  }")
	assert.Error(t, err) // 算力上限

	vm = NewVM()
	vm.Config.OpCountLimit = 30000
	err = vm.Run("i = 0; while 1 {}")
	assert.Error(t, err) // 算力上限

	vm = NewVM()
	vm.Config.OpCountLimit = 30000
	err = vm.Run("i = 0; while1 {}") // nolint
	assert.True(t, vm.RestInput == " {}", vm.RestInput)
}

func TestWhileContinueBreak(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a = 0; while a < 5 { a = a+1; continue; a=a+10 }; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(5)))
	}

	vm = NewVM()
	err = vm.Run("a = 0; while a < 5 { a = a+1; a=a+10; continue }; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(11)))
	}

	vm = NewVM()
	err = vm.Run("a = 1; while a < 5 { break; a = a+1; a=a+10 }; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("a = 1; while a < 5 { a = a+1; break; a=a+10 }; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}
}

func TestLineBreak(t *testing.T) {
	vm := NewVM()
	err := vm.Run("if 1 {} 2")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("1; if 1 {} 2")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("1; if 1 {}; 2")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}
}

func TestItemSetBug(t *testing.T) {
	// 由于言诺在2022/9/9提交，此用例之前的输出内容为[3,3,3]
	vm := NewVM()
	err := vm.Run("a = [0,0,0]; i=0; while i<3 { a[i] = i+1; i=i+1 }  a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewArrayVal(ni(1), ni(2), ni(3))))
	}
}

func TestCompareExpr(t *testing.T) {
	tests := []struct {
		expr  string
		value *VMValue
	}{
		{"1>0", ni(1)},
		{"1>=0", ni(1)},
		{"1==0", ni(0)},
		{"1==1", ni(1)},
		{"1<0", ni(0)},
		{"1<=0", ni(0)},
		{"1!=0", ni(1)},

		// 带空格
		{"1 > 0", ni(1)},

		// 中断
		{"5＝+2", ni(5)},
	}

	for _, i := range tests {
		vm := NewVM()
		err := vm.Run(i.expr)
		assert.NoError(t, err, i.expr)
		assert.True(t, valueEqual(vm.Ret, i.value), i.expr)
	}
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

	vm = NewVM()
	vm.Attrs.Store("a", ni(1))
	err = vm.Run("a == 1 ? 'A', a == 2 ? 'B'")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("A")))
	}

	vm = NewVM()
	vm.Attrs.Store("a", ni(2))
	err = vm.Run("a == 1 ? 'A', a == 2 ? 'B', a == 3 ? 'C'")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("B")))
	}

	vm = NewVM()
	vm.Attrs.Store("a", ni(3))
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
	err = vm.Run("- 1")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(-1)))
	}

	vm = NewVM()
	err = vm.Run("--1")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("++1")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("-+1")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("+-1")
	assert.Error(t, err)

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
		assert.True(t, vm.RestInput == " 2")
	}
}

func TestRecursion(t *testing.T) {
	vm := NewVM()
	vm.Config.OpCountLimit = 30000
	err := vm.Run("&a = a + 1")
	assert.NoError(t, err)

	err = vm.Run("a")
	assert.Error(t, err) // 算力上限
}

func TestArray(t *testing.T) {
	vm := NewVM()
	err := vm.Run("[1,2,3]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewArrayVal(ni(1), ni(2), ni(3))))
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

	vm = NewVM()
	err = vm.Run("a = [1,2,3]; a[1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("b[1]")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("b[0][0]")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("[[1]][0][0]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("([[2]])[0][0]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("a = [0,0,0]; a[0] = 1; a[0]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("a[0] = 1")
	assert.Error(t, err)
}

func TestArrayMethod(t *testing.T) {
	vm := NewVM()
	err := vm.Run("[1,2,3].sum()")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(6)))
	}

	vm = NewVM()
	err = vm.Run("[1,2,3].len()")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	vm = NewVM()
	err = vm.Run("[1,2,3].pop()")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	vm = NewVM()
	err = vm.Run("[1,2,3].shift()")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("a = [1,2,3]; a.push(4); a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, na(ni(1), ni(2), ni(3), ni(4))))
	}
}

func TestReturn(t *testing.T) {
	vm := NewVM()
	err := vm.Run("func test(n) { return 1; 2 }; test(11)")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}
}

func TestComputed(t *testing.T) {
	vm := NewVM()
	err := vm.Run("&a = d1+2; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	vm = NewVM()
	err = vm.Run("&a = []+2; a")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("&a = undefined; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewNullVal()))
	}
}

func TestComputed2(t *testing.T) {
	vm := NewVM()
	err := vm.Run("&a = d1 + this.x")
	assert.NoError(t, err)

	err = vm.Run("&a.x = 2")
	assert.NoError(t, err)

	err = vm.Run("a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	// vm = NewVM()
	err = vm.Run("&a.x")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}
}

func TestFunction(t *testing.T) {
	vm := NewVM()
	err := vm.Run("func a() { 123 }; a()")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(123)))
	}

	vm = NewVM()
	err = vm.Run("func a(d,b,c) { return this.b }; a(1,2,3)")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("func a(d,b,c) { this.b }; a(1,2)")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("func a(d,b,c) { this.b }; a(1,2,3,4,5)")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("func a() { 2 / 0 }; a()")
	assert.Error(t, err)
}

func TestFunctionRecursion(t *testing.T) {
	vm := NewVM()
	err := vm.Run(`
func foo(n) {
	if (n < 2) {
		return foo(n + 1)
	}
	return 123
}
foo(1)
`)
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(123)))
	}
}

func TestFunctionFib(t *testing.T) {
	vm := NewVM()
	err := vm.Run(`func fib(n) {
  this.n == 0 ? 0,
  this.n == 1 ? 1,
  this.n == 2 ? 1,
   1 ? fib(this.n-1)+fib(this.n-2)
}
fib(11)
`)
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(89)))
	}
}

func TestBytecodeToString(t *testing.T) {
	ops := []ByteCode{
		{typePushIntNumber, IntType(1)},
		{typePushFloatNumber, float64(1.2)},
		{typePushString, "abc"},

		{typeAdd, nil},
		{typeSubtract, nil},
		{typeMultiply, nil},
		{typeDivide, nil},
		{typeModulus, nil},
		{typeExponentiation, nil},
		{typeNullCoalescing, nil},

		{typeCompLT, nil},
		{typeCompLE, nil},
		{typeCompEQ, nil},
		{typeCompNE, nil},
		{typeCompGE, nil},
		{typeCompGT, nil},

		{typeLogicAnd, nil},
		{typeLogicOr, nil},

		{typeNop, nil},

		{typeBitwiseAnd, nil},
		{typeBitwiseOr, nil},

		{typeDiceInit, nil},
		{typeDiceSetTimes, nil},
		{typeDiceSetKeepLowNum, nil},
		{typeDiceSetKeepHighNum, nil},
		{typeDiceSetDropLowNum, nil},
		{typeDiceSetDropHighNum, nil},
		{typeDiceSetMin, nil},
		{typeDiceSetMax, nil},

		{typeJmp, IntType(0)},
		{typeJe, IntType(0)},
		{typeJne, IntType(0)},
	}

	for _, i := range ops {
		if i.CodeString() == "" {
			t.Errorf("Not work: %d", i.T)
		}
	}
}

func TestWriteCodeOverflow(t *testing.T) {
	vm := NewVM()
	_ = vm.Run("")
	for i := 0; i < 8193; i++ {
		vm.parser.cur.data.WriteCode(typeNop, nil)
	}
	if !vm.parser.cur.data.checkStackOverflow() {
		t.Errorf("Failed")
	}
}

func TestGetASM(t *testing.T) {
	vm := NewVM()
	_ = vm.Run("1+1")
	vm.GetAsmText()
}

func TestSliceGet(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a = [1,2,3,4]")
	assert.NoError(t, err)

	err = vm.Run("a[:]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewArrayVal(ni(1), ni(2), ni(3), ni(4))))
	}

	err = vm.Run("a[:2]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewArrayVal(ni(1), ni(2))))
	}

	err = vm.Run("a[0:-1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewArrayVal(ni(1), ni(2), ni(3))))
	}

	err = vm.Run("a[-3:-1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewArrayVal(ni(2), ni(3))))
	}

	err = vm.Run("a[-3:-1:1]")
	assert.Error(t, err)
	// 尚不支持分片步长
	// if assert.NoError(t, err) {
	//	assert.True(t, valueEqual(vm.Ret, NewArrayVal(ni(2), ni(3))))
	// }

	err = vm.Run("a[-3:-1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewArrayVal(ni(2), ni(3))))
	}

	err = vm.Run("b = a[-3:-1]; b[0] = 9; a[-3:-1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewArrayVal(ni(2), ni(3))))
	}

	vm = NewVM()
	err = vm.Run("'12345'[2:3]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("3")))
	}
}

func TestSliceSet(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a = [1,2,3,4]")
	assert.NoError(t, err)

	err = vm.Run("a[:] = [1,2,3]; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewArrayVal(ni(1), ni(2), ni(3))))
	}

	err = vm.Run("a = [1,2,3]; a[:1] = [4,5];a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewArrayVal(ni(4), ni(5), ni(2), ni(3))))
	}

	err = vm.Run("a = [1,2,3]; a[2:] = [4,5];a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewArrayVal(ni(1), ni(2), ni(4), ni(5))))
	}
}

func TestRange(t *testing.T) {
	vm := NewVM()
	err := vm.Run("[1..4]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewArrayVal(ni(1), ni(2), ni(3), ni(4))))
	}

	vm = NewVM()
	err = vm.Run("[4..1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewArrayVal(ni(4), ni(3), ni(2), ni(1))))
	}
}

func TestDictExpr(t *testing.T) {
	vm := NewVM()
	err := vm.Run("{'a': 1}") // nolint
	assert.NoError(t, err)

	vm = NewVM()
	err = vm.Run("a = {'a': 1}") // nolint
	assert.NoError(t, err)

	err = vm.Run("a.a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	err = vm.Run("a['a']")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}
}

func TestDictExpr2(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a = {'a': 1,}") // nolint
	assert.NoError(t, err)

	err = vm.Run("a.a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("c = 'c'; a = {c:1,'b':3}") // nolint
	// if assert.NoError(t, err) {
	// }
	err = vm.Run("a.c")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}
}

func TestIdExpr(t *testing.T) {
	vm := NewVM()
	vm.Attrs.Store("a:b", ni(3))
	err := vm.Run("a:b") // 如果读到a 余下a:b即为错误
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}
}

func TestStringExpr(t *testing.T) {
	vm := NewVM()
	err := vm.Run("\x1e xxx \x1e")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns(" xxx ")))
	}
}

func TestContinuousDiceExpr(t *testing.T) {
	vm := NewVM()
	err := vm.Run("10d1d1")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(10)))
	}
}

func TestCrash1(t *testing.T) {
	// 一种崩溃，崩溃条件是第二次调用vm.Run且第二次的tokens少于第一次的
	// 注：重构后已经不会崩溃
	vm := NewVM()
	err := vm.Run("aa + 2//asd")
	if assert.Error(t, err) {
		err := vm.Run("/")
		if assert.Error(t, err) {
			// assert.True(t, strings.Contains(err.Error(), "parse error near"))
			assert.True(t, strings.Contains(err.Error(), "no match found"))
		}
	}
}

func TestDiceCocExpr(t *testing.T) {
	vm := NewVM()
	vm.Config.EnableDiceCoC = true
	err := vm.Run("b1 + p1")
	if assert.NoError(t, err) {
		assert.Equal(t, "", vm.RestInput)
		assert.True(t, vm.Ret.MustReadInt() > 1)
	}

	err = vm.Run("b")
	if assert.NoError(t, err) {
		assert.Equal(t, "", vm.RestInput)
		assert.True(t, vm.Ret.MustReadInt() >= 1)
	}

	err = vm.Run("b技能") // rab技能，这种不予修改，由指令那边做支持
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewNullVal()))
	}
}

func TestDiceWodExpr(t *testing.T) {
	vm := NewVM()
	vm.Config.EnableDiceWoD = true
	err := vm.Run("8a11m10k1")
	if assert.NoError(t, err) {
		assert.Equal(t, "", vm.RestInput)
		assert.True(t, valueEqual(vm.Ret, ni(8)))
	}

	vm = NewVM()
	vm.Config.EnableDiceWoD = true
	err = vm.Run("20001a11m10k1")
	assert.Error(t, err)

	vm = NewVM()
	vm.Config.EnableDiceWoD = true
	err = vm.Run("8a1m10k1")
	assert.Error(t, err)

	vm = NewVM()
	vm.Config.EnableDiceWoD = true
	err = vm.Run("8a11m0k1")
	assert.Error(t, err)

	vm = NewVM()
	vm.Config.EnableDiceWoD = true
	err = vm.Run("8a11m10k0")
	assert.Error(t, err)
}

func TestDiceDoubleCrossExpr(t *testing.T) {
	// 没有很好的测试用例
	vm := NewVM()
	vm.Config.EnableDiceDoubleCross = true
	err := vm.Run("10c11m10")
	if assert.NoError(t, err) {
		assert.Equal(t, "", vm.RestInput)
		assert.True(t, vm.Ret.MustReadInt() <= 10)
	}

	vm = NewVM()
	vm.Config.EnableDiceDoubleCross = true
	err = vm.Run("20001c11m10")
	assert.Error(t, err)

	vm = NewVM()
	vm.Config.EnableDiceDoubleCross = true
	err = vm.Run("10c1m10")
	assert.Error(t, err)

	vm = NewVM()
	vm.Config.EnableDiceDoubleCross = true
	err = vm.Run("10c11m0")
	assert.Error(t, err)
}

func TestDiceFlagWodMacroExpr(t *testing.T) {
	vm := NewVM()
	err := vm.Run("// #EnableDice wod true\n")
	if assert.NoError(t, err) {
		vm.Config = vm.parser.cur.data.Config
		err := vm.Run("10a11")
		if assert.NoError(t, err) {
			assert.Equal(t, "", vm.RestInput)

			err := vm.Run("// #EnableDice wod false")
			vm.Config = vm.parser.cur.data.Config
			if assert.NoError(t, err) {
				err := vm.Run("10a11")
				if assert.NoError(t, err) {
					assert.Equal(t, "a11", vm.RestInput)
				}
			}
		}
	}
}

func TestDiceFlagCoCMacroExpr(t *testing.T) {
	vm := NewVM()
	vm.Config.EnableDiceCoC = false
	err := vm.Run("// #EnableDice coc true\nb2")
	if assert.NoError(t, err) {
		assert.Equal(t, "", vm.RestInput)
		assert.Equal(t, VMTypeInt, vm.Ret.TypeId)

		vm.Config.EnableDiceCoC = true
		err := vm.Run("// #EnableDice coc false\nb2")
		if assert.NoError(t, err) {
			assert.Equal(t, VMTypeNull, vm.Ret.TypeId)
		}
	}
}

func TestDiceFlagFateMacroExpr(t *testing.T) {
	vm := NewVM()
	err := vm.Run("// #EnableDice fate true")
	if assert.NoError(t, err) {
		vm.Config = vm.parser.cur.data.Config
		err := vm.Run("f")
		if assert.NoError(t, err) {
			assert.Equal(t, "", vm.RestInput)
			assert.Equal(t, VMTypeInt, vm.Ret.TypeId)

			err := vm.Run("// #EnableDice fate false")
			vm.Config = vm.parser.cur.data.Config
			if assert.NoError(t, err) {
				err := vm.Run("f")
				if assert.NoError(t, err) {
					assert.Equal(t, VMTypeNull, vm.Ret.TypeId)
				}
			}
		}
	}
}
func TestDiceFlagDoubleCrossMacroExpr(t *testing.T) {
	vm := NewVM()
	err := vm.Run("// #EnableDice doublecross true")
	vm.Config = vm.parser.cur.data.Config
	if assert.NoError(t, err) {
		err := vm.Run("2c5")
		if assert.NoError(t, err) {
			assert.Equal(t, "", vm.RestInput)
			assert.Equal(t, VMTypeInt, vm.Ret.TypeId)

			err := vm.Run("// #EnableDice doublecross false")
			vm.Config = vm.parser.cur.data.Config
			if assert.NoError(t, err) {
				err := vm.Run("2c5")
				if assert.NoError(t, err) {
					assert.Equal(t, "c5", vm.RestInput)
				}
			}
		}
	}
}

func TestComment(t *testing.T) {
	vm := NewVM()
	err := vm.Run("// test\na = 1;\na")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}
}

func TestComment2(t *testing.T) {
	// 发现注释间无法空行，予以修复
	vm := NewVM()
	err := vm.Run(`
//c1

   
//c2
a = 1; a`)
	assert.NoError(t, err)
	assert.Equal(t, vm.RestInput, "")
	assert.True(t, valueEqual(vm.Ret, ni(1)))
}

func TestDiceAndSpaceBug(t *testing.T) {
	// 一个错误的代码逻辑: 部分算符后需要跟sp1，导致f +1可以工作，但f+1不行
	// 但也不能让 f1 被解析为f，剩余文本1
	vm := NewVM()
	vm.Config.EnableDiceFate = true
	err := vm.Run("f +1")
	if assert.NoError(t, err) {
		assert.Equal(t, VMTypeInt, vm.Ret.TypeId)
	}

	err = vm.Run("f+1")
	if assert.NoError(t, err) {
		assert.Equal(t, VMTypeInt, vm.Ret.TypeId)
	}

	err = vm.Run("f1")
	if assert.NoError(t, err) {
		assert.Equal(t, "", vm.RestInput)
		assert.Equal(t, VMTypeNull, vm.Ret.TypeId)
	}
}

func TestDiceAndSpaceBug2(t *testing.T) {
	// 其他版本
	tests := [][]string{
		{"b +1", "b+1", "bX"},
		{"p +1", "p+1", "pX"},
		{"a10 +1", "a10+1", "a10x"},
		{"1c5 +1", "1c5+1", "x"},
	}

	for _, i := range tests {
		e1, e2, e3 := i[0], i[1], i[2]
		vm := NewVM()
		vm.Config.EnableDiceCoC = true
		vm.Config.EnableDiceWoD = true
		vm.Config.EnableDiceDoubleCross = true
		err := vm.Run(e1)
		if assert.NoError(t, err) {
			assert.Equal(t, VMTypeInt, vm.Ret.TypeId)
		}

		err = vm.Run(e2)
		if assert.NoError(t, err) {
			assert.Equal(t, VMTypeInt, vm.Ret.TypeId)
		}

		err = vm.Run(e3)
		if assert.NoError(t, err) {
			assert.Equal(t, "", vm.RestInput)
			assert.Equal(t, VMTypeNull, vm.Ret.TypeId)
		}
	}

	vm := NewVM()
	vm.Config.EnableDiceDoubleCross = true
	err := vm.Run("1c5d")
	if assert.NoError(t, err) {
		assert.Equal(t, "d", vm.RestInput)
	}

	vm = NewVM()
	vm.Config.EnableDiceWoD = true
	err = vm.Run("2a10x")
	if assert.NoError(t, err) {
		assert.Equal(t, "x", vm.RestInput)
	}
}

func TestBitwisePrecedence(t *testing.T) {
	vm := NewVM()
	err := vm.Run("1|2&4")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("(1|2)&4")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(0)))
	}
}

func TestLogicOp(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a = [1,2]; 5 || a.push(3); a ")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, na(ni(1), ni(2))))
	}

	vm = NewVM()
	err = vm.Run("a = [1,2]; 5 && a.push(3); a ")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, na(ni(1), ni(2), ni(3))))
	}
}

func TestFuncAbs(t *testing.T) {
	vm := NewVM()
	err := vm.Run("abs(-1)")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}
}

func TestLogicAnd(t *testing.T) {
	vm := NewVM()
	err := vm.Run("1 && 2 && 3")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	vm = NewVM()
	err = vm.Run("1 && 0 && 3")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(0)))
	}
}

func TestStackOverFlow(t *testing.T) {
	vm := NewVM()
	err := vm.Run("while 1 { 2 }")
	assert.Error(t, err)
}

func TestSliceUnicode(t *testing.T) {
	vm := NewVM()
	err := vm.Run("'中文测试'[1:3]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("文测")))
	}

	err = vm.Run("'中文测试'[-3:3]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("文测")))
	}
}

func TestDiceExprError(t *testing.T) {
	vm := NewVM()
	err := vm.Run("(-1)d5")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("('xxx')d5")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("3d(-10)")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("3d('xx')")
	assert.Error(t, err)
}

func TestDiceDH_DL(t *testing.T) {
	reResult := regexp.MustCompile(`\{(\d+) (\d+) \| (\d+)}`)

	vm := NewVM()
	for {
		err := vm.Run("3d1000dh1")
		if assert.NoError(t, err) {
			m := reResult.FindStringSubmatch(vm.GetDetailText())
			a1, _ := strconv.ParseInt(m[1], 10, 64)
			a2, _ := strconv.ParseInt(m[2], 10, 64)
			a3, _ := strconv.ParseInt(m[3], 10, 64)
			if a1 != a2 && a2 != a3 {
				// 三个输出数字不等，符合测试条件
				assert.True(t, a3 > a2 && a3 > a1)
				break
			}
		}
	}

	vm = NewVM()
	for {
		err := vm.Run("3d1000dl1")
		if assert.NoError(t, err) {
			m := reResult.FindStringSubmatch(vm.GetDetailText())
			a1, _ := strconv.ParseInt(m[1], 10, 64)
			a2, _ := strconv.ParseInt(m[2], 10, 64)
			a3, _ := strconv.ParseInt(m[3], 10, 64)
			if a1 != a2 && a2 != a3 {
				// 三个输出数字不等，符合测试条件
				assert.True(t, a3 < a2 && a3 < a1)
				break
			}
		}
	}
}

func TestIdentifier(t *testing.T) {
	vm := NewVM()
	err := vm.Run("$a = 1")
	if assert.NoError(t, err) {
		_ = vm.Run("$a")
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	err = vm.Run("`{$b}`")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, NewStrVal("null")))
	}
}

func TestIsDiceCalculateExists(t *testing.T) {
	vm := NewVM()
	err := vm.Parse("d100")
	if assert.NoError(t, err) {
		assert.True(t, vm.IsCalculateExists())
	}

	err = vm.Parse("100")
	if assert.NoError(t, err) {
		assert.False(t, vm.IsCalculateExists())
	}

	err = vm.Parse("1+2")
	if assert.NoError(t, err) {
		assert.True(t, vm.IsCalculateExists())
	}

	err = vm.Parse("f()")
	if assert.NoError(t, err) {
		assert.True(t, vm.IsCalculateExists())
	}
}

func TestIsDiceCalculateExists2(t *testing.T) {
	vm := NewVM()
	assert.Equal(t, vm.IsComputedLoaded, false)
	err := vm.Run("&a=4d1; a")
	if assert.NoError(t, err) {
		assert.Equal(t, vm.IsComputedLoaded, true)
	}

	err = vm.Run("1+1")
	if assert.NoError(t, err) {
		assert.Equal(t, vm.IsComputedLoaded, false)
	}
}

func TestIsDiceCalculateExists3(t *testing.T) {
	vm := NewVM()
	vm.GlobalValueLoadFunc = func(name string) *VMValue {
		if name == "a" {
			return NewComputedVal("4d1")
		}
		return nil
	}
	err := vm.Run("a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(ni(4), vm.Ret))
		assert.Equal(t, vm.IsComputedLoaded, true)
	}
}

func TestDiceExprIndexBug(t *testing.T) {
	// 12.1 于言诺发现，如 2d(3d1) 会被错误计算为 9[2d(3d1)=9=3+3+3,3d1=3]
	// 经查原因为Dice字节指令执行时，并未将骰子栈正确出栈
	reResult := regexp.MustCompile(`2d\(3d1\)=(\d+)\+(\d+),`)

	vm := NewVM()
	err := vm.Run("2d(3d1)")

	if assert.NoError(t, err) {
		assert.True(t, reResult.MatchString(vm.GetDetailText()))
	}
}

func TestStringGetItem(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a = '测试'; a[1]")

	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("试")))
	}

	err = vm.Run("a = '测试'; a[-1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("试")))
	}
}

func TestDiceExprKlBug(t *testing.T) {
	// 12.6 云陌发现，2d5kld4时有概率中间过程这样子：1[2d5kld4=1={1 | 2 3 4},2d5kl=4]
	// 原因也是骰子栈未正确出栈
	// 这个有一定运气成分(虽然很小)，所以跑5次

	for i := 0; i < 5; i++ {
		vm := NewVM()
		err := vm.Run("(1d1000kl)d1")

		if assert.NoError(t, err) {
			assert.False(t, strings.Contains(vm.GetDetailText(), "|"))
		}
	}
}

func TestIfElseExprBug1(t *testing.T) {
	// 12.6 于言诺 else后面必须跟一个空格
	vm := NewVM()
	err := vm.Run("if true {} else{}")

	if assert.NoError(t, err) {
		assert.Equal(t, "", vm.RestInput)
	}

	vm = NewVM()
	err = vm.Run("if true {} elseif 1{}")

	if assert.NoError(t, err) {
		// 注: 这里 elseif 会被当做变量 所以这里读到是undefined
		assert.Equal(t, " 1{}", vm.RestInput)
	}
}

func TestBlockExprBug(t *testing.T) {
	// 12.7 木落
	vm := NewVM()
	err := vm.Run("if 1 {} 1 2 3 4 5")

	if assert.NoError(t, err) {
		assert.Equal(t, " 2 3 4 5", vm.RestInput)
	}
}

func TestWhileExprBug(t *testing.T) {
	// 12.7 云陌
	// 故障原因是第二次解析while时，第一次的没有出栈，因此又被处理了一遍，这个会引起程序崩溃
	vm := NewVM()
	vm.Config.OpCountLimit = 30000
	err := vm.Run(`i = 1; while i < 2 {continue}`)
	assert.Error(t, err) // 算力超出

	err = vm.Run(`while i < 2 {i=i+1}`)
	if assert.NoError(t, err) {
		assert.Equal(t, "", vm.RestInput)
	}
}

func TestNameDetailBug(t *testing.T) {
	// "a = 1;a   " 时，过程为 "a = 1;1[a    =1]"，不应有空格
	vm := NewVM()
	err := vm.Run(`a = 1;a   `)
	if assert.NoError(t, err) {
		// TODO: 后面的空格
		assert.Equal(t, "a = 1;1   ", vm.GetDetailText())
	}
}

func TestLogicOrBug(t *testing.T) {
	// (0||0)+1 报错，原因是生成的代码里最后有一个jmp 1，跳过了1的push，导致栈里只有一个值
	vm := NewVM()
	err := vm.Run(`(0||0)+1`)
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	err = vm.Run(`(0||1)+1`)
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}
}

func TestAttrSetBug(t *testing.T) {
	// 这个问题是最后一个函数调用的第一个参数成了他自己
	// 例如下面这个例子中，str(a.x)中，str拿到的参数是 &{2 nfunction str}
	// 原因是 attr_set 在设置时未进行值复制，而是拿到了vm栈地址，栈被覆盖后问题就出现了
	// 2024/04/23 绑定时发现
	// ItemSet 也有同样问题
	vm := NewVM()
	err := vm.Run(`a = {}; a.x = 10; str(a.x)`)
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("10")))
	}
}

func TestItemSetBug2(t *testing.T) {
	vm := NewVM()
	err := vm.Run(`a = {}; a[1] = 10; str(a[1])`)
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("10")))
	}
}

func TestStackTop(t *testing.T) {
	vm := NewVM()
	_ = vm.Run(`1;2;3`)
	assert.Equal(t, vm.StackTop(), 3) // 暂时的设计是只在语句块弃栈

	_ = vm.Run(`4`)
	assert.Equal(t, vm.StackTop(), 1) // 二次运行清空栈

	_ = vm.Run(`while (i<10) { i=i+1; 1;2;3 }`)
	assert.Equal(t, vm.StackTop(), 0) // 语句块后空栈

	_ = vm.Run(`1;2; while (i<10) { i=i+1; 1;2;3 }`)
	assert.Equal(t, vm.StackTop(), 2) // 语句块弃栈不影响上级
}

func TestFStringDiceType4(t *testing.T) {
	vm := NewVM()
	vm.Config.DisableNDice = false
	err := vm.Run("`{db}`")
	assert.NoError(t, err)

	err = vm.Run("`{ddx}`")
	assert.NoError(t, err)
}

func TestFStringBlock(t *testing.T) {
	vm := NewVM()
	var err error
	err = vm.Run("`{% a=2; b=3 %}4`")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("34")))
	}

	err = vm.Run("`{  if b=3 {} }`")
	assert.NoError(t, err)
	// assert.Contains(t, err.Error(), "关键字作为变量名")

	err = vm.Run("`{ a=1;b=2 }`")
	assert.NoError(t, err)
	// assert.Contains(t, err.Error(), "无法处理字符 ;")
}

func TestFStringIf(t *testing.T) {
	vm := NewVM()
	err := vm.Run("`{ if }`")
	// assert.Contains(t, err.Error(), "{} 内必须是一个表达式")
	assert.Contains(t, err.Error(), "stmtIf:")
}

func TestFStringStackOverflowBug(t *testing.T) {
	// `{1} {2} {% 3;4;5;6 %}`
	// 的结果会成为 345 或 456(根据版本不同)
	// 这是栈不平衡的体现
	vm := NewVM()
	err := vm.Run("`{1} {2} {% 3;4;5;6 %}`")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("1 2 6")))
	}
}

func TestFStringStackOverflowBug2(t *testing.T) {
	// `{1} {2} {% if false {} %}`
	// 会报错，因为没有任何返回
	vm := NewVM()
	err := vm.Run("`{1} {2} {% if false {} %}`")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("1 2 ")))
	}
}

func TestIfError(t *testing.T) {
	vm := NewVM()
	err := vm.Run("if 1 ")
	assert.Contains(t, err.Error(), "不符合if语法")
}

func TestFStringV1IfCompatible(t *testing.T) {
	// `1 {% if 1 {'test'} %} 2`
	// 在v1中会返回1  2，中间的if语句执行后栈中是空的
	// 但是v2改为不进行栈平衡，所以会得到1 test 2，这个兼容选项用于模拟这一行为
	vm := NewVM()
	err := vm.Run("`1 {% if 1 {'test'} %} 2`")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("1 test 2")))
	}

	vm.Config.EnableV1IfCompatible = true
	err = vm.Run("`1 {% if 1 {'test'} %} 2`")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("1  2")))
		assert.Equal(t, vm.V1IfCompatibleCount, 1)
	}
}

func TestNegExpr(t *testing.T) {
	vm := NewVM()
	err := vm.Run("-1 + 5")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(4)))
	}
}

func TestNegExpr2(t *testing.T) {
	vm := NewVM()
	err := vm.Run("-1-5")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(-6)))
	}
}

func TestFStringCRBug(t *testing.T) {
	// 2024.6.15 白鱼
	// 遇到的问题是自定义文本中的换行正常但\n不转义[解析中为\\n]
	// 经测试发现如果混用 \n 和 \\n 混用，则后面的 \\n 不转义
	vm := NewVM()
	err := vm.Run("`AAAAA\n1234\\n5678`")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("AAAAA\n1234\n5678")))
	}

	vm = NewVM()
	err = vm.Run("\x1eAAAAA\n1234\\n5678\x1e")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("AAAAA\n1234\n5678")))
	}
}

func TestDicePushDefaultExpr(t *testing.T) {
	vm := NewVM()
	vm.Config.DefaultDiceSideExpr = "12d1 - 11"
	err := vm.Run("d") // 1d1
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	// 测试缓存
	vm.Config.DefaultDiceSideExpr = "12d1 - 11"
	err = vm.Run("d") // 1d1
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	// 无缓存，默认为d100，同时测试缓存失效情况
	vm.Config.DefaultDiceSideExpr = ""
	err = vm.Run("2d") // 2d100
	if assert.NoError(t, err) {
		assert.True(t, vm.Ret.MustReadInt() >= 2)
	}
}

func TestDetailTextComputed(t *testing.T) {
	vm := NewVM()
	vm.Attrs.Store("a", NewComputedVal("4d1"))
	err := vm.Run("a")
	if assert.NoError(t, err) {
		assert.Equal(t, "4[a=4d1=4]", vm.GetDetailText())
	}
}

func TestDetailTextComputed2(t *testing.T) {
	vm := NewVM()
	vm.Attrs.Store("a", NewComputedVal("4d(1d1)"))
	err := vm.Run("a")
	if assert.NoError(t, err) {
		assert.Equal(t, "4[a=4d(1d1)=4]", vm.GetDetailText())
	}
}

func TestDetailText1(t *testing.T) {
	vm := NewVM()
	err := vm.Run("(6d1)d1")
	if assert.NoError(t, err) {
		assert.Equal(t, "6[(6d1)d1=1+1+1+1+1+1,6d1=6]", vm.GetDetailText())
	}
}

func TestDetailTextRule13(t *testing.T) {
	vm := NewVM()
	err := vm.Run("d1")
	if assert.NoError(t, err) {
		// 简易式子，可以吃掉所有detail，所以为0。否则会出现d1=1=1，吃掉后表现为d1=1
		assert.Equal(t, "", vm.GetDetailText())
	}
}

func TestDetailText2(t *testing.T) {
	vm := NewVM()
	err := vm.Run("2d1")
	if assert.NoError(t, err) {
		assert.Equal(t, "2[2d1=1+1]", vm.GetDetailText())
	}
}

func TestDetailText3(t *testing.T) {
	vm := NewVM()
	err := vm.Run("(2d1)d1")
	if assert.NoError(t, err) {
		assert.Equal(t, "2[(2d1)d1=1+1,2d1=2]", vm.GetDetailText())
	}
}

func TestDetailText4(t *testing.T) {
	vm := NewVM()
	vm.Config.DiceMaxMode = true
	err := vm.Run("d + 2d")
	if assert.NoError(t, err) {
		assert.Equal(t, "100[D100] + 200[2D100=100+100]", vm.GetDetailText())
	}
}

func TestDetailText5(t *testing.T) {
	vm := NewVM()
	vm.Config.DiceMaxMode = true
	err := vm.Run("2dk1")
	if assert.NoError(t, err) {
		assert.Equal(t, "100[2D100kh1={100 | 100}]", vm.GetDetailText())
	}
}

func TestDetailText6(t *testing.T) {
	vm := NewVM()
	vm.Config.DiceMaxMode = true
	err := vm.Run("d + 1")
	if assert.NoError(t, err) {
		assert.Equal(t, "100[D100] + 1", vm.GetDetailText())
	}
}

func TestDetailText7(t *testing.T) {
	vm := NewVM()
	vm.Config.DiceMaxMode = true
	err := vm.Run("d")
	if assert.NoError(t, err) {
		assert.Equal(t, "100[D100]", vm.GetDetailText())
	}
}

func TestDiceAdvantage(t *testing.T) {
	vm := NewVM()
	vm.Config.DefaultDiceSideExpr = "1"
	err := vm.Run("d优势")
	if assert.NoError(t, err) {
		assert.Equal(t, "1[d优势={1 | 1}]", vm.GetDetailText())
	}
}

func TestDiceAdvantage2(t *testing.T) {
	vm := NewVM()
	vm.Config.DefaultDiceSideExpr = "1"
	err := vm.Run("3d优势")
	if assert.NoError(t, err) {
		assert.Equal(t, "3[3D1=1+1+1]", vm.GetDetailText())
	}
}
