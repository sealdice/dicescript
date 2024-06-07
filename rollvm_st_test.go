package dicescript

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStBasic(t *testing.T) {
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

func TestStBasicMod(t *testing.T) {
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

func TestStBasicStX0(t *testing.T) {
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

func TestStBasicStX1(t *testing.T) {
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
