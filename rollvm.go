/*
  Copyright [2022] fy0748@gmail.com

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package dicescript

import (
	"errors"
	"strconv"
)

func (e *DiceScriptParser) checkStackOverflow() bool {
	if e.Error != nil {
		return true
	}
	if e.Top >= len(e.Code) {
		need := len(e.Code) * 2
		if need <= 8192 {
			newCode := make([]ByteCode, need)
			copy(newCode, e.Code)
			e.Code = newCode
		} else {
			e.Error = errors.New("E1:指令虚拟机栈溢出，请不要发送过长的指令")
			return true
		}
	}
	return false
}

func (e *DiceScriptParser) AddOperator(operator CodeType) int {
	if e.checkStackOverflow() {
		return -1
	}
	code, top := e.Code, e.Top
	e.Top++
	code[top].T = operator
	return e.Top
}

func (e *DiceScriptParser) AddLeftValueMark() {
	if e.checkStackOverflow() {
		return
	}
	code, top := e.Code, e.Top
	e.Top++
	code[top].T = TypeLeftValueMark
}

func (e *DiceScriptParser) AddValue(value string) {
	// 实质上的压栈命令
	if e.checkStackOverflow() {
		return
	}
	code, top := e.Code, e.Top
	e.Top++
	code[top].Value, _ = strconv.ParseInt(value, 10, 64)
}
