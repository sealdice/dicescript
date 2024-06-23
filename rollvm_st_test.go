package dicescript

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type checkItem struct {
	Name  string
	Value *VMValue
	Extra *VMValue
	Type  string
	Op    string
}

func (item *checkItem) check(t *testing.T, _type string, name string, val *VMValue, extra *VMValue, op string, detail string) {
	assert.Equal(t, name, item.Name)
	assert.True(t, valueEqual(val, item.Value))
	if item.Extra != nil {
		assert.True(t, valueEqual(extra, item.Extra))
	}
	if item.Op != "" {
		assert.Equal(t, op, item.Op)
	}
	assert.Equal(t, _type, item.Type)
}

func TestStSetBasic(t *testing.T) {
	vm := NewVM()

	vm.Config.CallbackSt = func(_type string, name string, val *VMValue, extra *VMValue, op string, detail string) {
		// fmt.Println("!!", _type, name, val, extra, op, detail)
		assert.Equal(t, name, "A")
		assert.True(t, valueEqual(val, ni(1)))
		assert.Equal(t, _type, "set")
	}

	err := vm.Run(`^stA:1`)
	assert.NoError(t, err)
}

func TestStSet(t *testing.T) {
	vm := NewVM()
	items := []checkItem{
		checkItem{Name: "力量", Value: ni(60), Type: "set"},
		checkItem{Name: "敏捷", Value: ni(70), Type: "set"},
	}

	index := 0
	vm.Config.CallbackSt = func(_type string, name string, val *VMValue, extra *VMValue, op string, detail string) {
		items[index].check(t, _type, name, val, extra, op, detail)
		index += 1
	}

	err := vm.Run(`^st力量60敏捷70`)
	assert.NoError(t, err)
}

func TestStSet2(t *testing.T) {
	vm := NewVM()
	items := []checkItem{
		checkItem{Name: "力量", Value: ni(60), Type: "set"},
		checkItem{Name: "敏捷", Value: ni(70), Type: "set"},
	}

	index := 0
	vm.Config.CallbackSt = func(_type string, name string, val *VMValue, extra *VMValue, op string, detail string) {
		items[index].check(t, _type, name, val, extra, op, detail)
		index += 1
	}

	err := vm.Run(`^st力量60 敏捷70`)
	assert.NoError(t, err)
}

func TestStSet3(t *testing.T) {
	vm := NewVM()
	items := []checkItem{
		checkItem{Name: "智力", Value: ni(80), Type: "set"},
		checkItem{Name: "知识", Value: ni(90), Type: "set"},
	}

	index := 0
	vm.Config.CallbackSt = func(_type string, name string, val *VMValue, extra *VMValue, op string, detail string) {
		items[index].check(t, _type, name, val, extra, op, detail)
		index += 1
	}

	err := vm.Run(`^st智力:80 知识=90`)
	assert.NoError(t, err)
}

func TestStSetCompute(t *testing.T) {
	vm := NewVM()
	items := []checkItem{
		checkItem{Name: "射击", Value: NewComputedVal("1d6"), Type: "set"},
		checkItem{Name: "射击", Value: NewComputedVal("(1d6)"), Type: "set"},
	}

	index := 0
	vm.Config.CallbackSt = func(_type string, name string, val *VMValue, extra *VMValue, op string, detail string) {
		// fmt.Println("!!!", _type, name, val, extra, op, detail, val.ToRepr()) // 注: 最后一个值在IDEA中的输出可能不正常，中间加了空格，实际没有
		items[index].check(t, _type, name, val, extra, op, detail)
		index += 1
	}

	// err := vm.Run(`^st&射击=1d6 `)
	err := vm.Run(`^st&射击=1d6      &射击=(1d6)`)
	assert.NoError(t, err)
	assert.Equal(t, index, 2)
}

func TestStModBasic(t *testing.T) {
	vm := NewVM()

	vm.Config.CallbackSt = func(_type string, name string, val *VMValue, extra *VMValue, op string, detail string) {
		// fmt.Println("!!", _type, name, val, extra, op, detail)
		assert.Equal(t, name, "A")
		assert.True(t, valueEqual(val, ni(2)))
		assert.Equal(t, _type, "mod")
	}

	err := vm.Run(`^stA+2`)
	assert.NoError(t, err)
}

func TestStSetX0Basic(t *testing.T) {
	vm := NewVM()

	vm.Config.CallbackSt = func(_type string, name string, val *VMValue, extra *VMValue, op string, detail string) {
		// fmt.Println("!!", _type, name, val, extra, op, detail)
		assert.Equal(t, name, "A")
		assert.True(t, valueEqual(val, ni(3)))
		assert.Equal(t, _type, "set.x0")
	}

	err := vm.Run(`^stA*:3`)
	assert.NoError(t, err)
}

func TestStSetX1Basic(t *testing.T) {
	vm := NewVM()

	vm.Config.CallbackSt = func(_type string, name string, val *VMValue, extra *VMValue, op string, detail string) {
		// fmt.Println("!!", _type, name, val, extra, op, detail)
		assert.Equal(t, name, "A")
		assert.True(t, valueEqual(val, ni(3)))
		assert.True(t, valueEqual(extra, nf(2.1)))
		assert.Equal(t, _type, "set.x1")
	}

	err := vm.Run(`^stA*2.1: 3`)
	assert.NoError(t, err)
}

func TestStMod(t *testing.T) {
	vm := NewVM()

	items := []checkItem{
		checkItem{Name: "力量", Value: ni(3), Type: "mod"},
		checkItem{Name: "敏捷", Value: ni(3), Type: "mod"},
	}

	index := 0
	vm.Config.CallbackSt = func(_type string, name string, val *VMValue, extra *VMValue, op string, detail string) {
		// fmt.Println("!!", _type, name, val, extra, op, detail)
		items[index].check(t, _type, name, val, extra, op, detail)
		index += 1
	}

	err := vm.Run(`^st力量+3d1 敏捷+=3 `)
	assert.NoError(t, err)
}

func TestStMod1(t *testing.T) {
	vm := NewVM()
	items := []checkItem{
		checkItem{Name: "力量", Value: ni(1), Type: "mod"},
		checkItem{Name: "力量", Value: ni(4), Type: "mod"},
		checkItem{Name: "力量", Value: ni(6), Type: "mod"},
		checkItem{Name: "力量", Value: ni(2), Type: "mod", Op: "-"},
	}

	index := 0
	vm.Config.CallbackSt = func(_type string, name string, val *VMValue, extra *VMValue, op string, detail string) {
		// fmt.Println("!!", _type, name, val, extra, op, detail)
		items[index].check(t, _type, name, val, extra, op, detail)
		index += 1
	}

	err := vm.Run(`^st力量+1力量+4d1力量+4d1+2力量-4d1+2`)
	assert.NoError(t, err)
}

func TestStMod2(t *testing.T) {
	vm := NewVM()

	items := []checkItem{
		checkItem{Name: "力量123", Value: ni(3), Type: "mod"},
	}

	index := 0
	vm.Config.CallbackSt = func(_type string, name string, val *VMValue, extra *VMValue, op string, detail string) {
		// fmt.Println("!!", _type, name, val, extra, op, detail)
		items[index].check(t, _type, name, val, extra, op, detail)
		index += 1
	}

	err := vm.Run(`^st'力量123'+=3`)
	assert.NoError(t, err)
}

func TestStModMinus(t *testing.T) {
	vm := NewVM()

	items := []checkItem{
		checkItem{Name: "力量", Value: ni(4), Type: "mod", Op: "-"},
	}

	index := 0
	vm.Config.CallbackSt = func(_type string, name string, val *VMValue, extra *VMValue, op string, detail string) {
		// fmt.Println("!!", _type, name, val, extra, op, detail)
		items[index].check(t, _type, name, val, extra, op, detail)
		index += 1
	}

	err := vm.Run(`^st力量-3d1-1 `)
	assert.NoError(t, err)
}

func TestStModMinus2(t *testing.T) {
	vm := NewVM()

	items := []checkItem{
		checkItem{Name: "力量", Value: ni(2), Type: "mod", Op: "-="},
	}

	index := 0
	vm.Config.CallbackSt = func(_type string, name string, val *VMValue, extra *VMValue, op string, detail string) {
		// fmt.Println("!!", _type, name, val, extra, op, detail)
		items[index].check(t, _type, name, val, extra, op, detail)
		index += 1
	}

	err := vm.Run(`^st力量-=3d1-1 `)
	assert.NoError(t, err)
}
