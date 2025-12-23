//go:build !js && !tinygo
// +build !js,!tinygo

package dicescript

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FmtPrintf 输出调试信息
func FmtPrintf(format string, args ...any) {
	fmt.Printf(format, args...)
}

// FmtPrintln 输出一行信息
func FmtPrintln(args ...any) {
	fmt.Println(args...)
}

// FmtSprintf 格式化字符串
func FmtSprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

// FmtErrorf wraps fmt.Errorf
func FmtErrorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// StrconvParseInt wraps strconv.ParseInt
func StrconvParseInt(s string, base int, bitSize int) (int64, error) {
	return strconv.ParseInt(s, base, bitSize)
}

// StrconvFormatInt wraps strconv.FormatInt
func StrconvFormatInt(i int64, base int) string {
	return strconv.FormatInt(i, base)
}

// StrconvFormatFloat wraps strconv.FormatFloat
func StrconvFormatFloat(f float64, fmt byte, prec int, bitSize int) string {
	return strconv.FormatFloat(f, fmt, prec, bitSize)
}

func StrconvParseFloat(s string, bitSize int) (float64, error) {
	return strconv.ParseFloat(s, bitSize)
}

// JSONMarshal 封装json.Marshal
func JSONMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// JSONUnmarshal 封装json.Unmarshal
func JSONUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
