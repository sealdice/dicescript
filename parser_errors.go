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
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// 错误消息语言选项常量
const (
	ParseErrorLanguageBilingual = 0 // 双语（默认）
	ParseErrorLanguageChinese   = 1 // 仅中文
	ParseErrorLanguageEnglish   = 2 // 仅英文
)

// parseErrorLanguage 当前错误消息语言设置
var parseErrorLanguage = ParseErrorLanguageBilingual

// bilingualMsg 双语消息
type bilingualMsg struct {
	cn, en string
}

// 错误消息映射
var errMsgs = map[string]bilingualMsg{
	"empty":           {"输入为空", "Empty input"},
	"invalidStart":    {"表达式不能以 '%c' 开头", "Expression cannot start with '%c'"},
	"missingRParen":   {"缺少右括号 ')'", "Missing closing parenthesis ')'"},
	"missingRBrace":   {"缺少右花括号 '}'", "Missing closing brace '}'"},
	"missingRBracket": {"缺少右方括号 ']'", "Missing closing bracket ']'"},
	"unclosedString":  {"字符串未闭合", "Unclosed string literal"},
	"missingExpr":     {"'%c' 后需要表达式", "Expression expected after '%c'"},
	"incomplete":      {"表达式不完整", "Incomplete expression"},
	"unexpectedChar":  {"无法识别的字符 '%c'", "Unexpected character '%c'"},
	"syntax":          {"语法错误", "Syntax error"},
}

func init() {
	// 注册错误格式化钩子
	ErrorFormatter = formatFriendlyError
}

// SetParseErrorLanguage 设置解析错误消息的语言
func SetParseErrorLanguage(lang int) {
	parseErrorLanguage = lang
}

// formatFriendlyError 生成友好的错误消息
func formatFriendlyError(pos position, input []byte, expected []string) error {
	if len(input) == 0 {
		return fmtErr(pos, input, errMsgs["empty"], 0)
	}

	var char rune
	if pos.offset < len(input) {
		char, _ = utf8.DecodeRune(input[pos.offset:])
	}

	// 判断错误类型
	var msg bilingualMsg
	var fmtChar rune

	inputToCheck := input
	if pos.offset < len(input) {
		inputToCheck = input[:pos.offset]
	}

	switch {
	case pos.offset == 0 && !isValidStartChar(char):
		msg, fmtChar = errMsgs["invalidStart"], char

	case findUnclosedBracketBytes(inputToCheck) == '(':
		msg = errMsgs["missingRParen"]

	case findUnclosedBracketBytes(inputToCheck) == '{':
		msg = errMsgs["missingRBrace"]

	case findUnclosedBracketBytes(inputToCheck) == '[':
		msg = errMsgs["missingRBracket"]

	case pos.offset > 0 && isOperatorChar(getPrevNonSpaceChar(string(input), pos.offset)):
		prevChar := getPrevNonSpaceChar(string(input), pos.offset)
		// 排除闭合括号
		if prevChar != ')' && prevChar != ']' && prevChar != '}' {
			msg, fmtChar = errMsgs["missingExpr"], prevChar
		} else {
			msg = errMsgs["syntax"]
		}

	case pos.offset >= len(input):
		msg = errMsgs["incomplete"]

	case char == '"' || char == '\'' || char == '`' || char == '\x1e':
		msg = errMsgs["unclosedString"]

	case !isValidStartChar(char) && !isValidIdentChar(char):
		msg, fmtChar = errMsgs["unexpectedChar"], char

	default:
		msg = errMsgs["syntax"]
	}

	return fmtErr(pos, input, msg, fmtChar)
}

// fmtErr 格式化错误输出
func fmtErr(pos position, input []byte, msg bilingualMsg, char rune) error {
	var sb strings.Builder

	// 标题
	switch parseErrorLanguage {
	case ParseErrorLanguageChinese:
		sb.WriteString("语法错误\n")
	case ParseErrorLanguageEnglish:
		sb.WriteString("Syntax Error\n")
	default:
		sb.WriteString("语法错误 Syntax Error\n")
	}

	// 上下文（如果有输入）
	if len(input) > 0 {
		sb.WriteString("  |\n")
		line := getLineAtBytes(input, pos.line)
		sb.WriteString(fmt.Sprintf("  |  %s\n", line))

		// 指示符
		pointerPos := pos.col - 1
		if pointerPos < 0 {
			pointerPos = 0
		}
		pointer := strings.Repeat(" ", pointerPos) + "^"
		sb.WriteString(fmt.Sprintf("  |  %s\n", pointer))
		sb.WriteString("  |\n")
	}

	// 格式化消息
	cn, en := msg.cn, msg.en
	if char != 0 {
		cn = fmt.Sprintf(cn, char)
		en = fmt.Sprintf(en, char)
	}

	// 位置和消息
	switch parseErrorLanguage {
	case ParseErrorLanguageChinese:
		sb.WriteString(fmt.Sprintf("  位置 %d:%d - %s", pos.line, pos.col, cn))
	case ParseErrorLanguageEnglish:
		sb.WriteString(fmt.Sprintf("  Pos %d:%d - %s", pos.line, pos.col, en))
	default:
		sb.WriteString(fmt.Sprintf("  位置 %d:%d - %s\n", pos.line, pos.col, cn))
		sb.WriteString(fmt.Sprintf("  Pos %d:%d - %s", pos.line, pos.col, en))
	}

	return errors.New(sb.String())
}

// getLineAtBytes 获取指定行的内容
func getLineAtBytes(input []byte, line int) string {
	lines := strings.Split(string(input), "\n")
	if line > 0 && line <= len(lines) {
		result := lines[line-1]
		// 如果行太长，截取
		if len(result) > 60 {
			result = result[:57] + "..."
		}
		return result
	}
	result := string(input)
	if len(result) > 60 {
		result = result[:57] + "..."
	}
	return result
}

// isValidStartChar 检查字符是否可以作为表达式的开头
func isValidStartChar(r rune) bool {
	if r >= '0' && r <= '9' {
		return true
	}
	if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
		return true
	}
	switch r {
	case '_', '$', '(', '[', '{', '"', '\'', '`', '\x1e', '+', '-', '.', '&':
		return true
	case '优', '劣', '（', '【': // 中文字符
		return true
	}
	// CJK统一汉字
	if r >= 0x4E00 && r <= 0x9FFF {
		return true
	}
	return false
}

// isValidIdentChar 检查字符是否可以作为标识符的一部分
func isValidIdentChar(r rune) bool {
	if isValidStartChar(r) {
		return true
	}
	if r >= '0' && r <= '9' {
		return true
	}
	return false
}

// isOperatorChar 检查是否是操作符字符
func isOperatorChar(r rune) bool {
	switch r {
	case '+', '-', '*', '/', '%', '^', '=', '<', '>', '!', '&', '|', '?', ':', ',':
		return true
	case '＋', '－', '＊', '／': // 全角操作符
		return true
	}
	return false
}

// getPrevNonSpaceChar 获取前一个非空白字符
func getPrevNonSpaceChar(input string, offset int) rune {
	for i := offset - 1; i >= 0; i-- {
		r, _ := utf8.DecodeRuneInString(input[i:])
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return r
		}
	}
	return 0
}

// findUnclosedBracketBytes 查找未闭合的括号（字节版本）
func findUnclosedBracketBytes(input []byte) rune {
	stack := []rune{}
	inString := false
	stringChar := rune(0)

	for _, b := range string(input) {
		r := b
		// 处理字符串
		if !inString && (r == '"' || r == '\'' || r == '`' || r == '\x1e') {
			inString = true
			stringChar = r
			continue
		}
		if inString {
			if r == stringChar {
				inString = false
			}
			continue
		}

		// 处理括号
		switch r {
		case '(', '{', '[':
			stack = append(stack, r)
		case ')':
			if len(stack) > 0 && stack[len(stack)-1] == '(' {
				stack = stack[:len(stack)-1]
			}
		case '}':
			if len(stack) > 0 && stack[len(stack)-1] == '{' {
				stack = stack[:len(stack)-1]
			}
		case ']':
			if len(stack) > 0 && stack[len(stack)-1] == '[' {
				stack = stack[:len(stack)-1]
			}
		}
	}

	if len(stack) > 0 {
		return stack[len(stack)-1]
	}
	return 0
}
