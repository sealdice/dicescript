package dicescript

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTypesMethodArraySum(t *testing.T) {
	d := NewArrayVal(ni(1), nf(2.2), ni(3))
	v := funcArraySum(nil, d, nil)
	assert.Equal(t, v.ToString(), "6.2")
}

func TestTypesMethodArrayShuttle(t *testing.T) {
	d := NewArrayVal(ni(1), ni(2), ni(3), ni(4))
	v := funcArrayShuttle(nil, d, nil)
	// 不知道怎么写测试，因为总是有概率打乱后与原本一致
	assert.Equal(t, v.Length(nil), IntType(4))
}

func TestTypesMethodArrayRand(t *testing.T) {
	d := NewArrayVal(ni(1), ni(1), ni(1), ni(1))
	v := funcArrayRand(nil, d, nil)
	assert.Equal(t, v.MustReadInt(), IntType(1))
}

func TestTypesMethodArrayRandSize(t *testing.T) {
	d := NewArrayVal(ni(1), ni(1), ni(1), ni(1))
	v := funcArrayRandSize(nil, d, []*VMValue{ni(1)})
	assert.Equal(t, v.Length(nil), IntType(1))
}

func TestTypesMethodDictKeys(t *testing.T) {
	d := NewDictValWithArrayMust(ns("a"), ni(1), ns("b"), ni(2))
	v := funcDictKeys(nil, d.V(), nil)
	assert.True(t, valueEqual(v, na(ns("a"), ns("b"))) || valueEqual(v, na(ns("b"), ns("a"))))
}

func TestTypesMethodDictValues(t *testing.T) {
	d := NewDictValWithArrayMust(ns("a"), ni(1), ns("b"), ni(2))
	v := funcDictValues(nil, d.V(), nil)
	assert.True(t, valueEqual(v, na(ni(1), ni(2))) || valueEqual(v, na(ni(2), ni(1))))
}

func TestTypesMethodDictItems(t *testing.T) {
	d := NewDictValWithArrayMust(ns("a"), ni(1), ns("b"), ni(2))
	v := funcDictItems(nil, d.V(), nil)
	assert.True(t, valueEqual(v, na(na(ns("a"), ni(1)), na(ns("b"), ni(2)))) ||
		valueEqual(v, na(na(ns("b"), ni(2)), na(ns("a"), ni(1)))))
}

func TestTypesMethodDictLen(t *testing.T) {
	d := NewDictValWithArrayMust(ns("a"), ni(1), ns("b"), ni(2))
	v := funcDictLen(nil, d.V(), nil)
	assert.Equal(t, v.MustReadInt(), IntType(2))
}
