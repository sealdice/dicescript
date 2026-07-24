package dicescript

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestValueMapStoreRestoresExpungedEntry(t *testing.T) {
	var values ValueMap
	values.Store("restored", ni(1))
	values.Range(func(string, *VMValue) bool { return true })
	values.Delete("restored")
	values.Store("other", ni(2))
	values.Store("restored", ni(3))

	value, ok := values.Load("restored")
	assert.True(t, ok)
	assert.True(t, valueEqual(value, ni(3)))
	assert.Equal(t, 2, values.Length())
}

func TestValueMapLoadOrStoreStateTransitions(t *testing.T) {
	var fresh ValueMap
	actual, loaded := fresh.LoadOrStore("key", ni(1))
	assert.False(t, loaded)
	assert.True(t, valueEqual(actual, ni(1)))
	fresh.Range(func(string, *VMValue) bool { return true })
	actual, loaded = fresh.LoadOrStore("key", ni(2))
	assert.True(t, loaded)
	assert.True(t, valueEqual(actual, ni(1)))

	var restored ValueMap
	restored.Store("key", ni(1))
	restored.Range(func(string, *VMValue) bool { return true })
	restored.Delete("key")
	restored.Store("other", ni(2))
	actual, loaded = restored.LoadOrStore("key", ni(3))
	assert.False(t, loaded)
	assert.True(t, valueEqual(actual, ni(3)))
	assert.True(t, valueEqual(restored.MustLoad("key"), ni(3)))
}

func TestValueMapDeleteAndRangeDeletedEntries(t *testing.T) {
	var values ValueMap
	values.Store("deleted", ni(1))
	values.Store("kept", ni(2))
	values.Range(func(string, *VMValue) bool { return true })
	values.Delete("deleted")
	values.Delete("missing")

	visited := 0
	values.Range(func(key string, value *VMValue) bool {
		visited++
		assert.Equal(t, "kept", key)
		assert.True(t, valueEqual(value, ni(2)))
		return false
	})
	assert.Equal(t, 1, visited)
	_, ok := values.LoadAndDelete("deleted")
	assert.False(t, ok)
}

func TestValueMapJSONErrors(t *testing.T) {
	var values ValueMap
	values.Store("nil", nil)
	data, err := values.ToJSON()
	assert.Error(t, err)
	assert.Nil(t, data)

	err = values.UnmarshalJSON([]byte("not json"))
	assert.Error(t, err)
}
