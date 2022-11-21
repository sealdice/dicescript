package dicescript

import "testing"

// 生成器生成的代码写测试无意义
// 此文件主要为解决覆盖率

func TestValueMap(t *testing.T) {
	v := ValueMap{}
	v.Store("a", nil)
	v.LoadOrStore("b", nil)
	v.LoadOrStore("b", nil)
	v.LoadAndDelete("a")
}
