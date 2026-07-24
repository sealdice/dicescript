package dicescript

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomDiceStreamNavigation(t *testing.T) {
	var stream CustomDiceStream
	stream.init([]byte("x甲12!"), 1)

	r, ok := stream.Peek()
	assert.True(t, ok)
	assert.Equal(t, '甲', r)
	assert.Zero(t, stream.Consumed())

	r, ok = stream.Read()
	assert.True(t, ok)
	assert.Equal(t, '甲', r)
	assert.Equal(t, "甲", stream.Current())
	assert.Equal(t, "12!", stream.Remaining())
	assert.True(t, stream.Unread())
	assert.False(t, stream.Unread())

	_, _ = stream.Read()
	digits, ok := stream.ReadDigits()
	assert.True(t, ok)
	assert.Equal(t, "12", digits)
	stream.Commit()
	assert.Equal(t, "甲12", stream.Current())

	digits, ok = stream.ReadDigits()
	assert.False(t, ok)
	assert.Empty(t, digits)
	r, ok = stream.Read()
	assert.True(t, ok)
	assert.Equal(t, '!', r)
	_, ok = stream.Peek()
	assert.False(t, ok)
	_, ok = stream.Read()
	assert.False(t, ok)

	stream.ResetAttempt()
	assert.Zero(t, stream.Consumed())
	assert.Equal(t, "甲12!", stream.Remaining())
}

func TestCustomDiceStreamInvalidUTF8AndReadExprErrors(t *testing.T) {
	var stream CustomDiceStream
	stream.init([]byte{0xff}, 0)

	r, ok := stream.Peek()
	assert.True(t, ok)
	assert.Equal(t, rune(0xff), r)
	r, ok = stream.Read()
	assert.True(t, ok)
	assert.Equal(t, rune(0xff), r)
	assert.True(t, stream.Unread())

	stream.init(nil, 0)
	value, matched, err := stream.ReadExpr("")
	assert.NoError(t, err)
	assert.False(t, matched)
	assert.Nil(t, value)

	stream.init([]byte("1"), 0)
	value, matched, err = stream.ReadExpr("not-a-rule")
	assert.Error(t, err)
	assert.False(t, matched)
	assert.Nil(t, value)
	assert.Zero(t, stream.Consumed())
}
