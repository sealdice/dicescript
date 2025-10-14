package dicescript

import (
	"strings"
	"unicode/utf8"
)

// CustomDiceStream 提供逐字符访问，用于自定义骰子解析函数。
type CustomDiceStream struct {
	data   []byte
	start  int
	offset int
	runes  []int
}

func (s *CustomDiceStream) init(data []byte, start int) {
	s.data = data
	s.start = start
	s.offset = 0
	s.runes = s.runes[:0]
}

// ResetAttempt 撤销当前尝试，光标回到起始位置。
func (s *CustomDiceStream) ResetAttempt() {
	s.offset = 0
	s.runes = s.runes[:0]
}

// Peek 返回下一个字符，但不移动光标。
func (s *CustomDiceStream) Peek() (rune, bool) {
	idx := s.start + s.offset
	if idx >= len(s.data) {
		return 0, false
	}
	r, size := utf8.DecodeRune(s.data[idx:])
	if r == utf8.RuneError && size == 1 {
		return rune(s.data[idx]), true
	}
	return r, true
}

// Read 读取下一个字符并前进。
func (s *CustomDiceStream) Read() (rune, bool) {
	idx := s.start + s.offset
	if idx >= len(s.data) {
		return 0, false
	}
	r, size := utf8.DecodeRune(s.data[idx:])
	if r == utf8.RuneError && size == 1 {
		r = rune(s.data[idx])
		size = 1
	}
	s.offset += size
	s.runes = append(s.runes, size)
	return r, true
}

// Unread 回退最近一次 Read 的字符。
func (s *CustomDiceStream) Unread() bool {
	if len(s.runes) == 0 {
		return false
	}
	size := s.runes[len(s.runes)-1]
	s.runes = s.runes[:len(s.runes)-1]
	s.offset -= size
	if s.offset < 0 {
		s.offset = 0
	}
	return true
}

// Commit 确认当前消费的字符数。
func (s *CustomDiceStream) Commit() {}

// Consumed 返回已消费的字节数。
func (s *CustomDiceStream) Consumed() int {
	return s.offset
}

// Current 返回已消费的文本内容。
func (s *CustomDiceStream) Current() string {
	return string(s.data[s.start : s.start+s.offset])
}

// Remaining 返回剩余文本，主要用于调试。
func (s *CustomDiceStream) Remaining() string {
	return string(s.data[s.start+s.offset:])
}

// ReadDigits 连续读取数字字符，返回字符串和是否成功读取至少一个字符。
func (s *CustomDiceStream) ReadDigits() (string, bool) {
	var buf []rune
	for {
		r, ok := s.Peek()
		if !ok || r < '0' || r > '9' {
			break
		}
		s.Read()
		buf = append(buf, r)
	}
	if len(buf) == 0 {
		return "", false
	}
	return string(buf), true
}

// ReadExpr 使用指定的语法入口（默认为 exprRoot）进行解析，并将捕获的文本作为计算值返回。
// 仅在解析成功时才会前移游标。
func (s *CustomDiceStream) ReadExpr(entry string) (*VMValue, bool, error) {
	if entry == "" {
		entry = "exprRoot"
	}

	absStart := s.start + s.offset
	if absStart >= len(s.data) {
		return nil, false, nil
	}

	parser := newParser("", s.data[absStart:], memoized(true))
	parser.entrypoint = entry

	data := parser.cur.data
	data.code = make([]ByteCode, 64)
	data.codeIndex = 0
	data.pendingCustomDice = nil

	if _, err := parser.parse(nil); err != nil {
		return nil, false, err
	}

	consumed := parser.pt.offset
	if consumed <= 0 {
		return nil, false, nil
	}

	matched := s.data[absStart : absStart+consumed]
	expr := strings.TrimSpace(string(matched))
	if expr == "" {
		return nil, false, nil
	}

	for idx := 0; idx < consumed; {
		r, size := utf8.DecodeRune(s.data[absStart+idx:])
		if r == utf8.RuneError && size == 1 {
			size = 1
		}
		idx += size
		s.runes = append(s.runes, size)
	}
	s.offset += consumed

	computed := NewComputedVal(expr)
	return computed, true, nil
}
