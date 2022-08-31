/*
  Copyright 2022 fy <fy0748@gmail.com>

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

import "testing"

type compareTestData []struct {
	v1       *VMValue
	v2       *VMValue
	excepted *VMValue
}

func valueEqual(a *VMValue, b *VMValue) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.TypeId == b.TypeId {
		return a.Value == b.Value
	}
	return false
}

var ni = VMValueNewInt64
var nf = VMValueNewFloat64
var ns = VMValueNewStr

func TestCompare(t *testing.T) {
	// lt 小于
	var compLTTest = compareTestData{
		// int, int
		{ni(0), ni(0), ni(0)}, // 0 < 0, false
		{ni(0), ni(2), ni(1)}, // 0 < 2, true
		{ni(2), ni(0), ni(0)}, // 2 < 0, false

		// int float
		{ni(0), nf(0), ni(0)}, // 0 < 0, false
		{ni(0), nf(2), ni(1)}, // 0 < 2, true
		{ni(2), nf(0), ni(0)}, // 2 < 0, false

		// float int
		{nf(0), ni(0), ni(0)}, // 0 < 0, false
		{nf(0), ni(2), ni(1)}, // 0 < 2, true
		{nf(2), ni(0), ni(0)}, // 2 < 0, false

		// float float
		{nf(0), nf(0), ni(0)}, // 0 < 0, false
		{nf(0), nf(2), ni(1)}, // 0 < 2, true
		{nf(2), nf(0), ni(0)}, // 2 < 0, false

		// int str
		{ni(0), ns("2"), nil}, // 0 < '2', ERR
	}

	for _, i := range compLTTest {
		r := (*VMValue).CompLT(i.v1, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("CompareLE(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// le 小于等于
	var compLETest = compareTestData{
		// int, int
		{ni(0), ni(0), ni(1)}, // 0 <= 0, true
		{ni(0), ni(2), ni(1)}, // 0 <= 2, true
		{ni(2), ni(0), ni(0)}, // 2 <= 0, false

		// int float
		{ni(0), nf(0), ni(1)}, // 0 <= 0, true
		{ni(0), nf(2), ni(1)}, // 0 <= 2, true
		{ni(2), nf(0), ni(0)}, // 2 <= 0, false

		// float int
		{nf(0), ni(0), ni(1)}, // 0 <= 0, true
		{nf(0), ni(2), ni(1)}, // 0 <= 2, true
		{nf(2), ni(0), ni(0)}, // 2 <= 0, false

		// float float
		{nf(0), nf(0), ni(1)}, // 0 <= 0, true
		{nf(0), nf(2), ni(1)}, // 0 <= 2, true
		{nf(2), nf(0), ni(0)}, // 2 <= 0, false

		// int str
		{ni(0), ns("2"), nil}, // 0 <= '2', ERR
	}

	for _, i := range compLETest {
		r := (*VMValue).CompLE(i.v1, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("CompareLE(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// ge 大于等于
	var compGETest = compareTestData{
		// int, int
		{ni(0), ni(0), ni(1)}, // 0 >= 0, true
		{ni(0), ni(2), ni(0)}, // 0 >= 2, false
		{ni(2), ni(0), ni(1)}, // 2 >= 0, true

		// int float
		{ni(0), nf(0), ni(1)}, // 0 >= 0, true
		{ni(0), nf(2), ni(0)}, // 0 >= 2, false
		{ni(2), nf(0), ni(1)}, // 2 >= 0, true

		// float int
		{nf(0), ni(0), ni(1)}, // 0 >= 0, true
		{nf(0), ni(2), ni(0)}, // 0 >= 2, false
		{nf(2), ni(0), ni(1)}, // 2 >= 0, true

		// float float
		{nf(0), nf(0), ni(1)}, // 0 >= 0, true
		{nf(0), nf(2), ni(0)}, // 0 >= 2, false
		{nf(2), nf(0), ni(1)}, // 2 >= 0, true

		// int str
		{ni(0), ns("2"), nil}, // 0 >= '2', ERR
	}

	for _, i := range compGETest {
		r := (*VMValue).CompGE(i.v1, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("CompareGE(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// gt 大于
	var compGTTest = compareTestData{
		// int, int
		{ni(0), ni(0), ni(0)}, // 0 > 0, false
		{ni(0), ni(2), ni(0)}, // 0 > 2, false
		{ni(2), ni(0), ni(1)}, // 2 > 0, true

		// int float
		{ni(0), nf(0), ni(0)}, // 0 > 0, false
		{ni(0), nf(2), ni(0)}, // 0 > 2, false
		{ni(2), nf(0), ni(1)}, // 2 > 0, true

		// float int
		{nf(0), ni(0), ni(0)}, // 0 > 0, false
		{nf(0), ni(2), ni(0)}, // 0 > 2, false
		{nf(2), ni(0), ni(1)}, // 2 > 0, true

		// float float
		{nf(0), nf(0), ni(0)}, // 0 > 0, false
		{nf(0), nf(2), ni(0)}, // 0 > 2, false
		{nf(2), nf(0), ni(1)}, // 2 > 0, true

		// int str
		{ni(0), ns("2"), nil}, // 0 > '2', ERR
	}

	for _, i := range compGTTest {
		r := (*VMValue).CompGT(i.v1, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("CompareGT(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// EQ
	theSame := ni(123)
	var compEQTest = compareTestData{
		{theSame, theSame, ni(1)},
		// int, int
		{ni(0), ni(0), ni(1)},  // 0 == 0, true
		{ni(-1), ni(1), ni(0)}, // -1 == 1, false
		// int, float
		{ni(0), nf(0), ni(1)}, // 0 == 0, true
		{ni(0), nf(1), ni(0)}, // 0 == 1, false
		// float, int
		{nf(1), ni(0), ni(0)}, // 1 == 0, false
		// int, str
		{ni(0), ns(""), ni(0)}, // 0 == '', false
	}
	for _, i := range compEQTest {
		r := (*VMValue).CompEQ(i.v1, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("CompareEQ(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

}

func TestAdditive(t *testing.T) {
	// + add
	var addTest = compareTestData{
		// int, int
		{ni(1), ni(2), ni(3)}, // 1+2=3
		// int, float
		{ni(1), nf(2), nf(3)}, // 1+2=3
		// float, int
		{nf(1), ni(2), nf(3)}, // 1+2=3
		// float, flaot
		{nf(1), nf(2), nf(3)}, // 1+2=3
		// str, str
		{ns("aa"), ns("bb"), ns("aabb")}, // 'aa'+'bb'='aabb'
	}

	for _, i := range addTest {
		r := (*VMValue).Add(i.v1, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("Add(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// - sub
	var subTest = compareTestData{
		// int, int
		{ni(3), ni(2), ni(1)}, // 3-2=1
		// int, float
		{ni(3), nf(2), nf(1)}, // 3-2=1
		// float, int
		{nf(3), ni(2), nf(1)}, // 3-2=1
		// float, flaot
		{nf(3), nf(2), nf(1)}, // 3-2=1
	}

	for _, i := range subTest {
		r := (*VMValue).Sub(i.v1, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("Sub(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// * multiply
	var subMul = compareTestData{
		// int, int
		{ni(3), ni(2), ni(6)}, // 3*2=6
		// int, float
		{ni(3), nf(2), nf(6)}, // 3*2=6
		// float, int
		{nf(3), ni(2), nf(6)}, // 3*2=6
		// float, flaot
		{nf(3), nf(2), nf(6)}, // 3*2=6
	}

	for _, i := range subMul {
		r := (*VMValue).Multiply(i.v1, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("Mul(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// * div
	var subDiv = compareTestData{
		// int, int
		{ni(3), ni(2), ni(1)}, // 3/2=1
		// int, float
		{ni(3), nf(2), nf(1.5)}, // 3/2=1.5
		// float, int
		{nf(3), ni(2), nf(1.5)}, // 3/2=1.5
		// float, flaot
		{nf(3), nf(2), nf(1.5)}, // 3/2=1.5
		// TODO: 被除数为0
	}

	for _, i := range subDiv {
		r := (*VMValue).Divide(i.v1, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("Div(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}

	// * mod
	var subMod = compareTestData{
		// int, int
		{ni(2), ni(3), ni(2)}, // 2%3=2
		// int, float
		{ni(3), nf(2), nil},
		// TODO: 被除数为0
	}

	for _, i := range subMod {
		r := (*VMValue).Modulus(i.v1, i.v2)
		if !valueEqual(r, i.excepted) {
			t.Errorf("Mod(%s, %s) = %s; expected %s", i.v1.ToString(), i.v2.ToString(), r.ToString(), i.excepted.ToString())
		}
	}
}