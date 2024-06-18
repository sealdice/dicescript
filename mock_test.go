package dicescript

import "testing"

// 有些东西没有过测试的意义，但是到目前为止，go还不支持在测试中忽略指定函数
// 因此这个文件用来水掉没意义的函数

func TestMockByteCodeString(t *testing.T) {
	for i := 0; i < 87; i++ {
		c := &ByteCode{T: CodeType(i), Value: IntType(1)}
		switch c.T {
		case typePushFloatNumber:
			c.Value = 1.1
		case typePushString:
			c.Value = ""
		case typePushComputed:
			c.Value = NewComputedVal("1")
		case typePushFunction:
			c.Value = NewFunctionValRaw(&FunctionData{Expr: "1"})
		case typeLoadName, typeLoadNameWithDetail, typeLoadNameRaw, typeInvokeSelf, typeAttrSet, typeAttrGet:
			c.Value = "name"
		case typeDetailMark:
			c.Value = BufferSpan{}
		}
		_ = c.CodeString()
	}
}
