//go:build !js && !tinygo
// +build !js,!tinygo

package dicescript

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFmtPrintf(t *testing.T) {
	FmtPrintf("test %d\n", 123)
}

func TestFmtPrintln(t *testing.T) {
	FmtPrintln("test", 123)
}

func TestFmtSprintf(t *testing.T) {
	result := FmtSprintf("test %d", 123)
	assert.Equal(t, "test 123", result)
}

func TestFmtErrorf(t *testing.T) {
	err := FmtErrorf("test error: %s", "failed")
	assert.Error(t, err)
	assert.Equal(t, "test error: failed", err.Error())
}

func TestStrconvParseInt(t *testing.T) {
	val, err := StrconvParseInt("123", 10, 64)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), val)

	_, err = StrconvParseInt("abc", 10, 64)
	assert.Error(t, err)
}

func TestStrconvFormatInt(t *testing.T) {
	result := StrconvFormatInt(123, 10)
	assert.Equal(t, "123", result)

	result = StrconvFormatInt(255, 16)
	assert.Equal(t, "ff", result)
}

func TestStrconvFormatFloat(t *testing.T) {
	result := StrconvFormatFloat(3.14159, 'f', 2, 64)
	assert.Equal(t, "3.14", result)
}

func TestStrconvParseFloat(t *testing.T) {
	val, err := StrconvParseFloat("3.14", 64)
	assert.NoError(t, err)
	assert.Equal(t, 3.14, val)

	_, err = StrconvParseFloat("abc", 64)
	assert.Error(t, err)
}

func TestJSONMarshal(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
		"age":  123,
	}
	bytes, err := JSONMarshal(data)
	assert.NoError(t, err)
	assert.NotNil(t, bytes)
}

func TestJSONUnmarshal(t *testing.T) {
	jsonData := []byte(`{"name":"test","age":123}`)
	var result map[string]interface{}
	err := JSONUnmarshal(jsonData, &result)
	assert.NoError(t, err)
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, float64(123), result["age"])

	err = JSONUnmarshal([]byte(`invalid json`), &result)
	assert.Error(t, err)
}
