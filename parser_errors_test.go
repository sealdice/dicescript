// Copyright 2022 fy <fy0748@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dicescript

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidStartChar(t *testing.T) {
	validChars := []rune{'a', 'Z', '0', '9', '_', '$', '(', '"', '\'', '`', '+', '-', '力'}
	for _, r := range validChars {
		if !isValidStartChar(r) {
			t.Errorf("expected %q to be a valid start character", r)
		}
	}

	invalidChars := []rune{'/', '@', '#', ';', ')', ']', '}'}
	for _, r := range invalidChars {
		if isValidStartChar(r) {
			t.Errorf("expected %q to NOT be a valid start character", r)
		}
	}
}

func TestIsOperatorChar(t *testing.T) {
	operators := []rune{'+', '-', '*', '/', '%', '^', '=', '<', '>', '!', '&', '|'}
	for _, r := range operators {
		if !isOperatorChar(r) {
			t.Errorf("expected %q to be an operator character", r)
		}
	}

	nonOperators := []rune{'a', '1', '(', '[', '_'}
	for _, r := range nonOperators {
		if isOperatorChar(r) {
			t.Errorf("expected %q to NOT be an operator character", r)
		}
	}
}

func TestFindUnclosedBracketBytes(t *testing.T) {
	tests := []struct {
		input string
		want  rune
	}{
		{"(1+2", '('},
		{"{a:1", '{'},
		{"[1,2", '['},
		{"(1+2)", 0},
		{"{a:1}", 0},
		{"[1,2]", 0},
		{"((1+2)", '('},
		{"\"(\"", 0}, // 括号在字符串内
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := findUnclosedBracketBytes([]byte(tt.input))
			if got != tt.want {
				t.Errorf("findUnclosedBracketBytes(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIntegrationWithVM_InvalidStart(t *testing.T) {
	vm := NewVM()
	err := vm.Run("/")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "语法错误")
		assert.Contains(t, err.Error(), "表达式不能以")
		assert.Contains(t, err.Error(), "Expression cannot start with")
	}
}

func TestIntegrationWithVM_MissingExpr(t *testing.T) {
	vm := NewVM()
	// 注意：在这个语法中 "1 +" 会被解析为 "1"（尾部操作符被忽略）
	// 测试一个确实会产生错误的表达式
	err := vm.Run("(1 +")
	if assert.Error(t, err) {
		// 这会触发缺少右括号的错误
		assert.Contains(t, err.Error(), "语法错误")
	}
}

func TestIntegrationWithVM_MissingParen(t *testing.T) {
	vm := NewVM()
	err := vm.Run("(1+2")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "语法错误")
		assert.Contains(t, err.Error(), "缺少右括号")
		assert.Contains(t, err.Error(), "Missing closing parenthesis")
	}
}

func TestLanguageOptions_English(t *testing.T) {
	vm := NewVM()
	vm.Config.ParseErrorLanguage = ParseErrorLanguageEnglish

	err := vm.Run("/")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Syntax Error")
		assert.Contains(t, err.Error(), "Pos")
		assert.NotContains(t, err.Error(), "语法错误")
		assert.NotContains(t, err.Error(), "位置")
	}
}

func TestLanguageOptions_Chinese(t *testing.T) {
	vm := NewVM()
	vm.Config.ParseErrorLanguage = ParseErrorLanguageChinese

	err := vm.Run("/")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "语法错误")
		assert.Contains(t, err.Error(), "位置")
		assert.NotContains(t, err.Error(), "Syntax Error")
		assert.NotContains(t, err.Error(), "Position")
	}
}

func TestLanguageOptions_Bilingual(t *testing.T) {
	vm := NewVM()
	vm.Config.ParseErrorLanguage = ParseErrorLanguageBilingual

	err := vm.Run("/")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "语法错误")
		assert.Contains(t, err.Error(), "Syntax Error")
		assert.Contains(t, err.Error(), "位置")
		assert.Contains(t, err.Error(), "Pos")
	}
}

func TestErrorContext_ShowsCodeAndPointer(t *testing.T) {
	vm := NewVM()
	// 使用一个确实会产生错误的表达式
	err := vm.Run("/test")
	if assert.Error(t, err) {
		errStr := err.Error()
		// 应该包含代码上下文和指示符
		assert.Contains(t, errStr, "/test")
		assert.Contains(t, errStr, "^")
		assert.Contains(t, errStr, "|")
	}
}

func TestLongLineHandling(t *testing.T) {
	vm := NewVM()
	// 超长输入，以非法字符开头
	longInput := "/" + strings.Repeat("a", 100)
	err := vm.Run(longInput)
	if assert.Error(t, err) {
		errStr := err.Error()
		// 应该被截断，不应该显示完整的100个a
		assert.True(t, len(errStr) < 500, "error message should be truncated")
		assert.Contains(t, errStr, "...")
	}
}

func TestGetLineAtBytes(t *testing.T) {
	// 测试单行
	line := getLineAtBytes([]byte("hello world"), 1)
	assert.Equal(t, "hello world", line)

	// 测试多行
	multiLine := []byte("line1\nline2\nline3")
	assert.Equal(t, "line1", getLineAtBytes(multiLine, 1))
	assert.Equal(t, "line2", getLineAtBytes(multiLine, 2))
	assert.Equal(t, "line3", getLineAtBytes(multiLine, 3))

	// 测试超长行截断
	longLine := []byte(strings.Repeat("x", 100))
	result := getLineAtBytes(longLine, 1)
	assert.True(t, len(result) <= 60, "long line should be truncated")
	assert.True(t, strings.HasSuffix(result, "..."), "should end with ...")
}

func TestGetPrevNonSpaceChar(t *testing.T) {
	tests := []struct {
		input  string
		offset int
		want   rune
	}{
		{"1 + ", 4, '+'},
		{"1+", 2, '+'},
		{"1  +  ", 6, '+'},
		{"", 0, 0},
		{"   ", 3, 0},
	}

	for _, tt := range tests {
		got := getPrevNonSpaceChar(tt.input, tt.offset)
		if got != tt.want {
			t.Errorf("getPrevNonSpaceChar(%q, %d) = %q, want %q", tt.input, tt.offset, got, tt.want)
		}
	}
}
