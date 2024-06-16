package dicescript

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValueMap(t *testing.T) {
	v := ValueMap{}
	v.Store("a", nil)
	v.LoadOrStore("b", nil)
	v.LoadOrStore("b", nil)
	v.LoadAndDelete("a")
}

func TestValueMapSize(t *testing.T) {
	v := ValueMap{}
	v.Store("a", ni(1))
	v.Store("b", ni(2))
	v.Store("c", ni(3))
	assert.Equal(t, v.Length(), 3)
}

func TestValueMapSize2(t *testing.T) {
	// 注: 此处Load和Store是为了让ValueMap将只读表和dirty表设为不同状态，这是一种中间态
	v := ValueMap{}
	v.Store("a", ni(1))
	v.Store("b", ni(2))
	v.Store("c", ni(3))
	v.Load("c")
	v.Load("c")
	v.Load("c")
	v.Store("c", ni(3))
	v.Store("d", ni(4))
	// fmt.Println(1, v.read)
	// fmt.Println(2, v.dirty)
	assert.Equal(t, v.Length(), 4)
}

func TestValueMapClear(t *testing.T) {
	// 注: 此处Load和Store是为了让ValueMap将只读表和dirty表设为不同状态
	v := ValueMap{}
	v.Store("a", ni(1))
	v.Store("b", ni(2))
	v.Store("c", ni(3))
	v.Store("d", ni(4))
	v.Range(func(key string, value *VMValue) bool {
		return true
	})
	assert.Equal(t, 0, len(v.dirty)) // 此时全在 read 表中
	assert.Equal(t, v.Length(), 4)

	v.Clear()
	assert.Equal(t, v.Length(), 0)
}
