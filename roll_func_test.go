package dicescript

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRollCommon(t *testing.T) {
	ret, _ := RollCommon(nil, 5, 1, nil, nil, 0, 0, 0, 0)
	assert.Equal(t, ret, IntType(5))
}

func TestRollCommonMax(t *testing.T) {
	ret, _ := RollCommon(nil, 5, 10, nil, nil, 0, 0, 0, 1)
	assert.Equal(t, ret, IntType(50))
}

func TestRollCommonMin(t *testing.T) {
	ret, _ := RollCommon(nil, 5, 10, nil, nil, 0, 0, 0, -1)
	assert.Equal(t, ret, IntType(5))
}

func TestRollDoubleCross(t *testing.T) {
	ret, _, _, _ := RollDoubleCross(nil, 11, 10, 10, 0) // pool默认为10，10c11 = 10c11m10
	assert.True(t, ret <= 10)
}

func TestRollWoD(t *testing.T) {
	ret, _, _, _ := RollWoD(nil, 11, 8, 10, 1, true, 0) // 8a11m10k1
	assert.Equal(t, IntType(8), ret)
}
